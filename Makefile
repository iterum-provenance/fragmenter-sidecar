.PHONY: FORCE

NAME=fragmenter-sidecar

link: FORCE 
	@echo "Trying to link the executable to your path:"
	sudo ln -fs "${PWD}/build/${NAME}" /usr/bin/${NAME}
	@echo "Use ${NAME} to test the program and 'make clean' to remove"

clean: FORCE
	sudo rm /usr/bin/${NAME}

build: FORCE 
	go mod edit -dropreplace=github.com/iterum-provenance/iterum-go
	go mod edit -dropreplace=github.com/iterum-provenance/sidecar
	go build -o ./build/${NAME} -modfile go.mod

local: FORCE
	go mod edit -replace=github.com/iterum-provenance/iterum-go=$(GOPATH)/src/github.com/iterum-provenance/iterum-go
	go mod edit -replace=github.com/iterum-provenance/sidecar=$(GOPATH)/src/github.com/iterum-provenance/sidecar
	go build -o ./build/$(NAME)