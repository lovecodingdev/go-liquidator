package instructions

import (
	// "fmt"
	// "sync"
	// "context"
	// "math"
	// "math/big"
	"bytes"
	"encoding/binary"

	"go-liquidator/global"
	// . "go-liquidator/models/layouts"

	// "go-liquidator/utils"

	// "github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

/// Repay borrowed liquidity to a reserve to receive collateral at a discount from an unhealthy
/// obligation. Requires a refreshed obligation and reserves.
/// Accounts expected by this instruction:
///   0. `[writable]` Source liquidity token account.
///                     Minted by repay reserve liquidity mint.
///                     $authority can transfer $liquidity_amount.
///   1. `[writable]` Destination collateral token account.
///                     Minted by withdraw reserve collateral mint.
///   2. `[writable]` Destination liquidity token account.
///   3. `[writable]` Repay reserve account - refreshed.
///   4. `[writable]` Repay reserve liquidity supply SPL Token account.
///   5. `[writable]` Withdraw reserve account - refreshed.
///   6. `[writable]` Withdraw reserve collateral SPL Token mint.
///   7. `[writable]` Withdraw reserve collateral supply SPL Token account.
///   8. `[writable]` Withdraw reserve liquidity supply SPL Token account.
///   9. `[writable]` Withdraw reserve liquidity fee receiver account.
///   10 `[writable]` Obligation account - refreshed.
///   11 `[]` Lending market account.
///   12 `[]` Derived lending market authority.
///   13 `[signer]` User transfer authority ($authority).
///   14 `[]` Token program id.

func LiquidateObligationAndRedeemReserveCollateralIx(
	config global.Config,
	liquidityAmount uint64,
	sourceLiquidity string,
	destinationCollateral string,
	destinationRewardLiquidity string,
	repayReserve string,
	repayReserveLiquiditySupply string,
	withdrawReserve string,
	withdrawReserveCollateralMint string,
	withdrawReserveCollateralSupply string,
	withdrawReserveLiquiditySupply string,
	withdrawReserveFeeReceiver string,
	obligation string,
	lendingMarket string,
	lendingMarketAuthority string,
	transferAuthority string,
) types.Instruction {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, byte(LiquidateObligationAndRedeemReserveCollateral))
	binary.Write(buf, binary.LittleEndian, liquidityAmount)

	keys := []types.AccountMeta{
		{PubKey: common.PublicKeyFromString(sourceLiquidity), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(destinationCollateral), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(destinationRewardLiquidity), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(repayReserve), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(repayReserveLiquiditySupply), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(withdrawReserve), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(withdrawReserveCollateralMint), IsSigner: false, IsWritable: true},

		{PubKey: common.PublicKeyFromString(withdrawReserveCollateralSupply), IsSigner: false, IsWritable: true},

		{PubKey: common.PublicKeyFromString(withdrawReserveLiquiditySupply), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(withdrawReserveFeeReceiver), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(obligation), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(lendingMarket), IsSigner: false, IsWritable: false},
		{PubKey: common.PublicKeyFromString(lendingMarketAuthority), IsSigner: false, IsWritable: false},
		{PubKey: common.PublicKeyFromString(transferAuthority), IsSigner: true, IsWritable: false},
		{PubKey: common.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	return types.Instruction{
		ProgramID: common.PublicKeyFromString(config.ProgramID),
		Accounts:  keys,
		Data:      buf.Bytes(),
	}
}
