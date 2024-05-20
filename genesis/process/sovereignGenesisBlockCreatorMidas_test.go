//go:build !race

package process

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/multiversx/mx-chain-go/vm/factory"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	nodeMock "github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	stateAcc "github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func createSovereignGenesisBlockCreatorMidas(t *testing.T) (ArgsGenesisBlockCreator, *sovereignGenesisBlockCreatorMidas) {
	arg := createSovereignMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}

	trieStorageManagers := createTrieStorageManagers()
	arg.Accounts, _ = createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		arg.RunTypeComponents.AccountsCreator(),
		trieStorageManagers[dataRetriever.UserAccountsUnit.String()],
		&testscommon.PubkeyConverterMock{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreatorMidas(gbc)
	return arg, sgbc
}

func TestNewSovereignGenesisBlockCreatorMidas(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		gbc := createGenesisBlockCreator(t)
		sgbc, err := NewSovereignGenesisBlockCreatorMidas(gbc)
		require.Nil(t, err)
		require.NotNil(t, sgbc)
	})

	t.Run("nil genesis block creator, should return error", func(t *testing.T) {
		t.Parallel()

		sgbc, err := NewSovereignGenesisBlockCreatorMidas(nil)
		require.Equal(t, errNilGenesisBlockCreator, err)
		require.Nil(t, sgbc)
	})
}

func TestSovereignGenesisBlockCreatorMidas_CreateGenesisBlocksEmptyBlocks(t *testing.T) {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.StartEpochNum = 1
	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreatorMidas(gbc)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Equal(t, map[uint32]data.HeaderHandler{
		core.SovereignChainShardId: &block.SovereignChainHeader{
			Header: &block.Header{
				ShardID: core.SovereignChainShardId,
			},
		},
	}, blocks)
}

func TestSovereignGenesisBlockCreatorMidas_CreateGenesisBaseProcess(t *testing.T) {
	args, sgbc := createSovereignGenesisBlockCreatorMidas(t)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Len(t, blocks, 1)
	require.Contains(t, blocks, core.SovereignChainShardId)

	indexingData := sgbc.GetIndexingData()
	require.Len(t, indexingData, 1)

	numDNSTypeScTxs := 2 * 256 // there are 2 contracts in testdata/smartcontracts.json
	numDefaultTypeScTxs := 1
	reqNumDeployInitialScTxs := numDNSTypeScTxs + numDefaultTypeScTxs

	sovereignIdxData := indexingData[core.SovereignChainShardId]
	require.Len(t, sovereignIdxData.DeployInitialScTxs, reqNumDeployInitialScTxs)
	require.Len(t, sovereignIdxData.DeploySystemScTxs, 4)
	require.Len(t, sovereignIdxData.DelegationTxs, 3)
	require.Len(t, sovereignIdxData.StakingTxs, 0)
	require.Greater(t, len(sovereignIdxData.ScrsTxs), 3)

	addr1 := "a00102030405060708090001020304050607080900010203040506070809000a"
	addr2 := "b00102030405060708090001020304050607080900010203040506070809000b"
	addr3 := "c00102030405060708090001020304050607080900010203040506070809000c"

	accountsDB := args.Accounts
	acc1 := getAccount(t, accountsDB, addr1)
	acc2 := getAccount(t, accountsDB, addr2)
	acc3 := getAccount(t, accountsDB, addr3)

	balance1 := big.NewInt(5000)
	balance2 := big.NewInt(2000)
	balance3 := big.NewInt(0)
	requireTokenExists(t, acc1, []byte(sovereignNativeToken), balance1, args.Core.InternalMarshalizer())
	requireTokenExists(t, acc2, []byte(sovereignNativeToken), balance2, args.Core.InternalMarshalizer())
	requireTokenExists(t, acc3, []byte(sovereignNativeToken), balance3, args.Core.InternalMarshalizer())
}

