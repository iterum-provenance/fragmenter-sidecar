package main

import (
	"net"

	"github.com/iterum-provenance/sidecar/socket"
	"github.com/iterum-provenance/sidecar/transmit"
	"github.com/prometheus/common/log"
)

// senderHandler is a handler function for a socket that sends files to the fragmenter
func senderHandler(socket socket.Socket, conn net.Conn) {
	defer socket.Stop()
	defer conn.Close()

	// Wait for the list of files to come off the queue.
	msg := <-socket.Channel
	files := *msg.(*filelist)

	// Send the msg over the connection
	err := transmit.EncodeSend(conn, &files)

	// Error handling
	switch err.(type) {
	case nil:
	case *transmit.SerializationError:
		log.Fatalf("Could not encode message due to '%v', skipping message", err)
	case *transmit.ConnectionError:
		log.Warnf("Closing connection towards fragmenter due to '%v'", err)
	default:
		log.Errorf("%v, closing connection", err)
	}

}

// receiverHandler is a handler for a socket that receives fragmented file list from the fragmenter
func receiverHandler(socket socket.Socket, conn net.Conn) {
	defer socket.Stop()
	defer conn.Close()
	defer close(socket.Channel)

	for {
		fragment := filelist{}
		err := transmit.DecodeRead(conn, &fragment)

		// Error handling
		switch err.(type) {
		case nil:
		case *transmit.SerializationError:
			log.Fatalf("Could not decode message due to '%v', skipping message", err)
			return
		case *transmit.ConnectionError:
			log.Warnf("Closing connection from due to '%v'", err)
			return
		default:
			log.Fatalf("%v, closing connection", err)
			return
		}

		socket.Channel <- &fragment
	}
}
