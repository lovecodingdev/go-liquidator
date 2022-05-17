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

	// "github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"
	// "github.com/btcsuite/btcd/btcutil/base58"

)

func CalculateRefreshedObligation(obligation Obligation, reserves []AccountWithReserve, tokensOracle []global.OracleToken) (RefreshedObligation, error) {
  depositedValue := big.NewInt(0)
  borrowedValue := big.NewInt(0)
  allowedBorrowValue := big.NewInt(0)
  unhealthyBorrowValue := big.NewInt(0)
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

		collateralExchangeRate := GetCollateralExchangeRate(reserve)
		marketValue := big.NewInt(0)
		marketValue = marketValue.Mul(big.NewInt(int64(deposit.DepositedAmount)), WAD)
		marketValue = marketValue.Div(marketValue, collateralExchangeRate)
		marketValue = marketValue.Mul(marketValue, oracleToken.Price)
		marketValue = marketValue.Div(marketValue, oracleToken.Decimals)

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
	DepositedValue *big.Int
	BorrowedValue *big.Int
	AllowedBorrowValue *big.Int
	UnhealthyBorrowValue *big.Int
	Deposits []Deposit
	Borrows []Borrow
	UtilizationRatio float32
}

type Borrow struct {
  BorrowReserve string
  BorrowAmountWads *big.Int
  MarketValue *big.Int
  MintAddress string
  Symbol string
};

type Deposit struct {
  DepositReserve string
  DepositAmount *big.Int
  MarketValue *big.Int
  Symbol string
};