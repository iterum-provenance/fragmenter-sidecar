package config

import (
	"encoding/json"
	"path"

	"github.com/iterum-provenance/fragmenter/data"
)

// Config is the struct holding configurable information
// This can be set via the environment variable ITERUM_CONFIG
// Config selectors contains all the config files used throughout this pipeline run
type Config struct {
	ConfigFiles map[string][]string `json:"config_files_all"` // nillable
}

// FromString converts a string value into an instance of Config and also does validation
func (conf *Config) FromString(stringified string) (err error) {
	err = json.Unmarshal([]byte(stringified), conf)
	if err != nil {
		return err
	}
	return conf.Validate()
}

// Validate does validation of a config struct,
// ensuring that it's members contain valid values
func (conf Config) Validate() error {
	return nil
}

// ReturnMatchingFiles returns the list of files matching the ConfigSelectors of conf
func (conf Config) ReturnMatchingFiles(commitFiles data.Filelist) (matches data.Filelist) {
	matches = data.Filelist([]string{})

	if len(commitFiles) == 0 {
		return matches
	}

	// Match each file against each selector
	for _, file := range commitFiles {
		filename := path.Dir(file)
		for _, selectors := range conf.ConfigFiles {
			for _, selector := range selectors {
				// path.Dir because a file from a commit is always: filename.extension/commit-hash-of-dataset
				if selector == filename {
					matches = append(matches, file)
				}
			}
		}
	}
	return matches
}
