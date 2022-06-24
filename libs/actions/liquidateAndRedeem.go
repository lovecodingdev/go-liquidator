package actions

import (
	// "fmt"
	// "sync"
	"context"
	"encoding/base64"
	"fmt"

	// "math"
	// "math/big"

	"go-liquidator/global"
	"go-liquidator/libs"
	. "go-liquidator/models/instructions"
	. "go-liquidator/models/layouts"
	"go-liquidator/utils"

	// "go-liquidator/utils"

	"github.com/google/go-cmp/cmp"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/program/assotokenprog"
	"github.com/portto/solana-go-sdk/rpc"

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
	uniqReserveAddresses = utils.RemoveDuplicateStr(uniqReserveAddresses)

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

	recent, err := c.GetRecentBlockhash(context.TODO())
	if err != nil {
		return fmt.Errorf("get recent block hash error, err: %v\n", err)
	}
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{payer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        payer.PublicKey,
			RecentBlockhash: recent.Blockhash,
			Instructions:    ixs,
		}),
	})
	st, _ := c.SimulateTransactionWithConfig(context.TODO(), tx, client.SimulateTransactionConfig{
		Commitment: rpc.CommitmentConfirmed,
	})
	if st.Err != nil {
		return fmt.Errorf("transaction simulate error")
	}

	sig, err := c.SendTransactionWithConfig(context.TODO(), tx, client.SendTransactionConfig{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentConfirmed,
	})
	fmt.Println(sig, err)

	err = utils.ConfirmTransaction(sig, c)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("failed to send tx, err: %v", err)
	}

	return nil
}

func TestLiquidate(
	c *client.Client,
	payer types.Account,
) {
	rawTx := "Ad4zuw8JOX9KLt0IGq8va40hQXYxkG/tus2Ekpp3LEplpNJgYTr0fZ1sWl3y4Re06Z8YysYdFDYhalu58RQByAIBAA0Z6Xkn+tjqP1sq+H9Q2SnDyrzTgk1KfZxKQn16b3+9cicsyOrxNJY/DlpMLc6ikNpF9llAUIJlXFP08XLmMF0DAy+i8sb1Busg9MHxG78eSRr2AK/4FfA5L85ckYmrWVs+MQCTR3TtKup33ecWjcQBW1q0qoseEVx4yDI+u8tY9ORBnFgt1wFS6gnO1AZ4Jjc7YGOZUcscGTsPGMU2tUtNN0h5L0gP01Mqv3ew77vr0QDiJ2YipzzHUSersr+3PcPBTVfc7BSIeVLXNlgFW5+Pdfhcrqs8tueYQZJHb9zlqXtul0iUCXFwDIz0xUcYRKbyPbqa/fJWwF3bum9ZiuqkHJD6sfSY4fNzPsbvZRUjnRPcS93vyQJm4w1eYWM2rbYAntF0r0JELmchUiZpzVUNFFLmIyaOEqKPSmJnycKql3qzVAcoaB9/XYEzm1mWfUlMy5Xd7nctmzGEAbnmpNqSwLqClYFhQAdhV3to7I3Ht7kdIep2QzcQL52g/+OP2y/qAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAjJFu3QlTmWpjMP/SjdEPXn1J+RORJdQrTBFOLAG8hvDOzHsTv+PoomuqMlUwBYy4tdkkIzlRNaGW97xEb/2ErjJclj04kifG7PRApFI4NgwtaE5na/xCEBI572Nvp+FmXGDa+TfXc0GI3sx0QDltQKsYzIl7TAmg1hJPMi5GeKKvhWnBfDtO4nEp79MhTmazMc4TX/koRahCZkYlPFD3mu6nN/XmuJp8Db7aXm0uYp3KSexdu9rcZZXaHIfHx8uTqoCDGHMR5cSgTRhzhU4lKlqbACyHtDPwnmNH5qenJSgabi5haq1MqRQkN6FV/zdy+bLfvxzoKZbBvkgNdtz7sBoMQhhqYMn0FUFdNhEGKpuEMM1Ldqn/X9YFSzO6yOIcGp9UXGMd0yShWY5hpHV62i164o5tLbVxzVVshAAAAAAan1RcZLFxRIYzJTD1K8X9Y2u4Im6H9ROPb2YoAAAAABt324ddloZPZy+FGzut5rBy0he1fWzeROoz1hX7/AKn4Wd+NXaopFYWSXDDUhNWf0Ordek5CpDiBl5SJvHNgVAYUBAUNEBYBAxQECRMRFgEDFAQBFgUJAQcPBwAGAAIMGBcADwcABAAVDBgXABQPCwYECQcFAgoDCAEOEgAYCRHZ4CUAAAAAAA=="

	bytesTx, _ := base64.StdEncoding.DecodeString(rawTx)
	tx, _ := types.TransactionDeserialize(bytesTx)

	recentBlockhashResponse, err := c.GetRecentBlockhash(context.TODO())
	tx.Message.RecentBlockHash = recentBlockhashResponse.Blockhash
	data, err := tx.Message.Serialize()
	tx.Signatures[0] = payer.Sign(data)

	sig, err := c.SendTransactionWithConfig(context.TODO(), tx, client.SendTransactionConfig{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentConfirmed,
	})

	fmt.Println(sig, err)
}
