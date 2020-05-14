package main

import (
	"sync"

	"github.com/prometheus/common/log"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
)

// Upload is a struct mapping an idv file name/path
// to a remote file description as it is stored in minio
type Upload struct {
	File     string
	FileDesc desc.RemoteFileDesc
}

// UploadMap maps files to their remote file description
// If a file is in here it means it was uploaded to minio
type UploadMap map[string]desc.RemoteFileDesc

// FragmentMap maps files to a list of fragmenter-sidecar internal fragments (a list of files)
// Meaning that an idv file name/path maps to a list of fragments its used in
type FragmentMap map[string][]filelist

// Tracker is a type used to monitor which fragments and files are uploaded and can start to be processed
type Tracker struct {
	Files          filelist
	Uploaded       chan Upload
	Fragmented     chan transmit.Serializable // filelist
	Completed      chan transmit.Serializable // desc.RemoteFragmentDesc
	fragments      FragmentMap
	uploads        UploadMap
	strictOrdering bool
}

// NewTracker instantiates a new tracker and attaches itself to the passed channels
func NewTracker(uploaded chan Upload, fragmented, completedFragment chan transmit.Serializable, allFiles filelist) Tracker {
	tracker := Tracker{
		allFiles,
		uploaded,
		fragmented,
		completedFragment,
		make(FragmentMap),
		make(UploadMap),
		false,
	}

	// Initialize each file to be attached to no fragments
	for _, file := range tracker.Files {
		tracker.fragments[file] = []filelist{}
	}
	return tracker
}

// IsUploaded checks whether a filelist is completely uploaded to Minio
func (t Tracker) IsUploaded(fragment filelist) bool {
	for _, file := range fragment {
		if _, ok := t.uploads[file]; !ok {
			return false
		}
	}
	return true
}

// toRemoteFragmentDesc transforms a fully uploaded list of files into a RemoteFragmentDesc
// This can be posted on the MQPublisher. It does a fatal log if any of the files is not yet uploaded
func (t Tracker) toRemoteFragmentDesc(fragment filelist) desc.RemoteFragmentDesc {
	fragmentDesc := desc.RemoteFragmentDesc{
		Metadata: desc.RemoteMetadata{
			FragmentID:   desc.NewIterumID(),
			Predecessors: []desc.IterumID{},
		},
	}
	for _, file := range fragment {
		if _, ok := t.uploads[file]; !ok {
			log.Fatalf("Error: cannot convert non-uploaded fragment into RemoteFragmentDesc. missing file: '%v'\n", file)
		}
		fragmentDesc.Files = append(fragmentDesc.Files, t.uploads[file])
	}
	return fragmentDesc
}

// processFileUpload takes an uploaded file and stores that it was uploaded
// If any fragment containing this file is now fully uploaded, send that fragment to the MQ
func (t *Tracker) processFileUpload(upload Upload) {
	if _, ok := t.uploads[upload.File]; ok {
		log.Fatalf("Multiple files with same destination detected: '%v'\n", upload.File)
	}
	// Store that this file has been uploaded
	t.uploads[upload.File] = upload.FileDesc

	// if there is strict ordering imposed, we don't need to perform these checks, since they are covered in the main loop
	if !t.strictOrdering {
		// If this upload caused any of the fragments that this file is in now is complete: publish it
		for _, fragment := range t.fragments[upload.File] {
			if t.IsUploaded(fragment) {
				fragmentDesc := t.toRemoteFragmentDesc(fragment)
				t.Completed <- &fragmentDesc
			}
		}
	}
	if len(t.uploads) == len(t.Files) {
		close(t.Uploaded)
	}
}

func (t *Tracker) processFragmentDescription(fragment filelist) {
	// Add this fragment to the list of fragments of each file
	for _, file := range fragment {
		if _, ok := t.fragments[file]; !ok {
			log.Fatalf("Returned fragment contained file not originally in the list of files: '%v'\n", file)
		}
		t.fragments[file] = append(t.fragments[file], fragment)
	}
	// If this fragment is already complete: publish it
	if !t.strictOrdering && t.IsUploaded(fragment) {
		fragmentDesc := t.toRemoteFragmentDesc(fragment)
		t.Completed <- &fragmentDesc
	}
}

// StartBlocking starts the process of tracking files and uploads
// On upload it checks whether a fragment was completed, if so it's pushed to the MQ publisher
// On fragment it checks whether all its files were already uploaded, if so, it's pushed to the MQ publisher
func (t Tracker) StartBlocking() {
	defer close(t.Completed)
	// strictOrdering variables
	orderedFragments := []filelist{}

	for t.Uploaded != nil || t.Fragmented != nil {
		select {
		case upload, ok := <-t.Uploaded: // On file uploaded to minio
			if !ok {
				log.Infoln("Uploaded all files")
				t.Uploaded = nil
				break
			}
			t.processFileUpload(upload)
		case fragmentptr, ok := <-t.Fragmented: // On fragment returned from fragmenter
			if !ok {
				log.Infoln("Fragmenter fragmented all files")
				t.Fragmented = nil
				break
			}
			fragment := *fragmentptr.(*filelist)
			t.processFragmentDescription(fragment)
			if t.strictOrdering {
				orderedFragments = append(orderedFragments, fragment)
			}
		}
	}

	if t.strictOrdering {
		for _, fragment := range orderedFragments {
			if !t.IsUploaded(fragment) {
				log.Fatalln("All files uploaded, yet fragment is incomplete")
			} else {
				fragmentDesc := t.toRemoteFragmentDesc(fragment)
				t.Completed <- &fragmentDesc
			}
		}
	}
	log.Infoln("Tracker completed")
}

// Start is an asyncrhonous alternative to StartBlocking by spawning a goroutine
func (t Tracker) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		t.StartBlocking()
	}()
}
