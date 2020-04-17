package main

import (
	"encoding/json"

	"github.com/iterum-provenance/sidecar/transmit"
)

type filelist []string

// Serialize tries to transform `fl` into a json encoded bytearray. Errors on failure
func (fl *filelist) Serialize() (data []byte, err error) {
	data, err = json.Marshal(fl)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `fl`. Errors on failure
func (fl *filelist) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, fl)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}
