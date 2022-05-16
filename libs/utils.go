package libs

import (
	"fmt"
	// "sync"
	"context"
	// "math"
	// "math/big"

	"go-liquidator/global"
	. "go-liquidator/models/layouts"
	"go-liquidator/utils"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"

)

func GetObligations(c *client.Client, config global.Config, lendingMarket string) []AccountWithObligation {
	cfg := rpc.GetProgramAccountsConfig{
		Encoding: rpc.GetProgramAccountsConfigEncodingBase64,
		Commitment: rpc.CommitmentConfirmed,
		Filters: []rpc.GetProgramAccountsConfigFilter{
			{
				MemCmp: &rpc.GetProgramAccountsConfigFilterMemCmp{
					Offset: 10,
					Bytes:  lendingMarket,
				},
			},
			{
				DataSize: OBLIGATION_LEN,
			},
		},
	}

  resp, err := c.RpcClient.GetProgramAccountsWithConfig(context.TODO(), config.ProgramID, cfg)
  if err != nil {
    fmt.Println(err)
		return []AccountWithObligation{}
  }

	var obligations []AccountWithObligation
	for _, account := range resp.Result {
		info, _ := utils.RpcProgramAccountInfoToClientAccountInfo(account.Account)
		accountWithObligation := ObligationParser(account.Pubkey, info)
		obligations = append(obligations, accountWithObligation)
	}

	return obligations
}

func GetReserves(c *client.Client, config global.Config, lendingMarket string) []AccountWithReserve {
	cfg := rpc.GetProgramAccountsConfig{
		Encoding: rpc.GetProgramAccountsConfigEncodingBase64,
		Commitment: rpc.CommitmentConfirmed,
		Filters: []rpc.GetProgramAccountsConfigFilter{
			{
				MemCmp: &rpc.GetProgramAccountsConfigFilterMemCmp{
					Offset: 10,
					Bytes:  lendingMarket,
				},
			},
			{
				DataSize: RESERVE_LEN,
			},
		},
	}

  resp, err := c.RpcClient.GetProgramAccountsWithConfig(context.TODO(), config.ProgramID, cfg)
  if err != nil {
    fmt.Println(err)
		return []AccountWithReserve{}
  }

	var reserves []AccountWithReserve
	for _, account := range resp.Result {
		info, _ := utils.RpcProgramAccountInfoToClientAccountInfo(account.Account)
		AccountWithReserve := ReserveParser(account.Pubkey, info)
		reserves = append(reserves, AccountWithReserve)
	}

	return reserves
}

func GetWalletTokenData(c *client.Client, config global.Config, wallet types.Account, mintAddress string, symbol string) global.WalletTokenData {
	return global.WalletTokenData {
		-1,
		-1,
		symbol,
	}
}