startainer, a container starter utility
=======================================

This is a utility script, especially for Windows systems, to start Docker containers without having to type or copy/paste all the parameters each time. On Linux and OSX, aliases can be used. On Windows, ehm... not. 

Other than that, this was a good intro to the GO programming language ;-)


**Note**: this was renamed from `docker-starter` to `startainer` to account for the possibility to use different container runtimes other than `docker`. The default configuration file is now  (mac/linux) `~/.startainer.yaml`, (windows) `~\startainer.yaml` instead of `dockerstarter.yaml`. 


First setup
-----------

### Executables
Pre-built executables are contained within the [`builds/` directory](builds/). Copy the one for your architecture within a folder covered by the PATH environment variable.

**Tip**: rename the downloaded executable to something both mnemonical for you and short, so that you have to type less in at every execution. E.g.: `cs` (container start).

### Configuration file
The utility relies on a configuration file describing which and how a container need to be started. **By default**, this file is within your home directory: 

- on mac/linux: `~/.startainer.yaml`
- on windows: `~\startainer.yaml` 

**Note:** the name and path of the configuration file can be altered using the `-c` command line parameter.

A default config file is already provided: [startainer.yaml](src/startainer.yaml): you can copy it to your home folder (on Linux/OSX, rename it with as `.startainer.yaml`).

**Syntax of the configuration file**

```yaml
# Settings are OPTIONAL
settings:
  # If you are not using docker, set here the name of your container manager.
  # This setting is optional and will default to 'docker'
  runtime: podman

<config-name>:
  image: <name of the image to be pulled>
  message: This gets printed-out to the user just before container run/start. It is useful to communicate stuff like mapped ports and shared volumes.
  run: #list of command-line parameters for the 'docker run' command. One on each item. Example
    - --rm
    - -d
    - -name=<config-name>
    - -v=~/:/share
    - <image>
  exec: #optional, list of command-line parameters for the 'docker exec' command. If not provided, 'docker exec -ti <config-name> /bin/bash' will be used
<config-name2>: 
  #....

<docker-compose-name>:
  message: This gets printed-out to the user just before "compose up". It is useful to communicate stuff like mapped ports and shared volumes.
  # path to the compose yaml file
  compose: ~/path/to/compose/file/docker-compose.yml
  up: #list of command-line parameters for the "compose up" command. One on each item. Example
    - -d
    - --wait
    - --wait-timeout
    - 45

```

**A sample configuration file**

```yaml
alpine: # config name
  message: Alpine container available!
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

myawesomecomposeproject:
  message: "Some message to be printed when starting the stack"
  # path to the compose yaml file
  compose: ~/myawesomecomposeproject/docker-compose.yml
  # optional list of parameters to be provided to the "compose up" command
  up:
    - -d
```

**A sample configuration file using the podman container manager**

```yaml
settings:
  runtime: podman

alpine: # config name
  message: Alpine container available!
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
```



### Important configuration topics:

- The name of the configuration definition and the `--name` parameter of the container **MUST** be the same, otherwise the tool will not find the container anymore while it is running.
- When starting a container, the tool attaches its console's standard-out, -in and -err to the "docker run" command.
- When starting a docker compose stack, the tool attaches its console's standard-out, -in and -err to the "compose up" command.
- If you specify the `image` configuration, the tool will try a `docker pull` (or `podman pull` if you set a different runtime)
- Container run/exec configurations can be provided on a single line using format `-x=VALUE` (the `=` sign MUST be there).

### Bash Completion

If you want to type even less – assuming `bash-completion` is already installed on your system – create the file
`/etc/bash_completion.d/startainer.bash` and add the following:

```
_startainer()
{
    COMPREPLY=($(egrep  "^[a-z]" ~/.startainer.yaml | cut -d: -f1))
}

complete -F _startainer startainer
```

**Note:** The last word in the file (here: `startainer`) has to match the name of your binary. So if you followed along, it could be `cs` or `dstart` instead.


Usage
-----
The syntax is: 

```bash
    startainer [-c <config-file-name.yaml>] [-l] <config-name> [additional optional parameters for the 'run'  or 'up' command]
```

Any command-line parameters after the name of the definition are provided to the container through the `run` or `up` command. 

The tool will: 

1. check wheter the config-name references a container or compose definition
2. if container:
    1. checks the status of the container: missing, stopped, running.
    2. if running: execute a `docker exec` (or `podman exec` if you configured such runtime)
    3. if not running, it checks if the referenced container is stopped.
    4. if stopped: execute a `docker start` (or `podman start` if you configured such runtime)
    5. if the container is not found, then execute a `docker run` (or `podman run` if you configured such runtime)
3. if docker compose definition:
    1. checks whether the compose stack is: missing, stopped, running.
    2. if running: do nothing
    3. if stopped: execute a `docker compose up`, which will restart the existing containers
    4. if missing, execute a `docker compose up`, which will startup the containers

### Command-line flags

- `-c Full path` : (optional) full path to the configuration file. If not provided, `~/.startainer.yaml` is used;
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
  startainer -l
  
    > Reading configuration file '~/.startainer.yaml'
    > The available container definitions are:
      - alpine       (container status: running)
      - splunk80     (container status: missing)
      - splunk81     (container status: stopped)
```

```bash
  # get status of a container and list its configurations.
  startainer -l splunk80
  
    > Reading configuration file '~/.startainer'
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
  startainer splunk81

  # start the container definition called "de-utils"
  startainer de-utils

  # start the container definition called "de-utils" and provide additional startup parameters
  startainer de-utils build.py
```

```bash
    # start the container splunk80 based on the configuration file ./config.yaml
    startainer -c ./config.yaml splunk80
```


Internal working
----------------
The tool:
- reads the name of the container definition from the command line;
- accesses the configuration file and reads the container definition;
- checks if a container with that name is already existing, and whether it is running;
- if the container does not exist: the `run` command is used; an interactive shell is attached to it and volumes are mounted;
- if the container exists but is not running: the `start` command is used and a shell is attached to it;
- if the container exists but is running, the `exec` command is used to attach a new shell inside it.


Development
-----------
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
