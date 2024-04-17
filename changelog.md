Change log
==========

## v3.0.0 - 2023-04-23
Added support for `docker compose` stacks.

## v2.0.1 - 2022-04-10
Fixes for podman runtime execution. It looks like `alias podman=docker` is not true when it comes to error messages and exit codes.

## v2.0.0 - 2022-04-10 
- Tool renamed to `startainer` to account for possibility to use different runtimes other than `docker`.
- New functionalities: 
    - config file: added `settings.runtime` to specify alternate container runtime.
    - config file: added `<containerdefinition>.message` to print a custom message to the user after starting a container. 

## v1.5.1 - 2021-05-12
- Changed default configuration file on linux from to include .yaml extension
- Refactored variable names so that they conform to golang conventions
- Added make targets for OSX M1 (arm) architectures

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
