package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var APIURL = "https://gitlab.com/api/v4"

type Config struct {
	Token string `json:"token"`
	Path  string `json:"path"`
}

func getGroupIds(token string, groupNames []string) []string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", APIURL+"/groups", nil)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	qp := url.Values{}
	qp.Add("per_page", "100")
	req.URL.RawQuery = qp.Encode()
	req.Header.Add("PRIVATE-TOKEN", token)
	response, err := client.Do(req)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// https://stackoverflow.com/a/22346593/11133327
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	var groups interface{}
	if err := d.Decode(&groups); err != nil {
		log.Fatal(err)
	}

	gArr, ok := groups.([]interface{})
	if !ok {
		log.Fatal("expected an array of objects")
	}
	var groupIds []string
	for k, v := range gArr {
		g, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(gArr[k]))
		}
		for _, n := range groupNames {
			if g["name"] == n || g["full_path"] == n {
				groupIds = append(groupIds, fmt.Sprint(g["id"]))
			}
		}
	}

	return groupIds
}

func getGroupProjects(token string, groupNames []string) []interface{} {
	ids := getGroupIds(token, groupNames)
	var projects []interface{}
	for _, id := range ids {
		client := &http.Client{}
		req, err := http.NewRequest("GET", APIURL+"/groups/"+id+"/projects", nil)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		req.Header.Add("PRIVATE-TOKEN", token)
		qp := url.Values{}
		qp.Add("per_page", "100")
		qp.Add("membership", "true")
		qp.Add("archived", "false")
		req.URL.RawQuery = qp.Encode()
		response, err := client.Do(req)
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		var groupProjects interface{}
		json.Unmarshal(body, &groupProjects)
		gpArr, ok := groupProjects.([]interface{})
		if !ok {
			log.Fatal("expected an array of objects")
		}
		projects = append(projects, gpArr...)
	}

	return projects
}

func getAllProjects(token string) []interface{} {
	client := &http.Client{}
	req, err := http.NewRequest("GET", APIURL+"/projects", nil)
	if err != nil {
		log.Fatal(err)
	}

	qp := url.Values{}
	qp.Add("per_page", "100")
	qp.Add("membership", "true")
	qp.Add("archived", "false")
	req.URL.RawQuery = qp.Encode()

	req.Header.Add("PRIVATE-TOKEN", token)
	response, err := client.Do(req)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var projects interface{}
	json.Unmarshal(body, &projects)

	// Ensure that we have array of objects.
	pArr, ok := projects.([]interface{})
	if !ok {
		log.Fatal("expected an array of objects")
	}

	return pArr
}

func parseHomeDirSymbols(path string) string {
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

func generateConfig(configPath string) *os.File {
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
		path = parseHomeDirSymbols(path)
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

	if err := ioutil.WriteFile(configPath, b, 0644); err != nil {
		log.Fatal("ERROR: Failed to write config to file: ", err)
	}
	fmt.Println("Wrote config to ", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal("ERROR: Failed to access config: ", err)
	}

	return file
}

func gitCloneOrPull(url string, path string, wg *sync.WaitGroup) {
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

	if stdout.String() == "Already up to date.\n" {
		fmt.Printf("%v up to date\n", path)
	} else {
		fmt.Print(stdout.String())
	}
}

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
			file = generateConfig(configFile)
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
	conf.Path = parseHomeDirSymbols(conf.Path)

	var projects []interface{}
	flag.Parse()
	groups := flag.Args()

	if len(groups) > 0 {
		projects = getGroupProjects(conf.Token, groups)
	} else {
		projects = getAllProjects(conf.Token)
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
		go gitCloneOrPull(url, projectPath, &wg)
	}

	wg.Wait()
}
