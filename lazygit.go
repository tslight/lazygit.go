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
	APIURL = "https://gitlab.com/api/v4"
	Groups = flag.String("g", "", "GitLab Groups to work with")
)

type Config struct {
	Token string
	Path  string
}

func getProjects(token string) []interface{} {
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
	flag.Parse()

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

	projects := getProjects(conf.Token)
	for k, v := range projects {
		p, ok := v.(map[string]interface{})
		if !ok {
			log.Fatalf("expected type map[string]interface{}, got %s", reflect.TypeOf(projects[k]))
		}
		fmt.Printf("Cloning %s to %s/%s...\n", p["ssh_url_to_repo"], conf.Path, p["path_with_namespace"])
	}
}
