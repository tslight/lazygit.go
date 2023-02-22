package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

var APIURL = "https://api.github.com"

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
