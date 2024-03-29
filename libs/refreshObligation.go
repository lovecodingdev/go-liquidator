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

	"github.com/samber/lo"
	// "github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"
	// "github.com/btcsuite/btcd/btcutil/base58"
)

const RISKY_OBLIGATION_THRESHOLD = 78

// This function doesn't actually refresh the obligation within the blockchain
// but does offchain calculation which mimics the actual refreshObligation instruction
// to optimize of transaction fees.
func CalculateRefreshedObligation(obligation Obligation, reserves []AccountWithReserve, tokensOracle []global.OracleToken) (RefreshedObligation, error) {
	depositedValue := new(big.Rat)
	borrowedValue := new(big.Rat)
	allowedBorrowValue := new(big.Rat)
	unhealthyBorrowValue := new(big.Rat)
	deposits := []Deposit{}
	borrows := []Borrow{}
	utilizationRatio := float64(0)

	//todo: CalculateRefreshedObligation

	for _, deposit := range obligation.Deposits {
		oracleToken, _ := lo.Find(tokensOracle, func(token global.OracleToken) bool {
			return token.ReserveAddress == deposit.DepositReserve
		})
		if oracleToken == (global.OracleToken{}) {
			return RefreshedObligation{}, fmt.Errorf("Missing token info for reserve %s, \nskipping this obligation. \nPlease restart liquidator to fetch latest configs from /v1/config", deposit.DepositReserve)
		}

		_reserve, _ := lo.Find(reserves, func(r AccountWithReserve) bool {
			return r.Pubkey == deposit.DepositReserve
		})
		reserve := _reserve.Info

		collateralExchangeRate := GetCollateralExchangeRate(reserve)
		marketValue := new(big.Rat)
		marketValue = marketValue.Mul(big.NewRat(int64(deposit.DepositedAmount), 1), WAD)
		marketValue = marketValue.Quo(marketValue, collateralExchangeRate)
		marketValue = marketValue.Mul(marketValue, oracleToken.Price)
		marketValue = marketValue.Quo(marketValue, oracleToken.Decimals)

		loanToValueRate := GetLoanToValueRate(reserve)
		liquidationThresholdRate := GetLiquidationThresholdRate(reserve)

		depositedValue = depositedValue.Add(depositedValue, marketValue)
		allowedBorrowValue = allowedBorrowValue.Add(allowedBorrowValue, new(big.Rat).Mul(marketValue, loanToValueRate))
		unhealthyBorrowValue = unhealthyBorrowValue.Add(unhealthyBorrowValue, new(big.Rat).Mul(marketValue, liquidationThresholdRate))

		deposits = append(deposits, Deposit{
			deposit.DepositReserve,
			big.NewRat(int64(deposit.DepositedAmount), 1),
			marketValue,
			oracleToken.MintAddress,
			oracleToken.Symbol,
		})
	}

	for _, borrow := range obligation.Borrows {
		oracleToken, _ := lo.Find(tokensOracle, func(token global.OracleToken) bool {
			return token.ReserveAddress == borrow.BorrowReserve
		})
		if oracleToken == (global.OracleToken{}) {
			return RefreshedObligation{}, fmt.Errorf("Missing token info for reserve %s, \nskipping this obligation. \nPlease restart liquidator to fetch latest configs from /v1/config", borrow.BorrowReserve)
		}

		_reserve, _ := lo.Find(reserves, func(r AccountWithReserve) bool {
			return r.Pubkey == borrow.BorrowReserve
		})
		reserve := _reserve.Info

		borrowAmountWadsWithInterest := getBorrrowedAmountWadsWithInterest(
			reserve.Liquidity.CumulativeBorrowRateWads,
			borrow.CumulativeBorrowRateWads,
			borrow.BorrowedAmountWads,
		)
		marketValue := new(big.Rat)
		marketValue = marketValue.Mul(borrowAmountWadsWithInterest, oracleToken.Price)
		marketValue = marketValue.Quo(marketValue, oracleToken.Decimals)

		borrowedValue = borrowedValue.Add(borrowedValue, marketValue)

		borrows = append(borrows, Borrow{
			borrow.BorrowReserve,
			borrow.BorrowedAmountWads,
			marketValue,
			oracleToken.MintAddress,
			oracleToken.Symbol,
		})
	}

	if depositedValue.Cmp(new(big.Rat)) != 0 {
		_utilizationRatio := new(big.Rat)
		_utilizationRatio = _utilizationRatio.Quo(borrowedValue, depositedValue)
		_utilizationRatio = _utilizationRatio.Mul(_utilizationRatio, big.NewRat(100, 1))
		utilizationRatio, _ = _utilizationRatio.Float64()
	}

	return RefreshedObligation{
		depositedValue,
		borrowedValue,
		allowedBorrowValue,
		unhealthyBorrowValue,
		deposits,
		borrows,
		utilizationRatio,
	}, nil
}

func getBorrrowedAmountWadsWithInterest(
	reserveCumulativeBorrowRateWads *big.Rat,
	obligationCumulativeBorrowRateWads *big.Rat,
	obligationBorrowAmountWads *big.Rat,
) *big.Rat {
	switch reserveCumulativeBorrowRateWads.Cmp(obligationCumulativeBorrowRateWads) {
	case -1:
		{
			// less than
			fmt.Printf(
				"Interest rate cannot be negative.\nreserveCumulativeBorrowRateWadsNum: %s | \nobligationCumulativeBorrowRateWadsNum: %s\n",
				reserveCumulativeBorrowRateWads.FloatString(2),
				obligationCumulativeBorrowRateWads.FloatString(2),
			)
			return obligationBorrowAmountWads
		}
	case 0:
		{
			// do nothing when equal
			return obligationBorrowAmountWads
		}
	case 1:
		{
			// greater than
			compoundInterestRate := new(big.Rat).Quo(reserveCumulativeBorrowRateWads, obligationCumulativeBorrowRateWads)
			return compoundInterestRate.Mul(obligationBorrowAmountWads, compoundInterestRate)
		}
	default:
		{
			fmt.Printf(
				"Error: getBorrrowedAmountWadsWithInterest() identified invalid comparator.\nreserveCumulativeBorrowRateWadsNum: %s |\nobligationCumulativeBorrowRateWadsNum: %s\n",
				reserveCumulativeBorrowRateWads.FloatString(2),
				obligationCumulativeBorrowRateWads.FloatString(2),
			)
			return obligationBorrowAmountWads
		}
	}
}

type RefreshedObligation struct {
	DepositedValue       *big.Rat
	BorrowedValue        *big.Rat
	AllowedBorrowValue   *big.Rat
	UnhealthyBorrowValue *big.Rat
	Deposits             []Deposit
	Borrows              []Borrow
	UtilizationRatio     float64
}

type Borrow struct {
	BorrowReserve    string
	BorrowAmountWads *big.Rat
	MarketValue      *big.Rat
	MintAddress      string
	Symbol           string
}

type Deposit struct {
	DepositReserve string
	DepositAmount  *big.Rat
	MarketValue    *big.Rat
	MintAddress    string
	Symbol         string
}
