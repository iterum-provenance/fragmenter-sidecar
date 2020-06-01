package config

import (
	"encoding/json"
	"regexp"

	data "github.com/iterum-provenance/fragmenter/data"
	"github.com/prometheus/common/log"
)

// Config is the struct holding configurable information
// This can be set via the environment variable ITERUM_CONFIG
type Config struct {
	ConfigSelectors []*regexp.Regexp // nillable
}

// FromString converts a string value into an instance of Config and also does validation
func (conf *Config) FromString(stringified string) (err error) {
	var placeholder struct {
		ConfigFiles []string `json:"config_files"`
	}
	err = json.Unmarshal([]byte(stringified), &placeholder)
	if err != nil {
		return err
	}
	conf.ConfigSelectors = []*regexp.Regexp{}
	for _, selector := range placeholder.ConfigFiles {
		regex, err := regexp.Compile(selector)
		if err != nil {
			return err
		}
		conf.ConfigSelectors = append(conf.ConfigSelectors, regex)
	}
	return conf.Validate()
}

// Validate does validation of a config struct,
// ensuring that it's members contain valid values
func (conf Config) Validate() error {
	return nil
}

// ReturnMatchingFiles returns the list of files matching the ConfigSelectors of conf
func (conf Config) ReturnMatchingFiles(files data.Filelist) (matches data.Filelist) {
	matches = data.Filelist([]string{})
	fullRegexp := ""
	first := true
	// Create 1 big regex
	for _, regex := range conf.ConfigSelectors {
		if first {
			fullRegexp += "(" + regex.String() + ")"
		} else {
			fullRegexp += "| (" + regex.String() + ")"
		}
	}
	if fullRegexp == "" {
		return
	}
	// Compile the big regex
	regex, err := regexp.Compile(fullRegexp)
	if err != nil {
		log.Errorln(err)
		return
	}
	// Match each file against the large regexp once
	for _, file := range files {
		if regex.MatchString(file) {
			matches = append(matches, file)
		}
	}
	return matches
}
