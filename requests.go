package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/iterum-provenance/cli/idv"
	"github.com/iterum-provenance/sidecar/data"
	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/fragmenter/daemon"
	"github.com/iterum-provenance/fragmenter/minio"
	"github.com/iterum-provenance/fragmenter/util"
)

var (
	errNotFound = errors.New("Error: Daemon responded with 404, resource not found")
)

// _get takes a url to fire a get request upon and a pointer to an interface to store the result in
// It returns an error on failure of either http.Get, Reading response or Unmarshalling json body
func _get(url string, target interface{}) (err error) {
	defer util.ReturnErrOnPanic(&err)()

	resp, err := http.Get(url)
	util.PanicOnErr(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.PanicOnErr(err)

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		return errNotFound
	default:
		return fmt.Errorf("Error: GET failed, daemon responded with statuscode %v", resp.StatusCode)
	}

	err = json.Unmarshal([]byte(body), target)
	util.PanicOnErr(err)

	return
}

// getCommitFiles pulls a specific commmit based on its hash and dataset and passed daemonURL and returns a list of files
func getCommitFiles(config daemon.Config) (files filelist, err error) {
	commit := idv.Commit{}
	err = _get(config.DaemonURL+"/"+config.Dataset+"/commit/"+config.CommitHash, &commit)
	return filelist(commit.Files), err
}

func pullAndUploadFile(minio minio.Config, daemon daemon.Config, filePath string, retries int) (remoteFile data.RemoteFileDesc, err error) {
	defer util.ReturnErrOnPanic(&err)()
	if !minio.IsConnected() {
		return remoteFile, fmt.Errorf("Minio client not initialized, cannot pull and send data")
	}

	// Get the data
	resp, err := http.Get(daemon.DaemonURL + "/" + daemon.Dataset + "/file/" + filePath)
	util.PanicOnErr(err)
	defer resp.Body.Close()

	// Check to see if we already own this bucket
	exists, errBucketExists := minio.Client.BucketExists(minio.TargetBucket)
	if errBucketExists != nil {
		return remoteFile, fmt.Errorf("Upload failed due to failure of bucket existence checking: '%v'", errBucketExists)
	} else if !exists {
		log.Infof("Bucket '%v' does not exist, creating...\n", minio.TargetBucket)
		errMakeBucket := minio.Client.MakeBucket(minio.TargetBucket, "")
		if errMakeBucket != nil {
			if retries > 0 { // retry a number of times
				time.Sleep(1 * time.Second)
				log.Infof("Failed to create bucket '%v', retrying pullAndUpload...\n", minio.TargetBucket)
				return pullAndUploadFile(minio, daemon, filePath, retries-1)
			}
			return remoteFile, fmt.Errorf("Failed to create bucket '%v' due to: '%v'", minio.TargetBucket, errMakeBucket)
		}
	}

	_, err = minio.Client.PutObject(minio.TargetBucket, filePath, resp.Body, resp.ContentLength, minio.PutOptions)
	util.PanicOnErr(err)

	remoteFile = data.RemoteFileDesc{
		Name:       filepath.Dir(filePath),
		RemotePath: filePath,
		Bucket:     minio.TargetBucket,
	}

	return remoteFile, err
}
