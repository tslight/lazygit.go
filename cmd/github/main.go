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
	"github.com/tslight/lazygit.go/pkg/github"
)

var Version = "unknown"
var version = flag.Bool("v", false, "print version info")
var status = flag.Bool("s", false, "only show unstaged local changes")
var config = filepath.Join(common.ConfDir(), "github.json")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s%s\n\n", os.Args[0], common.ConfigUsage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *version {
		fmt.Println(Version)
		return
	}
	conf := common.GetConfig(config)
	github.AddSSHKey(conf.Token)
	repos := github.GetAllRepos(conf.Token)

	common.AddKnownHosts("github.com")

	var wg sync.WaitGroup
	wg.Add(len(repos))

	for k, v := range repos {
		p, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(repos[k]))
		}
		url := fmt.Sprint(p["ssh_url"])
		repoPath := filepath.Join(conf.Path, fmt.Sprint(p["name"]))
		if *status {
			go common.GitStatus(repoPath, &wg)
		} else {
			go common.GitCloneOrPull(url, repoPath, &wg)
		}
	}

	wg.Wait()
}
