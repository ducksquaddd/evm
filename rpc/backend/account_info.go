package backend

import (
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

    	// Direct keeper call with proper SDK context (no gRPC overhead)
	fmt.Printf("🔍 Creating query context for height: %d\n", blockNum.Int64())
	
	// Use the secure context factory to create a read-only SDK context
	if b.queryCtxFactory == nil {
		fmt.Printf("⚠️ queryCtxFactory is nil, falling back to EVM state query\n")
		// Fallback to original EVM state query
		req := &evmtypes.QueryBalanceRequest{
			Address: address.String(),
		}
		
		grpcCtx := rpctypes.ContextWithHeight(blockNum.Int64())
		res, err := b.queryClient.Balance(grpcCtx, req)
		if err != nil {
			return nil, err
		}
		
		val, ok := sdkmath.NewIntFromString(res.Balance)
		if !ok {
			return nil, errors.New("invalid balance")
		}
		
		if val.IsNegative() {
			return nil, errors.New("couldn't fetch balance. Node state is pruned")
		}
		
		// ELYS SCALING: Apply same scaling to EVM fallback for consistency
		scalingFactor := sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)) // 10^12
		scaledVal := val.Mul(scalingFactor)
		fmt.Printf("🔄 Scaled EVM balance for MetaMask: %s (added 12 zeros)\n", scaledVal.String())
		
		return (*hexutil.Big)(scaledVal.BigInt()), nil
	}
	
	// Create proper SDK context for the specific height
	queryCtx := b.queryCtxFactory(blockNum.Int64())
	fmt.Printf("✅ Created SDK context for direct keeper call\n")
	
	// Direct bank keeper call (fast, no gRPC overhead)
	balance := b.bankKeeper.GetBalance(queryCtx, cosmosAddr, b.baseDenom)
	val := balance.Amount
	fmt.Printf("✅ Bank balance found via direct keeper call: %s %s (raw)\n", val.String(), b.baseDenom)

    if val.IsNegative() {
        return nil, errors.New("couldn't fetch balance. Node state is pruned")
    }

    // ELYS SCALING: Convert 6-decimal ELYS to 18-decimal for MetaMask compatibility
    // MetaMask expects 18 decimals for native tokens, but ELYS uses 6 decimals
    // So we multiply by 10^12 to add the missing 12 decimal places
    scalingFactor := sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)) // 10^12
    scaledVal := val.Mul(scalingFactor)
    
    fmt.Printf("🔄 Scaled balance for MetaMask: %s (added 12 zeros)\n", scaledVal.String())

    return (*hexutil.Big)(scaledVal.BigInt()), nil
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
