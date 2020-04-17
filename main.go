package main

import (
	"runtime"

	"github.com/iterum-provenance/fragmenter/daemon"
	"github.com/iterum-provenance/fragmenter/env"
	"github.com/iterum-provenance/fragmenter/minio"
	"github.com/iterum-provenance/fragmenter/util"
	"github.com/iterum-provenance/sidecar/messageq"
	"github.com/iterum-provenance/sidecar/socket"
	"github.com/iterum-provenance/sidecar/transmit"
)

func main() {
	// Initiate pipe to fragmenter channels
	fragmenterBufferSize := 10
	toFragmenterChannel := make(chan transmit.Serializable, fragmenterBufferSize)
	fromFragmenterChannel := make(chan transmit.Serializable, fragmenterBufferSize)
	toFragmenterSocket := "./build/tf.sock"
	fromFragmenterSocket := "./build/ff.sock"

	// Start pipe to fragmenter
	pipe := socket.NewPipe(fromFragmenterSocket, toFragmenterSocket,
		fromFragmenterChannel, toFragmenterChannel,
		receiverHandler, senderHandler)
	pipe.Start()

	// Define and connect to minio storage and configure for remote Daemon
	daemonConfig := daemon.NewDaemonConfigFromEnv()
	minioConfig, err := minio.NewMinioConfigFromEnv()
	util.PanicOnErr(err)
	err = minioConfig.Connect()
	util.PanicOnErr(err)

	// Get the target commit
	files, err := getCommitFiles(daemonConfig)
	util.PanicOnErr(err)

	// Send the file list to the fragmenter
	pipe.ToTarget <- &files

	uploadedBufferSize := len(files)
	uploaded := make(chan Upload, uploadedBufferSize)

	toMQBufferSize := len(files)
	toMQChannel := make(chan transmit.Serializable, toMQBufferSize)

	// Start downloading files from daemon and upload them to minio
	dataMover := NewDataMover(minioConfig, daemonConfig, files, uploaded)
	dataMover.Start()

	tracker := NewTracker(uploaded, fromFragmenterChannel, toMQChannel, files)
	tracker.Start()

	mqSender, err := messageq.NewSender(toMQChannel, env.MQBrokerURL, env.MQOutputQueue)
	util.Ensure(err, "MessageQueue sender succesfully created and listening")
	mqSender.Start()

	runtime.Goexit()
}
