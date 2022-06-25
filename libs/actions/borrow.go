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
	"go-liquidator/utils"

	// "go-liquidator/utils"

	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/rpc"

	"github.com/portto/solana-go-sdk/types"
	"github.com/samber/lo"
)

func Borrow(
	c *client.Client,
	config global.Config,
	payer types.Account,
	liquidityAmount uint64,
	borrowTokenSymbol string,
	lendingMarket global.Market,
	obligation AccountWithObligation,
) error {
	//TODO: LiquidateAndRedeem
	ixs := []types.Instruction{}

	borrowReserve, _ := lo.Find(lendingMarket.Reserves, func(r global.Reserve) bool {
		return r.Asset == borrowTokenSymbol
	})

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
	uniqReserveAddresses = append(uniqReserveAddresses, borrowReserve.Address)
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

	borrowTokenInfo := libs.GetTokenInfo(config, borrowTokenSymbol)

	// get account that will be repaying the reserve liquidity
	borrowAccount, _, _ := common.FindAssociatedTokenAddress(
		payer.PublicKey,
		common.PublicKeyFromString(borrowTokenInfo.MintAddress),
	)

	ixs = append(ixs, BorrowObligationLiquidityIx(
		config,
		liquidityAmount,
		borrowReserve.LiquidityAddress,
		borrowAccount.ToBase58(),
		borrowReserve.Address,
		borrowReserve.LiquidityFeeReceiverAddress,
		obligation.Pubkey,
		lendingMarket.Address,
		lendingMarket.AuthorityAddress,
		payer.PublicKey.ToBase58(),
		"",
	))

	recent, err := c.GetRecentBlockhash(context.TODO())
	if err != nil {
		return fmt.Errorf("get recent block hash error, err: %v", err)
	}
	tx, _ := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{payer},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        payer.PublicKey,
			RecentBlockhash: recent.Blockhash,
			Instructions:    ixs,
		}),
	})
	// st, _ := c.SimulateTransactionWithConfig(context.TODO(), tx, client.SimulateTransactionConfig{
	// 	Commitment: rpc.CommitmentConfirmed,
	// })
	// fmt.Println(st.Err)
	// if st.Err != nil {
	// 	return fmt.Errorf("transaction simulate error")
	// }

	sig, err := c.SendTransactionWithConfig(context.TODO(), tx, client.SendTransactionConfig{
		SkipPreflight:       true,
		PreflightCommitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return err
	}
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
