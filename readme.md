# PAG VPN starter

This is a utility script, especially for Windows systems, to start the PAGVPN container without having to type or copy/paste all the parameters each time. On Linux and OSX, aliases can be used. On Windows, ehm... not. 

Other than that, this was a good intro to the GO programming language ;-)

## Usage
Executables are contained within the `build` directory. Copy the one for your architecture within a folder covered by the PATH environment variable and just start it. 

Any command-line parameters sent to the `pagvpn-starter` executable are provided to the container. So these are all possible: 

    pagvpn-starter
    pagvpn-starter 123456 (token code)
    pagvpn-starter -h
    pagvpn-starter -i
    pagvpn-starter -r

and, so on. 

## Internal working

- If the container named `pagvpn` does not exist: the `run` command is used; an interactive shell is attached to it and volumes are mounted;
- If the container named `pagvpn` exists but is not running: the `start` command is used and a shell is attached to it;
- If the container named `pagvpn` exists but is running, the `exec` command is used to attach a new shell inside it.

## Development
You need `docker` and `make` available within your system to build this effortlessy: the build system is based on a golang container.

See the `Makefile` for instructions to build. 

Available make commands: 

- `make all`: performs all the build chain: clean pull build_win build_osx build_linux
- `make dev`: launches an interactive golang container to support development.
- `make clean`: cleans-up old builds
- `make pull`: downloads the necessary docker golang container
- `make build_osx`: builds the executable for OSX
- `make build_win`: builds the executable for Windows
- `make build_linux`: builds the executable for Linux

