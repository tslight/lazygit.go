package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var ConfigUsage string = `

With no arguments will clone or pull all projects that can be accessed with
your API token to a specified directory.

The API token & directory belong in a JSON configuration file, which will live
in one of the following locations depending on your OS:

macOS:     $HOME/Library/Application Support/lazygit
Linux/BSD: $XDG_CONFIG_HOME/lazygit (usually $HOME/.config)
Windows:   %APPDATA%\lazygit (usually C:\Users\%USER%\AppData\Roaming)
Fallback:  $HOME/.config

If a JSON configuration file doesn't exist you will be prompted to enter an API
token and a directory. Those choices will be saved to a JSON file the
aforementioned directory. `

type Config struct {
	Token string `json:"token"`
	Path  string `json:"path"`
}

func ConfDir() string {
	var dir string

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd":
		dir = os.Getenv("XDG_CONFIG_HOME")
	case "windows":
		dir = os.Getenv("APPDATA")
	case "darwin":
		dir = filepath.Join(home, "/Library/Application Support")
	}

	if dir == "" {
		dir = filepath.Join(home, ".config")
	}

	return filepath.Join(dir, "lazygit")
}

func mkConfDir(path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func AbsHomeDir(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if strings.HasPrefix(path, "~") {
		return filepath.Join(home, path[1:])
	}
	if strings.HasPrefix(path, "$HOME") {
		return filepath.Join(home, path[5:])
	}
	if strings.HasPrefix(path, "%USERPROFILE%") {
		return filepath.Join(home, path[13:])
	}
	return path
}

func GenerateConfig(configPath string) *os.File {
	var token string
	var path string
	mkConfDir(ConfDir())

	fmt.Printf("Creating configuration json file at %s\n", configPath)

	for {
		fmt.Print("GitLab API token: ")
		fmt.Scanln(&token)
		if token != "" {
			break
		}
	}

	for {
		fmt.Print("Directory to clone to: ")
		fmt.Scanln(&path)
		path = AbsHomeDir(path)
		if stat, err := os.Stat(path); err == nil && stat.IsDir() {
			break
		}
		fmt.Printf("%s doesn't exist or isn't a directory!\n", path)
	}

	config := Config{token, path}
	b, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		log.Fatal("ERROR: Failed to marshal config to json: ", err)
	}

	if err := os.WriteFile(configPath, b, 0600); err != nil {
		log.Fatal("ERROR: Failed to write config to file: ", err)
	}
	fmt.Println("Wrote config to ", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal("ERROR: Failed to access config: ", err)
	}

	return file
}

func GetConfig(path string) Config {
	var file *os.File

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			file = GenerateConfig(path)
		} else {
			log.Fatal("ERROR: ", err)
		}
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	conf := Config{}
	if err := decoder.Decode(&conf); err != nil {
		log.Fatal("ERROR:", err)
	}
	conf.Path = AbsHomeDir(conf.Path)

	return conf
}
