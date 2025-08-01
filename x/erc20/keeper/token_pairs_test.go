package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	testconstants "github.com/cosmos/evm/testutil/constants"
	utiltx "github.com/cosmos/evm/testutil/tx"
	"github.com/cosmos/evm/x/erc20/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGetTokenPairs() {
	var (
		ctx    sdk.Context
		expRes []types.TokenPair
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no pair registered", func() { expRes = testconstants.ExampleTokenPairs },
		},
		{
			"1 pair registered",
			func() {
				pair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)
				expRes = testconstants.ExampleTokenPairs
				expRes = append(expRes, pair)
			},
		},
		{
			"2 pairs registered",
			func() {
				pair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
				pair2 := types.NewTokenPair(utiltx.GenerateAddress(), "coin2", types.OWNER_MODULE)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair2)
				expRes = testconstants.ExampleTokenPairs
				expRes = append(expRes, []types.TokenPair{pair, pair2}...)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx = suite.network.GetContext()

			tc.malleate()
			res := suite.network.App.Erc20Keeper.GetTokenPairs(ctx)

			suite.Require().ElementsMatch(expRes, res, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestGetTokenPairID() {
	baseDenom, err := sdk.GetBaseDenom()
	suite.Require().NoError(err, "failed to get base denom")

	pair := types.NewTokenPair(utiltx.GenerateAddress(), baseDenom, types.OWNER_MODULE)

	testCases := []struct {
		name  string
		token string
		expID []byte
	}{
		{"nil token", "", nil},
		{"valid hex token", utiltx.GenerateAddress().Hex(), []byte{}},
		{"valid hex token", utiltx.GenerateAddress().String(), []byte{}},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx := suite.network.GetContext()

		suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)

		id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, tc.token)
		if id != nil {
			suite.Require().Equal(tc.expID, id, tc.name)
		} else {
			suite.Require().Nil(id)
		}
	}
}

