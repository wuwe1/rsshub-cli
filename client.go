package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mmcdole/gofeed"
)

type RoutesResult struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    map[string]struct {
		Routes []string `json:"routes"`
	} `json:"data"`
}

func fetchRoutes() RoutesResult {
	res, err := http.Get("https://rsshub.app/api/routes")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result RoutesResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}

	if result.Status != 0 {
		log.Fatal(result.Message)
	}
	return result
}

func fetchFeed(url string) gofeed.Feed {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		log.Fatal(err)
	}
	return *feed
}
