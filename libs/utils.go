package libs

import (
	"fmt"
	// "sync"
	"context"
	"math"

	// "math/big"

	"go-liquidator/global"
	. "go-liquidator/models/layouts"
	"go-liquidator/utils"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/tokenprog"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
	"github.com/samber/lo"
)

// Converts amount to human (rebase with decimals)
func ToHuman(config global.Config, amount string, symbol string) string {
	// decimals := GetDecimals(config, symbol)
	// return toHumanDec(amount, decimals);
	return amount
}

func GetDecimals(config global.Config, symbol string) uint8 {
	tokenInfo := GetTokenInfo(config, symbol)
	return tokenInfo.Decimals
}

// Returns token info from config
func GetTokenInfo(config global.Config, symbol string) global.Asset {
	asset, _ := lo.Find(config.Assets, func(_asset global.Asset) bool {
		return _asset.Symbol == symbol
	})
	return asset
}

func GetObligations(c *client.Client, config global.Config, lendingMarket string) []AccountWithObligation {
	cfg := rpc.GetProgramAccountsConfig{
		Encoding:   rpc.GetProgramAccountsConfigEncodingBase64,
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
		Encoding:   rpc.GetProgramAccountsConfigEncodingBase64,
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

func GetWalletTokenData(c *client.Client, config global.Config, wallet types.Account, mintAddress string, symbol string) (global.WalletTokenData, error) {
	walletTokenData := global.WalletTokenData{
		Balance:     -1,
		BalanceBase: -1,
		Symbol:      symbol,
	}

	userTokenAccount, _, _ := common.FindAssociatedTokenAddress(wallet.PublicKey, common.PublicKeyFromString(mintAddress))

	getAccountInfoResponse, err := c.GetAccountInfo(context.TODO(), userTokenAccount.ToBase58())
	if err != nil {
		return walletTokenData, fmt.Errorf("failed to get account info, err: %v", err)
	}

	tokenAccount, err := tokenprog.TokenAccountFromData(getAccountInfoResponse.Data)
	if err != nil {
		return walletTokenData, fmt.Errorf("failed to parse data to a token account, err: %v", err)
	}

	decimals := GetDecimals(config, symbol)
	walletTokenData.Balance = float64(tokenAccount.Amount) / math.Pow(10, float64(decimals))
	walletTokenData.BalanceBase = int64(tokenAccount.Amount)

	return walletTokenData, nil
}
