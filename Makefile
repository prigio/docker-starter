IMAGE_NAME=golang:1.16-rc
CONTAINER_NAME=golang-dockerstarter
#Environment settings for cross compilation
#Ref: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04
ENV_OSX=-e GOOS=darwin -e GOARCH=amd64
ENV_WIN=-e GOOS=windows -e GOARCH=amd64
ENV_LIN=-e GOOS=linux -e GOARCH=amd64
#this is there the src files are located, within the container
#the name of the directory might be used by GO for the name of the executable
WORKDIR=/usr/src/docker-starter
#this is where build files are to be stored, within the container
BUILDSDIR=/usr/local/bin
VOL_SRC="${PWD}/src:${WORKDIR}"
VOL_BUILDS="${PWD}/builds:${BUILDSDIR}"
#the libraries here are populated by the go container itself
VOL_LIBS="${PWD}/go_build_libs:/go"

default: osx

all: clean pull build_all

clean:
	@echo "> NOT removing locally saved build libraries in go_build_libs"
	@echo "> Removing copy of readme.md from src/ folder"
	rm src/readme.md 2>/dev/null
	@echo "> Removing dev container"
	docker stop ${CONTAINER_NAME} 2>/dev/null | true
	docker rm ${CONTAINER_NAME} 2>/dev/null | true
	@echo "> Deleting built executables"
	find builds/ -type f -delete

pull:
	# this command might require a "docker login" to be performed
	docker pull $(IMAGE_NAME) | true

build_all:
	@echo "> Copying readme.md to src/ folder"
	cp readme.md src/
	@echo "> Compiling executable for all targets within ${BUILDSDIR}/ using src/Makefile"
	docker run --rm --name ${CONTAINER_NAME} -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -w ${WORKDIR} ${IMAGE_NAME} make all

osx:
	@echo "> Copying readme.md to src/ folder"
	cp readme.md src/
	@echo "> Compiling executable for OSX within ${BUILDSDIR}/osx/ using src/Makefile"
	docker run --rm --name ${CONTAINER_NAME} -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -w ${WORKDIR} ${IMAGE_NAME} make osx

win:
	@echo "> Copying readme.md to src/ folder"
	cp readme.md src/
	@echo "> Compiling executable for Windows within ${BUILDSDIR}/windows/ using src/Makefile"
	docker run --rm --name ${CONTAINER_NAME} -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -w ${WORKDIR} ${IMAGE_NAME} make win
linux:
	@echo "> Copying readme.md to src/ folder"
	cp readme.md src/
	@echo "> Compiling executable for Linux within ${BUILDSDIR}/linux/ using src/Makefile"
	docker run --rm --name ${CONTAINER_NAME} -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -w ${WORKDIR} ${IMAGE_NAME} make linux

dev:
	@echo "> Copying readme.md to src/ folder"
	cp readme.md src/
	@echo "> Starting interactive container to perform local test"
	@echo "> You can execute 'go run main.go'"
	docker run --rm -ti -v ${VOL_SRC} -v ${VOL_BUILDS} -v ${VOL_LIBS} -w ${WORKDIR} ${IMAGE_NAME}
