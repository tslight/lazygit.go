package main

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/tslight/lazygit.go/pkg/common"
	"github.com/tslight/lazygit.go/pkg/github"
)

var Version = "unknown"

func main() {
	conf := common.GetConfig(".lazygithub.json")
	repos := github.GetAllRepos(conf.Token)

	var wg sync.WaitGroup
	wg.Add(len(repos))

	for k, v := range repos {
		p, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(repos[k]))
		}
		url := fmt.Sprint(p["ssh_url"])
		projectPath := filepath.Join(conf.Path, fmt.Sprint(p["name"]))
		go common.GitCloneOrPull(url, projectPath, &wg)
	}

	wg.Wait()
}
