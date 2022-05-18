package libs

import (
	"fmt"
	// "sync"
	// "context"
	// "math"
	"math/big"
  // "bytes"
	// "encoding/base64"
	// "encoding/binary"
	// "encoding/hex"

	"go-liquidator/global"
	. "go-liquidator/models/layouts"
	. "go-liquidator/utils"

	// "github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"
	// "github.com/btcsuite/btcd/btcutil/base58"

)

func CalculateRefreshedObligation(obligation Obligation, reserves []AccountWithReserve, tokensOracle []global.OracleToken) (RefreshedObligation, error) {
  depositedValue := big.NewFloat(0)
  borrowedValue := big.NewFloat(0)
  allowedBorrowValue := big.NewFloat(0)
  unhealthyBorrowValue := big.NewFloat(0)
  deposits := []Deposit{}
  borrows := []Borrow{}
  utilizationRatio := float32(0)

	//todo: CalculateRefreshedObligation

	for _, deposit := range obligation.Deposits {
		var oracleToken global.OracleToken
		for _, token := range tokensOracle {
			if(token.ReserveAddress == deposit.DepositReserve){
				oracleToken = token
				break
			}
		}
		if oracleToken == (global.OracleToken{}) {
			return RefreshedObligation{}, fmt.Errorf(
				`Missing token info for reserve %s, 
				skipping this obligation. Please restart liquidator to fetch latest configs from /v1/config`,
				deposit.DepositReserve,
			)
		}

		var reserve Reserve
		for _, r := range reserves {
			if(r.Pubkey == deposit.DepositReserve){
				reserve = r.Info
				break
			}
		}
		fmt.Println("refresh", JsonFromObject(reserve))

		// collateralExchangeRate := GetCollateralExchangeRate(reserve)
		// marketValue := big.NewFloat(0)
		// marketValue = marketValue.Mul(big.NewFloat(int64(deposit.DepositedAmount)), WAD)
		// marketValue = marketValue.Div(marketValue, collateralExchangeRate)
		// marketValue = marketValue.Mul(marketValue, oracleToken.Price)
		// marketValue = marketValue.Div(marketValue, oracleToken.Decimals)

	}

	return RefreshedObligation {
		depositedValue,
		borrowedValue,
		allowedBorrowValue,
		unhealthyBorrowValue,
		deposits,
		borrows,
		utilizationRatio,
	}, nil
}

type RefreshedObligation struct {
	DepositedValue *big.Float
	BorrowedValue *big.Float
	AllowedBorrowValue *big.Float
	UnhealthyBorrowValue *big.Float
	Deposits []Deposit
	Borrows []Borrow
	UtilizationRatio float32
}

type Borrow struct {
  BorrowReserve string
  BorrowAmountWads *big.Float
  MarketValue *big.Float
  MintAddress string
  Symbol string
};

type Deposit struct {
  DepositReserve string
  DepositAmount *big.Float
  MarketValue *big.Float
  Symbol string
};