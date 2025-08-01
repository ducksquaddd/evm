package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// EVMKeeper exposes the required EVM keeper interface required for ante handlers
type EVMKeeper interface {
	statedb.Keeper

	NewEVM(ctx sdk.Context, msg core.Message, cfg *statedb.EVMConfig, tracer *tracing.Hooks,
		stateDB vm.StateDB) *vm.EVM
	DeductTxCostsFromUserBalance(ctx sdk.Context, fees sdk.Coins, from common.Address) error
	SpendableCoin(ctx sdk.Context, addr common.Address) *uint256.Int
	ResetTransientGasUsed(ctx sdk.Context)
	GetTxIndexTransient(ctx sdk.Context) uint64
	GetParams(ctx sdk.Context) evmtypes.Params
	// GetBaseFee returns the BaseFee param from the fee market module
	// adapted according to the evm denom decimals
	GetBaseFee(ctx sdk.Context) *big.Int
	// GetMinGasPrice returns the MinGasPrice param from the fee market module
	// adapted according to the evm denom decimals
	GetMinGasPrice(ctx sdk.Context) math.LegacyDec
}

// FeeMarketKeeper exposes the required feemarket keeper interface required for ante handlers
type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (params feemarkettypes.Params)
	AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error)
	GetBaseFeeEnabled(ctx sdk.Context) bool
	GetBaseFee(ctx sdk.Context) math.LegacyDec
}

type ProtoTxProvider interface {
	GetProtoTx() *tx.Tx
}
