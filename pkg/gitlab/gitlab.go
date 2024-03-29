package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"reflect"
	"strings"

	"github.com/tslight/lazygit.go/pkg/common"
)

var APIURL = "https://gitlab.com/api/v4"

func AddSSHKey(token string) {
	sshPubKey := common.GetSSHPubKey()
	user, err := user.Current()
	if err != nil {
		log.Fatal(err.Error())
	}
	username := user.Username

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err.Error())
	}

	body := url.Values{}
	body.Set("title", username+"@"+hostname+" is a lazygit!")
	body.Set("key", string(sshPubKey[:]))

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		APIURL+"/user/keys",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("PRIVATE-TOKEN", token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 201 {
		resBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(string(resBody), `{"message":{"fingerprint_sha256":["has already been taken"]}}`) {
			log.Print("Already uploaded this public SSH key to GitLab.")
		} else {
			log.Print("GitLab API Response Status: ", response.Status)
			log.Print(string(resBody))
		}
	} else {
		log.Print("Successfully uploaded public SSH key to GitLab :-)")
	}
}

func getGroupIds(token string, groupNames []string) []string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", APIURL+"/groups", nil)
	if err != nil {
		log.Fatal(err)
	}
	qp := url.Values{}
	qp.Add("per_page", "100")
	req.URL.RawQuery = qp.Encode()
	req.Header.Add("PRIVATE-TOKEN", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
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

func GetGroupProjects(token string, groupNames []string) []interface{} {
	ids := getGroupIds(token, groupNames)
	var projects []interface{}
	for _, id := range ids {
		client := &http.Client{}
		req, err := http.NewRequest("GET", APIURL+"/groups/"+id+"/projects", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("PRIVATE-TOKEN", token)
		qp := url.Values{}
		qp.Add("per_page", "100")
		qp.Add("membership", "true")
		qp.Add("archived", "false")
		req.URL.RawQuery = qp.Encode()
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		var groupProjects interface{}
		if err := json.Unmarshal(body, &groupProjects); err != nil {
			log.Fatal(err)
		}
		gpArr, ok := groupProjects.([]interface{})
		if !ok {
			log.Fatal("expected an array of objects")
		}
		projects = append(projects, gpArr...)
	}

	if len(projects) < 1 {
		log.Printf("No projects found for %v groups :-(", groupNames)
	}
	return projects
}

func GetAllProjects(token string) []interface{} {
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
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var projects interface{}
	if err := json.Unmarshal(body, &projects); err != nil {
		log.Fatal(err)
	}

	// Ensure that we have array of objects.
	pArr, ok := projects.([]interface{})
	if !ok {
		log.Fatal("expected an array of objects")
	}

	return pArr
}
