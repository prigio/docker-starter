# Change log

## v1.5 - 2021-02-18
- Project moved to <https://github.com>
- Added config-file syntax to documentation
- 
## v1.4 - 2021-02-18
- Added `-changelog` command line parameter to print this file out.
- Removed self-made coloring package which was not working on windows and adopted `https://github.com/fatih/color` instead.

## v1.3 - 2021-02-01
- Added possibility to list configuration for one specified container. E.g. `docker-starter -l splunk80`
- Added command-line flag - `-readme` to print the complete documentation.
- Renamed command-line flags:
    - `-q` -> `-quiet`
    - `-v` -> `-version`
- Updated from golang v1.15 to v1.16.

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
- if the parameters of the `run` config do not contain a name for the container, the name of the configuration is used as a container name

## 2021-01-11 

- Modified sample `dockerstarter.yaml`
- Added (optional) configuration `image` to a container configuration to specify the image name
- Added facilities to check if an image exists and to `docker pull` it if missing
