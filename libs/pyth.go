package libs

import (
	// "fmt"
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"math/big"
	"sync"

	. "go-liquidator/global"
	"go-liquidator/libs/AggregatorState"

	"github.com/portto/solana-go-sdk/client"
	"github.com/samber/lo"
	// "github.com/portto/solana-go-sdk/common"
)

const (
	NULL_ORACLE            = "nu11111111111111111111111111111111111111111"
	SWITCHBOARD_V1_ADDRESS = "DtmE9D2CSB4L5D6A15mraeEjrGMm6auWVzgaD8hK2tZM"
	SWITCHBOARD_V2_ADDRESS = "SW1TCH7qEPTdLsDHRgPuMQjbQxKdH2aBStViMFnt64f"
)

func getTokenOracleData(wg *sync.WaitGroup, c *client.Client, config Config, oracles []OracleAsset, reserve Reserve, oracleTokens *[]OracleToken) {
	defer wg.Done()

	oracle, _ := lo.Find(oracles, func(o OracleAsset) bool {
		return o.Asset == reserve.Asset
	})

	var price float64
	if oracle.PriceAddress != "" && oracle.PriceAddress != NULL_ORACLE {
		result, _ := c.GetAccountInfo(context.TODO(), oracle.PriceAddress)
		buf := bytes.NewReader(result.Data)
		var _price uint64
		var exponent int32
		buf.Seek(20, 0)
		binary.Read(buf, binary.LittleEndian, &exponent)
		buf.Seek(208, 0)
		binary.Read(buf, binary.LittleEndian, &_price)
		price = float64(_price) * math.Pow(10, float64(exponent))
	} else {
		result, _ := c.GetAccountInfo(context.TODO(), oracle.SwitchboardFeedAddress)
		owner := result.Owner
		if owner == SWITCHBOARD_V1_ADDRESS {
			agg := AggregatorState.DecodeDelimited(result.Data[1:])
			price = agg.LastRoundResult.Result
		} else if owner == SWITCHBOARD_V2_ADDRESS {

		}
	}

	assetConfig, _ := lo.Find(config.Assets, func(asset Asset) bool {
		return asset.Symbol == oracle.Asset
	})

	Symbol := oracle.Asset
	ReserveAddress := reserve.Address
	MintAddress := assetConfig.MintAddress
	Decimals := big.NewRat(int64(math.Pow(10, float64(assetConfig.Decimals))), 1)

	exponent := int64(100000000)
	Price := big.NewRat(int64(price*float64(exponent)), exponent)

	oracleToken := OracleToken{
		Symbol:         Symbol,
		ReserveAddress: ReserveAddress,
		MintAddress:    MintAddress,
		Decimals:       Decimals,
		Price:          Price,
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
