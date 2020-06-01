package data

import (
	"encoding/json"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
)

// FragmenterInput is the message format sent from the sidecar to the fragmenter
type FragmenterInput struct {
	ConfigFiles []desc.LocalFileDesc `json:"config_files"`
	DataFiles   Filelist             `json:"data_files"`
}

// Serialize tries to transform `fi` into a json encoded bytearray. Errors on failure
func (fi *FragmenterInput) Serialize() (data []byte, err error) {
	data, err = json.Marshal(fi)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `fi`. Errors on failure
func (fi *FragmenterInput) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, fi)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}
