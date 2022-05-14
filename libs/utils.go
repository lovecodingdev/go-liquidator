package libs

import (
	"fmt"
	// "sync"
	"context"
	// "math"
	// "math/big"

	. "go-liquidator/global"
	. "go-liquidator/models/layouts"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	// "github.com/portto/solana-go-sdk/common"

)

func GetObligations(c *client.Client, config Config, lendingMarket string) {
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
		return
  }

	var obligations []AccountWithObligation
	for _, account := range resp.Result {
		accountWithObligation := ObligationParser(account.Pubkey, account.Account)
		obligations = append(obligations, accountWithObligation)
	}

	return obligations
}