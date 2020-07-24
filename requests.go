package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/iterum-provenance/iterum-go/daemon"
	"github.com/iterum-provenance/iterum-go/util"
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
func getCommitFiles(config daemon.Config) (files []string, err error) {
	commit := struct {
		Parent      string   `json:"parent"`
		Branch      string   `json:"branch"`
		Hash        string   `json:"hash"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Files       []string `json:"files"`
	}{}
	err = _get(config.DaemonURL+"/"+config.Dataset+"/commit/"+config.CommitHash, &commit)
	return commit.Files, err
}
