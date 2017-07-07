package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const pageSize = 100

var (
	repo  string
	token string
)

type stargazer struct {
	StarredAt string `json:"starred_at"`
}

func main() {
	token = os.Getenv("GITHUB_TOKEN")
	repo = os.Args[1]
	if !strings.Contains(repo, "/") {
		log.Fatalln("you need to pass a repo in the owner/name format")
	}
	var page = 1
	for {
		url := fmt.Sprintf(
			"https://api.github.com/repos/%s/stargazers?page=%d&per_page=%d",
			repo, page, pageSize,
		)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Add("Accept", "application/vnd.github.v3.star+json")
		if token != "" {
			req.Header.Add("Authorization", "token "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		var stargazers []stargazer
		if err := json.NewDecoder(resp.Body).Decode(&stargazers); err != nil {
			log.Fatalln(err)
		}
		if len(stargazers) == 0 {
			return
		}
		for i, star := range stargazers {
			fmt.Printf("%v\t%v\n", i+((page-1)*pageSize), star.StarredAt)
		}
		page++
	}
}
