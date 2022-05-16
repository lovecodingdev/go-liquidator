package actions

import (
	// "fmt"
	// "sync"
	// "context"
	// "math"
	// "math/big"

	"go-liquidator/global"
	. "go-liquidator/models/layouts"
	// "go-liquidator/utils"

	"github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"

)

func LiquidateAndRedeem(
	c *client.Client,
  config global.Config,
  payer types.Account,
  liquidityAmount int64,
  repayTokenSymbol string,
  withdrawTokenSymbol string,
  lendingMarket global.Market,
  obligation AccountWithObligation,
) {
	//TODO: LiquidateAndRedeem
}