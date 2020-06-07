package main

import (
	"sync"

	"github.com/iterum-provenance/fragmenter/data"
	"github.com/iterum-provenance/iterum-go/daemon"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/prometheus/common/log"
)

// DataMover is the structure responsible for pulling data from the daemon and into minio
type DataMover struct {
	MinioConfig  minio.Config
	DaemonConfig daemon.Config
	Files        data.Filelist
	Completed    chan Upload
}

// NewDataMover initializes a new datamover
func NewDataMover(mc minio.Config, dc daemon.Config, files data.Filelist, completed chan Upload) DataMover {
	return DataMover{mc, dc, files, completed}
}

// StartBlocking starts the process of pulling files from the daemon and storing them in minio
func (dm DataMover) StartBlocking() {
	var wg sync.WaitGroup
	numWorkers := 10

	filesToUploadChannel := make(chan string, len(dm.Files))

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(filesToUpload <-chan string, completions chan<- Upload) {
			defer wg.Done()
			for fileName := range filesToUpload {
				remoteFile, err := pullAndUploadFile(dm.MinioConfig, dm.DaemonConfig, fileName, 5)
				if err != nil {
					log.Fatalln(err)
				}
				completions <- Upload{fileName, remoteFile}
			}
		}(filesToUploadChannel, dm.Completed)
	}

	for _, file := range dm.Files {
		filesToUploadChannel <- file
	}
	close(filesToUploadChannel)

	wg.Wait()
}

// Start is an asyncrhonous alternative to StartBlocking spawning a goroutine
func (dm DataMover) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		dm.StartBlocking()
	}()
}
