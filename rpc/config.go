package rpc

import (
	cmn "github.com/cosmos/evm/precompiles/common"
)

// RPCConfig holds the configuration and dependencies for the JSON-RPC server
// ELYS MODIFICATION: Added to allow clean module integration
type RPCConfig struct {
	// BankKeeper is used for querying bank module balances in GetBalance RPC calls
	BankKeeper cmn.BankKeeper
	// BaseDenom is the native token denomination (e.g., "uelys")
	BaseDenom string
}

// Global RPC configuration instance
// ELYS MODIFICATION: This allows the EVM module to be configured during app initialization
var globalRPCConfig *RPCConfig



// SetRPCConfig sets the global RPC configuration with required dependencies
func SetRPCConfig(config *RPCConfig) {
	globalRPCConfig = config
}

// GetRPCConfig returns the current RPC configuration
func GetRPCConfig() *RPCConfig {
	return globalRPCConfig
}

// GetBankKeeper returns the configured BankKeeper or nil if not set
func GetBankKeeper() cmn.BankKeeper {
	if globalRPCConfig != nil {
		return globalRPCConfig.BankKeeper
	}
	return nil
}

// GetBaseDenom returns the configured base denomination or default if not set
func GetBaseDenom() string {
	if globalRPCConfig != nil && globalRPCConfig.BaseDenom != "" {
		return globalRPCConfig.BaseDenom
	}
	return "uelys" // Default fallback
}

// IsConfigured returns true if the RPC system has been properly configured
func IsConfigured() bool {
	return globalRPCConfig != nil && globalRPCConfig.BankKeeper != nil
}
