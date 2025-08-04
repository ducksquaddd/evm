package rpc

import (
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
func SetupEVMRPC(bankKeeper cmn.BankKeeper) {
	config := &RPCConfig{
		BankKeeper: bankKeeper,
	}
	SetRPCConfig(config)
}

// Alternative: SetupEVMRPCWithConfig allows for more advanced configuration
// ELYS MODIFICATION: Extensible configuration for future enhancements
func SetupEVMRPCWithConfig(config *RPCConfig) {
	SetRPCConfig(config)
}