package main

import (
	"net"

	"github.com/iterum-provenance/sidecar/socket"
	"github.com/iterum-provenance/sidecar/transmit"
	"github.com/prometheus/common/log"
)

// senderHandler is a handler function for a socket that sends files to the fragmenter
func senderHandler(socket socket.Socket, conn net.Conn) {
	defer conn.Close()
	for {
		// Wait for the next job to come off the queue.
		msg := <-socket.Channel

		files := *msg.(*filelist)

		// Send the msg over the connection
		err := transmit.EncodeSend(conn, &files)

		// Error handling
		switch err.(type) {
		case *transmit.SerializationError:
			log.Warnf("Could not encode message due to '%v', skipping message", err)
			continue
		case *transmit.ConnectionError:
			log.Warnf("Closing connection towards due to '%v'", err)
			return
		default:
			log.Errorf("%v, closing connection", err)
			return
		case nil:
		}
	}
}

// receiverHandler is a handler for a socket that receives fragmented file list from the fragmenter
func receiverHandler(socket socket.Socket, conn net.Conn) {
	defer conn.Close()
	for {
		fragment := filelist{}
		err := transmit.DecodeRead(conn, &fragment)

		// Error handling
		switch err.(type) {
		case *transmit.SerializationError:
			log.Warnf("Could not decode message due to '%v', skipping message", err)
			continue
		case *transmit.ConnectionError:
			log.Warnf("Closing connection towards due to '%v'", err)
			return
		default:
			log.Errorf("%v, closing connection", err)
			return
		case nil:
		}

		socket.Channel <- &fragment
	}
}
