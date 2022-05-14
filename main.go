package main

import (
	// "context"
	"fmt"
	"log"
	"os"
	"encoding/json"
	"io/ioutil"

	. "go-liquidator/config"
	. "go-liquidator/libs"
	// . "go-liquidator/models/layouts"

	"github.com/joho/godotenv"
	"github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

func main() {
	// ObligationDataDecode()
	// return

  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

	config := GetConfig()
	fmt.Printf("config %s\n", config.ProgramID)

	ENV_APP := os.Getenv("APP")
	clusterUrl := ENDPOINTS[ENV_APP]
	c := client.NewClient(clusterUrl)

	keypairFile, _ := os.Open("keypair.json")
	defer keypairFile.Close()
	byteValue, _ := ioutil.ReadAll(keypairFile)

	var keypair []byte;
	json.Unmarshal([]byte(byteValue), &keypair)

	payer, _ := types.AccountFromBytes(keypair)
	fmt.Printf(" app: %s\n clusterUrl: %s\n wallet: %s\n", ENV_APP, clusterUrl, payer.PublicKey.ToBase58())

	ENV_MARKET := os.Getenv("MARKET")
	for epoch := 0; epoch<1; epoch++ {
		for _, market := range config.Markets {
			if ENV_MARKET != "" && ENV_MARKET != market.Address {
        continue;
      }

			tokensOracle := GetTokensOracleData(c, config, market.Reserves);
			fmt.Println(tokensOracle)
			fmt.Println("\n")

			allObligations := GetObligations(c, config, market.Address);
			fmt.Println(tokensOracle)
			fmt.Println("\n")

		}	
	}
}
