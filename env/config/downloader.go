package config

import (
	"sync"

	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/iterum-go/daemon"
	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/minio"
)

// FDownloader is the structure responsible for downloading
// the config files used by the fragmenter and other transformations as well
type FDownloader struct {
	AllFiles     []string
	Config       *Config
	DaemonConfig daemon.Config
	MinioConfig  minio.Config
	finished     chan desc.LocalFileDesc
}

// NewFDownloader instantiates a new config downloader without starting it
func NewFDownloader(files []string, config *Config,
	daemon daemon.Config, minio minio.Config) FDownloader {

	return FDownloader{
		AllFiles:     files,
		Config:       config,
		DaemonConfig: daemon,
		MinioConfig:  minio,
		finished:     nil,
	}
}

// StartBlocking starts the process of downloading the config files
func (cd *FDownloader) StartBlocking() {
	toDownload := []string{}
	if cd.Config != nil {
		toDownload = cd.Config.ReturnMatchingFiles(cd.AllFiles)
		cd.finished = make(chan desc.LocalFileDesc, len(toDownload))
	}

	log.Infof("Starting to download %v config files", len(toDownload))
	wg := &sync.WaitGroup{}
	// Start the downloading of each config file
	for _, file := range toDownload {
		wg.Add(1)
		go func(f string) {
			fileDesc, err := downloadConfigFileFromDaemon(cd.DaemonConfig, f)
			if err != nil {
				log.Fatalf("Could not download config file due to '%v'", err)
			}
			cd.finished <- fileDesc
			wg.Done()
		}(file)
	}
	// Wait for the downloading to finish
	wg.Wait()
	close(cd.finished)
	log.Infof("Finished downloading config files")

	configFiles := []desc.LocalFileDesc{}
	// Channel cd.finished is already closed so loop will terminate once all messages are processed
	for fileDesc := range cd.finished {
		configFiles = append(configFiles, fileDesc)
	}
	log.Infof("Finishing up config.FDownloader")

	// Start uploading the config files of other transformations to minio concurrently as well
	cfgUploader := NewFUploader(configFiles, cd.MinioConfig)
	cfgUploader.Start(wg)
	wg.Wait()
}

// Start is an asyncrhonous alternative to StartBlocking by spawning a goroutine
func (cd *FDownloader) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		cd.StartBlocking()
	}()
}
