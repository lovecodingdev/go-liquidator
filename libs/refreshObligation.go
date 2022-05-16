package libs

import (
	// "fmt"
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

func CalculateRefreshedObligation(obligation Obligation, reserves []AccountWithReserve, tokensOracle []global.OracleToken) RefreshedObligation {
  depositedValue := big.NewInt(0)
  borrowedValue := big.NewInt(0)
  allowedBorrowValue := big.NewInt(0)
  unhealthyBorrowValue := big.NewInt(0)
  deposits := []Deposit{}
  borrows := []Borrow{}
  utilizationRatio := float32(0)

	//todo: CalculateRefreshedObligation

	return RefreshedObligation {
		depositedValue,
		borrowedValue,
		allowedBorrowValue,
		unhealthyBorrowValue,
		deposits,
		borrows,
		utilizationRatio,
	}
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