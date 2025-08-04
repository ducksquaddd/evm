package rpc

import (
	"fmt"

	cmn "github.com/cosmos/evm/precompiles/common"
)

// SetupEVMRPC configures the EVM RPC server with necessary dependencies
// This should be called during app initialization, after all keepers are created
// ELYS MODIFICATION: Clean integration function for blockchain projects
//
// Example usage in your app.go:
//
//	// After BankKeeper is initialized
//	rpc.SetupEVMRPC(app.BankKeeper)
//
// SetupEVMRPC configures the EVM RPC with bank keeper, base denomination, and context factory
// ELYS MODIFICATION: Enhanced to accept query context factory for direct keeper calls
func SetupEVMRPC(bankKeeper cmn.BankKeeper, baseDenom string, queryCtxFactory QueryContextFactory) error {
	// Validate inputs
	if bankKeeper == nil {
		return fmt.Errorf("bank keeper cannot be nil")
	}
	
	if baseDenom == "" {
		return fmt.Errorf("base denomination cannot be empty")
	}
	
	if queryCtxFactory == nil {
		return fmt.Errorf("query context factory cannot be nil")
	}
	
	// Set the global RPC configuration
	config := &RPCConfig{
		BankKeeper:          bankKeeper,
		BaseDenom:           baseDenom,
		QueryContextFactory: queryCtxFactory,
	}
	
	SetRPCConfig(config)
	return nil
}

// Alternative: SetupEVMRPCWithConfig allows for more advanced configuration
// ELYS MODIFICATION: Extensible configuration for future enhancements
func SetupEVMRPCWithConfig(config *RPCConfig) {
	SetRPCConfig(config)
}