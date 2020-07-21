package main

import (
	"sync"

	"github.com/iterum-provenance/fragmenter/data"
	"github.com/iterum-provenance/fragmenter/env/config"
	"github.com/iterum-provenance/iterum-go/daemon"
	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/prometheus/common/log"
)

// ConfigDownloader is the structure responsible for downloading
// the config files used by the fragmenter to properly fragment data files
type ConfigDownloader struct {
	AllFiles     data.Filelist
	Config       *config.Config
	DaemonConfig daemon.Config
	MinioConfig  minio.Config
	finished     chan desc.LocalFileDesc
}

// NewConfigDownloader instantiates a new config downloader without starting it
func NewConfigDownloader(files data.Filelist, config *config.Config,
	daemon daemon.Config, minio minio.Config) ConfigDownloader {

	return ConfigDownloader{
		AllFiles:     files,
		Config:       config,
		DaemonConfig: daemon,
		MinioConfig:  minio,
		finished:     nil,
	}
}

// StartBlocking starts the process of downloading the config files
func (cd *ConfigDownloader) StartBlocking() {
	toDownload := data.Filelist([]string{})
	if cd.Config != nil {
		toDownload = cd.Config.ReturnMatchingFiles(cd.AllFiles)
		cd.finished = make(chan desc.LocalFileDesc, len(toDownload))
	}

	log.Infof("Starting to dowload %v fragmenter config files", len(toDownload))
	wg := &sync.WaitGroup{}
	// Start the downloading of each config file
	for _, file := range toDownload {
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

	configFiles := []desc.LocalFileDesc{}
	// Channel cd.finished is already closed so loop will terminate once all messages are processed
	for fileDesc := range cd.finished {
		configFiles = append(configFiles, fileDesc)
	}
	log.Infof("Finishing up ConfigDownloader")

	// Start uploading the config files of other transformations to minio concurrently as well
	cfgUploader := NewConfigUploader(configFiles, cd.MinioConfig)
	cfgUploader.Start(wg)
	wg.Wait()
}

// Start is an asyncrhonous alternative to StartBlocking by spawning a goroutine
func (cd *ConfigDownloader) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		cd.StartBlocking()
	}()
}
