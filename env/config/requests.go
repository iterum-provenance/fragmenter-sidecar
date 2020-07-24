package config

import (
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/iterum-provenance/iterum-go/daemon"
	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/process"
	"github.com/iterum-provenance/iterum-go/util"
)

func downloadConfigFileFromDaemon(daemon daemon.Config, filePath string) (local desc.LocalFileDesc, err error) {
	defer util.ReturnErrOnPanic(&err)()

	// Get the data
	url := daemon.DaemonURL + "/" + daemon.Dataset + "/file/" + filePath
	resp, err := http.Get(url)
	util.PanicIfErr(err, "")
	defer resp.Body.Close()

	// Create location to save it
	err = os.MkdirAll(process.ConfigPath, os.ModePerm)
	path := path.Join(process.ConfigPath, filepath.Dir(filePath))

	util.PanicIfErr(err, "")
	out, err := os.Create(path)
	util.PanicIfErr(err, "")
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	util.PanicIfErr(err, "")

	// return a local file description describing the downloaded file
	return desc.LocalFileDesc{
		Name:      filepath.Dir(filePath),
		LocalPath: path,
	}, nil
}
