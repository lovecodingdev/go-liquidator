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

/// Borrow liquidity from a reserve by depositing collateral tokens. Requires a refreshed
/// obligation and reserve.
///
/// Accounts expected by this instruction:
///
///   0. `[writable]` Source borrow reserve liquidity supply SPL Token account.
///   1. `[writable]` Destination liquidity token account.
///                     Minted by borrow reserve liquidity mint.
///   2. `[writable]` Borrow reserve account - refreshed.
///   3. `[writable]` Borrow reserve liquidity fee receiver account.
///                     Must be the fee account specified at InitReserve.
///   4. `[writable]` Obligation account - refreshed.
///   5. `[]` Lending market account.
///   6. `[]` Derived lending market authority.
///   7. `[signer]` Obligation owner.
///   8. `[]` Clock sysvar.
///   9. `[]` Token program id.
///   10 `[optional, writable]` Host fee receiver account.
func BorrowObligationLiquidityIx(
	config global.Config,
	liquidityAmount uint64,
	sourceLiquidity string,
	destinationLiquidity string,
	borrowReserve string,
	borrowReserveLiquidityFeeReceiver string,
	obligation string,
	lendingMarket string,
	lendingMarketAuthority string,
	obligationOwner string,
	hostFeeReceiver string,
) types.Instruction {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, byte(BorrowObligationLiquidity))
	binary.Write(buf, binary.LittleEndian, liquidityAmount)

	keys := []types.AccountMeta{
		{PubKey: common.PublicKeyFromString(sourceLiquidity), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(destinationLiquidity), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(borrowReserve), IsSigner: false, IsWritable: true},
		{
			PubKey:     common.PublicKeyFromString(borrowReserveLiquidityFeeReceiver),
			IsSigner:   false,
			IsWritable: true,
		},
		{PubKey: common.PublicKeyFromString(obligation), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(lendingMarket), IsSigner: false, IsWritable: false},
		{PubKey: common.PublicKeyFromString(lendingMarketAuthority), IsSigner: false, IsWritable: false},
		{PubKey: common.PublicKeyFromString(obligationOwner), IsSigner: true, IsWritable: false},
		{PubKey: common.SysVarClockPubkey, IsSigner: false, IsWritable: false},
		{PubKey: common.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	if hostFeeReceiver != "" {
		keys = append(keys, types.AccountMeta{
			PubKey: common.PublicKeyFromString(hostFeeReceiver), IsSigner: false, IsWritable: true,
		})
	}

	return types.Instruction{
		ProgramID: common.PublicKeyFromString(config.ProgramID),
		Accounts:  keys,
		Data:      buf.Bytes(),
	}
}
