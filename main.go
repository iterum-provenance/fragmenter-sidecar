package main

import (
	"sync"

	"github.com/iterum-provenance/iterum-go/daemon"
	envcomm "github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"

	"github.com/iterum-provenance/sidecar/lineage"
	"github.com/iterum-provenance/sidecar/messageq"
	"github.com/iterum-provenance/sidecar/socket"

	"github.com/iterum-provenance/fragmenter/env"
)

func main() {
	var wg sync.WaitGroup

	// Initiate pipe to fragmenter channels
	fragmenterBufferSize := 10
	toFragmenterChannel := make(chan transmit.Serializable, 1)
	fromFragmenterChannel := make(chan transmit.Serializable, fragmenterBufferSize)
	toFragmenterSocket := env.FragmenterInputSocket
	fromFragmenterSocket := env.FragmenterOutputSocket

	// Start pipe to fragmenter
	pipe := socket.NewPipe(fromFragmenterSocket, toFragmenterSocket,
		fromFragmenterChannel, toFragmenterChannel,
		receiverHandler, senderHandler)
	pipe.Start(&wg)

	// Define and connect to minio storage and configure for remote Daemon
	daemonConfig := daemon.NewDaemonConfigFromEnv()
	minioConfig, err := minio.NewMinioConfigFromEnv()
	util.PanicIfErr(err, "")
	err = minioConfig.Connect()
	util.PanicIfErr(err, "")

	// Get the target commit
	files, err := getCommitFiles(daemonConfig)
	util.PanicIfErr(err, "")

	// Download config and then send the file list to the fragmenter
	configDownloader := NewConfigDownloader(files, env.Config, daemonConfig, minioConfig, pipe.ToTarget)
	configDownloader.Start(&wg)

	uploadedBufferSize := len(files)
	uploaded := make(chan Upload, uploadedBufferSize)

	toMQBufferSize := len(files)
	toMQChannel := make(chan transmit.Serializable, toMQBufferSize)

	mqLineageBridgeBufferSize := 10
	mqLineageBridge := make(chan transmit.Serializable, mqLineageBridgeBufferSize)

	// Start downloading files from daemon and upload them to minio
	dataMover := NewDataMover(minioConfig, daemonConfig, files, uploaded)
	dataMover.Start(&wg)

	tracker := NewTracker(uploaded, fromFragmenterChannel, toMQChannel, files)
	tracker.Start(&wg)

	mqSender, err := messageq.NewSender(toMQChannel, mqLineageBridge, envcomm.MQBrokerURL, envcomm.MQOutputQueue)
	util.Ensure(err, "MessageQueue sender succesfully created and listening")
	mqSender.Start(&wg)

	lineageTracker := lineage.NewTracker(envcomm.ProcessName, envcomm.ManagerURL, envcomm.PipelineHash, mqLineageBridge)
	lineageTracker.Start(&wg)

	wg.Wait()
}
