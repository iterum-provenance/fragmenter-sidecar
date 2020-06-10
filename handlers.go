package main

import (
	"encoding/binary"
	"net"
	"os"

	"github.com/iterum-provenance/fragmenter/data"
	desc "github.com/iterum-provenance/iterum-go/descriptors"

	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/sidecar/socket"
	"github.com/prometheus/common/log"
)

// senderHandler is a handler function for a socket that sends files to the fragmenter
func senderHandler(socket socket.Socket, conn net.Conn) {
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
	err := EncodeSend2(conn, msg)
	if err != nil {
		errHandler(err)
		return
	}
	err = EncodeSend2(conn, &kill)
	if err != nil {
		errHandler(err)
		return
	}

	socket.Stop()
}

// EncodeSend2 is temporary.
func EncodeSend2(conn net.Conn, obj transmit.Serializable) (err error) {
	// Encoding
	data, err := obj.Serialize()
	if err != nil {
		return
	}

	size := make([]byte, transmit.FragmentSizeLength)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))

	log.Infoln(string(size))
	log.Infoln(uint32(len(data)))
	log.Infoln(len(data))

	data = append(size, data...)

	// Sending
	bytesWrittenTotal := 0
	lwb := 0
	maxChunkSize := 131072
	upb := maxChunkSize
	dataLen := len(data)
	if maxChunkSize > dataLen {
		upb = dataLen
	} else {
		upb = maxChunkSize
	}
	for bytesWrittenTotal != dataLen {

		bytesWritten, err := conn.Write(data[lwb:upb])
		lwb = upb
		if maxChunkSize+upb > dataLen {
			upb = dataLen
		} else {
			upb += maxChunkSize
		}
		if err != nil {
			return transmit.ErrConnection(err)
		}
		bytesWrittenTotal += bytesWritten
		log.Infof("Written %d / %d bytes", bytesWrittenTotal, dataLen)
	}

	return
}

// receiverHandler is a handler for a socket that receives fragmented file list from the fragmenter
func receiverHandler(socket socket.Socket, conn net.Conn) {
	defer conn.Close()
	for {
		encMsg, err := transmit.ReadMessage(conn)

		// Error handling
		switch err.(type) {
		case nil:
		case *transmit.ConnectionError:
			log.Warnf("Closing connection from due to '%v'", err)
			return
		default:
			log.Fatalf("%v, closing connection", err)
			return
		}

		// If it is a subfragment
		subfrag := data.Subfragment{}
		errFragment := subfrag.Deserialize(encMsg)
		if errFragment == nil {
			socket.Channel <- &subfrag
			continue
		}

		// If it is a kill_message
		kill := desc.KillMessage{}
		errKill := kill.Deserialize(encMsg)

		if errKill != nil {
			os.Exit(-1)
			// log.Fatalf("Could not decode message due to '%v'", util.ReturnFirstErr(errFragment, errKill))
		} else {
			defer socket.Stop()
			defer close(socket.Channel)
			return
		}
	}
}
