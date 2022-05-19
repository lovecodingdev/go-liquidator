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

/// Accrue interest and update market price of liquidity on a reserve.
/// Accounts expected by this instruction:
///   0. `[writable]` Reserve account.
///   1. `[]` Clock sysvar.
///   2. `[optional]` Reserve liquidity oracle account.
///                     Required if the reserve currency is not the lending market quote
///                     currency.

func RefreshReserveInstruction(
	config global.Config,
	reserve string,
	oracle string,
	switchboardFeedAddress string,
) types.Instruction {
	var buf bytes.Buffer
	buf.WriteByte(byte(RefreshReserve))

	keys := []types.AccountMeta{
		{PubKey: common.PublicKeyFromString(reserve), IsSigner: false, IsWritable: true},
	}

	if oracle != "" {
		keys = append(keys, types.AccountMeta{
			PubKey: common.PublicKeyFromString(oracle), IsSigner: false, IsWritable: false,
		})
	}

	if switchboardFeedAddress != "" {
		keys = append(keys, types.AccountMeta{
			PubKey: common.PublicKeyFromString(switchboardFeedAddress), IsSigner: false, IsWritable: false,
		})
	}

	keys = append(keys, types.AccountMeta{
		PubKey: common.SysVarClockPubkey, IsSigner: false, IsWritable: false,
	})

	return types.Instruction{
		ProgramID: common.PublicKeyFromString(config.ProgramID),
		Accounts:  keys,
		Data:      buf.Bytes(),
	}
}
