package libs

import (
	"fmt"
	// "sync"
	"context"
	// "math"
	// "math/big"

	"go-liquidator/global"
	. "go-liquidator/models/layouts"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"

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
		accountWithObligation := ObligationParser(account.Pubkey, account.Account)
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
		AccountWithReserve := ReserveParser(account.Pubkey, account.Account)
		reserves = append(reserves, AccountWithReserve)
	}

	return reserves
}