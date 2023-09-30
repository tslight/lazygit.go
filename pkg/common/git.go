package common

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

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
		log.Printf("%s FAILED!: %s", cmd.String(), err.Error())
	}

	fmt.Print(stderr.String())

	if stdout.String() != "Already up to date.\n" {
		fmt.Print(stdout.String())
	}
}

func AddKnownHosts(host string) {
	var cmd *exec.Cmd
	cmd = exec.Command("ssh-keyscan", "-H", host, ">>", "~/.ssh/known_hosts")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("%s FAILED!: %s", cmd.String(), err.Error())
	}

	// fmt.Print(stdout.String())
	// fmt.Print(stderr.String())
}
