package handler

import (
	"net"
	"os"

	"github.com/prometheus/common/log"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/socket"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"

	"github.com/iterum-provenance/fragmenter/data"
)

// Receiver is a handler for a socket that receives fragmented file list from the fragmenter
func Receiver(socket socket.Socket, conn net.Conn) {
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
			log.Errorf("Could not decode message due to '%v'", util.ReturnFirstErr(errFragment, errKill))
			os.Exit(-1)
		} else {
			defer socket.Stop()
			defer close(socket.Channel)
			return
		}
	}
}
