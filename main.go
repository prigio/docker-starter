package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

const (
	VERSION string = "3.0.0" // version of the script. To be updated after each sensible update
	// Common status codes for containers/images
	MISSING             string = "missing"
	STOPPED             string = "stopped"
	RUNNING             string = "running"
	IMAGE_EXISTING      string = "image_existing"
	ERROR               string = "error"
	COMPOSEFILENOTFOUND string = "compose file missing"
)

// Read the content of the text file and save it within the script.
// This happens AT COMPILE TIME
// For this reason the makefile within the repository copies the readme.md to src/ ;-)
//
//go:embed readme.md
var Readme string

//go:embed changelog.md
var ChangeLog string

// functions defining color styles
var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
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
	case COMPOSEFILENOTFOUND:
		return red(status)
	default:
		return status
	}
}

func ListConfigs(containerManagerCmd string) {
	// create empty map of string->boolean
	statusMap := make(map[string]bool)
	// populate the map of all the available container definitions
	containerDefinitions := viper.AllKeys()
	sort.Strings(containerDefinitions)
	// "settings" MUST NOT be analyzed, as this is use as global configuration
	// threfore, it is marked as already analyzed here :-)
	statusMap["settings"] = true
	log.Print("The available container definitions are:")
	for _, key := range containerDefinitions {
		// key looks like: 'pagvpn.run', 'pagvpn.exec', 'splunk80.run', ...
		definition := strings.SplitN(key, ".", 2)[0]
		// if this definition has not been analyzed yet
		if _, ok := statusMap[definition]; !ok {
			if ConfigType(definition) == CONFTYPECONTAINER {
				// get info about docker container configuration
				status, err := ContainerStatus(containerManagerCmd, definition, false)
				if err != nil {
					log.Fatal(err)
				}
				// print out the definition
				log.Printf("  - %-15s (container status: %s)", definition, styleStatus(status))
			} else {
				// get info about docker docker compose configuration
				status, err := ComposeStatus(containerManagerCmd, definition, false)
				if err != nil {
					log.Fatal(err)
				}
				// print out the definition
				log.Printf("  - %-15s (compose status: %s)", definition, styleStatus(status))
			}
			// tracks that this definition was already printed out
			statusMap[definition] = true
		}

	}
}

// readConfig initializes viper and reads the given configuration file
func readConfig(configFile string) error {
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
			return fmt.Errorf("config file '%s' not found", configFile)
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("fatal error when opening config file: '%s'. %s", configFile, err)
		}
	}
	return nil
}

func main() {
	var (
		defaultConfigFile   string
		containerManagerCmd string
		definitionName      string
		configFile          string
		flagListConfigs     bool
		flagVersion         bool
		flagReadme          bool
		flagChangeLog       bool
		flagQuiet           bool
		flagDown            bool
		flagNoColor         bool
		additionalArgs      []string
	)

	// Remove date&time from logging format
	//	https://golang.org/pkg/log/#SetFlags
	log.SetFlags(0)
	// Prepend a "> " to each output line
	log.SetPrefix("> ")

	if runtime.GOOS == "windows" {
		// in case we are running on windows, the container manager executable needs the ".exe" extension
		containerManagerCmd = "docker.exe"
		defaultConfigFile = "~/startainer.yaml"
	} else {
		containerManagerCmd = "docker"
		defaultConfigFile = "~/.startainer.yaml"
	}

	// Define command line parameters
	// https://gobyexample.com/command-line-flags
	flag.StringVar(&configFile, "c", defaultConfigFile, "`Full path` to a configuration file")
	flag.BoolVar(&flagListConfigs, "l", false, "If provided without any additional parameters, the script lists all the available container definitions and the status of the corresponding container, then exits. If provided with the name of a container definition the script displays the container status and its configurations.")
	flag.BoolVar(&flagDown, "down", false, "If provided, stops the container or compose stack")
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
		log.SetOutput(io.Discard)
	}

	log.Printf("Reading configuration file '%s'", configFile)
	configFile, _ = ExpandPath(configFile)
	if err := readConfig(configFile); err != nil {
		log.Fatal(err.Error())
	}

	// Check if the runtime setting is present within the configuration file.
	// The setting can be used to replace the standard docker with, for instance, podman.
	// However, the new container manager MUST implement the same command-line behavior of docker
	// because this tool relies on the command-line to issue commands.
	if viper.IsSet("settings.runtime") && viper.GetString("settings.runtime") != "docker" && viper.GetString("settings.runtime") != "docker.exe" {
		containerManagerCmd = viper.GetString("settings.runtime")

		if runtime.GOOS == "windows" && !strings.HasSuffix(containerManagerCmd, ".exe") {
			// in case we are running on windows, the container manager executable needs the ".exe" extension
			containerManagerCmd = containerManagerCmd + ".exe"
		}
		log.Printf("Set %s as container runtime", containerManagerCmd)
	}

	if flagListConfigs && flag.NArg() > 0 {
		ListSingleContainer(containerManagerCmd, flag.Arg(0))
		return
	} else if flagListConfigs {
		// the user asked to list all available configurations. Do that and exit
		ListConfigs(containerManagerCmd)
		return
	}

	if flag.NArg() == 0 {
		log.Fatal("Specify the name of a container as defined within the configuration file, or `-l` to list all definitions")
	}

	definitionName = flag.Arg(0)
	//.Args() is an array of the remaining parameters provided, which do not have a name
	if flag.NArg() > 1 {
		additionalArgs = flag.Args()[1:]
	}

	switch ConfigType(definitionName) {
	case CONFTYPECONTAINER:
		ManageContainer(containerManagerCmd, definitionName, additionalArgs)
	case CONFTYPECOMPOSE:
		ManageCompose(containerManagerCmd, definitionName, additionalArgs)
	default:
		log.Fatalf("Impossible to discern type of configuration for '%s'", definitionName)
	}

}
