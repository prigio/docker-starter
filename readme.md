# Docker Starter utility

This is a utility script, especially for Windows systems, to start docker containers without having to type or copy/paste all the parameters each time. On Linux and OSX, aliases can be used. On Windows, ehm... not. 

Other than that, this was a good intro to the GO programming language ;-)


## First setup

### Executables
Pre-built executables are contained within the [`builds/` directory](builds/). Copy the one for your architecture within a folder covered by the PATH environment variable.

### Configuration file
The utility relies on a configuration file describing which and how a container need to be started. 

**By default**, this file is within your home directory: `~/.dockerstarter.yaml` (on windows: `~\dockerstarter.yaml`). **Note:** the name and path of the configuration file can be altered using the `-c` command line parameter.

A default config file is already provided: [dockerstarter.yaml](src/dockerstarter.yaml): you can copy it to your home folder (on Linux/OSX, rename it with as `.dockerstarter.yaml`).
A sample configuration is the following, defining the command-line parameters for the most important docker commands: run, exec and start

```yaml
#Splunk v8.1.1 container
splunk81:
  run:
    - -d
    - --name
    - splunk81
    - --hostname
    - splunk81-docker
    - -p
    - 38081:8000
    - -v
    - .:/exchange
    - -e
    - SPLUNK_START_ARGS=--accept-license
    - -e
    - SPLUNK_PASSWORD=splunked
    - splunk/splunk:8.1.1
  exec:
    - -ti
    - splunk81
    - /bin/bash
  start:
    - -ai
    - splunk81
```

### Important configuration topics:

- The name of the definition and the `--name` parameter of the container **MUST** be the same, otherwise the tool will not find the container anymore.
- As of now, the tool attaches its the console's standard-out, -in and -err to the container.
- The tool **DOES NOT** perform a `docker pull`: you must perform it once before using a container definition


## Usage
The syntax is: 

```bash
    docker-starter [-c <config-file-name.yaml>] <container-definition-name> [additional optinal parameters for 'docker run']
```

Example:

```bash
    # start the container definition called "splunk81"
    docker-starter splunk81

    # start the container definition called "pagvpn"
    docker-starter pagvpn

    # start the container splunk80 based on the configuration file ./config.yaml
    docker-starter -c ./config.yaml splunk80

    # start the container definition called "pagvpn" and provide a VPN token code already
    docker-starter pagvpn 123456
```

Any command-line parameters after the name of the definition are provided to the container within docker run. So these are all possible: 


## Internal working
The tool:
- reads the name of the container definition from the command line
- accesses the configuration file and reads the container definition
- checks if a container with that name is already existing, and wether its running. 
- if the container does not exist: the `run` command is used; an interactive shell is attached to it and volumes are mounted;
- if the container exists but is not running: the `start` command is used and a shell is attached to it;
- if the container exists but is running, the `exec` command is used to attach a new shell inside it.

## Development
You need `docker` and `make` available within your system to build this effortlessy: the build system is based on a `golang` container.

See the `Makefile` for instructions to build.

Available make commands: 

- `make pull`: downloads the necessary docker golang container
- `make all`: performs all the build chain: clean pull build_win build_osx build_linux
- `make osx`: builds the executable for OSX
- `make win`: builds the executable for Windows
- `make linux`: builds the executable for Linux
- `make clean`: cleans-up old builds
- `make dev`: launches an interactive golang container to support development.

