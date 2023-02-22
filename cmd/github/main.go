package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/tslight/lazygit.go/pkg/common"
	"github.com/tslight/lazygit.go/pkg/github"
)

var Version = "unknown"

func main() {
	var file *os.File

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configFile := home + "/.lazygithub.json"
	file, err = os.Open(configFile)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			file = common.GenerateConfig(configFile)
		} else {
			log.Fatal("ERROR: ", err)
		}
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	conf := common.Config{}
	if err := decoder.Decode(&conf); err != nil {
		log.Fatal("ERROR:", err)
	}
	conf.Path = common.AbsHomeDir(conf.Path)

	repos := github.GetAllRepos(conf.Token)

	var wg sync.WaitGroup
	wg.Add(len(repos))

	for k, v := range repos {
		p, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(repos[k]))
		}
		url := fmt.Sprint(p["ssh_url"])
		projectPath := filepath.Join(conf.Path, fmt.Sprint(p["full_name"]))
		go common.GitCloneOrPull(url, projectPath, &wg)
	}

	wg.Wait()
}
