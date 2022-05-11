package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"encoding/json"
)

var (
	OBLIGATION_LEN			= 1300
	RESERVE_LEN					= 619
	LENDING_MARKET_LEN	= 290
	ENDPOINTS						= map[string]string{
		"production": "https://solana-api.projectserum.com",
		"devnet": "https://api.devnet.solana.com",
	}
)

var eligibleApps = []string{"production", "devnet"}

func getConfig() Config {
	// attemptCount := 0
  // backoffFactor := 1
  // maxAttempt := 10

	envApp := os.Getenv("APP")

	for {
		response, err := http.Get("https://api.solend.fi/v1/config?deployment="+envApp)	
		if err != nil {
			fmt.Print(err.Error())
			continue
		}

		responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
			log.Fatal(err)
			continue
    }

		var responseObject Config
		json.Unmarshal(responseData, &responseObject)

		return responseObject
	}

}