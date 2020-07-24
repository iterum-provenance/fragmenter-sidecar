// Package config contains extra custom/optional configuration for the fragmenter-sidecar.
// This is passed as a single stringified JSON object in an environment variable.
//
// config.go defines this structure and  deals with finding and matching all files
// in a data set, to be used as configuration files.
// downloader.go downloads all configuration files of all the steps
// uploader.go uplaods the configuration files to a special MinIO bucket
package config
