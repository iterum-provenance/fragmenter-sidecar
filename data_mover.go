package main

import (
	"log"
	"sync"

	"github.com/iterum-provenance/fragmenter/daemon"
	"github.com/iterum-provenance/fragmenter/minio"
)

// DataMover is the structure responsible for pulling data from the daemon and into minio
type DataMover struct {
	MinioConfig  minio.Config
	DaemonConfig daemon.Config
	Files        filelist
	Completed    chan Upload
}

// NewDataMover initializes a new datamover
func NewDataMover(mc minio.Config, dc daemon.Config, files filelist, completed chan Upload) DataMover {
	return DataMover{mc, dc, files, completed}
}

// StartBlocking starts the process of pulling files from the daemon and storing them in minio
func (dm DataMover) StartBlocking() {
	var wg sync.WaitGroup
	for _, file := range dm.Files {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			remoteFile, err := pullAndUploadFile(dm.MinioConfig, dm.DaemonConfig, fileName, 5)
			if err != nil {
				log.Fatalln(err)
			}
			dm.Completed <- Upload{fileName, remoteFile}
		}(file)
	}
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
