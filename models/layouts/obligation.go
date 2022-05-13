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
	"encoding/hex"

	// . "go-liquidator/global"

	// "github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"

)

const OBLIGATION_LEN = 1300;

type Obligation struct {
  Version uint8
  LastUpdate LastUpdate
  LendingMarket string
  Owner string
  // @FIXME: check usages
  Deposits []ObligationCollateral
  // @FIXME: check usages
  borrows []ObligationLiquidity
  DepositedValue big.Int // decimals
  BorrowedValue big.Int // decimals
  AllowedBorrowValue big.Int // decimals
  UnhealthyBorrowValue big.Int // decimals
}

type ObligationCollateral struct {
  DepositReserve string
  DepositedAmount big.Int
  MarketValue big.Int // decimals
}

type ObligationLiquidity struct {
  BorrowReserve string
  CumulativeBorrowRateWads big.Int // decimals
  BorrowedAmountWads big.Int // decimals
  MarketValue big.Int // decimals
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
  po.LendingMarket = hex.EncodeToString(_pubkey[:])
  binary.Read(buf, binary.LittleEndian, &_pubkey)
  po.Owner = hex.EncodeToString(_pubkey[:])
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

  fmt.Println(po, fmt.Sprintf("%x", po.DepositedValue) )
}

func ObligationParser (pubkey string, info rpc.GetProgramAccountsAccount) {
  data := info.Data.([]any)
	dec, _ := base64.StdEncoding.DecodeString(data[0].(string))
  fmt.Println(data[0].(string))
  fmt.Println(dec)
  buf := bytes.NewBuffer(dec)

  var _pubkey [32]byte
  var po ProtoObligation
  binary.Read(buf, binary.LittleEndian, &po.Version)
  binary.Read(buf, binary.LittleEndian, &po.LastUpdate.Slot)
  binary.Read(buf, binary.LittleEndian, &po.LastUpdate.Stale)
  binary.Read(buf, binary.LittleEndian, _pubkey)
  po.LendingMarket = string(_pubkey[:])
  fmt.Println(_pubkey)

  // const buffer = Buffer.from(info.data);
  // const {
  //   version,
  //   lastUpdate,
  //   lendingMarket,
  //   owner,
  //   depositedValue,
  //   borrowedValue,
  //   allowedBorrowValue,
  //   unhealthyBorrowValue,
  //   depositsLen,
  //   borrowsLen,
  //   dataFlat,
  // } = ObligationLayout.decode(buffer) as ProtoObligation;

  // if (lastUpdate.slot.isZero()) {
  //   return null;
  // }

  // const depositsBuffer = dataFlat.slice(
  //   0,
  //   depositsLen * ObligationCollateralLayout.span,
  // );
  // const deposits = BufferLayout.seq(
  //   ObligationCollateralLayout,
  //   depositsLen,
  // ).decode(depositsBuffer) as ObligationCollateral[];

  // const borrowsBuffer = dataFlat.slice(
  //   depositsBuffer.length,
  //   depositsLen * ObligationCollateralLayout.span
  //     + borrowsLen * ObligationLiquidityLayout.span,
  // );
  // const borrows = BufferLayout.seq(
  //   ObligationLiquidityLayout,
  //   borrowsLen,
  // ).decode(borrowsBuffer) as ObligationLiquidity[];

  // const obligation = {
  //   version,
  //   lastUpdate,
  //   lendingMarket,
  //   owner,
  //   depositedValue,
  //   borrowedValue,
  //   allowedBorrowValue,
  //   unhealthyBorrowValue,
  //   deposits,
  //   borrows,
  // } as Obligation;

  // const details = {
  //   pubkey,
  //   account: {
  //     ...info,
  //   },
  //   info: obligation,
  // };

  // return details;
};
