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

type stargazer struct {
	StarredAt string `json:"starred_at"`
}

var repo string

func main() {
	repo = os.Args[1]
	if !strings.Contains(repo, "/") {
		log.Fatalln("you need to pass a repo in the owner/name format")
	}
	var page = 1
	for {
		req, err := http.NewRequest(http.MethodGet, urlFor(page), nil)
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Add("Accept", "application/vnd.github.v3.star+json")
		req.Header.Add("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))
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

func urlFor(page int) string {
	return fmt.Sprintf(
		`https://api.github.com/repos/%v/stargazers?page=%v&per_page=%v`,
		repo,
		page,
		pageSize,
	)
}
