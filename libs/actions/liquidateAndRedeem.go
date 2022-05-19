package actions

import (
	// "fmt"
	// "sync"
	"context"
	"fmt"

	// "math"
	// "math/big"

	"go-liquidator/global"
	"go-liquidator/libs"
	. "go-liquidator/models/instructions"
	. "go-liquidator/models/layouts"

	"github.com/google/go-cmp/cmp"

	// "go-liquidator/utils"

	"github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/assotokenprog"
	"github.com/portto/solana-go-sdk/types"
	"github.com/samber/lo"
)

func LiquidateAndRedeem(
	c *client.Client,
	config global.Config,
	payer types.Account,
	liquidityAmount uint64,
	repayTokenSymbol string,
	withdrawTokenSymbol string,
	lendingMarket global.Market,
	obligation AccountWithObligation,
) error {
	//TODO: LiquidateAndRedeem
	ixs := []types.Instruction{}

	depositReserves := []string{}
	for _, deposit := range obligation.Info.Deposits {
		depositReserves = append(depositReserves, deposit.DepositReserve)
	}

	borrowReserves := []string{}
	for _, borrow := range obligation.Info.Borrows {
		borrowReserves = append(borrowReserves, borrow.BorrowReserve)
	}

	uniqReserveAddresses := []string{}
	uniqReserveAddresses = append(uniqReserveAddresses, depositReserves...)
	uniqReserveAddresses = append(uniqReserveAddresses, borrowReserves...)

	for _, reserveAddress := range uniqReserveAddresses {
		reserveInfo, _ := lo.Find(lendingMarket.Reserves, func(r global.Reserve) bool {
			return r.Address == reserveAddress
		})

		oracleInfo, _ := lo.Find(config.Oracles.Assets, func(oa global.OracleAsset) bool {
			return oa.Asset == reserveInfo.Asset
		})

		refreshReserveIx := RefreshReserveInstruction(
			config,
			reserveAddress,
			oracleInfo.PriceAddress,
			oracleInfo.SwitchboardFeedAddress,
		)
		ixs = append(ixs, refreshReserveIx)
	}

	refreshObligationIx := RefreshObligationInstruction(
		config,
		obligation.Pubkey,
		depositReserves,
		borrowReserves,
	)
	ixs = append(ixs, refreshObligationIx)

	repayTokenInfo := libs.GetTokenInfo(config, repayTokenSymbol)

	// get account that will be repaying the reserve liquidity
	repayAccount, _, _ := common.FindAssociatedTokenAddress(
		payer.PublicKey,
		common.PublicKeyFromString(repayTokenInfo.MintAddress),
	)

	repayReserve, _ := lo.Find(lendingMarket.Reserves, func(r global.Reserve) bool {
		return r.Asset == repayTokenSymbol
	})
	withdrawReserve, _ := lo.Find(lendingMarket.Reserves, func(r global.Reserve) bool {
		return r.Asset == withdrawTokenSymbol
	})
	withdrawTokenInfo := libs.GetTokenInfo(config, withdrawTokenSymbol)

	rewardedWithdrawalCollateralAccount, _, _ := common.FindAssociatedTokenAddress(
		payer.PublicKey,
		common.PublicKeyFromString(withdrawReserve.CollateralMintAddress),
	)
	rewardedWithdrawalCollateralAccountInfo, err := c.GetAccountInfo(context.TODO(), rewardedWithdrawalCollateralAccount.ToBase58())
	if err != nil {
		return fmt.Errorf("failed to get account info, err: %v", err)
	}
	if cmp.Equal(rewardedWithdrawalCollateralAccountInfo, (client.AccountInfo{})) {
		createUserCollateralAccountIx := assotokenprog.CreateAssociatedTokenAccount(assotokenprog.CreateAssociatedTokenAccountParam{
			Funder:                 payer.PublicKey,
			Owner:                  payer.PublicKey,
			Mint:                   common.PublicKeyFromString(withdrawReserve.CollateralMintAddress),
			AssociatedTokenAccount: rewardedWithdrawalCollateralAccount,
		})
		ixs = append(ixs, createUserCollateralAccountIx)
	}

	rewardedWithdrawalLiquidityAccount, _, _ := common.FindAssociatedTokenAddress(
		payer.PublicKey,
		common.PublicKeyFromString(withdrawTokenInfo.MintAddress),
	)
	rewardedWithdrawalLiquidityAccountInfo, err := c.GetAccountInfo(context.TODO(), rewardedWithdrawalLiquidityAccount.ToBase58())
	if err != nil {
		return fmt.Errorf("failed to get account info, err: %v", err)
	}
	if cmp.Equal(rewardedWithdrawalLiquidityAccountInfo, (client.AccountInfo{})) {
		createUserCollateralAccountIx := assotokenprog.CreateAssociatedTokenAccount(assotokenprog.CreateAssociatedTokenAccountParam{
			Funder:                 payer.PublicKey,
			Owner:                  payer.PublicKey,
			Mint:                   common.PublicKeyFromString(withdrawTokenInfo.MintAddress),
			AssociatedTokenAccount: rewardedWithdrawalLiquidityAccount,
		})
		ixs = append(ixs, createUserCollateralAccountIx)
	}

	ixs = append(ixs, LiquidateObligationAndRedeemReserveCollateralIx(
		config,
		liquidityAmount,
		repayAccount.ToBase58(),
		rewardedWithdrawalCollateralAccount.ToBase58(),
		rewardedWithdrawalLiquidityAccount.ToBase58(),
		repayReserve.Address,
		repayReserve.LiquidityAddress,
		withdrawReserve.Address,
		withdrawReserve.CollateralMintAddress,
		withdrawReserve.CollateralSupplyAddress,
		withdrawReserve.LiquidityAddress,
		withdrawReserve.LiquidityFeeReceiverAddress,
		obligation.Pubkey,
		lendingMarket.Address,
		lendingMarket.AuthorityAddress,
		payer.PublicKey.ToBase58(),
	))

	recentBlockhashResponse, err := c.GetRecentBlockhash(context.Background())
	if err != nil {
		return fmt.Errorf("get recent block hash error, err: %v\n", err)
	}
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{payer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        payer.PublicKey,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
			Instructions:    ixs,
		}),
	})
	sig, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		return fmt.Errorf("failed to send tx, err: %v", err)
	}
	fmt.Println(sig)
	return nil
}
