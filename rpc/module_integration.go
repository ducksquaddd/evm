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
// SetupEVMRPC configures the EVM RPC with bank keeper and base denomination
// ELYS MODIFICATION: Enhanced to accept base denomination for proper configuration
func SetupEVMRPC(bankKeeper cmn.BankKeeper, baseDenom string) error {
	// Validate inputs
	if bankKeeper == nil {
		return fmt.Errorf("bank keeper cannot be nil")
	}
	
	if baseDenom == "" {
		return fmt.Errorf("base denomination cannot be empty")
	}
	
	// Set the global RPC configuration
	config := &RPCConfig{
		BankKeeper: bankKeeper,
		BaseDenom:  baseDenom,
	}
	
	SetRPCConfig(config)
	return nil
}

// Alternative: SetupEVMRPCWithConfig allows for more advanced configuration
// ELYS MODIFICATION: Extensible configuration for future enhancements
func SetupEVMRPCWithConfig(config *RPCConfig) {
	SetRPCConfig(config)
}