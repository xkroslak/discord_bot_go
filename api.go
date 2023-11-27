package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ChuckNorrisJoke struct {
	Categories []string `json:"categories"`
	CreatedAt  string   `json:"created_at"`
	IconURL    string   `json:"icon_url"`
	ID         string   `json:"id"`
	UpdatedAt  string   `json:"updated_at"`
	URL        string   `json:"url"`
	Value      string   `json:"value"`
}

type DriversJson struct {
	MRData struct {
		DriverTable struct {
			Drivers []struct {
				Code            string `json:"code"`
				DateOfBirth     string `json:"dateOfBirth"`
				DriverID        string `json:"driverId"`
				FamilyName      string `json:"familyName"`
				GivenName       string `json:"givenName"`
				Nationality     string `json:"nationality"`
				PermanentNumber string `json:"permanentNumber"`
				URL             string `json:"url"`
			} `json:"Drivers"`
			Season string `json:"season"`
		} `json:"DriverTable"`
		Limit  string `json:"limit"`
		Offset string `json:"offset"`
		Series string `json:"series"`
		Total  string `json:"total"`
		URL    string `json:"url"`
		Xmlns  string `json:"xmlns"`
	} `json:"MRData"`
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

func getDrivers(year string) (driversJson DriversJson) {
	url := fmt.Sprintf("http://ergast.com/api/f1/%s/drivers.json", year)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	var data DriversJson
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	return data
}
