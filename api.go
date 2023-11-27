package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type ChuckNorrisJoke struct {
	Categories []string    `json:"categories"`
	CreatedAt  string   `json:"created_at"`
	IconURL    string   `json:"icon_url"`
	ID         string   `json:"id"`
	UpdatedAt  string   `json:"updated_at"`
	URL        string   `json:"url"`
	Value      string   `json:"value"`
}

func getJoke() (joke string) {
	resp, err := http.Get("https://api.chucknorris.io/jokes/random")
	if err != nil {
		log.Println("Error making the request:", err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var chuckNorrisJoke ChuckNorrisJoke
	if err := json.Unmarshal(body, &chuckNorrisJoke); err != nil {
		log.Println("Can not unmarshal Json")
	}
	
	return chuckNorrisJoke.Value
}

