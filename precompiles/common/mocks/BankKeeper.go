// Code generated by mockery v2.53.4. DO NOT EDIT.

package mocks

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper is an autogenerated mock type for the BankKeeper type
type BankKeeper struct {
	mock.Mock
}

// GetBalance provides a mock function with given fields: ctx, addr, denom
func (_m *BankKeeper) GetBalance(ctx context.Context, addr types.AccAddress, denom string) types.Coin {
	ret := _m.Called(ctx, addr, denom)

	if len(ret) == 0 {
		panic("no return value specified for GetBalance")
	}

	var r0 types.Coin
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, string) types.Coin); ok {
		r0 = rf(ctx, addr, denom)
	} else {
		r0 = ret.Get(0).(types.Coin)
	}

	return r0
}

// GetDenomMetaData provides a mock function with given fields: ctx, denom
func (_m *BankKeeper) GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool) {
	ret := _m.Called(ctx, denom)

	if len(ret) == 0 {
		panic("no return value specified for GetDenomMetaData")
	}

	var r0 banktypes.Metadata
	var r1 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) (banktypes.Metadata, bool)); ok {
		return rf(ctx, denom)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) banktypes.Metadata); ok {
		r0 = rf(ctx, denom)
	} else {
		r0 = ret.Get(0).(banktypes.Metadata)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) bool); ok {
		r1 = rf(ctx, denom)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// GetSupply provides a mock function with given fields: ctx, denom
func (_m *BankKeeper) GetSupply(ctx context.Context, denom string) types.Coin {
	ret := _m.Called(ctx, denom)

	if len(ret) == 0 {
		panic("no return value specified for GetSupply")
	}

	var r0 types.Coin
	if rf, ok := ret.Get(0).(func(context.Context, string) types.Coin); ok {
		r0 = rf(ctx, denom)
	} else {
		r0 = ret.Get(0).(types.Coin)
	}

	return r0
}

// IterateAccountBalances provides a mock function with given fields: ctx, account, cb
func (_m *BankKeeper) IterateAccountBalances(ctx context.Context, account types.AccAddress, cb func(types.Coin) bool) {
	_m.Called(ctx, account, cb)
}

// IterateTotalSupply provides a mock function with given fields: ctx, cb
func (_m *BankKeeper) IterateTotalSupply(ctx context.Context, cb func(types.Coin) bool) {
	_m.Called(ctx, cb)
}

// SendCoins provides a mock function with given fields: ctx, fromAddr, toAddr, amt
func (_m *BankKeeper) SendCoins(ctx context.Context, fromAddr types.AccAddress, toAddr types.AccAddress, amt types.Coins) error {
	ret := _m.Called(ctx, fromAddr, toAddr, amt)

	if len(ret) == 0 {
		panic("no return value specified for SendCoins")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, types.AccAddress, types.Coins) error); ok {
		r0 = rf(ctx, fromAddr, toAddr, amt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDenomMetaData provides a mock function with given fields: ctx, denomMetaData
func (_m *BankKeeper) SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata) {
	_m.Called(ctx, denomMetaData)
}

// SpendableCoin provides a mock function with given fields: ctx, addr, denom
func (_m *BankKeeper) SpendableCoin(ctx context.Context, addr types.AccAddress, denom string) types.Coin {
	ret := _m.Called(ctx, addr, denom)

	if len(ret) == 0 {
		panic("no return value specified for SpendableCoin")
	}

	var r0 types.Coin
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, string) types.Coin); ok {
		r0 = rf(ctx, addr, denom)
	} else {
		r0 = ret.Get(0).(types.Coin)
	}

	return r0
}

// NewBankKeeper creates a new instance of BankKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBankKeeper(t interface {
	mock.TestingT
	Cleanup(func())
}) *BankKeeper {
	mock := &BankKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
