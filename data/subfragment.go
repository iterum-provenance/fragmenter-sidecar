package data

import (
	"encoding/json"
	"errors"

	"github.com/iterum-provenance/iterum-go/transmit"
)

// Subfragment is the incomplete structure returned from a fragmenter, which is then further enriched
// and completed to a full FragmentDesc by the fragmenter-sidecar before sending it on to the message queue
type Subfragment struct {
	Files    []string               `json:"files"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Serialize tries to transform `sf` into a json encoded bytearray. Errors on failure
func (sf *Subfragment) Serialize() (data []byte, err error) {
	data, err = json.Marshal(sf)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `sf`. Errors on failure
func (sf *Subfragment) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, sf)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	if len(sf.Files) == 0 {
		return transmit.ErrSerialization(errors.New("Invalid subfragment value, cannot contain 0 files"))
	}
	return
}
