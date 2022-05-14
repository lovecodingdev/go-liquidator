package layouts

import (
	"fmt"
	// "sync"
	// "context"
	// "math"
	"math/big"
  "bytes"
	"encoding/base64"
	"encoding/binary"
	// "encoding/hex"

	// . "go-liquidator/global"

	// "github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"
	"github.com/btcsuite/btcd/btcutil/base58"
)

const RESERVE_LEN = 619
var WAD = big.NewInt(1000000000000000000)

var INITIAL_COLLATERAL_RATIO = 1
var INITIAL_COLLATERAL_RATE = new(big.Int).Mul(big.NewInt(int64(INITIAL_COLLATERAL_RATIO)), WAD)

type AccountWithReserve struct {
  Pubkey string
  Account rpc.GetProgramAccountsAccount
  Info Reserve
}

type Reserve struct {
	Version uint8
  LastUpdate LastUpdate
  LendingMarket string
	Liquidity ReserveLiquidity
  Collateral ReserveCollateral
  Config ReserveConfig
	_Padding [256]byte
}

type ReserveLiquidity struct {
  MintPubkey string
  MintDecimals uint8
  SupplyPubkey string
  // @FIXME oracle option
  OracleOption uint32
  PythOraclePubkey string
  SwitchboardOraclePubkey string
  AvailableAmount uint64
  BorrowedAmountWads *big.Int
  CumulativeBorrowRateWads *big.Int
  MarketPrice *big.Int
}

type ReserveCollateral struct {
  MintPubkey string
  MintTotalSupply uint64
  SupplyPubkey string
}

type ReserveConfig struct {
  OptimalUtilizationRate uint8
  LoanToValueRatio uint8
  LiquidationBonus uint8
  LiquidationThreshold uint8
  MinBorrowRate uint8
  OptimalBorrowRate uint8
  MaxBorrowRate uint8
  Fees struct {
    BorrowFeeWad uint64
		FlashLoanFeeWad uint64
    HostFeePercentage uint8
  }
  DepositLimit uint64
  BorrowLimit uint64
  FeeReceiver string
}

func ReserveDataDecode(){
  data := "AXrL+AcAAAAAADOzHsTv+PoomuqMlUwBYy4tdkkIzlRNaGW97xEb/2ErC2K6B09yLJ1BFPLY9woAxmACM3ub+QyHNlem0gHbTIAJI+JVRPDAkseR0HH3owasEz4CGAvw+IIrp5g9mI9PyTnCKJpqQ9LOkcb1XK7DcPSsw4ou1Hf1iBMzTG0DdJ/ypKbeiI8OUOyQIEV4zU4acs6F+XCU9Ns+yj5o80ZLgQZH7YT0MT62BADKdNpBT0RVv+rKXJlBAAAAuFQPnh/M6g0AAAAAAAAAAADA24UvmNHiAgAAAAAAAAAiIksJe0Y5gR6fkskcQ3aCbPDlaPzQknKEh1oGeSZfUPDgKX7OugQA0+M5BldCMw5dVHNhPp87+Ixa65/6HtkBSmQjGHsVj5hQSwVVAB7IAIDGpH6NAwAAgFPue6gKABQAAI1J/RoHAP//////////aLxo/vvcyp0E/w/d/Zw7pj83SeKdA3iiHd+qhDp+dioAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="  
	dec, _ := base64.StdEncoding.DecodeString(data)
  fmt.Println(dec)
  buf := bytes.NewBuffer(dec)

  var _pubkey [32]byte
  var _uint128 [16]byte

  var reserve Reserve

  binary.Read(buf, binary.LittleEndian, &reserve.Version)
  binary.Read(buf, binary.LittleEndian, &reserve.LastUpdate.Slot)
  binary.Read(buf, binary.LittleEndian, &reserve.LastUpdate.Stale)

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.LendingMarket = base58.Encode(_pubkey[:])

	//Liquidity
  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.MintPubkey = base58.Encode(_pubkey[:])

	binary.Read(buf, binary.LittleEndian, &reserve.Liquidity.MintDecimals)

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.SupplyPubkey = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.PythOraclePubkey = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.SwitchboardOraclePubkey = base58.Encode(_pubkey[:])

	binary.Read(buf, binary.LittleEndian, &reserve.Liquidity.AvailableAmount)

	binary.Read(buf, binary.LittleEndian, &_uint128)
  reserve.Liquidity.BorrowedAmountWads = bigIntFromBytes(_uint128[:])

	binary.Read(buf, binary.LittleEndian, &_uint128)
  reserve.Liquidity.CumulativeBorrowRateWads = bigIntFromBytes(_uint128[:])

	binary.Read(buf, binary.LittleEndian, &_uint128)
  reserve.Liquidity.MarketPrice = bigIntFromBytes(_uint128[:])

	//Collateral
  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Collateral.MintPubkey = base58.Encode(_pubkey[:])

	binary.Read(buf, binary.LittleEndian, &reserve.Collateral.MintTotalSupply)

	binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Collateral.SupplyPubkey = base58.Encode(_pubkey[:])

	//Config
	binary.Read(buf, binary.LittleEndian, &reserve.Config.OptimalUtilizationRate)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.LoanToValueRatio)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.LiquidationBonus)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.LiquidationThreshold)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.MinBorrowRate)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.OptimalBorrowRate)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.MaxBorrowRate)

	//Config.Fees
	binary.Read(buf, binary.LittleEndian, &reserve.Config.Fees.BorrowFeeWad)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.Fees.FlashLoanFeeWad)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.Fees.HostFeePercentage)

	binary.Read(buf, binary.LittleEndian, &reserve.Config.DepositLimit)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.BorrowLimit)

	binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Config.FeeReceiver = base58.Encode(_pubkey[:])

	buf.Next(256)

  fmt.Println(reserve)
}

