package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/iterum-provenance/fragmenter/data"
	"github.com/iterum-provenance/iterum-go/daemon"
	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/util"
	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/cli/idv"
)

var (
	errNotFound = errors.New("Error: Daemon responded with 404, resource not found")
)

// _get takes a url to fire a get request upon and a pointer to an interface to store the result in
// It returns an error on failure of either http.Get, Reading response or Unmarshalling json body
func _get(url string, target interface{}) (err error) {
	defer util.ReturnErrOnPanic(&err)()

	resp, err := http.Get(url)
	util.PanicIfErr(err, "")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.PanicIfErr(err, "")

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		return errNotFound
	default:
		return fmt.Errorf("Error: GET failed, daemon responded with statuscode %v", resp.StatusCode)
	}

	err = json.Unmarshal([]byte(body), target)
	util.PanicIfErr(err, "")

	return
}

// getCommitFiles pulls a specific commmit based on its hash and dataset and passed daemonURL
// it returns the list of files associated with this commmit
func getCommitFiles(config daemon.Config) (files data.Filelist, err error) {
	commit := idv.Commit{}
	err = _get(config.DaemonURL+"/"+config.Dataset+"/commit/"+config.CommitHash, &commit)
	return data.Filelist(commit.Files), err
}

func pullAndUploadFile(minio minio.Config, daemon daemon.Config, filePath string, retries int) (remoteFile desc.RemoteFileDesc, err error) {
	defer util.ReturnErrOnPanic(&err)()
	if !minio.IsConnected() {
		return remoteFile, fmt.Errorf("Minio client not initialized, cannot pull and send data")
	}

	// Get the data
	resp, err := http.Get(daemon.DaemonURL + "/" + daemon.Dataset + "/file/" + filePath)
	util.PanicIfErr(err, "")

	localFile := desc.LocalFileDesc{
		Name:      filepath.Dir(filePath),
		LocalPath: filePath,
	}
	files, err := minio.PutFileFromReader(resp.Body, resp.ContentLength, localFile)
	for idx := 0; idx < retries && err != nil; idx++ {
		// Get the data again. The stream/filehandle is removed after each attempt.
		resp, err := http.Get(daemon.DaemonURL + "/" + daemon.Dataset + "/file/" + filePath)
		util.PanicIfErr(err, "")

		files, err = minio.PutFileFromReader(resp.Body, resp.ContentLength, localFile)
		time.Sleep(1 * time.Second)
	}
	return files, err
}

func downloadConfigFileFromDaemon(daemon daemon.Config, filePath string) (local desc.LocalFileDesc, err error) {
	defer util.ReturnErrOnPanic(&err)()

	// Get the data
	url := daemon.DaemonURL + "/" + daemon.Dataset + "/file/" + filePath
	log.Infof("Url: %v", url)
	resp, err := http.Get(url)
	util.PanicIfErr(err, "")
	defer resp.Body.Close()
	log.Infof("Status: %v", resp.Status)
	log.Infof("ContentLength: %v", resp.ContentLength)

	if env.ProcessConfigPath == env.DataVolumePath {
		log.Fatalf("EnvironmentError: '%v' is not a valid value for ITERUM_CONFIG_PATH", env.ProcessConfigPath)
	}
	err = os.MkdirAll(env.ProcessConfigPath, os.ModePerm)
	path := path.Join(env.ProcessConfigPath, filepath.Dir(filePath))

	util.PanicIfErr(err, "")
	out, err := os.Create(path)
	util.PanicIfErr(err, "")
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	util.PanicIfErr(err, "")

	return desc.LocalFileDesc{
		Name:      filepath.Dir(filePath),
		LocalPath: path,
	}, nil
}
