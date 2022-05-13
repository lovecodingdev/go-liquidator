package libs

import (
	"fmt"
	"sync"
	"context"
	"math"
	"math/big"

	. "go-liquidator/global"

	"github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/common"

)

const (
	NULL_ORACLE = "nu11111111111111111111111111111111111111111"
	SWITCHBOARD_V1_ADDRESS = "DtmE9D2CSB4L5D6A15mraeEjrGMm6auWVzgaD8hK2tZM"
	SWITCHBOARD_V2_ADDRESS = "SW1TCH7qEPTdLsDHRgPuMQjbQxKdH2aBStViMFnt64f"
)

func getTokenOracleData(wg *sync.WaitGroup, c *client.Client, config Config,   oracles []OracleAsset,	reserve Reserve, rAssets *[]ReserveAsset) {
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

	if oracle.PriceAddress != "" && oracle.PriceAddress != NULL_ORACLE {
    result, _ := c.GetAccountInfo(context.TODO(), oracle.PriceAddress);
		fmt.Println(result.Lamports)
    // price = parsePriceData(result!.data).price;
  } else {
    // const pricePublicKey = new PublicKey(oracle.switchboardFeedAddress);
    // const info = await connection.getAccountInfo(pricePublicKey);
    // const owner = info?.owner.toString();
    // if (owner === SWITCHBOARD_V1_ADDRESS) {
    //   const result = AggregatorState.decodeDelimited((info?.data as Buffer)?.slice(1));
    //   price = result?.lastRoundResult?.result;
    // } else if (owner === SWITCHBOARD_V2_ADDRESS) {
    //   if (!switchboardV2) {
    //     switchboardV2 = await SwitchboardProgram.loadMainnet(connection);
    //   }
    //   const result = switchboardV2.decodeLatestAggregatorValue(info!);
    //   price = result?.toNumber();
    // } else {
    //   console.error('unrecognized switchboard owner address: ', owner);
    // }
  }

	Symbol := "BNB"
	ReserveAddress := reserve.Address
	MintAddress := assetConfig.MintAddress
	Decimals := big.NewInt(int64(math.Pow(10, float64(assetConfig.Decimals))))
	Price := big.NewInt(1234)

	rAsset := ReserveAsset {
		Symbol,
		ReserveAddress,
		MintAddress,
		Decimals,
		Price,
	}

	*rAssets = append(*rAssets, rAsset)
}

func GetTokensOracleData(c *client.Client, config Config, reserves []Reserve) []ReserveAsset {
	var rAssets []ReserveAsset

	var wg sync.WaitGroup
	oracles := config.Oracles.Assets
	for _, reserve := range reserves {
		wg.Add(1)
		go getTokenOracleData(&wg, c, config, oracles, reserve, &rAssets)
	}
	wg.Wait()

	return rAssets
}