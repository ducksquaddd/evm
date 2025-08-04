# EVM Module Integration Guide for ELYS Project

This guide shows how to integrate the modified cosmos/evm module into your ELYS blockchain project to enable `eth_getBalance` calls that return actual bank module balances instead of EVM state.

## Integration Steps

### 1. Update go.mod

Add the replace directive in your ELYS project's go.mod:

```go
replace github.com/cosmos/evm => github.com/ducksquaddd/evm v0.3.0-elys
```

Then run:
```bash
go mod tidy
```

### 2. Import EVM Module in Your App

In your `app.go` file, import the EVM module:

```go
import (
    // ... your existing imports ...
    "github.com/cosmos/evm/rpc"
    evmkeeper "github.com/cosmos/evm/x/vm/keeper"
    evmtypes "github.com/cosmos/evm/x/vm/types"
    // ... other EVM imports as needed ...
)
```

### 3. Configure EVM RPC During App Initialization

After your BankKeeper is initialized, configure the EVM RPC system:

```go
func NewElysApp(
    logger log.Logger,
    db dbm.DB,
    traceStore io.Writer,
    loadLatest bool,
    appOpts servertypes.AppOptions,
    baseAppOptions ...func(*baseapp.BaseApp),
) *ElysApp {
    // ... your existing app initialization ...

    // Initialize BankKeeper (this is probably already in your code)
    app.BankKeeper = bankkeeper.NewBaseKeeper(
        appCodec,
        runtime.NewKVStoreService(keys[banktypes.StoreKey]),
        app.AccountKeeper,
        BlockedAddresses(),
        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
        logger,
    )

    // ... other keeper initialization ...

    // ELYS MODIFICATION: Configure EVM RPC to use your BankKeeper
    // This enables eth_getBalance to return actual uelys balances
    rpc.SetupEVMRPC(app.BankKeeper)

    // ... rest of your app initialization ...

    return app
}
```

### 4. Alternative: Advanced Configuration

If you need more control over the configuration:

```go
// Instead of rpc.SetupEVMRPC(app.BankKeeper), use:
rpcConfig := &rpc.RPCConfig{
    BankKeeper: app.BankKeeper,
    // Add more configuration fields as needed in the future
}
rpc.SetupEVMRPCWithConfig(rpcConfig)
```

## What This Enables

After integration:

1. **eth_getBalance** calls will return actual **uelys** balances from your bank module
2. The balance represents the single source of truth from your native token
3. No changes needed to your existing EVM transaction processing
4. Backward compatible - if RPC configuration is not set, it falls back to EVM state

## Example Usage

Once integrated, Ethereum clients can query balances:

```javascript
// This will now return actual uelys balance instead of 0
const balance = await web3.eth.getBalance("0x1234..."); 
console.log(`Balance: ${web3.utils.fromWei(balance, 'ether')} ELYS`);
```

## Testing Integration

1. Fund an account with uelys using your bank module
2. Query the balance using eth_getBalance RPC call
3. Verify it returns the correct uelys amount (not 0)

## Benefits of This Approach

- **Clean Architecture**: No need to modify cosmos/evm core files
- **Module Pattern**: Follows standard Cosmos SDK module integration practices  
- **Easy Maintenance**: Simple to update when cosmos/evm releases new versions
- **Extensible**: Easy to add more RPC configuration options in the future
- **Backward Compatible**: Works with existing EVM functionality

## Support

If you encounter any issues during integration, the key integration points are:

1. `rpc.SetupEVMRPC(app.BankKeeper)` - called after BankKeeper initialization
2. `go.mod` replace directive pointing to your fork
3. Ensure your BankKeeper implements the `cmn.BankKeeper` interface

The RPC system will automatically use your configured BankKeeper for all balance queries.