func TestSovereignGenesisBlockCreatorMidas_initGenesisAccounts(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	t.Run("cannot load system account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreatorMidas(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(container, core.SystemAccountAddress) {
					return nil, localErr
				}
				return nil, nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("cannot save system account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreatorMidas(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				if bytes.Equal(account.AddressBytes(), core.SystemAccountAddress) {
					return localErr
				}
				return nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("cannot load esdt sc account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreatorMidas(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				if bytes.Equal(container, core.ESDTSCAddress) {
					return nil, localErr
				}
				return &mock.UserAccountMock{}, nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("cannot save esdt sc account, should return error", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreatorMidas(t)
		sgbc.arg.Accounts = &state.AccountsStub{
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				if bytes.Equal(account.AddressBytes(), core.ESDTSCAddress) {
					return localErr
				}
				return nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Equal(t, localErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		_, sgbc := createSovereignGenesisBlockCreatorMidas(t)
		contractsToUpdate := map[string]struct{}{
			string(vm.StakingSCAddress):           {},
			string(vm.ValidatorSCAddress):         {},
			string(vm.GovernanceSCAddress):        {},
			string(vm.ESDTSCAddress):              {},
			string(vm.DelegationManagerSCAddress): {},
			string(vm.FirstDelegationSCAddress):   {},
		}

		expectedCodeMetaData := &vmcommon.CodeMetadata{
			Readable: true,
		}

		sgbc.arg.Accounts = &state.AccountsStub{
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				if bytes.Equal(account.AddressBytes(), core.SystemAccountAddress) {
					return nil
				}

				addrStr := string(account.AddressBytes())
				_, found := contractsToUpdate[addrStr]
				require.True(t, found)

				userAcc := account.(*state.AccountWrapMock)
				require.NotEmpty(t, userAcc.GetCode())
				require.NotEmpty(t, userAcc.GetOwnerAddress())
				require.Equal(t, expectedCodeMetaData.ToBytes(), userAcc.GetCodeMetadata())

				delete(contractsToUpdate, addrStr)

				return nil
			},
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return &state.AccountWrapMock{
					Address: container,
				}, nil
			},
		}

		err := sgbc.initGenesisAccounts()
		require.Nil(t, err)
		require.Empty(t, contractsToUpdate)
	})
}

func TestSovereignGenesisBlockCreatorMidas_setSovereignStakedDataMidas(t *testing.T) {
	t.Parallel()

	args := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))

	acc := &state.AccountWrapMock{
		Balance: big.NewInt(1),
	}
	acc.IncreaseNonce(4)
	args.Accounts = &state.AccountsStub{
		LoadAccountCalled: func(addr []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
	}
	initialNode := &sharding.InitialNode{
		Address: "addr",
	}
	nodesSpliter := &mock.NodesListSplitterStub{
		GetAllNodesCalled: func() []nodesCoordinator.GenesisNodeInfoHandler {
			return []nodesCoordinator.GenesisNodeInfoHandler{initialNode}
		},
	}
	expectedTx := &transaction.Transaction{
		Nonce:     acc.GetNonce(),
		Value:     big.NewInt(0),
		RcvAddr:   vm.ValidatorSCAddress,
		SndAddr:   factory.AbstractStakingSCAddress,
		GasPrice:  0,
		GasLimit:  math.MaxUint64,
		Data:      []byte("stake@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(initialNode.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis")) + "@" + hex.EncodeToString(initialNode.AddressBytes()) + "@" + hex.EncodeToString(new(big.Int).Set(args.GenesisNodePrice).Bytes())),
		Signature: nil,
	}
	processors := &genesisProcessors{
		txProcessor: &testscommon.TxProcessorStub{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				require.Equal(t, expectedTx, transaction)

				return vmcommon.Ok, nil
			},
		},
		queryService: &nodeMock.SCQueryServiceStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
				require.Equal(t, &process.SCQuery{
					ScAddress: vm.StakingSCAddress,
					FuncName:  "isStaked",
					Arguments: [][]byte{initialNode.PubKeyBytes()}}, query)

				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, holders.NewBlockInfo(nil, 0, nil), nil
			},
		},
	}

	txs, err := setSovereignStakedDataMidas(args, processors, nodesSpliter)
	require.Nil(t, err)
	require.Equal(t, []data.TransactionHandler{expectedTx}, txs)
}

func TestSovereignGenesisBlockCreatorMidas_InitSystemAccountCalled(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}

	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreatorMidas(gbc)
	require.NotNil(t, sgbc)

	acc, err := arg.Accounts.GetExistingAccount(core.SystemAccountAddress)
	require.Nil(t, acc)
	require.Equal(t, err, stateAcc.ErrAccNotFound)

	_, err = sgbc.CreateGenesisBlocks()
	require.Nil(t, err)

	acc, err = arg.Accounts.GetExistingAccount(core.SystemAccountAddress)
	require.NotNil(t, acc)
	require.Nil(t, err)
}
