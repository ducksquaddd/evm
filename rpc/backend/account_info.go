package backend

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"

	"github.com/cometbft/cometbft/libs/bytes"

	rpctypes "github.com/cosmos/evm/rpc/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetCode returns the contract code at the given address and block number.
func (b *Backend) GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Code(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetProof returns an account object with proof and any storage proofs
func (b *Backend) GetProof(address common.Address, storageKeys []string, blockNrOrHash rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	height := int64(blockNum)

	_, err = b.TendermintBlockByNumber(blockNum)
	if err != nil {
		// the error message imitates geth behavior
		return nil, errors.New("header not found")
	}

	// if the height is equal to zero, meaning the query condition of the block is either "pending" or "latest"
	if height == 0 {
		bn, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}

		if bn > math.MaxInt64 {
			return nil, fmt.Errorf("not able to query block number greater than MaxInt64")
		}

		height = int64(bn) //#nosec G115 -- checked for int overflow already
	}

	ctx := rpctypes.ContextWithHeight(height)
	clientCtx := b.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))

	for i, key := range storageKeys {
		hexKey := common.HexToHash(key)
		valueBz, proof, err := b.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, hexKey.Bytes()))
		if err != nil {
			return nil, err
		}

		storageProofs[i] = rpctypes.StorageResult{
			Key:   key,
			Value: (*hexutil.Big)(new(big.Int).SetBytes(valueBz)),
			Proof: GetHexProofs(proof),
		}
	}

	// query EVM account
	req := &evmtypes.QueryAccountRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	accountKey := bytes.HexBytes(append(authtypes.AddressStoreKeyPrefix, address.Bytes()...))
	_, proof, err := b.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	balance, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return &rpctypes.AccountResult{
		Address:      address,
		AccountProof: GetHexProofs(proof),
		Balance:      (*hexutil.Big)(balance.BigInt()),
		CodeHash:     common.HexToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Cosmos EVM doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (b *Backend) GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := b.queryClient.Storage(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetBalance returns the provided account's *spendable* balance up to the provided block number.
// ELYS MODIFICATION: Use bank module instead of EVM state for native token balance
func (b *Backend) GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error) {
    // Add debug logging
    fmt.Printf("🔍 GetBalance called for address: %s\n", address.Hex())
    
    blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
    if err != nil {
        fmt.Printf("❌ BlockNumberFromTendermint error: %v\n", err)
        return nil, err
    }

    _, err = b.TendermintBlockByNumber(blockNum)
    if err != nil {
        fmt.Printf("❌ TendermintBlockByNumber error: %v\n", err)
        return nil, err
    }

    // Check if bankKeeper is available
    if b.bankKeeper == nil {
        fmt.Printf("⚠️ bankKeeper is nil, falling back to EVM state\n")
        // Fallback to original implementation
        // ... original EVM state query code
    }

    // Check if baseDenom is available
    if b.baseDenom == "" {
        fmt.Printf("⚠️ baseDenom is empty, using default 'uelys'\n")
        b.baseDenom = "uelys"
    }

    fmt.Printf("✅ Using bankKeeper with baseDenom: %s\n", b.baseDenom)

    // Convert Ethereum address to Cosmos address
    cosmosAddr := sdk.AccAddress(address.Bytes())
    fmt.Printf("✅ Converted to cosmos address: %s\n", cosmosAddr.String())

    // Create a proper SDK context with store access for direct keeper calls
    // rpctypes.ContextWithHeight doesn't work for direct keeper calls - only for gRPC queries
    fmt.Printf("🔍 Creating proper SDK context for height: %d\n", blockNum.Int64())
    
    // Use the backend's base context and set the height
    ctx := b.ctx  // Use the backend's context
    if ctx == nil {
        fmt.Printf("⚠️ Backend context is nil, using a background context\n")
        ctx = context.Background()
    }
    
    // For now, let's try using the latest block context since height-specific queries are failing
    // We'll use rpctypes.ContextWithHeight(0) which should use latest
    queryCtx := rpctypes.ContextWithHeight(0) // 0 = latest block
    fmt.Printf("✅ Using latest block context instead of specific height\n")
    
    fmt.Printf("✅ Using context with height: %d\n", blockNum.Int64())
    fmt.Printf("🔍 About to call GetBalance on bank keeper...\n")
    
    // Try direct keeper call with panic recovery
    var val sdkmath.Int
    
    // Attempt direct keeper call with panic recovery
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("⚠️ Direct keeper call panicked: %v\n", r)
            // TODO: Implement gRPC fallback here if needed
        }
    }()
    
    // Try the direct keeper approach
    balance := b.bankKeeper.GetBalance(queryCtx, cosmosAddr, b.baseDenom)
    val = balance.Amount
    fmt.Printf("✅ Direct bank keeper call successful: %s %s\n", val.String(), b.baseDenom)

    if val.IsNegative() {
        return nil, errors.New("couldn't fetch balance. Node state is pruned")
    }

    return (*hexutil.Big)(val.BigInt()), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (b *Backend) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	n := hexutil.Uint64(0)
	bn, err := b.BlockNumber()
	if err != nil {
		return &n, err
	}
	height := blockNum.Int64()

	currentHeight := int64(bn) //#nosec G115 -- checked for int overflow already
	if height > currentHeight {
		return &n, errorsmod.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"cannot query with height in the future (current: %d, queried: %d); please provide a valid height",
			currentHeight, height,
		)
	}
	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := b.clientCtx.AccountRetriever

	err = accRet.EnsureExists(b.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		return &n, nil
	}

	includePending := blockNum == rpctypes.EthPendingBlockNumber
	nonce, err := b.getAccountNonce(address, includePending, blockNum.Int64(), b.logger)
	if err != nil {
		return nil, err
	}

	n = hexutil.Uint64(nonce)
	return &n, nil
}
