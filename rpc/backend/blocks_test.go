package backend

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"google.golang.org/grpc/metadata"

	"github.com/cometbft/cometbft/abci/types"
	cmtrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/evm/rpc/backend/mocks"
	ethrpc "github.com/cosmos/evm/rpc/types"
	utiltx "github.com/cosmos/evm/testutil/tx"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *BackendTestSuite) TestBlockNumber() {
	testCases := []struct {
		name           string
		registerMock   func()
		expBlockNumber hexutil.Uint64
		expPass        bool
	}{
		{
			"fail - invalid block header height",
			func() {
				var header metadata.MD
				height := int64(1)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsInvalidHeight(queryClient, &header, height)
			},
			0x0,
			false,
		},
		{
			"fail - invalid block header",
			func() {
				var header metadata.MD
				height := int64(1)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsInvalidHeader(queryClient, &header, height)
			},
			0x0,
			false,
		},
		{
			"pass - app state header height 1",
			func() {
				var header metadata.MD
				height := int64(1)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParams(queryClient, &header, height)
			},
			0x1,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			blockNumber, err := suite.backend.BlockNumber()

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expBlockNumber, blockNumber)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetBlockByNumber() {
	var (
		blockRes *cmtrpctypes.ResultBlockResults
		resBlock *cmtrpctypes.ResultBlock
	)
	msgEthereumTx, bz := suite.buildEthereumTx()

	testCases := []struct {
		name         string
		blockNumber  ethrpc.BlockNumber
		fullTx       bool
		baseFee      *big.Int
		validator    sdk.AccAddress
		tx           *evmtypes.MsgEthereumTx
		txBz         []byte
		registerMock func(ethrpc.BlockNumber, math.Int, sdk.AccAddress, []byte)
		expNoop      bool
		expPass      bool
	}{
		{
			"pass - tendermint block not found",
			ethrpc.BlockNumber(1),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(blockNum ethrpc.BlockNumber, _ math.Int, _ sdk.AccAddress, _ []byte) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, height)
			},
			true,
			true,
		},
		{
			"pass - block not found (e.g. request block height that is greater than current one)",
			ethrpc.BlockNumber(1),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(blockNum ethrpc.BlockNumber, _ math.Int, _ sdk.AccAddress, _ []byte) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlockNotFound(client, height)
			},
			true,
			true,
		},
		{
			"pass - block results error",
			ethrpc.BlockNumber(1),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(blockNum ethrpc.BlockNumber, _ math.Int, _ sdk.AccAddress, txBz []byte) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlock(client, height, txBz)
				RegisterBlockResultsError(client, blockNum.Int64())
			},
			true,
			true,
		},
		{
			"pass - without tx",
			ethrpc.BlockNumber(1),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(blockNum ethrpc.BlockNumber, baseFee math.Int, validator sdk.AccAddress, txBz []byte) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlock(client, height, txBz)
				blockRes, _ = RegisterBlockResults(client, blockNum.Int64())
				RegisterConsensusParams(client, height)

				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)
			},
			false,
			true,
		},
		{
			"pass - with tx",
			ethrpc.BlockNumber(1),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			msgEthereumTx,
			bz,
			func(blockNum ethrpc.BlockNumber, baseFee math.Int, validator sdk.AccAddress, txBz []byte) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlock(client, height, txBz)
				blockRes, _ = RegisterBlockResults(client, blockNum.Int64())
				RegisterConsensusParams(client, height)

				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)
			},
			false,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock(tc.blockNumber, math.NewIntFromBigInt(tc.baseFee), tc.validator, tc.txBz)

			block, err := suite.backend.GetBlockByNumber(tc.blockNumber, tc.fullTx)

			if tc.expPass {
				if tc.expNoop {
					suite.Require().Nil(block)
				} else {
					expBlock := suite.buildFormattedBlock(
						blockRes,
						resBlock,
						tc.fullTx,
						tc.tx,
						tc.validator,
						tc.baseFee,
					)
					suite.Require().Equal(expBlock, block)
				}
				suite.Require().NoError(err)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetBlockByHash() {
	var (
		blockRes *cmtrpctypes.ResultBlockResults
		resBlock *cmtrpctypes.ResultBlock
	)
	msgEthereumTx, bz := suite.buildEthereumTx()

	block := cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil)

	testCases := []struct {
		name         string
		hash         common.Hash
		fullTx       bool
		baseFee      *big.Int
		validator    sdk.AccAddress
		tx           *evmtypes.MsgEthereumTx
		txBz         []byte
		registerMock func(common.Hash, math.Int, sdk.AccAddress, []byte)
		expNoop      bool
		expPass      bool
	}{
		{
			"fail - tendermint failed to get block",
			common.BytesToHash(block.Hash()),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(hash common.Hash, _ math.Int, _ sdk.AccAddress, txBz []byte) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockByHashError(client, hash, txBz)
			},
			false,
			false,
		},
		{
			"noop - tendermint blockres not found",
			common.BytesToHash(block.Hash()),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(hash common.Hash, _ math.Int, _ sdk.AccAddress, txBz []byte) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockByHashNotFound(client, hash, txBz)
			},
			true,
			true,
		},
		{
			"noop - tendermint failed to fetch block result",
			common.BytesToHash(block.Hash()),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(hash common.Hash, _ math.Int, _ sdk.AccAddress, txBz []byte) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlockByHash(client, hash, txBz)

				RegisterBlockResultsError(client, height)
			},
			true,
			true,
		},
		{
			"pass - without tx",
			common.BytesToHash(block.Hash()),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			nil,
			nil,
			func(hash common.Hash, baseFee math.Int, validator sdk.AccAddress, txBz []byte) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlockByHash(client, hash, txBz)

				blockRes, _ = RegisterBlockResults(client, height)
				RegisterConsensusParams(client, height)

				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)
			},
			false,
			true,
		},
		{
			"pass - with tx",
			common.BytesToHash(block.Hash()),
			true,
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			msgEthereumTx,
			bz,
			func(hash common.Hash, baseFee math.Int, validator sdk.AccAddress, txBz []byte) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resBlock, _ = RegisterBlockByHash(client, hash, txBz)

				blockRes, _ = RegisterBlockResults(client, height)
				RegisterConsensusParams(client, height)

				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)
			},
			false,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock(tc.hash, math.NewIntFromBigInt(tc.baseFee), tc.validator, tc.txBz)

			block, err := suite.backend.GetBlockByHash(tc.hash, tc.fullTx)

			if tc.expPass {
				if tc.expNoop {
					suite.Require().Nil(block)
				} else {
					expBlock := suite.buildFormattedBlock(
						blockRes,
						resBlock,
						tc.fullTx,
						tc.tx,
						tc.validator,
						tc.baseFee,
					)
					suite.Require().Equal(expBlock, block)
				}
				suite.Require().NoError(err)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetBlockTransactionCountByHash() {
	_, bz := suite.buildEthereumTx()
	block := cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil)
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		hash         common.Hash
		registerMock func(common.Hash)
		expCount     hexutil.Uint
		expPass      bool
	}{
		{
			"fail - block not found",
			common.BytesToHash(emptyBlock.Hash()),
			func(hash common.Hash) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockByHashError(client, hash, nil)
			},
			hexutil.Uint(0),
			false,
		},
		{
			"fail - tendermint client failed to get block result",
			common.BytesToHash(emptyBlock.Hash()),
			func(hash common.Hash) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockByHash(client, hash, nil)
				suite.Require().NoError(err)
				RegisterBlockResultsError(client, height)
			},
			hexutil.Uint(0),
			false,
		},
		{
			"pass - block without tx",
			common.BytesToHash(emptyBlock.Hash()),
			func(hash common.Hash) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockByHash(client, hash, nil)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, height)
				suite.Require().NoError(err)
			},
			hexutil.Uint(0),
			true,
		},
		{
			"pass - block with tx",
			common.BytesToHash(block.Hash()),
			func(hash common.Hash) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockByHash(client, hash, bz)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, height)
				suite.Require().NoError(err)
			},
			hexutil.Uint(1),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			tc.registerMock(tc.hash)
			count := suite.backend.GetBlockTransactionCountByHash(tc.hash)
			if tc.expPass {
				suite.Require().Equal(tc.expCount, *count)
			} else {
				suite.Require().Nil(count)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetBlockTransactionCountByNumber() {
	_, bz := suite.buildEthereumTx()
	block := cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil)
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		blockNum     ethrpc.BlockNumber
		registerMock func(ethrpc.BlockNumber)
		expCount     hexutil.Uint
		expPass      bool
	}{
		{
			"fail - block not found",
			ethrpc.BlockNumber(emptyBlock.Height),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, height)
			},
			hexutil.Uint(0),
			false,
		},
		{
			"fail - tendermint client failed to get block result",
			ethrpc.BlockNumber(emptyBlock.Height),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, nil)
				suite.Require().NoError(err)
				RegisterBlockResultsError(client, height)
			},
			hexutil.Uint(0),
			false,
		},
		{
			"pass - block without tx",
			ethrpc.BlockNumber(emptyBlock.Height),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, nil)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, height)
				suite.Require().NoError(err)
			},
			hexutil.Uint(0),
			true,
		},
		{
			"pass - block with tx",
			ethrpc.BlockNumber(block.Height),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, bz)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, height)
				suite.Require().NoError(err)
			},
			hexutil.Uint(1),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			tc.registerMock(tc.blockNum)
			count := suite.backend.GetBlockTransactionCountByNumber(tc.blockNum)
			if tc.expPass {
				suite.Require().Equal(tc.expCount, *count)
			} else {
				suite.Require().Nil(count)
			}
		})
	}
}

