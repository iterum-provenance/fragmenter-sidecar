.PHONY: FORCE

NAME=fragmenter-sidecar


build: FORCE 
	go build -o ./build/${NAME}


link: FORCE 
	@echo "Trying to link the executable to your path:"
	sudo ln -fs "${PWD}/build/${NAME}" /usr/bin/${NAME}
	@echo "Use ${NAME} to test the program and 'make clean' to remove"

clean: FORCE
	sudo rm /usr/bin/${NAME}
	
