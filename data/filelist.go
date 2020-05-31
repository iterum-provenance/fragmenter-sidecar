package data

import (
	"encoding/json"

	"github.com/iterum-provenance/iterum-go/transmit"
)

// Filelist is a serializable array of strings containing files
type Filelist []string

// Serialize tries to transform `fl` into a json encoded bytearray. Errors on failure
func (fl *Filelist) Serialize() (data []byte, err error) {
	data, err = json.Marshal(fl)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `fl`. Errors on failure
func (fl *Filelist) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, fl)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}