func ReserveParser (pubkey string, info rpc.GetProgramAccountsAccount) AccountWithReserve {
  data := info.Data.([]any)
	dec, _ := base64.StdEncoding.DecodeString(data[0].(string))

  buf := bytes.NewBuffer(dec)

  var _pubkey [32]byte
  var _uint128 [16]byte

  var reserve Reserve

  binary.Read(buf, binary.LittleEndian, &reserve.Version)
  binary.Read(buf, binary.LittleEndian, &reserve.LastUpdate.Slot)
  binary.Read(buf, binary.LittleEndian, &reserve.LastUpdate.Stale)

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.LendingMarket = base58.Encode(_pubkey[:])

	//Liquidity
  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.MintPubkey = base58.Encode(_pubkey[:])

	binary.Read(buf, binary.LittleEndian, &reserve.Liquidity.MintDecimals)

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.SupplyPubkey = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.PythOraclePubkey = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Liquidity.SwitchboardOraclePubkey = base58.Encode(_pubkey[:])

	binary.Read(buf, binary.LittleEndian, &reserve.Liquidity.AvailableAmount)

	binary.Read(buf, binary.LittleEndian, &_uint128)
  reserve.Liquidity.BorrowedAmountWads = bigIntFromBytes(_uint128[:])

	binary.Read(buf, binary.LittleEndian, &_uint128)
  reserve.Liquidity.CumulativeBorrowRateWads = bigIntFromBytes(_uint128[:])

	binary.Read(buf, binary.LittleEndian, &_uint128)
  reserve.Liquidity.MarketPrice = bigIntFromBytes(_uint128[:])

	//Collateral
  binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Collateral.MintPubkey = base58.Encode(_pubkey[:])

	binary.Read(buf, binary.LittleEndian, &reserve.Collateral.MintTotalSupply)

	binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Collateral.SupplyPubkey = base58.Encode(_pubkey[:])

	//Config
	binary.Read(buf, binary.LittleEndian, &reserve.Config.OptimalUtilizationRate)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.LoanToValueRatio)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.LiquidationBonus)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.LiquidationThreshold)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.MinBorrowRate)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.OptimalBorrowRate)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.MaxBorrowRate)

	//Config.Fees
	binary.Read(buf, binary.LittleEndian, &reserve.Config.Fees.BorrowFeeWad)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.Fees.FlashLoanFeeWad)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.Fees.HostFeePercentage)

	binary.Read(buf, binary.LittleEndian, &reserve.Config.DepositLimit)
	binary.Read(buf, binary.LittleEndian, &reserve.Config.BorrowLimit)

	binary.Read(buf, binary.LittleEndian, &_pubkey)
  reserve.Config.FeeReceiver = base58.Encode(_pubkey[:])

	buf.Next(256)

  return AccountWithReserve {
    pubkey,
    info,
    reserve,
  }
};