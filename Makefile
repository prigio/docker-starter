IMAGE_NAME=golang:1.16
#this is there the src files are located, within the container
#the name of the directory might be used by GO for the name of the executable
WORKDIR=/usr/src/docker-starter
#this is where build files are to be stored, within the container
BUILDSDIR=/usr/local/bin
VOL_SRC="${PWD}:${WORKDIR}"
VOL_BUILDS="${PWD}/builds:${BUILDSDIR}"
#the libraries here are populated by the go container itself
VOL_LIBS="${PWD}/go_build_libs:/go"
#Ref: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

BUILD_CMD_DOCKER=docker run --rm -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -e GOOS -e GOARCH -w ${WORKDIR} ${IMAGE_NAME}
DEV_CMD_DOCKER=docker run --rm -ti -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -w ${WORKDIR} ${IMAGE_NAME}

default: osx

all: clean pull build_all

build_all: osx linux win osxm1

clean:
	@echo "> Deleting built executables"
	find builds/ -type f -delete

pull:
	# this command might require a "docker login" to be performed
	docker pull $(IMAGE_NAME) | true

osx:
	@echo "> Compiling executable for OSX within ${BUILDSDIR}/osx/"
	GOOS=darwin GOARCH=amd64 ${BUILD_CMD_DOCKER} go build -o ${BUILDSDIR}/osx/

win:
	@echo "> Compiling executable for Windows within ${BUILDSDIR}/windows/"
	GOOS=windows GOARCH=amd64 ${BUILD_CMD_DOCKER} go build -o ${BUILDSDIR}/windows/

linux:
	@echo "> Compiling executable for Linux within ${BUILDSDIR}/linux/"
	GOOS=linux GOARCH=amd64 ${BUILD_CMD_DOCKER} go build -o ${BUILDSDIR}/linux/

osxm1: 
	@echo "> Compiling executable for OSX M1 within ${BUILDSDIR}/osx_m1/"
	GOOS=darwin GOARCH=arm64 ${BUILD_CMD_DOCKER} go build -o ${BUILDSDIR}/osx_m1/

dev:
	@echo "> Starting interactive container to perform local test"
	@echo "> You can execute 'go run main.go'"
	${DEV_CMD_DOCKER}
