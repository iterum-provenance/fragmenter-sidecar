package config

import (
	"sync"

	"github.com/prometheus/common/log"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/minio"
)

// FUploader is the structure responsible for uploading
// all files marked as config files to minio
type FUploader struct {
	ToUpload    []desc.LocalFileDesc
	MinioConfig minio.Config
}

// NewFUploader instantiates a new config downloader without starting it
func NewFUploader(files []desc.LocalFileDesc, minio minio.Config) FUploader {
	return FUploader{
		ToUpload:    files,
		MinioConfig: minio,
	}
}

// StartBlocking starts the process of uploading the config files
func (cu *FUploader) StartBlocking() {
	log.Infof("Starting to upload %v config files", len(cu.ToUpload))
	wg := &sync.WaitGroup{}
	// Start the uploading of each config file
	for _, localFileDesc := range cu.ToUpload {
		wg.Add(1)
		go func(fdesc desc.LocalFileDesc) {
			_, err := cu.MinioConfig.PutConfigFile(fdesc)
			if err != nil {
				log.Fatalf("Could not upload config file to minio due to '%v'", err)
			}
			wg.Done()
		}(localFileDesc)
	}
	// Wait for the uploading to finish
	wg.Wait()
	log.Infof("Finished uploading config files")
}

// Start is an asyncrhonous alternative to StartBlocking by spawning a goroutine
func (cu *FUploader) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		cu.StartBlocking()
	}()
}
