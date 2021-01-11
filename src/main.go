package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sort"
	"bytes"
	"flag"
	//"regexp"
	"path/filepath"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"github.com/yalp/jsonpath"
	"github.com/spf13/viper"
)

const (
	MISSING string = "missing"
	STOPPED string = "stopped"
	RUNNING string = "running"
	IMAGE_EXISTING string = "image_existing"
	ERROR string = "error"
)

func run_docker_exec(docker_cmd string, container_name string, docker_exec_args []string) {
	log.Printf("Attaching an additional session to running container '%s'", container_name)
	// add "exec" at the beginning of the arguments
	docker_exec_args = append([]string{"exec"}, docker_exec_args...)
	cmd := exec.Command(docker_cmd, docker_exec_args...)	
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
			log.Printf("An error occurred when executing container\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
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
	return
}

func run_docker_start(docker_cmd string, container_name string, docker_start_args []string) {
	log.Printf("Restarting stopped container '%s'", container_name)
	docker_start_args = append([]string{"start"}, docker_start_args...)
	cmd := exec.Command(docker_cmd, docker_start_args...)
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
			log.Printf("An error occurred when starting container\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
			log.Print(errb.String())
			log.Fatal(err)
		case *exec.ExitError:
			// this is raised id the executed command does not return 0
			//exitError, _ := err.(*exec.ExitError)			
			log.Printf("An error occurred when starting container\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
			log.Print(errb.String())
			log.Fatal(err)			
	}
	return
}

/*
func prepare_docker_cli_arguments(args []string) ([]string, error) {
	if args == nil {
		return nil, nil
	} else if len(args) == 0 {
		return []string{}, nil
	}
	// the slice to be returned
	var prepared_args []string

	prev_conf := ""
	re := regexp.MustCompile(`^--?[\w_-]+`)
	for i, arg := range args {
		// if previous conf item is a volume definition flag
		// if there are no spaces within the config, it can be returned as is
		if ! strings.Contains(arg, " ") {
			prepared_args = append(prepared_args, arg)
		}

		if prev_conf == "-v" {
			// replace ~ and . with their local, absolute counterparts
			docker_run_args[i], _ = expand_path(curr_conf)
		}
		prev_conf = curr_conf
	}
}
// */

func run_docker_run(docker_cmd string, container_name string, docker_run_args []string) {
	log.Printf("Starting docker container '%s'", container_name)
	docker_run_args = append([]string{"run"}, docker_run_args...)

	// Replace ~ and . within volume definitions
	prev_conf := ""
	for i, curr_conf := range docker_run_args {
		// if previous conf item is a volume definition flag
		if prev_conf == "-v" {
			// replace ~ and . with their local, absolute counterparts
			docker_run_args[i], _ = expand_path(curr_conf)
		}
		prev_conf = curr_conf
	}
	
	cmd := exec.Command(docker_cmd, docker_run_args...)	
	// Redirect all input and output of the parent to the child process
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	// this is used to be able to read the stderr of the docker command
	var errb bytes.Buffer
	// cmd.Stderr = os.Stderr
	cmd.Stderr = &errb
	
	log.Printf("Container startup arguments are:\n%s", strings.Join(cmd.Args," "))
	
	// execute the command and wait for its completion
	err := cmd.Run()
	// check for errors, depending by their type
	switch err.(type) {
		case nil:
			// program terminates here in best case
		case *exec.Error:
			// check if the error was raised at the system level, such as if docker is not installed.
			log.Printf("An error occurred when starting container\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
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
				run_docker_exec(docker, container_name)
			} else { // */
			log.Printf("Unexpected error by executing 'docker run'. Exit code is %d", exitError.ExitCode())
			log.Print(errb.String())
			log.Fatal(exitError)
			//}		
	}
	return
}

func expand_path(path string) (string, error) {
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

	if (path == "~" || path == "~" + string(os.PathSeparator)) {
		return home, nil
	} else if (path == "." || path == "." + string(os.PathSeparator)) {
		return cwd, nil
	} else if strings.HasPrefix(path, "~" + string(os.PathSeparator)) {
		return filepath.Join(home, path[2:]), nil
	} else if strings.HasPrefix(path, "~") {
			return filepath.Join(home, path[1:]), nil
	} else if strings.HasPrefix(path, "." + string(os.PathSeparator)) {
		return filepath.Join(cwd, path[2:]), nil
	} else if strings.HasPrefix(path, ".") {
		return filepath.Join(cwd, path[1:]), nil
	} else {
		return path, nil
	}
}
func ListConfigs(docker_cmd string) {
		// create empty map of string->boolean
		statusMap := make(map[string]string)
		// create slice of strings
		var definitionsList []string
		// populate the map of all the available container definitions
		for _, key := range viper.AllKeys() {
			// key looks like: 'pagvpn.run', 'pagvpn.exec', 'splunk80.run', ...
			definition := strings.SplitN(key, ".", 2)[0]
			if _, ok := statusMap[definition]; ! ok {
				// get info about container
				status, err := ContainerStatus(docker_cmd, definition, false); 
				if err != nil {
					log.Fatal(err)
				}
				switch status {
					case MISSING:
						statusMap[definition] = Style(status, COLOR_RED)
					case STOPPED:
						statusMap[definition] = Style(status, COLOR_YELLOW)
					case RUNNING:
						statusMap[definition] = Style(status, COLOR_GREEN)
				}
	
				definitionsList = append(definitionsList, fmt.Sprintf("%-12s (container status: %s)", definition, statusMap[definition]))
			}
		}
		// print out the definitions
		sort.Strings(definitionsList)
		log.Printf("The available container definitions are:\n  - %s", strings.Join(definitionsList, "\n  - "))		
		return
} 

func DockerPull(docker_cmd string, image_name string, verbose bool) (err error) {	
	var outb, errb bytes.Buffer	
	log.Printf("Pulling image '%s'.\n\tIf this fails, you might have to manually perform 'docker login' or 'docker login <registry>'" , image_name)
	cmd := exec.Command(docker_cmd, "image", "pull", image_name)
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
			log.Printf("An error occurred when executing 'docker pull'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
			log.Print(errb.String())
			log.Fatal(err)
		case *exec.ExitError:
			// this is raised id the executed command does not return 0
			exitError, _ := err.(*exec.ExitError)
			switch {
				case exitError.ExitCode() == 1 && strings.Contains(errb.String(), "not found"):
					log.Printf("Image '%s' not found. Docker wrote:\n\t%s", image_name, errb.String())
					log.Fatal(exitError)
				default:
					// check if the error was raised at the system level, such as if docker is not installed.
					log.Printf("An error occurred when executing 'docker pull'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
					log.Print(errb.String())
					log.Fatal(exitError)
			}
	}
	return 
}

func ImageStatus(docker_cmd string, image_name string, verbose bool) (status string, err error) {
	var outb, errb bytes.Buffer
	if verbose {
		log.Printf("Retrieving information about image '%s'", image_name)
	}

	cmd := exec.Command(docker_cmd, "image", "inspect", image_name)
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
				log.Print("Error when reading 'docker inspect' output")				
				log.Fatal(err)
				return ERROR, err
			}
			if verbose {
				log.Print("Image already existing")
			}
			return IMAGE_EXISTING, nil			
		case *exec.Error:
		// check if the error was raised at the system level, such as if docker is not installed.
			log.Printf("An error occurred when executing 'docker inspect'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
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
					log.Printf("An error occurred when executing 'docker inspect'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
					log.Print(errb.String())
					log.Fatal(exitError)
			}
	}
	return ERROR, err
}

func ContainerStatus(docker_cmd string, container_name string, verbose bool) (status string, err error) {
	var outb, errb bytes.Buffer
	if verbose {
		log.Printf("Retrieving information about container '%s'", container_name)
	}

	cmd := exec.Command(docker_cmd, "inspect", container_name)
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
				log.Print("Error when reading 'docker inspect' output")
				log.Print("Is_running = %s", is_running)
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
			log.Printf("An error occurred when executing 'docker inspect'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
			log.Print(errb.String())
			log.Fatal(err)
		case *exec.ExitError:
			// this is raised id the executed command does not return 0
			exitError, _ := err.(*exec.ExitError)
			switch {
				case exitError.ExitCode() == 1 && strings.Contains(errb.String(), "Error: No such object"):
					// the container is missing, need to "run"
					return MISSING, nil
				default:
					// check if the error was raised at the system level, such as if docker is not installed.
					log.Printf("An error occurred when executing 'docker inspect'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
					log.Print(errb.String())
					log.Fatal(exitError)
			}
	}
	return ERROR, err
}

func IsIn(val string, list []string) bool {
	if list == nil || len(list) == 0 {
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
		docker_cmd string
		container_name string
		default_config_file string
		config_file string
		list_configs bool
		additional_args []string
	)
	//additional_args := []string{}
	
	// Remove date&time from logging format
	//	https://golang.org/pkg/log/#SetFlags
	log.SetFlags(0)
	// Prepend a "> " to each output line
	log.SetPrefix("> ")	

	if runtime.GOOS == "windows" {
		// in case we are running on windows, the docker executable needs the ".exe" extension
		docker_cmd = "docker.exe"
		default_config_file =  "~/dockerstarter.yaml"
	} else {
		docker_cmd = "docker"
		default_config_file = "~/.dockerstarter"
	}
	
	// Define command line parameters
	// https://gobyexample.com/command-line-flags
	flag.StringVar(&config_file, "c", default_config_file, "`Full path` to a configuration file")
	flag.BoolVar(&list_configs, "l", false, "If provided, the script lists all the available container definitions and the status of the corresponding container, then exits")
	// parse cmd-line parameters
	flag.Parse()

	log.Printf("Reading configuration file '%s'", config_file)
	config_file, _ = expand_path(config_file)
	
	// Read-in the configuration file
	viper.SetConfigType("yaml")
	viper.SetConfigName(filepath.Base(config_file))
	config_dir, _ := filepath.Split(config_file)
	if config_dir == "" {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(config_dir)
	}

	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Fatalf("Config file '%s' not found", config_file)
		} else {
			// Config file was found but another error was produced
			log.Fatalf("Fatal error when opening config file: '%s'. %s", config_file, err)
		}
	}
	
	if list_configs {
		// the user asked to list all available configurations. Do that and exit
		ListConfigs(docker_cmd)
		return
	}

	if flag.NArg() == 0 {
		log.Fatal("Specify the name of a container as defined within the configuration file, or `-l` to list all definitions")
	}
	
	//.Args() is an array of the remaining parameters provided, which do not have a name
	if flag.NArg() > 1 {
		additional_args = flag.Args()[1:]
	}

	container_name = flag.Arg(0)

	//log.Printf("Retrieving information about container '%s'", container_name)
	status, err := ContainerStatus(docker_cmd, container_name, true)
	if err != nil {
		log.Fatal(err)
	}
	switch status {
		case MISSING:
			// the container is missing, need to "run"
			if viper.IsSet(container_name + ".image") {
				// check if the image is actually available
				// if not, pull it.
				image_name := viper.GetString(container_name + ".image")
				if image_status, err := ImageStatus(docker_cmd, image_name, true); err != nil {
					log.Fatal("%v", err)
				} else {
					if image_status == MISSING {
						if err := DockerPull(docker_cmd, image_name, true); err != nil {
							log.Fatal("%v", err)
						}
					}
				}
			}
			if viper.IsSet(container_name + ".run") {
				docker_run_args := append(viper.GetStringSlice(container_name + ".run"), additional_args...)				
				
				// Append the command-line parameters the user provided to the docker run command, to the ones specified within the config file
				run_docker_run(docker_cmd, container_name, docker_run_args)
			} else {
				log.Fatal(Style("No configurations for 'docker run' are present within the config file", COLOR_RED))
			}
		case STOPPED:
			if viper.IsSet(container_name + ".start") {
				run_docker_start(docker_cmd, container_name, viper.GetStringSlice(container_name + ".start"))
			} else {
				log.Print("The container is stopped, but no configurations for 'docker start' are present within the config file. Defaulting to standard command")
				if IsIn("-d", viper.GetStringSlice(container_name + ".run")) {
					// The "run" command specifies detached mode (-d), thus, by default, we do not attach stdin and stdout when doing start
					run_docker_start(docker_cmd, container_name, []string{container_name})
				} else {
					run_docker_start(docker_cmd, container_name, []string{"-ai", container_name})
				}
			}
		case RUNNING:
			if viper.IsSet(container_name + ".exec") {
				run_docker_exec(docker_cmd, container_name, viper.GetStringSlice(container_name + ".exec"))
			} else {
				log.Print("The container is already running, but no configurations for 'docker exec' are present within the config file. Defaulting to standard command")
				run_docker_exec(docker_cmd, container_name, []string{"-ti", container_name, "/bin/bash"})
			}
	}
}

