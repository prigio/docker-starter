# Docker Starter utility

This is a utility script, especially for Windows systems, to start Docker containers without having to type or copy/paste all the parameters each time. On Linux and OSX, aliases can be used. On Windows, ehm... not. 

Other than that, this was a good intro to the GO programming language ;-)

## First setup

### Executables
Pre-built executables are contained within the [`builds/` directory](builds/). Copy the one for your architecture within a folder covered by the PATH environment variable.

### Configuration file
The utility relies on a configuration file describing which and how a container need to be started. 

**By default**, this file is within your home directory: `~/.dockerstarter.yaml` (on windows: `~\dockerstarter.yaml`). **Note:** the name and path of the configuration file can be altered using the `-c` command line parameter.

A default config file is already provided: [dockerstarter.yaml](src/dockerstarter.yaml): you can copy it to your home folder (on Linux/OSX, rename it with as `.dockerstarter.yaml`).

**Syntax of the configuration file.**

```yaml
<config-name>:
  image: <name of the image to be pulled>
  run: #list of command-line parameters for the 'docker run' command. One on each item
  exec: #optional, list of command-line parameters for the 'docker exec' command. If not provided, 'docker exec -ti <config-name> /bin/bash' will be used
<config-name2>: 
  #....
```

**A sample configuration file**

```yaml
alpine: # config name
  image: alpine:latest
  # command-line parameters for "docker run"
  # `run` is a list of flags or parameters: 
  #   - They are provided AS-IS to the command line, 
  #     so you need to take care of quoting values having emtpy spaces
  #   - volume mounts parameters: automatically expand '.', '..', and '~' 
  #   - parameters having arguments can be split on multiple lines. E.g.:
  #        - -v
  #        -  myfolder:containerfolder
  #   - the `--name` MUST be the same as the config name
  run:
    - --rm
    - -ti
    # if the --name is not provided, it is set automatically to the name of the config
    - --name=alpine
    - -v=.:/srv
    - --workdir=/srv
    - alpine:latest
    # you can also add an additional command here
    # - some_command
  
  # command-line parameters for "docker exec". Can be omitted.
  # `exec` is a list of flags or parameters. 
  # If not specified, `-ti <configname> /bin/bash` will be used
  exec:
    - -ti
    # name of the running container, MUST be same as the definition
    - alpine
    # by default, alpine has no bash
    - /bin/sh 
#next config name
centos: 
  image: centos:8
  run:
    - ...
```

### Important configuration topics:

- The name of the configuration definition and the `--name` parameter of the container **MUST** be the same, otherwise the tool will not find the container anymore while it is running.
- As of now, the tool attaches its console's standard-out, -in and -err to the container.
- If you specify the `image` configuration, the tool will try a `docker pull`
- Docker configurations can be provided on a single line using format `-x=VALUE` (the `=` sign MUST be there).

## Usage
The syntax is: 

```bash
    docker-starter [-c <config-file-name.yaml>] [-l] <config-name> [additional optional parameters for 'docker run']
```
Any command-line parameters after the name of the definition are provided to the container within docker run. 

The tool will: 

1. check if the config-name references a running container.
2. if running: execute a `docker exec`
3. if not running, it checks if the referenced container is stopped.
4. if stopped: execute a `docker start`
5. if the container is not found, then execute a `docker start`

### Command-line flags

- `-c Full path` : (optional) full path to the configuration file. If not provided, `~/.dockerstarter.yaml` is used;
- `-l` : (optional) if provided:
  - _without any additional parameters_: the script lists all the available container definitions and the status of the corresponding container, then exits;
  - _with the name of a container definition_: the script displays the container status and its configurations;
- `-quiet`: (optional) Activate quiet mode: do not emit any internal logging;
- `-version`: if provided, print out the script version and then exits;
- `-readme` : if provided, print out the complete documentation and then exits;
- `-changelog`: if provided, print out the complete changelog and then exits;
- `-no-color`: disable colored output


### Examples

```bash
  # get list of container definitions, and corresponding status
  docker-starter -l
  
    > Reading configuration file '~/.dockerstarter.yaml'
    > The available container definitions are:
      - alpine       (container status: running)
      - splunk80     (container status: missing)
      - splunk81     (container status: stopped)
```

```bash
  # get status of a container and list its configurations.
  docker-starter -l splunk80
  
    > Reading configuration file '~/.dockerstarter'
    > The container 'splunk80' is running
    > RUN configurations for the container:
        docker run
        -d
        --name=splunk80
        --hostname=splunk80-docker
        -p=38080:8000
        -v=.:/exchange
    ....
```

```bash
  # start the container definition called "splunk81"
  docker-starter splunk81

  # start the container definition called "de-utils"
  docker-starter de-utils

  # start the container definition called "de-utils" and provide additional startup parameters
  docker-starter de-utils build.py
```

```bash
    # start the container splunk80 based on the configuration file ./config.yaml
    docker-starter -c ./config.yaml splunk80
```


## Internal working
The tool:
- reads the name of the container definition from the command line;
- accesses the configuration file and reads the container definition;
- checks if a container with that name is already existing, and whether it is running;
- if the container does not exist: the `run` command is used; an interactive shell is attached to it and volumes are mounted;
- if the container exists but is not running: the `start` command is used and a shell is attached to it;
- if the container exists but is running, the `exec` command is used to attach a new shell inside it.

## Development
You need `docker` and `make` available within your system to build this effortlessy: the build system is based on a `golang` container.

A local folder, `go_build_libs`, which is NOT tracked within GIT is used by the go container managed through the Makefile to save the libraries used when compiling. This saves time each time the container is restarted.

See the `Makefile` for instructions to build.

Available make commands: 

- `make pull`: downloads the necessary docker golang container
- `make all`: performs all the build chain: clean pull build_win build_osx build_linux
- `make osx`: builds the executable for OSX
- `make osxm1`: builds the executable for OSX M1 ARM architecture
- `make win`: builds the executable for Windows
- `make linux`: builds the executable for Linux
- `make clean`: cleans-up old builds
- `make dev`: launches an interactive golang container to support development.
