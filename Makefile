IMAGE_NAME=golang:alpine
CONTAINER_NAME=godev
#Environment settings for cross compilation
#Ref: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04
ENV_OSX=-e GOOS=darwin -e GOARCH=amd64
ENV_WIN=-e GOOS=windows -e GOARCH=amd64
ENV_LIN=-e GOOS=linux -e GOARCH=amd64
#this is there the src files are located, within the container
WORKDIR=/usr/src/
#this is where build files are to be stored, within the container
BUILDSDIR=/usr/local/bin
VOL_SRC="${PWD}/src:${WORKDIR}"
VOL_BUILDS="${PWD}/builds:${BUILDSDIR}"

default: build_osx

all: clean pull build_win build_osx build_linux

clean:
	@echo "> Removing dev container"
	docker stop ${CONTAINER_NAME} 2>/dev/null | true
	docker rm ${CONTAINER_NAME} 2>/dev/null | true
	@echo "> Deleting built executables"
	find builds/ -type f -delete

pull:
	#docker login
	docker pull $(IMAGE_NAME) | true

build_osx:
	@echo "> Compiling executable for OSX within ${BUILDSDIR}/osx/"
	docker run --rm --name ${CONTAINER_NAME} ${ENV_OSX} -v ${VOL_SRC} -v ${VOL_BUILDS} -w ${WORKDIR} ${IMAGE_NAME} go build -o ${BUILDSDIR}/osx/

build_win:
	@echo "> Compiling executable for Windows within ${BUILDSDIR}/windows/"
	docker run --rm --name ${CONTAINER_NAME} ${ENV_WIN} -v ${VOL_SRC} -v ${VOL_BUILDS} -w ${WORKDIR} ${IMAGE_NAME} go build -o ${BUILDSDIR}/windows/

build_linux:
	@echo "> Compiling executable for Linux within ${BUILDSDIR}/linux/"
	docker run --rm --name ${CONTAINER_NAME} ${ENV_LIN} -v ${VOL_SRC} -v ${VOL_BUILDS} -w ${WORKDIR} ${IMAGE_NAME} go build -o ${BUILDSDIR}/linux/

dev:
	@echo "> Starting interactive container to perform local test"
	@echo "> You can execute 'go run main.go'"
	docker run --rm -ti -v ${VOL_SRC} -w ${WORKDIR} ${IMAGE_NAME}
