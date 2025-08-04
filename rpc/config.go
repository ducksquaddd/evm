package rpc

import (
	cmn "github.com/cosmos/evm/precompiles/common"
)

// RPCConfig holds the configuration and dependencies for the JSON-RPC server
// ELYS MODIFICATION: Added to allow clean module integration
type RPCConfig struct {
	// BankKeeper is used for querying bank module balances in GetBalance RPC calls
	BankKeeper cmn.BankKeeper
}

// Global RPC configuration instance
// ELYS MODIFICATION: This allows the EVM module to be configured during app initialization
var globalRPCConfig *RPCConfig

// SetRPCConfig sets the global RPC configuration with required dependencies
// This should be called during app initialization after keepers are created
// ELYS MODIFICATION: Clean integration point for ELYS project
func SetRPCConfig(config *RPCConfig) {
	globalRPCConfig = config
}

// GetRPCConfig returns the current RPC configuration
// Returns nil if not configured (for backward compatibility)
func GetRPCConfig() *RPCConfig {
	return globalRPCConfig
}

// GetBankKeeper returns the configured BankKeeper or nil if not set
// This provides a clean way to access the bank keeper from RPC handlers
func GetBankKeeper() cmn.BankKeeper {
	if globalRPCConfig != nil {
		return globalRPCConfig.BankKeeper
	}
	return nil
}