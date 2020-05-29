set -a
source ./test.env

make build
fragmenter-sidecar
