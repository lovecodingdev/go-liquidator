package instructions

import (
	// "fmt"
	// "sync"
	// "context"
	// "math"
	// "math/big"
	"bytes"

	"go-liquidator/global"
	// . "go-liquidator/models/layouts"

	// "go-liquidator/utils"

	// "github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/common"
	"github.com/portto/solana-go-sdk/types"
)

/// Refresh an obligation"s accrued interest and collateral and liquidity prices. Requires
/// refreshed reserves, as all obligation collateral deposit reserves in order, followed by all
/// liquidity borrow reserves in order.
/// Accounts expected by this instruction:
///   0. `[writable]` Obligation account.
///   1. `[]` Clock sysvar.
///   .. `[]` Collateral deposit reserve accounts - refreshed, all, in order.
///   .. `[]` Liquidity borrow reserve accounts - refreshed, all, in order.

func RefreshObligationInstruction(
	config global.Config,
	obligation string,
	depositReserves []string,
	borrowReserves []string,
) types.Instruction {
	var buf bytes.Buffer
	buf.WriteByte(byte(RefreshObligation))

	keys := []types.AccountMeta{
		{PubKey: common.PublicKeyFromString(obligation), IsSigner: false, IsWritable: true},
		{PubKey: common.SysVarClockPubkey, IsSigner: false, IsWritable: false},
	}

	for _, depositReserve := range depositReserves {
		keys = append(keys, types.AccountMeta{
			PubKey: common.PublicKeyFromString(depositReserve), IsSigner: false, IsWritable: false,
		})
	}

	for _, borrowReserve := range borrowReserves {
		keys = append(keys, types.AccountMeta{
			PubKey: common.PublicKeyFromString(borrowReserve), IsSigner: false, IsWritable: false,
		})
	}

	return types.Instruction{
		ProgramID: common.PublicKeyFromString(config.ProgramID),
		Accounts:  keys,
		Data:      buf.Bytes(),
	}
}
