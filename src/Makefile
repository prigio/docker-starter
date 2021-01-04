#Environment settings for cross compilation
#Ref: https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04
ENV_OSX=GOOS=darwin GOARCH=amd64
ENV_WIN=GOOS=windows GOARCH=amd64
ENV_LIN=GOOS=linux GOARCH=amd64
#this is where build files are to be stored, within the container.
#this path must be the same mapped through a volume within the container (done within makefile in the parent folder)
BUILDSDIR=/usr/local/bin

default: osx

all: clean win osx linux

clean:
	@echo "(container)> Deleting built executables"
	find ${BUILDSDIR}/ -type f -delete

osx:
	@echo "(container)> Compiling executable for OSX within ${BUILDSDIR}/osx/"
	${ENV_OSX} go build -o ${BUILDSDIR}/osx/

win:
	@echo "(container)> Compiling executable for Windows within ${BUILDSDIR}/windows/"
	${ENV_WIN} go build -o ${BUILDSDIR}/windows/

linux:
	@echo "(container)> Compiling executable for Linux within ${BUILDSDIR}/linux/"
	${ENV_LIN} go build -o ${BUILDSDIR}/linux/
