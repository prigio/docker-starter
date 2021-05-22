package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/yalp/jsonpath"
)

const (
	VERSION string = "1.5.1" // version of the script. To be updated after each sensible update
	// Common status codes for containers/images
	MISSING        string = "missing"
	STOPPED        string = "stopped"
	RUNNING        string = "running"
	IMAGE_EXISTING string = "image_existing"
	ERROR          string = "error"
)

// Read the content of the text file and save it within the script.
// This happens AT COMPILE TIME
// For this reason the makefile within the repository copies the readme.md to src/ ;-)
//go:embed readme.md
var Readme string

//go:embed changelog.md
var ChangeLog string

//functions defining color styles
var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

// styleStatus analyzes the status of a container
// and returns a colored stirng string corresponding to it
func styleStatus(status string) string {
	switch status {
	case MISSING:
		return red(status)
	case STOPPED:
		return yellow(status)
	case RUNNING:
		return green(status)
	default:
		return status
	}
}

func DockerExec(dockerCmd string, containerName string, docker_exec_args []string) {
	log.Printf("Attaching an additional session to running container '%s'", containerName)
	// add "exec" at the beginning of the arguments
	docker_exec_args = append([]string{"exec"}, docker_exec_args...)
	cmd := exec.Command(dockerCmd, docker_exec_args...)
	// Redirect all input and output of the parent to the child process
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	// this is used to be able to read the stderr of the docker command
	var errb bytes.Buffer
	cmd.Stderr = &errb
	log.Printf("Command line arguments are:\n  %s", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	switch err.(type) {
	case nil: // program terminates here in best case
	case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
		log.Printf("An error occurred when executing container\nCommand line arguments were:\n%s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// this is raised id the executed command does not return 0
		exitError, _ := err.(*exec.ExitError)
		// 137 Indicates failure as container received SIGKILL,
		// which happens when the user terminates the container from another shell or via docker stop
		// however, both of these cases are OK for use.
		switch {
		case exitError.ExitCode() == 137:
			log.Print("Container terminated.")
		default:
			// bash returns the last exticode on exit. So if the user performed a command within the container
			// and that command raised an error, then the user exits the shell (with ctrl+D)
			// the parent program intercepts the exit code of the program within the container
			// we cannot distinguish on them here....
			log.Printf("Session terminated. Exit code is %d. %s", exitError.ExitCode(), errb.String())
		}
	}
}

func DockerStart(dockerCmd string, containerName string, docker_start_args []string) {
	log.Printf("Restarting stopped container '%s'", containerName)
	docker_start_args = append([]string{"start"}, docker_start_args...)
	cmd := exec.Command(dockerCmd, docker_start_args...)
	// Redirect all input and output of the parent to the child process
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	// this is used to be able to read the stderr of the docker command
	var errb bytes.Buffer
	cmd.Stderr = &errb
	err := cmd.Run()
	switch err.(type) {
	case nil: // program terminates here in best case
	case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
		log.Printf("An error occurred when starting container\nCommand line arguments were:\n  %s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// this is raised id the executed command does not return 0
		//exitError, _ := err.(*exec.ExitError)
		log.Printf("An error occurred when starting container\nCommand line arguments were:\n  %s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	}
}

func DockerRun(dockerCmd string, containerName string, docker_run_args []string) {
	log.Printf("Starting docker container '%s'", containerName)

	// Replace ~ and . within volume definitions
	prev_conf := ""
	containerName_was_set := false
	for i, curr_conf := range docker_run_args {
		// if previous conf item is a volume definition flag
		if prev_conf == "-v" || prev_conf == "--volume" {
			// replace ~ and . with their local, absolute counterparts
			docker_run_args[i], _ = ExpandPath(curr_conf)
		} else if strings.HasPrefix(curr_conf, "-v=") || strings.HasPrefix(curr_conf, "-v ") {
			// format of config: '-v=host-path:container-path' or '-v host-path:container-path'
			expanded_path, _ := ExpandPath(curr_conf[3:])
			docker_run_args[i] = fmt.Sprintf("-v=%s", expanded_path)
		} else if strings.HasPrefix(curr_conf, "--volume=") || strings.HasPrefix(curr_conf, "--volume ") {
			// format of config: '-v=host-path:container-path' or '-v host-path:container-path'
			expanded_path, _ := ExpandPath(curr_conf[9:])
			// force the current config to use the format --conf=val to avoid having empty spaces within it
			docker_run_args[i] = fmt.Sprintf("--volume=%s", expanded_path)
		} else if prev_conf == "--mount" || strings.HasPrefix(curr_conf, "--mount=") || strings.HasPrefix(curr_conf, "--mount ") {
			if strings.HasPrefix(curr_conf, "--mount ") {
				// replace the empty space with the = sign
				curr_conf = "--mount=" + curr_conf[8:]
			}
			if path_pos := strings.Index(curr_conf, "source="); path_pos >= 0 {
				expanded_path, _ := ExpandPath(curr_conf[path_pos+7:])
				docker_run_args[i] = fmt.Sprintf("%ssource=%s", curr_conf[0:path_pos], expanded_path)
			} else if path_pos := strings.Index(curr_conf, "src="); path_pos >= 0 {
				expanded_path, _ := ExpandPath(curr_conf[path_pos+4:])
				docker_run_args[i] = fmt.Sprintf("%ssrc=%s", curr_conf[0:path_pos], expanded_path)
			}
		} else if curr_conf == "--name" || strings.HasPrefix(curr_conf, "--name=") {
			containerName_was_set = true
		}

		prev_conf = curr_conf
	}

	if containerName_was_set {
		// prepend the "run" parameter
		docker_run_args = append([]string{"run"}, docker_run_args...)
	} else {
		// prepend the "run" parameter, and force the container name
		docker_run_args = append([]string{"run", "--name=" + containerName}, docker_run_args...)
	}

	cmd := exec.Command(dockerCmd, docker_run_args...)
	// Redirect all input and output of the parent to the child process
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	// this is used to be able to read the stderr of the docker command
	var errb bytes.Buffer
	// cmd.Stderr = os.Stderr
	cmd.Stderr = &errb

	log.Printf("Container startup arguments are:\n  %s", strings.Join(cmd.Args, " "))

	// execute the command and wait for its completion
	err := cmd.Run()
	// check for errors, depending by their type
	switch err.(type) {
	case nil:
		// program terminates here in best case
	case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
		log.Printf("An error occurred when starting container. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// this is raised id the executed command does not return 0
		exitError, _ := err.(*exec.ExitError)
		/*if exitError.ExitCode() != 125 {
			log.Printf("Unexpected error. Exit code is %d", exitError.ExitCode())
			log.Fatal(exitError)
		} else if strings.Contains(errb.String(), "Conflict.") {
			// the container is already running, try using the exec command
			DockerExec(docker, containerName)
		} else { // */
		log.Printf("Unexpected error by executing 'docker run'. Exit code is %d", exitError.ExitCode())
		log.Print(errb.String())
		log.Fatal(exitError)
		//}
	}
}

func ExpandPath(path string) (string, error) {
	if len(path) == 0 || (path[0] != '~' && path[0] != '.') {
		return path, nil
	}
	// retrieve home dir of current user
	home, herr := homedir.Dir()
	if herr != nil {
		return "", herr
	}

	//retrieve current working directory
	cwd, cerr := os.Getwd()
	if cerr != nil {
		return "", cerr
	}

	if path == "~" || path == "~"+string(os.PathSeparator) {
		return home, nil
	} else if path == "." || path == "."+string(os.PathSeparator) {
		return cwd, nil
	} else if strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return filepath.Join(home, path[2:]), nil
	} else if strings.HasPrefix(path, "~") {
		return filepath.Join(home, path[1:]), nil
	} else if strings.HasPrefix(path, "."+string(os.PathSeparator)) {
		return filepath.Join(cwd, path[2:]), nil
	} else if strings.HasPrefix(path, ".") {
		return filepath.Join(cwd, path[1:]), nil
	} else {
		return path, nil
	}
}

func ListSingleContainer(dockerCmd string, containerName string) {
	var configsList []string
	status, err := ContainerStatus(dockerCmd, containerName, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("The container '%s' is %s", bold(containerName), styleStatus(status))
	configsList = viper.GetStringSlice(containerName + ".run")
	log.Printf("RUN configurations for the container:\n    %s run\n    %s\n", dockerCmd, strings.Join(configsList, "\n    "))

	if viper.IsSet(containerName + ".exec") {
		configsList = viper.GetStringSlice(containerName + ".exec")
		log.Printf("EXEC configurations for the container:\n    %s exec\n    %s\n", dockerCmd, strings.Join(configsList, "\n    "))
	}
	if viper.IsSet(containerName + ".start") {
		configsList = viper.GetStringSlice(containerName + ".start")
		log.Printf("EXEC configurations for the container:\n    %s start\n    %s\n", dockerCmd, strings.Join(configsList, "\n    "))
	}
}

func ListConfigs(dockerCmd string) {
	// create slice of strings
	var definitionsList []string
	// create empty map of string->boolean
	statusMap := make(map[string]bool)
	// populate the map of all the available container definitions
	for _, key := range viper.AllKeys() {
		// key looks like: 'pagvpn.run', 'pagvpn.exec', 'splunk80.run', ...
		definition := strings.SplitN(key, ".", 2)[0]
		if _, ok := statusMap[definition]; !ok {
			// get info about container
			status, err := ContainerStatus(dockerCmd, definition, false)
			if err != nil {
				log.Fatal(err)
			}
			statusMap[definition] = true
			definitionsList = append(definitionsList, fmt.Sprintf("%-12s (container status: %s)", definition, styleStatus(status)))
		}
	}
	// print out the definitions
	sort.Strings(definitionsList)
	log.Printf("The available container definitions are:\n  - %s", strings.Join(definitionsList, "\n  - "))
}

func DockerPull(dockerCmd string, image_name string, verbose bool) (err error) {
	var outb, errb bytes.Buffer
	log.Printf("Pulling image '%s'.\n  If this fails, you might have to manually perform 'docker login' or 'docker login <registry>'", image_name)
	cmd := exec.Command(dockerCmd, "image", "pull", image_name)
	if verbose {
		// redirect child's process output to StdOut so that user can see it.
		cmd.Stdout = os.Stdout
	} else {
		// keep stdout internal
		cmd.Stdout = &outb
	}
	// this is used to be able to read stderr of the docker command
	cmd.Stderr = &errb
	err = cmd.Run()
	switch err.(type) {
	case nil:
		log.Print("Image downloaded")
		return nil
	case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
		log.Printf("An error occurred when executing 'docker pull' . Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// this is raised id the executed command does not return 0
		exitError, _ := err.(*exec.ExitError)
		switch {
		case exitError.ExitCode() == 1 && strings.Contains(errb.String(), "not found"):
			log.Printf("Image '%s' not found. Docker wrote:\n  %s", image_name, errb.String())
			log.Fatal(exitError)
		default:
			// check if the error was raised at the system level, such as if docker is not installed.
			log.Printf("An error occurred when executing 'docker pull'. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
			log.Print(errb.String())
			log.Fatal(exitError)
		}
	}
	return nil
}

func ImageStatus(dockerCmd string, image_name string, verbose bool) (status string, err error) {
	var outb, errb bytes.Buffer
	if verbose {
		log.Printf("Retrieving information about image '%s'", image_name)
	}

	cmd := exec.Command(dockerCmd, "image", "inspect", image_name)
	// Redirect all input and output of the parent to the child process
	// this is used to be able to read the stdout and stderr of the docker command
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	switch err.(type) {
	case nil:
		// the container is present, need to check if it is running or not
		var inspect_output interface{}
		err = json.Unmarshal(outb.Bytes(), &inspect_output)
		if err != nil {
			log.Print("Impossible to convert output of 'docker inspect' to Json")
			log.Fatal(err)
			return ERROR, err
		}
		_, err = jsonpath.Read(inspect_output, "$[0].Created")
		if err != nil {
			log.Print("Error when reading 'docker image inspect' output")
			log.Fatal(err)
			return ERROR, err
		}
		if verbose {
			log.Print("Image already existing")
		}
		return IMAGE_EXISTING, nil
	case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
		log.Printf("An error occurred when executing 'docker image inspect'. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// this is raised id the executed command does not return 0
		exitError, _ := err.(*exec.ExitError)
		switch {
		case exitError.ExitCode() == 1 && (strings.Contains(errb.String(), "Error: No such image") || strings.Contains(errb.String(), "Error: No such object")):
			// the image is missing
			return MISSING, nil
		default:
			// check if the error was raised at the system level, such as if docker is not installed.
			log.Printf("An error occurred when executing 'docker inspect'. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
			log.Print(errb.String())
			log.Fatal(exitError)
		}
	}
	return ERROR, err
}

func ContainerStatus(dockerCmd string, containerName string, verbose bool) (status string, err error) {
	var outb, errb bytes.Buffer
	if verbose {
		log.Printf("Retrieving information about container '%s'", containerName)
	}

	cmd := exec.Command(dockerCmd, "container", "inspect", containerName)
	// Redirect all input and output of the parent to the child process
	// this is used to be able to read the stdout and stderr of the docker command
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	switch err.(type) {
	case nil:
		// the container is present, need to check if it is running or not
		var inspect_output interface{}
		var is_running interface{}
		err = json.Unmarshal(outb.Bytes(), &inspect_output)
		if err != nil {
			log.Print("Impossible to convert output of 'docker inspect' to Json")
			log.Fatal(err)
			return ERROR, err
		}
		is_running, err = jsonpath.Read(inspect_output, "$[0].State.Running")
		if err != nil {
			log.Print("Error when reading 'docker container inspect' output")
			log.Printf("Is_running = %v", is_running)
			log.Fatal(err)
			return ERROR, err
		} else {
			if is_running.(bool) {
				// container is running
				return RUNNING, nil
			} else {
				return STOPPED, nil
			}
		}
	case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
		log.Printf("An error occurred when executing 'docker container inspect'. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// this is raised id the executed command does not return 0
		exitError, _ := err.(*exec.ExitError)
		switch {
		case exitError.ExitCode() == 1 && (strings.Contains(errb.String(), "Error: No such container") || strings.Contains(errb.String(), "Error: No such object")):
			// the container is missing, need to "run"
			return MISSING, nil
		default:
			// check if the error was raised at the system level, such as if docker is not installed.
			log.Printf("An error occurred when executing 'docker inspect'. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
			log.Print(errb.String())
			log.Fatal(exitError)
		}
	}
	return ERROR, err
}

func IsIn(val string, list []string) bool {
	if len(list) == 0 {
		return false
	}
	for _, v := range list {

		if val == v {
			return true
		}
	}
	return false
}

func main() {
	var (
		defaultConfigFile string
		dockerCmd         string
		containerName     string
		configFile        string
		flagListConfigs   bool
		flagVersion       bool
		flagReadme        bool
		flagChangeLog     bool
		flagQuiet         bool
		flagNoColor       bool
		additionalArgs    []string
	)

	// Remove date&time from logging format
	//	https://golang.org/pkg/log/#SetFlags
	log.SetFlags(0)
	// Prepend a "> " to each output line
	log.SetPrefix("> ")

	if runtime.GOOS == "windows" {
		// in case we are running on windows, the docker executable needs the ".exe" extension
		dockerCmd = "docker.exe"
		defaultConfigFile = "~/dockerstarter.yaml"
	} else {
		dockerCmd = "docker"
		defaultConfigFile = "~/.dockerstarter.yaml"
	}

	// Define command line parameters
	// https://gobyexample.com/command-line-flags
	flag.StringVar(&configFile, "c", defaultConfigFile, "`Full path` to a configuration file")
	flag.BoolVar(&flagListConfigs, "l", false, "If provided without any additional parameters, the script lists all the available container definitions and the status of the corresponding container, then exits. If provided with the name of a container definition the script displays the container status and its configurations.")
	flag.BoolVar(&flagVersion, "version", false, "If provided, print out the script version and then exits")
	flag.BoolVar(&flagReadme, "readme", false, "If provided, print out the complete documentation and then exits")
	flag.BoolVar(&flagChangeLog, "changelog", false, "If provided, print out the complete changelog and then exits")
	flag.BoolVar(&flagQuiet, "quiet", false, "Activate quiet mode: do not emit any internal logging")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable colored output")

	// parse cmd-line parameters
	flag.Parse()

	if flagNoColor {
		color.NoColor = true // disables colorized output
	}
	if flagVersion {
		fmt.Printf("%s version %s\n", os.Args[0], VERSION)
		return
	}
	if flagReadme {
		fmt.Println(Readme)
		return
	}
	if flagChangeLog {
		fmt.Println(ChangeLog)
		return
	}
	if flagQuiet {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("Reading configuration file '%s'", configFile)
	configFile, _ = ExpandPath(configFile)

	// Read-in the configuration file
	viper.SetConfigType("yaml")
	viper.SetConfigName(filepath.Base(configFile))
	config_dir, _ := filepath.Split(configFile)
	if config_dir == "" {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(config_dir)
	}

	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Fatalf("Config file '%s' not found", configFile)
		} else {
			// Config file was found but another error was produced
			log.Fatalf("Fatal error when opening config file: '%s'. %s", configFile, err)
		}
	}

	if flagListConfigs && flag.NArg() > 0 {
		ListSingleContainer(dockerCmd, flag.Arg(0))
		return
	} else if flagListConfigs {
		// the user asked to list all available configurations. Do that and exit
		ListConfigs(dockerCmd)
		return
	}

	if flag.NArg() == 0 {
		log.Fatal("Specify the name of a container as defined within the configuration file, or `-l` to list all definitions")
	}

	//.Args() is an array of the remaining parameters provided, which do not have a name
	if flag.NArg() > 1 {
		additionalArgs = flag.Args()[1:]
	}

	containerName = flag.Arg(0)

	//log.Printf("Retrieving information about container '%s'", containerName)
	status, err := ContainerStatus(dockerCmd, containerName, true)
	if err != nil {
		log.Fatal(err)
	}
	switch status {
	case MISSING:
		// the container is missing, need to "run"
		if viper.IsSet(containerName + ".image") {
			// check if the image is actually available
			// if not, pull it.
			image_name := viper.GetString(containerName + ".image")
			if image_status, err := ImageStatus(dockerCmd, image_name, true); err != nil {
				log.Fatal("%v", err)
			} else {
				if image_status == MISSING {
					if err := DockerPull(dockerCmd, image_name, true); err != nil {
						log.Fatal("%v", err)
					}
				}
			}
		}
		if viper.IsSet(containerName + ".run") {
			docker_run_args := append(viper.GetStringSlice(containerName+".run"), additionalArgs...)

			// Append the command-line parameters the user provided to the docker run command, to the ones specified within the config file
			DockerRun(dockerCmd, containerName, docker_run_args)
		} else {
			log.Fatal(red("No configurations for 'docker run' are present within the config file"))
		}
	case STOPPED:
		if viper.IsSet(containerName + ".start") {
			DockerStart(dockerCmd, containerName, viper.GetStringSlice(containerName+".start"))
		} else {
			log.Print("The container is stopped, but no configurations for 'docker start' are present within the config file. Defaulting to standard command")
			if IsIn("-d", viper.GetStringSlice(containerName+".run")) {
				// The "run" command specifies detached mode (-d), thus, by default, we do not attach stdin and stdout when doing start
				DockerStart(dockerCmd, containerName, []string{containerName})
			} else {
				DockerStart(dockerCmd, containerName, []string{"-ai", containerName})
			}
		}
	case RUNNING:
		if viper.IsSet(containerName + ".exec") {
			DockerExec(dockerCmd, containerName, viper.GetStringSlice(containerName+".exec"))
		} else {
			log.Print("The container is already running, but no configurations for 'docker exec' are present within the config file. Defaulting to standard command")
			DockerExec(dockerCmd, containerName, []string{"-ti", containerName, "/bin/bash"})
		}
	}
}