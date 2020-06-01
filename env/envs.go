package env

import (
	"os"
	"regexp"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/util"

	"github.com/iterum-provenance/fragmenter/env/config"
)

const (
	inputSocketEnv  = "FRAGMENTER_INPUT"
	outputSocketEnv = "FRAGMENTER_OUTPUT"
)

// FragmenterInputSocket is the path to the socket used for fragmenter input
var FragmenterInputSocket = env.DataVolumePath + "/" + os.Getenv(inputSocketEnv)

// FragmenterOutputSocket is the path to the socket used for fragmenter output
var FragmenterOutputSocket = env.DataVolumePath + "/" + os.Getenv(outputSocketEnv)

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
	c := config.Config{ConfigSelectors: []*regexp.Regexp{}}
	if env.ProcessConfig == "" {
		log.Warnln("Fragmenter-sidecar was initialized without additional config, make sure that this was intended")
	} else {
		errConfig := c.FromString(env.ProcessConfig)
		if errConfig != nil {
			return errConfig
		}
	}
	Config = &c
	return nil
}

func init() {
	errIterum := env.VerifyIterumEnvs()
	errDaemon := env.VerifyDaemonEnvs()
	errMinio := env.VerifyMinioEnvs()
	errMessageq := env.VerifyMessageQueueEnvs()
	errSidecar := VerifyFragmenterSidecarEnvs()
	errSidecarConf := VerifyFragmenterSidecarConfig()

	err := util.ReturnFirstErr(errIterum, errMinio, errMessageq, errDaemon, errSidecar, errSidecarConf)
	if err != nil {
		log.Fatalln(err)
	}
}
