package main

import (
	//_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	CONFTYPECONTAINER string = "container"
	CONFTYPECOMPOSE   string = "compose"
	CONFTYPEUNKNOWN   string = "unknown"
)

// ConfigType returns the type of configuration, whether container, compose or unknown
func ConfigType(configName string) string {
	if viper.IsSet(configName + ".compose") {
		return CONFTYPECOMPOSE
	} else if viper.IsSet(configName + ".run") {
		return CONFTYPECONTAINER
	}
	return CONFTYPEUNKNOWN
}

/*
ExpandPath is a function that takes a file path as a string and expands it to an absolute path.
It returns the expanded path as a string, along with an error if any errors occur during the path expansion.
The function first checks whether the input path is an empty string, a relative path or an absolute path.
If the path is an absolute path or not starting with ~ or ., it is returned as is.
*/
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

/*
FileExists returns a boolean value, which is true if the file represented by 'path' exists,

	and false if the file does not exist.
*/
func FileExists(path string) bool {
	if strings.Trim(path, " \r\n") == "" {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

/* IsIn returns a boolean value, which is true if the val is found in the list, and false otherwise.
 */
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