func (suite *KeeperTestSuite) TestGetTokenPair() {
	baseDenom, err := sdk.GetBaseDenom()
	suite.Require().NoError(err, "failed to get base denom")

	pair := types.NewTokenPair(utiltx.GenerateAddress(), baseDenom, types.OWNER_MODULE)

	testCases := []struct {
		name string
		id   []byte
		ok   bool
	}{
		{"nil id", nil, false},
		{"valid id", pair.GetID(), true},
		{"pair not found", []byte{}, false},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx := suite.network.GetContext()

		suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)
		p, found := suite.network.App.Erc20Keeper.GetTokenPair(ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(pair, p, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteTokenPair() {
	tokenDenom := "random"

	var ctx sdk.Context
	pair := types.NewTokenPair(utiltx.GenerateAddress(), tokenDenom, types.OWNER_MODULE)
	id := pair.GetID()

	testCases := []struct {
		name     string
		id       []byte
		malleate func()
		ok       bool
	}{
		{"nil id", nil, func() {}, false},
		{"pair not found", []byte{}, func() {}, false},
		{"valid id", id, func() {}, true},
		{
			"delete tokenpair",
			id,
			func() {
				suite.network.App.Erc20Keeper.DeleteTokenPair(ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx = suite.network.GetContext()
		err := suite.network.App.Erc20Keeper.SetToken(ctx, pair)
		suite.Require().NoError(err)

		tc.malleate()
		p, found := suite.network.App.Erc20Keeper.GetTokenPair(ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(pair, p, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsTokenPairRegistered() {
	baseDenom, err := sdk.GetBaseDenom()
	suite.Require().NoError(err, "failed to get base denom")

	var ctx sdk.Context
	pair := types.NewTokenPair(utiltx.GenerateAddress(), baseDenom, types.OWNER_MODULE)

	testCases := []struct {
		name string
		id   []byte
		ok   bool
	}{
		{"valid id", pair.GetID(), true},
		{"pair not found", []byte{}, false},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx = suite.network.GetContext()

		suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)
		found := suite.network.App.Erc20Keeper.IsTokenPairRegistered(ctx, tc.id)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsERC20Registered() {
	var ctx sdk.Context
	addr := utiltx.GenerateAddress()
	pair := types.NewTokenPair(addr, "coin", types.OWNER_MODULE)

	testCases := []struct {
		name     string
		erc20    common.Address
		malleate func()
		ok       bool
	}{
		{"nil erc20 address", common.Address{}, func() {}, false},
		{"valid erc20 address", pair.GetERC20Contract(), func() {}, true},
		{
			"deleted erc20 map",
			pair.GetERC20Contract(),
			func() {
				suite.network.App.Erc20Keeper.DeleteTokenPair(ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx = suite.network.GetContext()

		err := suite.network.App.Erc20Keeper.SetToken(ctx, pair)
		suite.Require().NoError(err)

		tc.malleate()

		found := suite.network.App.Erc20Keeper.IsERC20Registered(ctx, tc.erc20)

		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsDenomRegistered() {
	var ctx sdk.Context
	addr := utiltx.GenerateAddress()
	pair := types.NewTokenPair(addr, "coin", types.OWNER_MODULE)

	testCases := []struct {
		name     string
		denom    string
		malleate func()
		ok       bool
	}{
		{"empty denom", "", func() {}, false},
		{"valid denom", pair.GetDenom(), func() {}, true},
		{
			"deleted denom map",
			pair.GetDenom(),
			func() {
				suite.network.App.Erc20Keeper.DeleteTokenPair(ctx, pair)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx = suite.network.GetContext()

		err := suite.network.App.Erc20Keeper.SetToken(ctx, pair)
		suite.Require().NoError(err)

		tc.malleate()

		found := suite.network.App.Erc20Keeper.IsDenomRegistered(ctx, tc.denom)

		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestGetTokenDenom() {
	var ctx sdk.Context
	tokenAddress := utiltx.GenerateAddress()
	tokenDenom := "token"

	testCases := []struct {
		name        string
		tokenDenom  string
		malleate    func()
		expError    bool
		errContains string
	}{
		{
			"denom found",
			tokenDenom,
			func() {
				pair := types.NewTokenPair(tokenAddress, tokenDenom, types.OWNER_MODULE)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, tokenAddress, pair.GetID())
			},
			true,
			"",
		},
		{
			"denom not found",
			tokenDenom,
			func() {
				address := utiltx.GenerateAddress()
				pair := types.NewTokenPair(address, tokenDenom, types.OWNER_MODULE)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, pair)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, address, pair.GetID())
			},
			false,
			fmt.Sprintf("token '%s' not registered", tokenAddress),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			ctx = suite.network.GetContext()

			tc.malleate()
			res, err := suite.network.App.Erc20Keeper.GetTokenDenom(ctx, tokenAddress)

			if tc.expError {
				suite.Require().NoError(err)
				suite.Require().Equal(res, tokenDenom)
			} else {
				suite.Require().Error(err, "expected an error while getting the token denom")
				suite.Require().ErrorContains(err, tc.errContains)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSetToken() {
	testCases := []struct {
		name     string
		pair1    types.TokenPair
		pair2    types.TokenPair
		expError bool
	}{
		{"same denom", types.NewTokenPair(common.HexToAddress("0x1"), "denom1", types.OWNER_MODULE), types.NewTokenPair(common.HexToAddress("0x2"), "denom1", types.OWNER_MODULE), true},
		{"same erc20", types.NewTokenPair(common.HexToAddress("0x1"), "denom1", types.OWNER_MODULE), types.NewTokenPair(common.HexToAddress("0x1"), "denom2", types.OWNER_MODULE), true},
		{"same pair", types.NewTokenPair(common.HexToAddress("0x1"), "denom1", types.OWNER_MODULE), types.NewTokenPair(common.HexToAddress("0x1"), "denom1", types.OWNER_MODULE), true},
		{"two different pairs", types.NewTokenPair(common.HexToAddress("0x1"), "denom1", types.OWNER_MODULE), types.NewTokenPair(common.HexToAddress("0x2"), "denom2", types.OWNER_MODULE), false},
	}
	for _, tc := range testCases {
		suite.SetupTest()
		ctx := suite.network.GetContext()

		err := suite.network.App.Erc20Keeper.SetToken(ctx, tc.pair1)
		suite.Require().NoError(err)
		err = suite.network.App.Erc20Keeper.SetToken(ctx, tc.pair2)
		if tc.expError {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}
	}
}
