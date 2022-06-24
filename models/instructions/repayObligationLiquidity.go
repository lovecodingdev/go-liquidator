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

/// Repay borrowed liquidity to a reserve. Requires a refreshed obligation and reserve.
///
/// Accounts expected by this instruction:
///
///   0. `[writable]` Source liquidity token account.
///                     Minted by repay reserve liquidity mint.
///                     $authority can transfer $liquidity_amount.
///   1. `[writable]` Destination repay reserve liquidity supply SPL Token account.
///   2. `[writable]` Repay reserve account - refreshed.
///   3. `[writable]` Obligation account - refreshed.
///   4. `[]` Lending market account.
///   5. `[]` Derived lending market authority.
///   6. `[signer]` User transfer authority ($authority).
///   7. `[]` Clock sysvar.
///   8. `[]` Token program id.
func RepayObligationLiquidityIx(
	config global.Config,
	liquidityAmount uint64,
	sourceLiquidity string,
	destinationLiquidity string,
	repayReserve string,
	obligation string,
	lendingMarket string,
	transferAuthority string,
) types.Instruction {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, byte(RepayObligationLiquidity))
	binary.Write(buf, binary.LittleEndian, liquidityAmount)

	keys := []types.AccountMeta{
		{PubKey: common.PublicKeyFromString(sourceLiquidity), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(destinationLiquidity), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(repayReserve), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(obligation), IsSigner: false, IsWritable: true},
		{PubKey: common.PublicKeyFromString(lendingMarket), IsSigner: false, IsWritable: false},
		{PubKey: common.PublicKeyFromString(transferAuthority), IsSigner: true, IsWritable: false},
		{PubKey: common.SysVarClockPubkey, IsSigner: false, IsWritable: false},
		{PubKey: common.TokenProgramID, IsSigner: false, IsWritable: false},
	}

	return types.Instruction{
		ProgramID: common.PublicKeyFromString(config.ProgramID),
		Accounts:  keys,
		Data:      buf.Bytes(),
	}
}
