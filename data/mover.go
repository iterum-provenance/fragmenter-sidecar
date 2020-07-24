package data

import (
	"fmt"
	"sync"

	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/iterum-go/daemon"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"
)

// Mover is the structure responsible for pulling data from the daemon and into minio
type Mover struct {
	MinioConfig  minio.Config
	DaemonConfig daemon.Config
	Files        []string
	Completed    chan transmit.Serializable // Upload
}

// NewMover initializes a new Mover
func NewMover(mc minio.Config, dc daemon.Config, files []string, completed chan transmit.Serializable) Mover {
	return Mover{mc, dc, files, completed}
}

// StartBlocking starts the process of pulling files from the daemon and storing them in minio
func (mover Mover) StartBlocking() {
	var wg sync.WaitGroup
	numWorkers := 50

	filesToUploadChannel := make(chan string, len(mover.Files))

	err := mover.MinioConfig.EnsureBucket(mover.MinioConfig.TargetBucket, 10)
	util.Ensure(err, fmt.Sprintf("Output bucket '%v' was created successfully", mover.MinioConfig.TargetBucket))

	// Spawn numWorkers uploader workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(filesToUpload <-chan string, completions chan<- transmit.Serializable) {
			defer wg.Done()
			for fileName := range filesToUpload {
				remoteFile, err := pullAndUploadFile(mover.MinioConfig, mover.DaemonConfig, fileName, 10)
				if err != nil {
					log.Fatalln(err)
				}
				upload := Upload{fileName, remoteFile}
				completions <- &upload
			}
		}(filesToUploadChannel, mover.Completed)
	}

	// Push all the work to the worker queue
	for _, file := range mover.Files {
		filesToUploadChannel <- file
	}
	close(filesToUploadChannel)

	wg.Wait()
}

// Start is an asyncrhonous alternative to StartBlocking spawning a goroutine
func (mover Mover) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		mover.StartBlocking()
	}()
}
