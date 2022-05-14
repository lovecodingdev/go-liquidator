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

const OBLIGATION_LEN = 1300;

type AccountWithObligation struct {
  Pubkey string
  Account rpc.GetProgramAccountsAccount
  Info Obligation
}

type Obligation struct {
  Version uint8
  LastUpdate LastUpdate
  LendingMarket string
  Owner string
  DepositedValue *big.Int // decimals
  BorrowedValue *big.Int // decimals
  AllowedBorrowValue *big.Int // decimals
  UnhealthyBorrowValue *big.Int // decimals
  Deposits []ObligationCollateral
  Borrows []ObligationLiquidity
}

type ObligationCollateral struct {
  DepositReserve string //32
  DepositedAmount uint64 //8
  MarketValue *big.Int // decimals
  _Padding [32]byte
}

type ObligationLiquidity struct {
  BorrowReserve string
  CumulativeBorrowRateWads *big.Int // decimals
  BorrowedAmountWads *big.Int // decimals
  MarketValue *big.Int // decimals
  _Padding [32]byte
}

type ProtoObligation struct {
  Version uint8
  LastUpdate LastUpdate
  LendingMarket string
  Owner string
  DepositedValue *big.Int // decimals
  BorrowedValue *big.Int // decimals
  AllowedBorrowValue *big.Int // decimals
  UnhealthyBorrowValue *big.Int // decimals
  _Padding [64]byte
  DepositsLen uint8
  BorrowsLen uint8
  DataFlat [1096]byte
}

