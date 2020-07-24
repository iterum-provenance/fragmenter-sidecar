package data

import (
	"encoding/json"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
)

// Upload is a struct mapping an idv file name/path
// to a remote file description as it is stored in minio
type Upload struct {
	File     string
	FileDesc desc.RemoteFileDesc
}

// Serialize tries to transform `upload` into a json encoded bytearray. Errors on failure
func (upload *Upload) Serialize() (data []byte, err error) {
	data, err = json.Marshal(upload)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `upload`. Errors on failure
func (upload *Upload) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, upload)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}
