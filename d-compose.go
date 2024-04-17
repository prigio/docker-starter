package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// dockerComposePSJsonOutput is used to unmarshal the output of `docker compose ps -a --format json“
// and check for status of the containers defined within the compose file
// by default, the JSON unmashaler ignores the values present within JSON but not in the struct
type dockerComposePSJsonOutput struct {
	ID     string `json:"ID"`
	Name   string `json:"Name"`
	State  string `json:"State"`
	Status string `json:"Status"`
}

func ManageCompose(containerManagerCmd string, composeConfName string, additionalArgs []string) {
	//log.Printf("Retrieving information about container '%s'", containerName)
	status, err := ComposeStatus(containerManagerCmd, composeConfName, true)
	if err != nil {
		log.Fatal(err)
	}
	switch status {
	case COMPOSEFILENOTFOUND:
		log.Fatalf("Configuration file for compose stack '%s' not found", composeConfName)
	case MISSING:
		// the containers for the compose file are stopped or missing to "up"
		ComposeUp(containerManagerCmd, composeConfName, viper.GetString(composeConfName+".message"))
	case STOPPED:
		// the containers for the compose file are stopped or missing to "up"
		ComposeUp(containerManagerCmd, composeConfName, viper.GetString(composeConfName+".message"))
	case RUNNING:
		log.Printf("The compose stack '%s' is already running", composeConfName)
	}
}

func ComposeStatus(containerManagerCmd string, composeConfName string, verbose bool) (status string, err error) {
	var outb, errb bytes.Buffer
	compose := viper.GetString(composeConfName + ".compose")

	fullpath, err := ExpandPath(compose)
	if err != nil {
		log.Printf("Impossible to expand path of file %s", compose)
		return ERROR, err
	}
	if !FileExists(fullpath) {
		return COMPOSEFILENOTFOUND, nil
	}

	composeDir := filepath.Dir(fullpath)
	composeFile := filepath.Base(compose)

	if verbose {
		log.Printf("Retrieving information about compose '%s', '%s'", composeConfName, compose)
	}
	cmd := exec.Command(containerManagerCmd, "compose", "-f", composeFile, "ps", "-a", "--format", "json")
	// execute the command in the folder where the compose-file is contained.
	cmd.Dir = composeDir

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	switch err.(type) {
	case nil:
		var compose_output []dockerComposePSJsonOutput
		err = json.Unmarshal(outb.Bytes(), &compose_output)
		if err != nil {
			return ERROR, err
		}
		if len(compose_output) == 0 {
			return MISSING, nil
		}
		for _, instance := range compose_output {
			if instance.State != "running" {
				return STOPPED, nil
			}
		}
		return RUNNING, nil
	case *exec.Error:
		// check if the error was raised at the system level, such as if container manager is not installed.
		log.Printf("System-error occurred when executing '%s compose ps'. Command line arguments were:\n  %s", containerManagerCmd, strings.Join(cmd.Args, " "))
		log.Print(errb.String())
		log.Fatal(err)
	case *exec.ExitError:
		// check if the error was raised at the command level, such as if command failed.
		exitError, _ := err.(*exec.ExitError)
		switch {
		/*
			docker compose -f <filename> ps <name>
				returns in case of missing compose file:
					return value: 14
					stderr: Error: stat /Users/pprigione/tmp/docker-compose.yml: no such file or directory
		*/
		case (exitError.ExitCode() == 14 && (strings.Contains(errb.String(), "no such file or directory"))):
			return COMPOSEFILENOTFOUND, nil
		default:
			log.Printf("Exit-error occurred when executing '%s compose ps'. Command line arguments were:\n  %s", containerManagerCmd, strings.Join(cmd.Args, " "))
			log.Print(errb.String())
			//log.Fatal(err)
		}
	}
	return ERROR, err
}

func ComposeUp(containerManagerCmd string, composeConfName string, message string) {
	log.Printf("Starting compose stack '%s'", composeConfName)

	var errb bytes.Buffer
	compose := viper.GetString(composeConfName + ".compose")

	fullpath, err := ExpandPath(compose)
	if err != nil {
		log.Fatalf("Impossible to expand path of file '%s'", compose)
		return
	}
	if !FileExists(fullpath) {
		log.Fatalf("Compose file not found '%s'", fullpath)
		return
	}

	composeDir := filepath.Dir(fullpath)
	composeFile := filepath.Base(compose)

	cmd := exec.Command(containerManagerCmd, "compose", "-f", composeFile, "up")
	// execute the command in the folder where the compose-file is contained.
	cmd.Dir = composeDir
	// Redirect all input and output of the parent to the child process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// this is used to be able to read the stderr of the container manager command
	cmd.Stderr = &errb

	log.Printf("Compose startup arguments are:\n  %s", strings.Join(cmd.Args, " "))
	if message != "" {
		log.Print(blue(message))
	}

	err = cmd.Run()
	switch err.(type) {
	case nil:
		// program terminates here in best case

	case *exec.Error:
		// check if the error was raised at the system level, such as if docker compose is not installed.
		log.Printf("An error occurred when starting compose stack. Command line arguments were:\n  %s", strings.Join(cmd.Args, " "))
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
			ContainerExec(containerManagerCmd, containerName)
		} else { */
		log.Printf("Unexpected error by executing '%s compose up'. Exit code is %d", containerManagerCmd, exitError.ExitCode())
		log.Print(errb.String())
		log.Fatal(exitError)
	}
}
