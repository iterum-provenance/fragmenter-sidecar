package handler

import (
	"net"

	"github.com/prometheus/common/log"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/socket"
	"github.com/iterum-provenance/iterum-go/transmit"
)

// Sender is a handler function for a socket that sends files to the fragmenter
func Sender(socket socket.Socket, conn net.Conn) {
	defer conn.Close()

	// Error handling
	errHandler := func(err error) {
		switch err.(type) {
		case nil:
		case *transmit.SerializationError:
			log.Fatalf("Could not encode message due to '%v'", err)
		case *transmit.ConnectionError:
			log.Warnf("Closing connection towards fragmenter due to '%v'", err)
		default:
			log.Errorf("%v, closing connection", err)
		}
	}

	// Wait for the list of files to come off the queue.
	msg := <-socket.Channel
	kill := desc.NewKillMessage()

	// Send the msgs over the connection
	err := transmit.EncodeSend(conn, msg)
	if err != nil {
		errHandler(err)
		return
	}
	err = transmit.EncodeSend(conn, &kill)
	if err != nil {
		errHandler(err)
		return
	}

	socket.Stop()
}
