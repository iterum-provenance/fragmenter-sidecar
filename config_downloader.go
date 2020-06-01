package main

import (
	"sync"

	"github.com/iterum-provenance/fragmenter/data"
	"github.com/iterum-provenance/iterum-go/daemon"
	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/prometheus/common/log"
)

// ConfigDownloader is the structure responsible for downloading
// the config files used by the fragmenter to properly fragment data files
type ConfigDownloader struct {
	AllFiles     data.Filelist
	ToDownload   data.Filelist
	DaemonConfig daemon.Config
	Completed    chan<- transmit.Serializable //*data.FragmenterInput
	finished     chan desc.LocalFileDesc
}

// NewConfigDownloader instantiates a new config downloader without starting it
func NewConfigDownloader(files data.Filelist, toDownload data.Filelist, daemon daemon.Config, completed chan transmit.Serializable) ConfigDownloader {
	return ConfigDownloader{
		AllFiles:     files,
		ToDownload:   toDownload,
		DaemonConfig: daemon,
		Completed:    completed,
		finished:     make(chan desc.LocalFileDesc, len(toDownload)),
	}
}

// StartBlocking starts the process of downloading the config files
func (cd *ConfigDownloader) StartBlocking() {
	defer close(cd.Completed)
	log.Infof("Starting to dowload %v fragmenter config files", len(cd.ToDownload))
	wg := &sync.WaitGroup{}
	// Start the downloading of each config file
	for _, file := range cd.ToDownload {
		wg.Add(1)
		go func(f string) {
			fileDesc, err := downloadConfigFileFromDaemon(cd.DaemonConfig, f)
			if err != nil {
				log.Fatalf("Could not download fragmenter config file due to '%v'", err)
			}
			cd.finished <- fileDesc
			wg.Done()
		}(file)
	}
	// Wait for the downloading to finish
	wg.Wait()
	close(cd.finished)
	log.Infof("Finished downloading fragmenter config files")

	// Channel is already closed so loop will terminate once all messages are processed
	configFiles := []desc.LocalFileDesc{}
	for fileDesc := range cd.finished {
		configFiles = append(configFiles, fileDesc)
	}

	fi := data.FragmenterInput{
		DataFiles:   cd.AllFiles,
		ConfigFiles: configFiles,
	}
	log.Infof("Finishing up ConfigDownloader")
	cd.Completed <- &fi
}

// Start is an asyncrhonous alternative to StartBlocking by spawning a goroutine
func (cd *ConfigDownloader) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		cd.StartBlocking()
	}()
}
