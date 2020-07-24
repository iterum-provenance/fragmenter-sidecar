package data

import (
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/iterum-provenance/iterum-go/daemon"
	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/util"
)

// pullAndUploadFile downloads a file from the daemon and uploads it to minio
func pullAndUploadFile(minio minio.Config, daemon daemon.Config, filePath string, retries int) (remoteFile desc.RemoteFileDesc, err error) {
	defer util.ReturnErrOnPanic(&err)()
	if !minio.IsConnected() {
		return remoteFile, fmt.Errorf("Minio client not initialized, cannot pull and send data")
	}

	// Get the data
	resp, err := http.Get(daemon.DaemonURL + path.Join("/", daemon.Dataset, "file", filePath))
	util.PanicIfErr(err, "")

	localFile := desc.LocalFileDesc{
		Name:      filepath.Dir(filePath),
		LocalPath: filePath,
	}
	files, err := minio.PutFileFromReader(resp.Body, resp.ContentLength, localFile, false)
	for idx := 0; idx < retries && err != nil; idx++ {
		// Get the data again. The stream/filehandle is removed after each attempt.
		resp, err := http.Get(daemon.DaemonURL + path.Join("/", daemon.Dataset, "file", filePath))
		util.PanicIfErr(err, "")

		files, err = minio.PutFileFromReader(resp.Body, resp.ContentLength, localFile, false)

		// Sleep for a max of 2 seconds, and at least 1
		time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
	}
	return files, err
}
