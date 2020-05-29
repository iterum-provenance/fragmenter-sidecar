package main

import (
	"encoding/json"

	"github.com/iterum-provenance/iterum-go/transmit"
)

// subfragment is the incomplete structure returned from a fragmenter, which is then further enriched
// and completed to a full FragmentDesc by the fragmenter-sidecar before sending it on to the message queue
type subfragment struct {
	Files    []string               `json:"files"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Serialize tries to transform `sf` into a json encoded bytearray. Errors on failure
func (sf *subfragment) Serialize() (data []byte, err error) {
	data, err = json.Marshal(sf)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `sf`. Errors on failure
func (sf *subfragment) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, sf)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}
