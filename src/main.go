package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"bytes"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"github.com/yalp/jsonpath"
)


func run_docker_exec(docker_cmd string, container_name string) {
	log.Print("Attaching an additional session to a running container")
	docker_exec_args := []string{"exec", "-ti", container_name, "/bin/bash" }
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
				case exitError.ExitCode() ==1 && strings.Contains(errb.String(), "is not running"):
					// the container is present but stopped, need to start it.
					run_docker_start(docker_cmd, container_name)
				case exitError.ExitCode() == 137: 
					log.Print("Container terminated.")
				default:
					// bash returns the last exticode on exit. So if the user performed a command within the container
					// and that command raised an error, then the user exits the shell (with ctrl+D)
					// the parent program intercepts the exit code of the program within the container
					// we cannot distinguish on them here....
					log.Printf("Session terminated. Exit code is %d", exitError.ExitCode())				
			} 
	}
	return
}

func run_docker_start(docker_cmd string, container_name string) {
	docker_start_args := []string{"start", "-ai", container_name,}
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

func run_docker_run(docker_cmd string, container_name string) {
	log.Printf("Starting docker container '%s'", container_name)
	var image = "git.cocus.com:5005/d/porsche-vpn"
	
	home, herr := homedir.Dir()
	if herr != nil {
		log.Fatal(herr)
	}
	// retrieve current working directory
	/*
	cwd, cerr := os.Getwd()
	if cerr != nil {
		log.Fatal(cerr)
	}
	*/

	//log.Println("Home dir is: " + home)
	//log.Println("Current dir is: " + cwd)
	// Prepare command-line parameters for the docker command

	// command-line parameters for docker run
    var docker_run_args = []string{
		"run", 
		"--rm", 
		"-ti", 
		"--name", container_name, 
		"--hostname", container_name,
		"--cap-add", "NET_ADMIN",
		"-e", "PAG_USER",
		"-e", "PAG_PIN",
		"-e", "PAG_CERT_PASS",
		"-e", "JIRA_USER",
		"-e", "JIRA_PASS",
		"-p", "8888:8888",
		"-p", "10022:22",
		"-p", "18000:8000",
		"-p", "18089:8089",
		"-v", "paghome:/root",
		"-v", home + ":/exchange",
		image,
	}
	
	// Append the command-line parameters the user provided to the docker run command.
	docker_run_args = append(docker_run_args, os.Args[1:]...)

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
}


func main() {
	var docker_cmd = "docker"
	var container_name = "pagvpn"
	
	// in case we are running on window, the docker executable is windows-based
	if  runtime.GOOS == "windows" {
		docker_cmd = "docker.exe"
	}
	
	log.Printf("Retrieving information about container '%s'", container_name)
	cmd := exec.Command(docker_cmd, "inspect", container_name)	
	// Redirect all input and output of the parent to the child process
	var outb, errb bytes.Buffer	
	// this is used to be able to read the stdout and stderr of the docker command
	cmd.Stdout = &outb	
	cmd.Stderr = &errb
	err := cmd.Run()
	switch err.(type) {
		case nil:
			// the container is present, need to check if it is running or not
			var inspect_output interface{}
			ierr := json.Unmarshal(outb.Bytes(), &inspect_output)
			if ierr != nil {
				log.Print("Impossible to convert output of 'docker inspect' to Json")
				log.Fatal(ierr)
			}
			is_running, jerr := jsonpath.Read(inspect_output, "$[0].State.Running")
			if jerr != nil {
				log.Print("Error when reading 'docker inspect' output")
				log.Print("Is_running = %s", is_running)
				log.Fatal(jerr)
			} else {
				if is_running.(bool) {
					run_docker_exec(docker_cmd, container_name)
				} else {
					run_docker_start(docker_cmd, container_name)
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
					run_docker_run(docker_cmd, container_name)				
				default:
					// check if the error was raised at the system level, such as if docker is not installed.
					log.Printf("An error occurred when executing 'docker inspect'\nCommand line arguments were:\n%s", strings.Join(cmd.Args," "))
					log.Print(errb.String())
					log.Fatal(exitError)
			}
	}
}

