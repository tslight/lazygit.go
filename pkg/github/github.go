package github

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
	"strings"

	"github.com/tslight/lazygit.go/pkg/common"
)

var APIURL = "https://api.github.com"

type GitHubSSHPubKey struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

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

	body := GitHubSSHPubKey{
		Title: username + "@" + hostname + " is a lazygit!",
		Key:   string(sshPubKey[:]),
	}
	b, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	sshPubKeyReader := bytes.NewReader(b)
	client := &http.Client{}
	req, err := http.NewRequest("POST", APIURL+"/user/keys", sshPubKeyReader)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", fmt.Sprint("token ", token))
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 201 {
		resBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(string(resBody), `"message":"key is already in use"`) {
			log.Print("Already uploaded this public SSH key to GitHub")
		} else {
			log.Print("GitHub API Response Status: ", response.Status)
			log.Print(string(resBody))
		}
	} else {
		log.Print("Successfully uploaded public SSH key to GitHub :-)")
	}
}

func GetAllRepos(token string) []interface{} {
	client := &http.Client{}
	req, err := http.NewRequest("GET", APIURL+"/user/repos", nil)
	if err != nil {
		log.Fatal(err)
	}

	qp := url.Values{}
	qp.Add("per_page", "100")
	qp.Add("affiliation", "owner")
	qp.Add("archived", "false")
	req.URL.RawQuery = qp.Encode()

	req.Header.Add("Authorization", fmt.Sprint("token ", token))
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var repos interface{}
	if err := json.Unmarshal(body, &repos); err != nil {
		log.Fatal(err)
	}

	// Ensure that we have array of objects.
	rArr, ok := repos.([]interface{})
	if !ok {
		log.Fatal("expected an array of objects")
	}

	return rArr
}
