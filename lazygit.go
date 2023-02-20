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
	"reflect"
)

var APIURL = "https://gitlab.com/api/v4"

type Config struct {
	Token string
	Path  string
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

	var projects interface{}
	json.Unmarshal(body, &projects)

	// Ensure that we have array of objects.
	pArr, ok := projects.([]interface{})
	if !ok {
		log.Fatal("expected an array of objects")
	}

	return pArr
}

func main() {
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

	var projects []interface{}
	flag.Parse()
	groups := flag.Args()

	if len(groups) > 0 {
		projects = getGroupProjects(conf.Token, groups)
	} else {
		projects = getAllProjects(conf.Token)
	}

	for k, v := range projects {
		p, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(projects[k]))
		}
		fmt.Printf("Cloning %s to %s/%s...\n", p["ssh_url_to_repo"], conf.Path, p["path_with_namespace"])
	}
}