func reverse(arr []byte) []byte{
	for i, j := 0, len(arr)-1; i<j; i, j = i+1, j-1 {
		 arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func bigIntFromBytes(bs []byte) *big.Int{
  return new(big.Int).SetBytes(reverse(bs))
}

func ObligationDataDecode(){
  data := "AX7XwAUAAAAAATOzHsTv+PoomuqMlUwBYy4tdkkIzlRNaGW97xEb/2Er1Wo+pWIdz6eZoSN1E1kgPfM9W+XufFUkDFv96iIJiMCteVfNQSq9KxMAAAAAAAAAxSOzwt5RXgAAAAAAAAAAAEGbAVqx381gDgAAAAAAAABXYawKm+4wVg8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEBbcvwdUngHTQhP3KK2EcTtbcXhy+sOBPKuWOvMSyN282wPw8AAAAAAK15V81BKr0rEwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACNlSqXrkQU25XOnSMYzHuJNOeei2iMWq3lwa8TdxeqjvRhvRm9D5w0AAAAAAAAAAIanUADo/gpvAAAAAAAAAADFI7PC3lFeAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="  
	dec, _ := base64.StdEncoding.DecodeString(data)
  fmt.Println(dec)
  buf := bytes.NewBuffer(dec)

  var _pubkey [32]byte
  var _uint128 [16]byte

  var po ProtoObligation

  binary.Read(buf, binary.LittleEndian, &po.Version)
  binary.Read(buf, binary.LittleEndian, &po.LastUpdate.Slot)
  binary.Read(buf, binary.LittleEndian, &po.LastUpdate.Stale)

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  po.LendingMarket = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  po.Owner = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.DepositedValue = bigIntFromBytes(_uint128[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.BorrowedValue = bigIntFromBytes(_uint128[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.AllowedBorrowValue = bigIntFromBytes(_uint128[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.UnhealthyBorrowValue = bigIntFromBytes(_uint128[:])

  buf.Next(64)

  binary.Read(buf, binary.LittleEndian, &po.DepositsLen)
  binary.Read(buf, binary.LittleEndian, &po.BorrowsLen)
  binary.Read(buf, binary.LittleEndian, &po.DataFlat)

  flatBuf := bytes.NewBuffer(po.DataFlat[:])

  var deposits []ObligationCollateral
	for d := 0; d < int(po.DepositsLen) ; d++ {
    var oc ObligationCollateral

    binary.Read(flatBuf, binary.LittleEndian, &_pubkey)
    oc.DepositReserve = base58.Encode(_pubkey[:])

    binary.Read(flatBuf, binary.LittleEndian, &oc.DepositedAmount)

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    oc.MarketValue = bigIntFromBytes(_uint128[:])
  
    flatBuf.Next(32)

    deposits = append(deposits, oc)
	}

  var borrows []ObligationLiquidity
	for d := 0; d < int(po.DepositsLen) ; d++ {
    var ol ObligationLiquidity

    binary.Read(flatBuf, binary.LittleEndian, &_pubkey)
    ol.BorrowReserve = base58.Encode(_pubkey[:])

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    ol.CumulativeBorrowRateWads = bigIntFromBytes(_uint128[:])

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    ol.BorrowedAmountWads = bigIntFromBytes(_uint128[:])

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    ol.MarketValue = bigIntFromBytes(_uint128[:])
  
    flatBuf.Next(32)

    borrows = append(borrows, ol)
	}

  obligation := Obligation {
    po.Version,
    po.LastUpdate,
    po.LendingMarket,
    po.Owner,
    po.DepositedValue,
    po.BorrowedValue,
    po.AllowedBorrowValue,
    po.UnhealthyBorrowValue,
    deposits,
    borrows,
  }

  fmt.Println(obligation)

}

func ObligationParser (pubkey string, info rpc.GetProgramAccountsAccount) AccountWithObligation {
  data := info.Data.([]any)
	dec, _ := base64.StdEncoding.DecodeString(data[0].(string))

  buf := bytes.NewBuffer(dec)

  var _pubkey [32]byte
  var _uint128 [16]byte

  var po ProtoObligation

  binary.Read(buf, binary.LittleEndian, &po.Version)
  binary.Read(buf, binary.LittleEndian, &po.LastUpdate.Slot)
  binary.Read(buf, binary.LittleEndian, &po.LastUpdate.Stale)

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  po.LendingMarket = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_pubkey)
  po.Owner = base58.Encode(_pubkey[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.DepositedValue = bigIntFromBytes(_uint128[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.BorrowedValue = bigIntFromBytes(_uint128[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.AllowedBorrowValue = bigIntFromBytes(_uint128[:])

  binary.Read(buf, binary.LittleEndian, &_uint128)
  po.UnhealthyBorrowValue = bigIntFromBytes(_uint128[:])

  buf.Next(64)

  binary.Read(buf, binary.LittleEndian, &po.DepositsLen)
  binary.Read(buf, binary.LittleEndian, &po.BorrowsLen)
  binary.Read(buf, binary.LittleEndian, &po.DataFlat)

  flatBuf := bytes.NewBuffer(po.DataFlat[:])

  var deposits []ObligationCollateral
	for d := 0; d < int(po.DepositsLen) ; d++ {
    var oc ObligationCollateral

    binary.Read(flatBuf, binary.LittleEndian, &_pubkey)
    oc.DepositReserve = base58.Encode(_pubkey[:])

    binary.Read(flatBuf, binary.LittleEndian, &oc.DepositedAmount)

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    oc.MarketValue = bigIntFromBytes(_uint128[:])
  
    flatBuf.Next(32)

    deposits = append(deposits, oc)
	}

  var borrows []ObligationLiquidity
	for d := 0; d < int(po.DepositsLen) ; d++ {
    var ol ObligationLiquidity

    binary.Read(flatBuf, binary.LittleEndian, &_pubkey)
    ol.BorrowReserve = base58.Encode(_pubkey[:])

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    ol.CumulativeBorrowRateWads = bigIntFromBytes(_uint128[:])

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    ol.BorrowedAmountWads = bigIntFromBytes(_uint128[:])

    binary.Read(flatBuf, binary.LittleEndian, &_uint128)
    ol.MarketValue = bigIntFromBytes(_uint128[:])
  
    flatBuf.Next(32)

    borrows = append(borrows, ol)
	}

  obligation := Obligation {
    po.Version,
    po.LastUpdate,
    po.LendingMarket,
    po.Owner,
    po.DepositedValue,
    po.BorrowedValue,
    po.AllowedBorrowValue,
    po.UnhealthyBorrowValue,
    deposits,
    borrows,
  }

  return AccountWithObligation {
    pubkey,
    info,
    obligation,
  }
};
