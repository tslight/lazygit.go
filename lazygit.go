package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
)

var (
	URL    = "https://gitlab.com"
	APIURL = URL + "/api/v4"
	Token  = flag.String("t", "", "GitLab API Token")
	Path   = flag.String("p", ".", "path to clone projects to")
)

type Project struct {
	URL  string
	Path string
}

type Config struct {
	Token string
	Path  string
}

func getProjects(token string, httpUrl bool) []Project {
	client := &http.Client{}
	req, err := http.NewRequest("GET", APIURL+"/projects", nil)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
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

	var objs interface{}
	json.Unmarshal(body, &objs)

	// Ensure that we have array of objects.
	objArr, ok := objs.([]interface{})
	if !ok {
		log.Fatal("expected an array of objects")
	}

	var projects []Project

	// Handle each object as a map[string]interface{}.
	for i, obj := range objArr {
		obj, ok := obj.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(objArr[i]))
		}
		if httpUrl {
			projects = append(projects, Project{
				URL:  fmt.Sprint((obj["http_url_to_repo"])),
				Path: fmt.Sprint((obj["path_with_namespace"]))})
		} else {
			projects = append(projects, Project{
				URL:  fmt.Sprint((obj["ssh_url_to_repo"])),
				Path: fmt.Sprint((obj["path_with_namespace"]))})
		}
	}

	return projects
}

func main() {
	var token string
	var path string

	flag.Parse()
	if *Token != "" {
		token = *Token
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(home + "/.lazygit.json")
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	conf := Config{}
	if err := decoder.Decode(&conf); err != nil {
		log.Fatal("ERROR:", err)
	}

	if conf.Token != "" {
		token = conf.Token
	}

	if conf.Path != "" {
		path = conf.Path
	} else {
		path = *Path
	}

	projects := getProjects(token, false)
	for _, v := range projects {
		fmt.Printf("Cloning %s to %s/%s...\n", v.URL, path, v.Path)
	}
}
