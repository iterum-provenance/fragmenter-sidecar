// Package env contains setup for important environment variables used by the fragmenter-sidecar.
// Thanks to the init function the usage of these environment variables is checked before they are used
// this ensures that if any of these have an invalid value the applications will crash immediately at the
// start of execution.
package env

import (
	"os"
	"path"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/process"
	"github.com/iterum-provenance/iterum-go/util"

	"github.com/iterum-provenance/fragmenter/env/config"
)

const (
	inputSocketEnv  = "FRAGMENTER_INPUT"
	outputSocketEnv = "FRAGMENTER_OUTPUT"
)

// FragmenterInputSocket is the path to the socket used for fragmenter input
var FragmenterInputSocket = path.Join(process.DataVolumePath, os.Getenv(inputSocketEnv))

// FragmenterOutputSocket is the path to the socket used for fragmenter output
var FragmenterOutputSocket = path.Join(process.DataVolumePath, os.Getenv(outputSocketEnv))

// Config , if it exists, contains additional configuration information for the fragmenter-sidecar
var Config *config.Config = nil

// VerifyFragmenterSidecarEnvs checks whether each of the environment variables returned a non-empty value
func VerifyFragmenterSidecarEnvs() error {
	if !strings.HasSuffix(FragmenterInputSocket, ".sock") {
		return env.ErrEnvironment(inputSocketEnv, FragmenterInputSocket)
	} else if !strings.HasSuffix(FragmenterOutputSocket, ".sock") {
		return env.ErrEnvironment(outputSocketEnv, FragmenterOutputSocket)
	}
	return nil
}

// VerifyFragmenterSidecarConfig verifies the config struct of the fragmenter sidecar
func VerifyFragmenterSidecarConfig() error {
	c := config.Config{}
	errConfig := c.FromString(process.Config)
	if errConfig != nil {
		return errConfig
	}
	Config = &c
	return nil
}

func init() {
	errSidecar := VerifyFragmenterSidecarEnvs()
	errSidecarConf := VerifyFragmenterSidecarConfig()

	err := util.ReturnFirstErr(errSidecar, errSidecarConf)
	if err != nil {
		log.Fatalln(err)
	}
}
