package main

import (
	"sync"
	"time"

	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/iterum-go/daemon"
	"github.com/iterum-provenance/iterum-go/lineage"
	mq "github.com/iterum-provenance/iterum-go/messageq"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/process"
	"github.com/iterum-provenance/iterum-go/socket"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"

	"github.com/iterum-provenance/fragmenter/data"
	"github.com/iterum-provenance/fragmenter/env"
	"github.com/iterum-provenance/fragmenter/env/config"
	"github.com/iterum-provenance/fragmenter/handler"
)

func main() {
	// log.Base().SetLevel("Debug")
	// log.Base().SetLevel("Info")
	startTime := time.Now()
	var wg sync.WaitGroup

	// ####################  Channel setup  #################### \\
	// We are going to send 1 message to the fragmenter containing commit files
	sidecarFragmenterBridge := make(chan transmit.Serializable, 1)

	// For each file moved from daemon to minio a message is send to the tracker
	moverTrackerBridgeBufferSize := 10
	moverTrackerBridge := make(chan transmit.Serializable, moverTrackerBridgeBufferSize)

	// For each completely uploaded fragment the tracker notifies the mqSender to post this fragment
	trackerMQBridgeBufferSize := 10
	trackerMQBridge := make(chan transmit.Serializable, trackerMQBridgeBufferSize)

	// For each subfragment coming from the fragmenter the `pipe` notifies the tracker of with this new fragment
	fragmenterTrackerBridgeBufferSize := 10
	fragmenterTrackerBridge := make(chan transmit.Serializable, fragmenterTrackerBridgeBufferSize)

	// For each uploaded fragment its lineage is tracked
	mqLineageBridgeBufferSize := 10
	mqLineageBridge := make(chan transmit.Serializable, mqLineageBridgeBufferSize)

	// Define and connect to minio storage and configure for remote Daemon
	daemonConfig := daemon.NewDaemonConfigFromEnv()
	minioConfig := minio.NewMinioConfigFromEnv()
	err := minioConfig.Connect()
	util.PanicIfErr(err, "")

	// ####################  Actual body  #################### \\

	// Get the target data commit
	files, err := getCommitFiles(daemonConfig)
	util.PanicIfErr(err, "")

	// Download all config (files) and upload config of other transformations to minio
	configDownloader := config.NewFDownloader(files, env.Config, daemonConfig, minioConfig)
	configDownloader.Start(&wg)

	// Downloads files from daemon and upload them to minio
	dataMover := data.NewMover(minioConfig, daemonConfig, files, moverTrackerBridge)
	dataMover.Start(&wg)

	// Connection between the fragmenter and the fragmenter-sidecar
	toFragmenterSocket := env.FragmenterInputSocket
	fromFragmenterSocket := env.FragmenterOutputSocket
	pipe := socket.NewPipe(fromFragmenterSocket, toFragmenterSocket,
		fragmenterTrackerBridge, sidecarFragmenterBridge,
		handler.Receiver, handler.Sender)
	pipe.Start(&wg)

	// Track each fragment description from the fragmenter and each file uploaded.
	// Once all files of a fragment are uploaded send it to the mqSender
	tracker := data.NewTracker(moverTrackerBridge, fragmenterTrackerBridge, trackerMQBridge, files)
	tracker.Start(&wg)

	// Publish completely uploaded fragment descriptions
	mqSender, err := mq.NewSender(trackerMQBridge, mqLineageBridge, mq.BrokerURL, mq.OutputQueue)
	util.Ensure(err, "MessageQueue sender succesfully created")
	mqSender.Start(&wg)

	// Track provenance information for each fragment published to the MQ by the mqSender
	lineageTracker := lineage.NewMqTracker(process.Name, process.PipelineHash, mq.BrokerURL, mqLineageBridge)
	lineageTracker.Start(&wg)

	// Kickstat all by sending the file list to the fragmenter
	pipe.ToTarget <- &handler.FragmenterInput{DataFiles: files}

	// #################### Finalize  #################### \\

	wg.Wait()
	log.Infof("Ran for %v", time.Now().Sub(startTime))
}