func (suite *BackendTestSuite) TestTendermintBlockByNumber() {
	var expResultBlock *cmtrpctypes.ResultBlock

	testCases := []struct {
		name         string
		blockNumber  ethrpc.BlockNumber
		registerMock func(ethrpc.BlockNumber)
		found        bool
		expPass      bool
	}{
		{
			"fail - client error",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, height)
			},
			false,
			false,
		},
		{
			"noop - block not found",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockNotFound(client, height)
				suite.Require().NoError(err)
			},
			false,
			true,
		},
		{
			"fail - blockNum < 0 with app state height error",
			ethrpc.BlockNumber(-1),
			func(_ ethrpc.BlockNumber) {
				var header metadata.MD
				appHeight := int64(1)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsError(queryClient, &header, appHeight)
			},
			false,
			false,
		},
		{
			"pass - blockNum < 0 with app state height >= 1",
			ethrpc.BlockNumber(-1),
			func(ethrpc.BlockNumber) {
				var header metadata.MD
				appHeight := int64(1)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParams(queryClient, &header, appHeight)

				tmHeight := appHeight
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultBlock, _ = RegisterBlock(client, tmHeight, nil)
			},
			true,
			true,
		},
		{
			"pass - blockNum = 0 (defaults to blockNum = 1 due to a difference between tendermint heights and geth heights)",
			ethrpc.BlockNumber(0),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultBlock, _ = RegisterBlock(client, height, nil)
			},
			true,
			true,
		},
		{
			"pass - blockNum = 1",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultBlock, _ = RegisterBlock(client, height, nil)
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			tc.registerMock(tc.blockNumber)
			resultBlock, err := suite.backend.TendermintBlockByNumber(tc.blockNumber)

			if tc.expPass {
				suite.Require().NoError(err)

				if !tc.found {
					suite.Require().Nil(resultBlock)
				} else {
					suite.Require().Equal(expResultBlock, resultBlock)
					suite.Require().Equal(expResultBlock.Block.Height, resultBlock.Block.Height)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestTendermintBlockResultByNumber() {
	var expBlockRes *cmtrpctypes.ResultBlockResults

	testCases := []struct {
		name         string
		blockNumber  int64
		registerMock func(int64)
		expPass      bool
	}{
		{
			"fail",
			1,
			func(blockNum int64) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockResultsError(client, blockNum)
			},
			false,
		},
		{
			"pass",
			1,
			func(blockNum int64) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockResults(client, blockNum)
				suite.Require().NoError(err)
				expBlockRes = &cmtrpctypes.ResultBlockResults{
					Height:     blockNum,
					TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock(tc.blockNumber)

			client := suite.backend.clientCtx.Client.(*mocks.Client)
			blockRes, err := client.BlockResults(suite.backend.ctx, &tc.blockNumber) //#nosec G601 -- fine for tests

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expBlockRes, blockRes)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestBlockNumberFromTendermint() {
	var resHeader *cmtrpctypes.ResultHeader

	_, bz := suite.buildEthereumTx()
	block := cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil)
	blockNum := ethrpc.NewBlockNumber(big.NewInt(block.Height))
	blockHash := common.BytesToHash(block.Hash())

	testCases := []struct {
		name         string
		blockNum     *ethrpc.BlockNumber
		hash         *common.Hash
		registerMock func(*common.Hash)
		expPass      bool
	}{
		{
			"error - without blockHash or blockNum",
			nil,
			nil,
			func(*common.Hash) {},
			false,
		},
		{
			"error - with blockHash, tendermint client failed to get block",
			nil,
			&blockHash,
			func(hash *common.Hash) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterHeaderByHashError(client, *hash, bz)
			},
			false,
		},
		{
			"pass - with blockHash",
			nil,
			&blockHash,
			func(hash *common.Hash) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resHeader, _ = RegisterHeaderByHash(client, *hash, bz)
			},
			true,
		},
		{
			"pass - without blockHash & with blockNumber",
			&blockNum,
			nil,
			func(*common.Hash) {},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			blockNrOrHash := ethrpc.BlockNumberOrHash{
				BlockNumber: tc.blockNum,
				BlockHash:   tc.hash,
			}

			tc.registerMock(tc.hash)
			blockNum, err := suite.backend.BlockNumberFromTendermint(blockNrOrHash)

			if tc.expPass {
				suite.Require().NoError(err)
				if tc.hash == nil {
					suite.Require().Equal(*tc.blockNum, blockNum)
				} else {
					expHeight := ethrpc.NewBlockNumber(big.NewInt(resHeader.Header.Height))
					suite.Require().Equal(expHeight, blockNum)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestBlockNumberFromTendermintByHash() {
	var resHeader *cmtrpctypes.ResultHeader

	_, bz := suite.buildEthereumTx()
	block := cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil)
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		hash         common.Hash
		registerMock func(common.Hash)
		expPass      bool
	}{
		{
			"fail - tendermint client failed to get block",
			common.BytesToHash(block.Hash()),
			func(hash common.Hash) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterHeaderByHashError(client, hash, bz)
			},
			false,
		},
		{
			"pass - block without tx",
			common.BytesToHash(emptyBlock.Hash()),
			func(hash common.Hash) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resHeader, _ = RegisterHeaderByHash(client, hash, bz)
			},
			true,
		},
		{
			"pass - block with tx",
			common.BytesToHash(block.Hash()),
			func(hash common.Hash) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				resHeader, _ = RegisterHeaderByHash(client, hash, bz)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			tc.registerMock(tc.hash)
			blockNum, err := suite.backend.BlockNumberFromTendermintByHash(tc.hash)
			if tc.expPass {
				expHeight := big.NewInt(resHeader.Header.Height)
				suite.Require().NoError(err)
				suite.Require().Equal(expHeight, blockNum)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestBlockBloom() {
	testCases := []struct {
		name          string
		blockRes      *cmtrpctypes.ResultBlockResults
		expBlockBloom ethtypes.Bloom
		expPass       bool
	}{
		{
			"fail - empty block result",
			&cmtrpctypes.ResultBlockResults{},
			ethtypes.Bloom{},
			false,
		},
		{
			"fail - non block bloom event type",
			&cmtrpctypes.ResultBlockResults{
				FinalizeBlockEvents: []types.Event{{Type: evmtypes.EventTypeEthereumTx}},
			},
			ethtypes.Bloom{},
			false,
		},
		{
			"fail - nonblock bloom attribute key",
			&cmtrpctypes.ResultBlockResults{
				FinalizeBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
						Attributes: []types.EventAttribute{
							{Key: evmtypes.AttributeKeyEthereumTxHash},
						},
					},
				},
			},
			ethtypes.Bloom{},
			false,
		},
		{
			"pass - block bloom attribute key",
			&cmtrpctypes.ResultBlockResults{
				FinalizeBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
						Attributes: []types.EventAttribute{
							{Key: evmtypes.AttributeKeyEthereumBloom},
						},
					},
				},
			},
			ethtypes.Bloom{},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			blockBloom, err := suite.backend.BlockBloom(tc.blockRes)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expBlockBloom, blockBloom)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetEthBlockFromTendermint() {
	msgEthereumTx, bz := suite.buildEthereumTx()
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		baseFee      *big.Int
		validator    sdk.AccAddress
		height       int64
		resBlock     *cmtrpctypes.ResultBlock
		blockRes     *cmtrpctypes.ResultBlockResults
		fullTx       bool
		registerMock func(math.Int, sdk.AccAddress, int64)
		expTxs       bool
		expPass      bool
	}{
		{
			"pass - block without tx",
			math.NewInt(1).BigInt(),
			sdk.AccAddress(common.Address{}.Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{Block: emptyBlock},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			false,
			func(baseFee math.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			false,
			true,
		},
		{
			"pass - block with tx - with BaseFee error",
			nil,
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			true,
			func(_ math.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFeeError(queryClient)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			true,
			true,
		},
		{
			"pass - block with tx - with ValidatorAccount error",
			math.NewInt(1).BigInt(),
			sdk.AccAddress(common.Address{}.Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			true,
			func(baseFee math.Int, _ sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccountError(queryClient)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			true,
			true,
		},
		{
			"pass - block with tx - with ConsensusParams error - BlockMaxGas defaults to max uint32",
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			true,
			func(baseFee math.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParamsError(client, height)
			},
			true,
			true,
		},
		{
			"pass - block with tx - with ShouldIgnoreGasUsed - empty txs",
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height: 1,
				TxsResults: []*types.ExecTxResult{
					{
						Code:    11,
						GasUsed: 0,
						Log:     "no block gas left to run tx: out of gas",
					},
				},
			},
			true,
			func(baseFee math.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			false,
			true,
		},
		{
			"pass - block with tx - non fullTx",
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			false,
			func(baseFee math.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			true,
			true,
		},
		{
			"pass - block with tx",
			math.NewInt(1).BigInt(),
			sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			int64(1),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			true,
			func(baseFee math.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock(math.NewIntFromBigInt(tc.baseFee), tc.validator, tc.height)

			block, err := suite.backend.RPCBlockFromTendermintBlock(tc.resBlock, tc.blockRes, tc.fullTx)

			var expBlock map[string]interface{}
			header := tc.resBlock.Block.Header
			gasLimit := int64(^uint32(0))                                                // for `MaxGas = -1` (DefaultConsensusParams)
			gasUsed := new(big.Int).SetUint64(uint64(tc.blockRes.TxsResults[0].GasUsed)) //nolint:gosec // G115 // won't exceed uint64

			root := common.Hash{}.Bytes()
			receipt := ethtypes.NewReceipt(root, false, gasUsed.Uint64())
			bloom := ethtypes.CreateBloom(receipt)

			ethRPCTxs := []interface{}{}

			if tc.expTxs {
				if tc.fullTx {
					rpcTx, err := ethrpc.NewRPCTransaction(
						msgEthereumTx.AsTransaction(),
						common.BytesToHash(header.Hash()),
						uint64(header.Height), //nolint:gosec // G115 // won't exceed uint64
						uint64(0),
						tc.baseFee,
						suite.backend.chainID,
					)
					suite.Require().NoError(err)
					ethRPCTxs = []interface{}{rpcTx}
				} else {
					ethRPCTxs = []interface{}{common.HexToHash(msgEthereumTx.Hash)}
				}
			}

			expBlock = ethrpc.FormatBlock(
				header,
				tc.resBlock.Block.Size(),
				gasLimit,
				gasUsed,
				ethRPCTxs,
				bloom,
				common.BytesToAddress(tc.validator.Bytes()),
				tc.baseFee,
			)

			if tc.expPass {
				suite.Require().Equal(expBlock, block)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestEthMsgsFromTendermintBlock() {
	msgEthereumTx, bz := suite.buildEthereumTx()

	testCases := []struct {
		name     string
		resBlock *cmtrpctypes.ResultBlock
		blockRes *cmtrpctypes.ResultBlockResults
		expMsgs  []*evmtypes.MsgEthereumTx
	}{
		{
			"tx in not included in block - unsuccessful tx without ExceedBlockGasLimit error",
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				TxsResults: []*types.ExecTxResult{
					{
						Code: 1,
					},
				},
			},
			[]*evmtypes.MsgEthereumTx(nil),
		},
		{
			"tx included in block - unsuccessful tx with ExceedBlockGasLimit error",
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				TxsResults: []*types.ExecTxResult{
					{
						Code: 1,
						Log:  ethrpc.ExceedBlockGasLimitError,
					},
				},
			},
			[]*evmtypes.MsgEthereumTx{msgEthereumTx},
		},
		{
			"pass",
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				TxsResults: []*types.ExecTxResult{
					{
						Code: 0,
						Log:  ethrpc.ExceedBlockGasLimitError,
					},
				},
			},
			[]*evmtypes.MsgEthereumTx{msgEthereumTx},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			msgs := suite.backend.EthMsgsFromTendermintBlock(tc.resBlock, tc.blockRes)
			suite.Require().Equal(tc.expMsgs, msgs)
		})
	}
}

func (suite *BackendTestSuite) TestHeaderByNumber() {
	var expResultBlock *cmtrpctypes.ResultBlock

	_, bz := suite.buildEthereumTx()

	testCases := []struct {
		name         string
		blockNumber  ethrpc.BlockNumber
		baseFee      *big.Int
		registerMock func(ethrpc.BlockNumber, math.Int)
		expPass      bool
	}{
		{
			"fail - tendermint client failed to get block",
			ethrpc.BlockNumber(1),
			math.NewInt(1).BigInt(),
			func(blockNum ethrpc.BlockNumber, _ math.Int) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, height)
			},
			false,
		},
		{
			"fail - block not found for height",
			ethrpc.BlockNumber(1),
			math.NewInt(1).BigInt(),
			func(blockNum ethrpc.BlockNumber, _ math.Int) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockNotFound(client, height)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"fail - block not found for height",
			ethrpc.BlockNumber(1),
			math.NewInt(1).BigInt(),
			func(blockNum ethrpc.BlockNumber, _ math.Int) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, nil)
				suite.Require().NoError(err)
				RegisterBlockResultsError(client, height)
			},
			false,
		},
		{
			"pass - without Base Fee, failed to fetch from prunned block",
			ethrpc.BlockNumber(1),
			nil,
			func(blockNum ethrpc.BlockNumber, _ math.Int) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultBlock, _ = RegisterBlock(client, height, nil)
				_, err := RegisterBlockResults(client, height)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFeeError(queryClient)
			},
			true,
		},
		{
			"pass - blockNum = 1, without tx",
			ethrpc.BlockNumber(1),
			math.NewInt(1).BigInt(),
			func(blockNum ethrpc.BlockNumber, baseFee math.Int) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultBlock, _ = RegisterBlock(client, height, nil)
				_, err := RegisterBlockResults(client, height)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			true,
		},
		{
			"pass - blockNum = 1, with tx",
			ethrpc.BlockNumber(1),
			math.NewInt(1).BigInt(),
			func(blockNum ethrpc.BlockNumber, baseFee math.Int) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultBlock, _ = RegisterBlock(client, height, bz)
				_, err := RegisterBlockResults(client, height)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			tc.registerMock(tc.blockNumber, math.NewIntFromBigInt(tc.baseFee))
			header, err := suite.backend.HeaderByNumber(tc.blockNumber)

			if tc.expPass {
				expHeader := ethrpc.EthHeaderFromTendermint(expResultBlock.Block.Header, ethtypes.Bloom{}, tc.baseFee)
				suite.Require().NoError(err)
				suite.Require().Equal(expHeader, header)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestHeaderByHash() {
	var expResultHeader *cmtrpctypes.ResultHeader

	_, bz := suite.buildEthereumTx()
	block := cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil)
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		hash         common.Hash
		baseFee      *big.Int
		registerMock func(common.Hash, math.Int)
		expPass      bool
	}{
		{
			"fail - tendermint client failed to get block",
			common.BytesToHash(block.Hash()),
			math.NewInt(1).BigInt(),
			func(hash common.Hash, _ math.Int) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterHeaderByHashError(client, hash, bz)
			},
			false,
		},
		{
			"fail - block not found for height",
			common.BytesToHash(block.Hash()),
			math.NewInt(1).BigInt(),
			func(hash common.Hash, _ math.Int) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterHeaderByHashNotFound(client, hash, bz)
			},
			false,
		},
		{
			"fail - block not found for height",
			common.BytesToHash(block.Hash()),
			math.NewInt(1).BigInt(),
			func(hash common.Hash, _ math.Int) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterHeaderByHash(client, hash, bz)
				suite.Require().NoError(err)
				RegisterBlockResultsError(client, height)
			},
			false,
		},
		{
			"pass - without Base Fee, failed to fetch from prunned block",
			common.BytesToHash(block.Hash()),
			nil,
			func(hash common.Hash, _ math.Int) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultHeader, _ = RegisterHeaderByHash(client, hash, bz)
				_, err := RegisterBlockResults(client, height)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFeeError(queryClient)
			},
			true,
		},
		{
			"pass - blockNum = 1, without tx",
			common.BytesToHash(emptyBlock.Hash()),
			math.NewInt(1).BigInt(),
			func(hash common.Hash, baseFee math.Int) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultHeader, _ = RegisterHeaderByHash(client, hash, nil)
				_, err := RegisterBlockResults(client, height)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			true,
		},
		{
			"pass - with tx",
			common.BytesToHash(block.Hash()),
			math.NewInt(1).BigInt(),
			func(hash common.Hash, baseFee math.Int) {
				height := int64(1)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				expResultHeader, _ = RegisterHeaderByHash(client, hash, bz)
				_, err := RegisterBlockResults(client, height)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries

			tc.registerMock(tc.hash, math.NewIntFromBigInt(tc.baseFee))
			header, err := suite.backend.HeaderByHash(tc.hash)

			if tc.expPass {
				expHeader := ethrpc.EthHeaderFromTendermint(*expResultHeader.Header, ethtypes.Bloom{}, tc.baseFee)
				suite.Require().NoError(err)
				suite.Require().Equal(expHeader, header)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestEthBlockByNumber() {
	msgEthereumTx, bz := suite.buildEthereumTx()
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		blockNumber  ethrpc.BlockNumber
		registerMock func(ethrpc.BlockNumber)
		expEthBlock  *ethtypes.Block
		expPass      bool
	}{
		{
			"fail - tendermint client failed to get block",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, height)
			},
			nil,
			false,
		},
		{
			"fail - block result not found for height",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, nil)
				suite.Require().NoError(err)
				RegisterBlockResultsError(client, blockNum.Int64())
			},
			nil,
			false,
		},
		{
			"pass - block without tx",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, nil)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, blockNum.Int64())
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				baseFee := math.NewInt(1)
				RegisterBaseFee(queryClient, baseFee)
			},
			ethtypes.NewBlock(
				ethrpc.EthHeaderFromTendermint(
					emptyBlock.Header,
					ethtypes.Bloom{},
					math.NewInt(1).BigInt(),
				),
				&ethtypes.Body{},
				nil,
				nil,
			),
			true,
		},
		{
			"pass - block with tx",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, height, bz)
				suite.Require().NoError(err)
				_, err = RegisterBlockResults(client, blockNum.Int64())
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				baseFee := math.NewInt(1)
				RegisterBaseFee(queryClient, baseFee)
			},
			ethtypes.NewBlock(
				ethrpc.EthHeaderFromTendermint(
					emptyBlock.Header,
					ethtypes.Bloom{},
					math.NewInt(1).BigInt(),
				),
				&ethtypes.Body{
					Transactions: []*ethtypes.Transaction{msgEthereumTx.AsTransaction()},
				},
				nil,
				trie.NewStackTrie(nil),
			),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock(tc.blockNumber)

			ethBlock, err := suite.backend.EthBlockByNumber(tc.blockNumber)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expEthBlock.Header(), ethBlock.Header())
				suite.Require().Equal(tc.expEthBlock.Uncles(), ethBlock.Uncles())
				suite.Require().Equal(tc.expEthBlock.ReceiptHash(), ethBlock.ReceiptHash())
				for i, tx := range tc.expEthBlock.Transactions() {
					suite.Require().Equal(tx.Data(), ethBlock.Transactions()[i].Data())
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestEthBlockFromTendermintBlock() {
	msgEthereumTx, bz := suite.buildEthereumTx()
	emptyBlock := cmttypes.MakeBlock(1, []cmttypes.Tx{}, nil, nil)

	testCases := []struct {
		name         string
		baseFee      *big.Int
		resBlock     *cmtrpctypes.ResultBlock
		blockRes     *cmtrpctypes.ResultBlockResults
		registerMock func(math.Int, int64)
		expEthBlock  *ethtypes.Block
		expPass      bool
	}{
		{
			"pass - block without tx",
			math.NewInt(1).BigInt(),
			&cmtrpctypes.ResultBlock{
				Block: emptyBlock,
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
			},
			func(baseFee math.Int, _ int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			ethtypes.NewBlock(
				ethrpc.EthHeaderFromTendermint(
					emptyBlock.Header,
					ethtypes.Bloom{},
					math.NewInt(1).BigInt(),
				),
				&ethtypes.Body{},
				nil,
				nil,
			),
			true,
		},
		{
			"pass - block with tx",
			math.NewInt(1).BigInt(),
			&cmtrpctypes.ResultBlock{
				Block: cmttypes.MakeBlock(1, []cmttypes.Tx{bz}, nil, nil),
			},
			&cmtrpctypes.ResultBlockResults{
				Height:     1,
				TxsResults: []*types.ExecTxResult{{Code: 0, GasUsed: 0}},
				FinalizeBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
						Attributes: []types.EventAttribute{
							{Key: evmtypes.AttributeKeyEthereumBloom},
						},
					},
				},
			},
			func(baseFee math.Int, _ int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			ethtypes.NewBlock(
				ethrpc.EthHeaderFromTendermint(
					emptyBlock.Header,
					ethtypes.Bloom{},
					math.NewInt(1).BigInt(),
				),
				&ethtypes.Body{Transactions: []*ethtypes.Transaction{msgEthereumTx.AsTransaction()}},
				nil,
				trie.NewStackTrie(nil),
			),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock(math.NewIntFromBigInt(tc.baseFee), tc.blockRes.Height)

			ethBlock, err := suite.backend.EthBlockFromTendermintBlock(tc.resBlock, tc.blockRes)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expEthBlock.Header(), ethBlock.Header())
				suite.Require().Equal(tc.expEthBlock.Uncles(), ethBlock.Uncles())
				suite.Require().Equal(tc.expEthBlock.ReceiptHash(), ethBlock.ReceiptHash())
				for i, tx := range tc.expEthBlock.Transactions() {
					suite.Require().Equal(tx.Data(), ethBlock.Transactions()[i].Data())
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
