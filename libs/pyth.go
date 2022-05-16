package libs

import (
	// "fmt"
	"sync"
	"context"
	"math"
	"math/big"
  "bytes"
	"encoding/binary"

	. "go-liquidator/global"
	"go-liquidator/libs/AggregatorState"

	"github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/common"

)

const (
	NULL_ORACLE = "nu11111111111111111111111111111111111111111"
	SWITCHBOARD_V1_ADDRESS = "DtmE9D2CSB4L5D6A15mraeEjrGMm6auWVzgaD8hK2tZM"
	SWITCHBOARD_V2_ADDRESS = "SW1TCH7qEPTdLsDHRgPuMQjbQxKdH2aBStViMFnt64f"
)

func getTokenOracleData(wg *sync.WaitGroup, c *client.Client, config Config,   oracles []OracleAsset,	reserve Reserve, oracleTokens *[]OracleToken) {
	defer wg.Done()
	
	var oracle OracleAsset
	for _, o := range oracles {
		if o.Asset == reserve.Asset {
			oracle = o
		}
	}

	var assetConfig Asset
	for _, asset := range config.Assets {
		if asset.Symbol == oracle.Asset {
			assetConfig = asset
		}
	}

	var price float64
	if oracle.PriceAddress != "" && oracle.PriceAddress != NULL_ORACLE {
    result, _ := c.GetAccountInfo(context.TODO(), oracle.PriceAddress);
		buf := bytes.NewReader(result.Data)
		var _price uint64
		var exponent int32
		buf.Seek(20, 0)
		binary.Read(buf, binary.LittleEndian, &exponent)
		buf.Seek(208, 0)
		binary.Read(buf, binary.LittleEndian, &_price)
		price = float64(_price) * math.Pow(10, float64(exponent))
  } else {
    result, _ := c.GetAccountInfo(context.TODO(), oracle.SwitchboardFeedAddress);
		owner := result.Owner
		if owner == SWITCHBOARD_V1_ADDRESS {
			agg := AggregatorState.DecodeDelimited(result.Data[1:])
			price = agg.LastRoundResult.Result
		} else if owner == SWITCHBOARD_V2_ADDRESS {

		}
  }

	Symbol := oracle.Asset
	ReserveAddress := reserve.Address
	MintAddress := assetConfig.MintAddress
	Decimals := big.NewInt(int64(math.Pow(10, float64(assetConfig.Decimals))))
	Price := big.NewFloat(price)

	oracleToken := OracleToken {
		Symbol,
		ReserveAddress,
		MintAddress,
		Decimals,
		Price,
	}

	*oracleTokens = append(*oracleTokens, oracleToken)
}

func GetTokensOracleData(c *client.Client, config Config, reserves []Reserve) []OracleToken {
	var oracleTokens []OracleToken

	var wg sync.WaitGroup
	oracles := config.Oracles.Assets
	for _, reserve := range reserves {
		wg.Add(1)
		go getTokenOracleData(&wg, c, config, oracles, reserve, &oracleTokens)
	}
	wg.Wait()

	return oracleTokens
}