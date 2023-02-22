package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func AbsHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(dirname, path[1:])
	}
	if strings.HasPrefix(path, "$HOME") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(dirname, path[5:])
	}
	if strings.HasPrefix(path, "%USERPROFILE%") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(dirname, path[13:])
	}
	return path
}

func GenerateConfig(configPath string) *os.File {
	var token string
	var path string
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

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configFile := filepath.Join(home, path)
	file, err = os.Open(configFile)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			file = GenerateConfig(configFile)
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

func GitCloneOrPull(url string, path string, wg *sync.WaitGroup) {
	defer wg.Done()
	var cmd *exec.Cmd
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		cmd = exec.Command("git", "-C", path, "pull")
	} else {
		cmd = exec.Command("git", "clone", url, path)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Print(stderr.String())

	if stdout.String() != "Already up to date.\n" {
		fmt.Print(stdout.String())
	}
}
