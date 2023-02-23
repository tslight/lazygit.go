package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/tslight/lazygit.go/pkg/common"
	"github.com/tslight/lazygit.go/pkg/gitlab"
)

var Version = "unknown"
var version = flag.Bool("v", false, "print version info")
var config = filepath.Join(common.ConfDir(), "gitlab.json")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [GROUP...] %s

Optional [GROUP...] arguments will only clone or pull the projects found in
those groups.

`, os.Args[0], common.ConfigUsage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *version {
		fmt.Println(Version)
		return
	}

	conf := common.GetConfig(config)
	var projects []interface{}
	groups := flag.Args()

	if len(groups) > 0 {
		projects = gitlab.GetGroupProjects(conf.Token, groups)
	} else {
		projects = gitlab.GetAllProjects(conf.Token)
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
