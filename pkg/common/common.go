package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func ConfDir() string {
	var dir string

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
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
			panic(err)
		}
	}
}

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
		if strings.Contains(err.Error(), "no such file or directory") {
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

func GitWalk(path string) {
	var gitDirName string
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				gitDir := filepath.Join(path, ".git")
				if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
					fmt.Println(path)
				}
				gitDirName = path
			}
			if strings.HasPrefix(path, gitDirName+"/") {
				return filepath.SkipDir
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func GitStatus(path string, wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("git", "-C", path, "status")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Fatal("ERROR: ", err)
	}

	fmt.Print(stderr.String())
	if !strings.HasSuffix(stdout.String(), "nothing to commit, working tree clean\n") {
		fmt.Printf("\033[1;33munstaged changes in %s:\033[0m\n", path)
		lines := strings.Split(stdout.String(), "\n")
		for _, l := range lines {
			if strings.HasPrefix(l, "\t") {
				fmt.Println(strings.Trim(l, "\t"))
			}
		}
	}
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
