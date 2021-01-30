# Change log

## v1.2 - 2021-01-30
- Added command-line flag `-q` to activate **quiet** mode, which disables the logging of the utility.

## v1.1 - 2021-01-14

- Fixed an issue which arises if a container definition is called as a container image: `docker inspect` was analyzing the image status instead of the container. Fixed by specializing the command using respectively: `docker image inspect ...` and `docker container inspect ...`.
- Improved logging for `docker exec`
- Added command line parameter `-v` to get script version
- Makefile: added local development folder to save go libraries downloaded when compiling.

## 2021-01-12

- Expansion of `.` and `~` in volume configurations is now also supported if the config is all on a single line. (e.g.: `-v=~/exchange`). Close #1
- Expansion of `.` and `~` in volume configurations is now also supported if the config is espressed with the `--mount` format (see <https://docs.docker.com/storage/bind-mounts/>)
- if the parameters of the `run` config do not contain a name for the container, the name of the configuration is uses as a container name

## 2021-01-11 

- Modified sample `dockerstarter.yaml`
- Added (optional) configuration `image` to a container configuration to specify the image name
- Added facilities to check if an image exists and to `docker pull` it if missing
