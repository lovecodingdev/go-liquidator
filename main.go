package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	. "go-liquidator/config"
	. "go-liquidator/libs"
	"go-liquidator/libs/actions"
	"go-liquidator/libs/jupag"
	. "go-liquidator/models/layouts"
	"go-liquidator/utils"

	"github.com/joho/godotenv"
	"github.com/portto/solana-go-sdk/client"

	// "github.com/portto/solana-go-sdk/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/portto/solana-go-sdk/types"
)

func main() {
	// ObligationDataDecode()
	// ReserveDataDecode()
	// return

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := GetConfig()
	fmt.Printf("config ProgramID: %s\n", config.ProgramID)

	ENV_APP := os.Getenv("APP")
	clusterUrl := os.Getenv("RPC")
	c := client.NewClient(clusterUrl)

	keypairFile, _ := os.Open("keypair-prod.json")
	defer keypairFile.Close()
	byteValue, _ := ioutil.ReadAll(keypairFile)

	var keypair []byte
	json.Unmarshal([]byte(byteValue), &keypair)

	payer, _ := types.AccountFromBytes(keypair)
	fmt.Printf(" app: %s\n clusterUrl: %s\n wallet: %s\n", ENV_APP, clusterUrl, payer.PublicKey.ToBase58())

	// actions.TestLiquidate(c, payer)
	// return

	// jupag.Swap(
	// 	"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
	// 	"So11111111111111111111111111111111111111112",
	// 	180_000_000,
	// 	payer,
	// 	c,
	// )
	// return

	ENV_MARKET := os.Getenv("MARKET")
	ENV_LIQUIDATION_MIN, _ := strconv.Atoi(os.Getenv("LIQUIDATION_MIN"))

	for epoch := 0; ; epoch++ {
		for _, market := range config.Markets {
			if ENV_MARKET != "" && ENV_MARKET != market.Address {
				continue
			}

			tokensOracle := GetTokensOracleData(c, config, market.Reserves)
			// fmt.Println(utils.JsonFromObject(tokensOracle))
			// fmt.Println("\n")

			allObligations := GetObligations(c, config, market.Address)
			// fmt.Println(utils.JsonFromObject(allObligations[0]))
			// fmt.Println("\n")

			allReserves := GetReserves(c, config, market.Address)
			// fmt.Println(utils.JsonFromObject(allReserves))
			// fmt.Println("\n")

			for _, obligation := range allObligations {
				for !cmp.Equal(obligation, (AccountWithObligation{})) {
					refreshed, err := CalculateRefreshedObligation(obligation.Info, allReserves, tokensOracle)
					// fmt.Println(utils.JsonFromObject(refreshed), err)
					if err != nil {
						break
					}

					borrowedValue := new(big.Rat).Quo(refreshed.BorrowedValue, WAD)
					_borrowedValue, _ := borrowedValue.Float64()
					unhealthyBorrowValue := new(big.Rat).Quo(refreshed.UnhealthyBorrowValue, WAD)
					_unhealthyBorrowValue, _ := unhealthyBorrowValue.Float64()
					if ENV_LIQUIDATION_MIN > 0 && (_unhealthyBorrowValue < float64(ENV_LIQUIDATION_MIN) || _borrowedValue < float64(ENV_LIQUIDATION_MIN)) {
						fmt.Printf(
							"Obligation %s is worth less than LIQUIDATION_MIN\n",
							obligation.Pubkey,
						)
						break
					}

					// Do nothing if obligation is healthy
					_cmp := refreshed.BorrowedValue.Cmp(refreshed.UnhealthyBorrowValue)
					if _cmp == -1 || _cmp == 0 {
						break
					}

					// select repay token that has the highest market value
					var selectedBorrow Borrow
					for _, borrow := range refreshed.Borrows {
						if selectedBorrow == (Borrow{}) {
							selectedBorrow = borrow
							continue
						}
						if borrow.MarketValue.Cmp(selectedBorrow.MarketValue) == 1 {
							selectedBorrow = borrow
						}
					}

					// select the withdrawal collateral token with the highest market value
					var selectedDeposit Deposit
					for _, deposit := range refreshed.Deposits {
						if selectedDeposit == (Deposit{}) {
							selectedDeposit = deposit
							continue
						}
						if deposit.MarketValue.Cmp(selectedDeposit.MarketValue) == 1 {
							selectedDeposit = deposit
						}
					}

					if selectedBorrow == (Borrow{}) || selectedDeposit == (Deposit{}) {
						// skip toxic obligations caused by toxic oracle data
						break
					}

					fmt.Printf(
						"Obligation %s is underwater\nborrowedValue: %s\nunhealthyBorrowValue: %s\nmarket address: %s\n",
						obligation.Pubkey,
						refreshed.BorrowedValue.FloatString(2),
						refreshed.UnhealthyBorrowValue.FloatString(2),
						market.Address,
					)

					fmt.Printf("Swaping all SOLs to %s", selectedBorrow.Symbol)
					jupag.SwapAllSolTo(selectedBorrow.MintAddress, payer, c)

					walletTokenData, err := GetWalletTokenData(c, config, payer, selectedBorrow.MintAddress, selectedBorrow.Symbol)
					fmt.Println(utils.JsonFromObject(walletTokenData), err)
					if walletTokenData.BalanceBase == 0 {
						fmt.Printf(
							"insufficient %s to liquidate obligation %s in market: %s \n\n",
							selectedBorrow.Symbol,
							obligation.Pubkey,
							market.Address,
						)
						break
					} else if walletTokenData.BalanceBase < 0 {
						fmt.Printf(
							"failed to get wallet balance for %s to liquidate obligation %s in market: %s.\nPotentially network error or token account does not exist in wallet\n\n",
							selectedBorrow.Symbol,
							obligation.Pubkey,
							market.Address,
						)
						break
					}

					// Set super high liquidation amount which acts as u64::MAX as program will only liquidate max
					// 50% val of all borrowed assets.
					err = actions.LiquidateAndRedeem(
						c,
						config,
						payer,
						uint64(walletTokenData.BalanceBase),
						selectedBorrow.Symbol,
						selectedDeposit.Symbol,
						market,
						obligation,
					)
					if err != nil {
						// fmt.Println(err)
						fmt.Printf("Swaping all %s(s) to SOL", selectedBorrow.Symbol)
						jupag.SwapSolFrom(selectedBorrow.MintAddress, payer, c)
						fmt.Printf("\n")
						break
					}

					fmt.Printf("Swaping all %s(s) to SOL", selectedDeposit.Symbol)
					jupag.SwapSolFrom(selectedDeposit.MintAddress, payer, c)

					postLiquidationObligation, _ := c.GetAccountInfo(context.TODO(), obligation.Pubkey)
					obligation = ObligationParser(obligation.Pubkey, postLiquidationObligation)
					fmt.Printf("\n")
				}
			}

			THROTTLE := os.Getenv("THROTTLE")
			if THROTTLE != "" {
				_throttle, _ := strconv.Atoi(THROTTLE)
				time.Sleep(time.Duration(_throttle) * time.Millisecond)
			}
		}
	}
}
