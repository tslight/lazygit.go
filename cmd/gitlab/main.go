package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/tslight/lazygit.go/cmd/common"
	"github.com/tslight/lazygit.go/cmd/gitlab"
)

func main() {
	var file *os.File

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configFile := home + "/.lazygit.json"
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

	var projects []interface{}
	flag.Parse()
	groups := flag.Args()

	if len(groups) > 0 {
		projects = gitlab.GetGroupProjects(conf.Token, groups)
	} else {
		projects = gitlab.getAllProjects(conf.Token)
	}

	var wg sync.WaitGroup
	wg.Add(len(projects))

	for k, v := range projects {
		p, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(projects[k]))
		}
		url := fmt.Sprint(p["ssh_url_to_repo"])
		projectPath := filepath.Join(conf.Path, fmt.Sprint(p["path_with_namespace"]))
		go common.GitCloneOrPull(url, projectPath, &wg)
	}

	wg.Wait()
}
