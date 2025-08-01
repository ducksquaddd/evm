package ante

import (
	storetypes "cosmossdk.io/store/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// BuildEvmExecutionCtx builds the context needed before executing an EVM transaction.
// It does the following:
// 1. Sets an empty KV gas config for gas to be calculated by opcodes
// and not kvstore actions
// 2. Setup an empty transient KV gas config for transient gas to be
// calculated by opcodes
func BuildEvmExecutionCtx(ctx sdktypes.Context) sdktypes.Context {
	// We need to setup an empty gas config so that the gas is consistent with Ethereum.
	return ctx.WithKVGasConfig(storetypes.GasConfig{}).
		WithTransientKVGasConfig(storetypes.GasConfig{})
}
