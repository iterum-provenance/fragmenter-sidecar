package env

import (
	"log"
	"os"
	"strings"

	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/util"
)

const (
	inputSocketEnv  = "FRAGMENTER_INPUT"
	outputSocketEnv = "FRAGMENTER_OUTPUT"
)

// FragmenterInputSocket is the path to the socket used for fragmenter input
var FragmenterInputSocket = env.DataVolumePath + "/" + os.Getenv(inputSocketEnv)

// FragmenterOutputSocket is the path to the socket used for fragmenter output
var FragmenterOutputSocket = env.DataVolumePath + "/" + os.Getenv(outputSocketEnv)

// VerifyFragmenterSidecarEnvs checks whether each of the environment variables returned a non-empty value
func VerifyFragmenterSidecarEnvs() error {
	if !strings.HasSuffix(FragmenterInputSocket, ".sock") {
		return env.ErrEnvironment(inputSocketEnv, FragmenterInputSocket)
	} else if !strings.HasSuffix(FragmenterOutputSocket, ".sock") {
		return env.ErrEnvironment(outputSocketEnv, FragmenterOutputSocket)
	}
	return nil
}

func init() {
	errIterum := env.VerifyIterumEnvs()
	errDaemon := env.VerifyDaemonEnvs()
	errMinio := env.VerifyMinioEnvs()
	errMessageq := env.VerifyMessageQueueEnvs()
	errSidecar := VerifyFragmenterSidecarEnvs()

	err := util.ReturnFirstErr(errIterum, errMinio, errMessageq, errDaemon, errSidecar)
	if err != nil {
		log.Fatalln(err)
	}
}
