package main

import (
	"log"

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
	for _, file := range dm.Files {
		go func(fileName string) {
			remoteFile, err := pullAndUploadFile(dm.MinioConfig, dm.DaemonConfig, fileName, 5)
			if err != nil {
				log.Fatalln(err)
			}
			dm.Completed <- Upload{fileName, remoteFile}
		}(file)
	}
}

// Start is an asyncrhonous alternative to StartBlocking spawning a goroutine
func (dm DataMover) Start() {
	go dm.StartBlocking()
}
