package processing_test

import (
	"context"
	"errors"
	"github.com/multiversx/mx-chain-core-go/core"
	commonFactory "github.com/multiversx/mx-chain-go/common/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	mockConsensus "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/mock"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/genesis"
	genesisMocks "github.com/multiversx/mx-chain-go/genesis/mock"
	mockCoreComp "github.com/multiversx/mx-chain-go/integrationTests/mock"
	testsMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	factoryMocks "github.com/multiversx/mx-chain-go/testscommon/factory"
	nodesSetupMock "github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	testState "github.com/multiversx/mx-chain-go/testscommon/state"
)

func TestNewProcessComponentsFactoryMidas(t *testing.T) {
	t.Parallel()

	t.Run("nil GasSchedule should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.GasSchedule = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilGasSchedule))
		require.Nil(t, pcf)
	})
	t.Run("nil Data should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilDataComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockChain should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = &testsMocks.DataComponentsStub{
			BlockChain: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockChainHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil DataPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = &testsMocks.DataComponentsStub{
			BlockChain: &testscommon.ChainHandlerStub{},
			DataPool:   nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilDataPoolsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil StorageService should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Data = &testsMocks.DataComponentsStub{
			BlockChain: &testscommon.ChainHandlerStub{},
			DataPool:   &dataRetriever.PoolsHolderStub{},
			Store:      nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStorageService))
		require.Nil(t, pcf)
	})
	t.Run("nil CoreData should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilCoreComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil EconomicsData should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilEconomicsData))
		require.Nil(t, pcf)
	})
	t.Run("nil GenesisNodesSetup should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler: &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:      nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilGenesisNodesSetupHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil AddressPubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler: &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:      &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:   nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilAddressPublicKeyConverter))
		require.Nil(t, pcf)
	})
	t.Run("nil EpochNotifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilEpochNotifier))
		require.Nil(t, pcf)
	})
	t.Run("nil ValidatorPubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilPubKeyConverter))
		require.Nil(t, pcf)
	})
	t.Run("nil InternalMarshalizer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       &testscommon.PubkeyConverterStub{},
			IntMarsh:            nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilInternalMarshalizer))
		require.Nil(t, pcf)
	})
	t.Run("nil Uint64ByteSliceConverter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.CoreData = &mock.CoreComponentsMock{
			EconomicsHandler:    &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:         &nodesSetupMock.NodesSetupStub{},
			AddrPubKeyConv:      &testscommon.PubkeyConverterStub{},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
			ValPubKeyConv:       &testscommon.PubkeyConverterStub{},
			IntMarsh:            &marshallerMock.MarshalizerStub{},
			UInt64ByteSliceConv: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilUint64ByteSliceConverter))
		require.Nil(t, pcf)
	})
	t.Run("nil Crypto should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Crypto = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilCryptoComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockSignKeyGen should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Crypto = &testsMocks.CryptoComponentsStub{
			BlKeyGen: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockSignKeyGen))
		require.Nil(t, pcf)
	})
	t.Run("nil State should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.State = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStateComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil AccountsAdapter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.State = &factoryMocks.StateComponentsMock{
			Accounts: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilAccountsAdapter))
		require.Nil(t, pcf)
	})
	t.Run("nil Network should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Network = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilNetworkComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil NetworkMessenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Network = &testsMocks.NetworkComponentsStub{
			Messenger: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilMessenger))
		require.Nil(t, pcf)
	})
	t.Run("nil InputAntiFloodHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Network = &testsMocks.NetworkComponentsStub{
			Messenger:      &p2pmocks.MessengerStub{},
			InputAntiFlood: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilInputAntiFloodHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil SystemSCConfig should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.SystemSCConfig = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilSystemSCConfig))
		require.Nil(t, pcf)
	})
	t.Run("nil BootstrapComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.BootstrapComponents = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBootstrapComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilShardCoordinator))
		require.Nil(t, pcf)
	})
	t.Run("nil EpochBootstrapParams should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:   &testscommon.ShardsCoordinatorMock{},
			BootstrapParams: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBootstrapParamsHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil StatusComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.StatusComponents = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStatusComponentsHolder))
		require.Nil(t, pcf)
	})
	t.Run("nil OutportHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.StatusComponents = &testsMocks.StatusComponentsStub{
			Outport: nil,
		}
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilOutportHandler))
		require.Nil(t, pcf)
	})
	t.Run("nil HistoryRepo should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.HistoryRepo = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilHistoryRepository))
		require.Nil(t, pcf)
	})
	t.Run("nil StatusCoreComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.StatusCoreComponents = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilStatusCoreComponents))
		require.Nil(t, pcf)
	})
	t.Run("nil RunTypeComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.RunTypeComponents = nil
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilRunTypeComponents))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.BlockProcessorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockProcessorCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil RequestHandlerCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.RequestHandlerFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilRequestHandlerCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil ScheduledTxsExecutionCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.ScheduledTxsExecutionFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilScheduledTxsExecutionCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockTrackerCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.BlockTrackerFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockTrackerCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil TransactionCoordinatorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.TransactionCoordinatorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilTransactionCoordinatorCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil HeaderValidatorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.HeaderValidatorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilHeaderValidatorCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil ForkDetectorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.ForkDetectorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilForkDetectorCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil ValidatorStatisticsProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.ValidatorStatisticsProcessorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilValidatorStatisticsProcessorCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil SCProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.SCProcessorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilSCProcessorCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil BlockChainHookHandlerCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.BlockChainHookHandlerFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBlockChainHookHandlerCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil BootstrapperFromStorageCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.BootstrapperFromStorageFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBootstrapperFromStorageCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil BootstrapperCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.BootstrapperFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilBootstrapperCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil EpochStartBootstrapperCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.EpochStartBootstrapperFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilEpochStartBootstrapperCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil AdditionalStorageServiceCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.AdditionalStorageServiceFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilAdditionalStorageServiceCreator))
		require.Nil(t, pcf)
	})
	t.Run("nil SmartContractResultPreProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.SCResultsPreProcessorFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilSCResultsPreProcessorCreator))
		require.Nil(t, pcf)
	})
	t.Run("invalid ConsensusModel should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.ConsensusModelType = consensus.ConsensusModelInvalid
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrInvalidConsensusModel))
		require.Nil(t, pcf)
	})
	t.Run("nil VmContainerMetaCreator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.VmContainerMetaFactory = nil
		args.RunTypeComponents = rtMock
		pcf, err := processComp.NewProcessComponentsFactoryMidas(args)
		require.True(t, errors.Is(err, errorsMx.ErrNilVmContainerMetaFactoryCreator))
		require.Nil(t, pcf)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pcf, err := processComp.NewProcessComponentsFactoryMidas(createMockProcessComponentsFactoryArgs())
		require.NoError(t, err)
		require.NotNil(t, pcf)
	})
}

func TestProcessComponentsFactoryMidas_Create(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("CreateCurrentEpochProvider fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.EpochStartConfig.RoundsPerEpoch = 0
		args.PrefConfigs.Preferences.FullArchive = true
		testCreateWithArgsMidas(t, args, "rounds per epoch")
	})
	t.Run("createNetworkShardingCollector fails due to invalid PublicKeyPeerId config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PublicKeyPeerId.Type = "invalid"
		testCreateWithArgsMidas(t, args, "cache type")
	})
	t.Run("createNetworkShardingCollector fails due to invalid PublicKeyShardId config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PublicKeyShardId.Type = "invalid"
		testCreateWithArgsMidas(t, args, "cache type")
	})
	t.Run("createNetworkShardingCollector fails due to invalid PeerIdShardId config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PeerIdShardId.Type = "invalid"
		testCreateWithArgsMidas(t, args, "cache type")
	})
	t.Run("prepareNetworkShardingCollector fails due to SetPeerShardResolver failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		netwCompStub, ok := args.Network.(*testsMocks.NetworkComponentsStub)
		require.True(t, ok)
		netwCompStub.Messenger = &p2pmocks.MessengerStub{
			SetPeerShardResolverCalled: func(peerShardResolver p2p.PeerShardResolver) error {
				return expectedErr
			},
		}
		testCreateWithArgsMidas(t, args, expectedErr.Error())
	})
	t.Run("prepareNetworkShardingCollector fails due to SetPeerValidatorMapper failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		netwCompStub, ok := args.Network.(*testsMocks.NetworkComponentsStub)
		require.True(t, ok)
		netwCompStub.InputAntiFlood = &testsMocks.P2PAntifloodHandlerStub{
			SetPeerValidatorMapperCalled: func(validatorMapper process.PeerValidatorMapper) error {
				return expectedErr
			},
		}
		testCreateWithArgsMidas(t, args, expectedErr.Error())
	})
	t.Run("newStorageRequester fails due to NewStorageServiceFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true
		args.Config.StoragePruning.NumActivePersisters = 0
		testCreateWithArgsMidas(t, args, "active persisters")
	})
	t.Run("newResolverContainerFactory fails due to NewPeerAuthenticationPayloadValidator failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.HeartbeatV2.HeartbeatExpiryTimespanInSec = 0
		testCreateWithArgsMidas(t, args, "expiry timespan")
	})
	t.Run("generateGenesisHeadersAndApplyInitialBalances fails due to invalid GenesisNodePrice should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.LogsAndEvents.SaveInStorageEnabled = false // coverage
		args.Config.DbLookupExtensions.Enabled = true          // coverage
		args.SystemSCConfig.StakingSystemSCConfig.GenesisNodePrice = "invalid"
		testCreateWithArgsMidas(t, args, "invalid genesis node price")
	})
	t.Run("newValidatorStatisticsProcessor fails due to nil genesis header should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.ImportDBConfig.IsImportDBMode = true // coverage
		dataCompStub, ok := args.Data.(*testsMocks.DataComponentsStub)
		require.True(t, ok)
		blockChainStub, ok := dataCompStub.BlockChain.(*testscommon.ChainHandlerStub)
		require.True(t, ok)
		blockChainStub.GetGenesisHeaderCalled = func() coreData.HeaderHandler {
			return nil
		}
		testCreateWithArgsMidas(t, args, errorsMx.ErrGenesisBlockNotInitialized.Error())
	})
	t.Run("indexGenesisBlocks fails due to GenerateInitialTransactions failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		rtMock := getRunTypeComponentsMock()
		rtMock.AccountParser = &mock.AccountsParserStub{
			GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*dataBlock.MiniBlock, map[uint32]*outportCore.TransactionPool, error) {
				return nil, nil, expectedErr
			},
		}
		args.RunTypeComponents = rtMock
		testCreateWithArgsMidas(t, args, expectedErr.Error())
	})
	t.Run("NewMiniBlocksPoolsCleaner fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedMiniBlocks = 0
		testCreateWithArgsMidas(t, args, "MaxRoundsToKeepUnprocessedData")
	})
	t.Run("NewTxsPoolsCleaner fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedTransactions = 0
		testCreateWithArgsMidas(t, args, "MaxRoundsToKeepUnprocessedData")
	})
	t.Run("createHardforkTrigger fails due to Decode failure should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.Hardfork.PublicKeyToListenFrom = "invalid key"
		testCreateWithArgsMidas(t, args, "PublicKeyToListenFrom")
	})
	t.Run("NewCache fails for vmOutput should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.VMOutputCacher.Type = "invalid"
		testCreateWithArgsMidas(t, args, "cache type")
	})
	t.Run("newShardBlockProcessor: attachProcessDebugger fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.Debug.Process.Enabled = true
		args.Config.Debug.Process.PollingTimeInSeconds = 0
		testCreateWithArgsMidas(t, args, "PollingTimeInSeconds")
	})
	t.Run("nodesSetupChecker.Check fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		coreCompStub.GenesisNodesSetupCalled = func() sharding.GenesisNodesSetupHandler {
			return &nodesSetupMock.NodesSetupStub{
				AllInitialNodesCalled: func() []nodesCoordinator.GenesisNodeInfoHandler {
					return []nodesCoordinator.GenesisNodeInfoHandler{
						&genesisMocks.GenesisNodeInfoHandlerMock{
							PubKeyBytesValue: []byte("no stake"),
						},
					}
				},
				GetShardConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
				GetMetaConsensusGroupSizeCalled: func() uint32 {
					return 2
				},
			}
		}
		args.CoreData = coreCompStub
		testCreateWithArgsMidas(t, args, "no one staked")
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to RootHash failure", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		stateCompMock := factoryMocks.NewStateComponentsMockFromRealComponent(args.State)
		realAccounts := stateCompMock.AccountsAdapter()
		stateCompMock.Accounts = &testState.AccountsStub{
			GetAllLeavesCalled: realAccounts.GetAllLeaves,
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
			CommitCalled: realAccounts.Commit,
		}
		args.State = stateCompMock

		pcf, _ := processComp.NewProcessComponentsFactoryMidas(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to GetAllLeaves failure", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		stateCompMock := factoryMocks.NewStateComponentsMockFromRealComponent(args.State)
		realAccounts := stateCompMock.AccountsAdapter()
		stateCompMock.Accounts = &testState.AccountsStub{
			GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
				close(leavesChannels.LeavesChan)
				leavesChannels.ErrChan.Close()
				return expectedErr
			},
			RootHashCalled: realAccounts.RootHash,
			CommitCalled:   realAccounts.Commit,
		}
		args.State = stateCompMock

		pcf, _ := processComp.NewProcessComponentsFactoryMidas(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to Unmarshal failure", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		stateCompMock := factoryMocks.NewStateComponentsMockFromRealComponent(args.State)
		realAccounts := stateCompMock.AccountsAdapter()
		stateCompMock.Accounts = &testState.AccountsStub{
			GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
				addrOk, _ := addrPubKeyConv.Decode("erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4")
				addrNOK, _ := addrPubKeyConv.Decode("erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp")
				leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage(addrOk, []byte("value")) // coverage
				leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage(addrNOK, []byte("value"))
				close(leavesChannels.LeavesChan)
				leavesChannels.ErrChan.Close()
				return nil
			},
			RootHashCalled: realAccounts.RootHash,
			CommitCalled:   realAccounts.Commit,
		}
		args.State = stateCompMock

		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		cnt := 0
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					cnt++
					if cnt == 1 {
						return nil // coverage, key_ok
					}
					return expectedErr
				},
			}
		}
		args.CoreData = coreCompStub
		pcf, _ := processComp.NewProcessComponentsFactoryMidas(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to error on GetAllLeaves", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		realStateComp := args.State
		args.State = &factoryMocks.StateComponentsMock{
			Accounts: &testState.AccountsStub{
				GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.WriteInChanNonBlocking(expectedErr)
					leavesChannels.ErrChan.Close()
					return nil
				},
				CommitCalled:   realStateComp.AccountsAdapter().Commit,
				RootHashCalled: realStateComp.AccountsAdapter().RootHash,
			},
			PeersAcc:             realStateComp.PeerAccounts(),
			Tries:                realStateComp.TriesContainer(),
			AccountsAPI:          realStateComp.AccountsAdapterAPI(),
			StorageManagers:      realStateComp.TrieStorageManagers(),
			MissingNodesNotifier: realStateComp.MissingTrieNodesNotifier(),
		}

		pcf, _ := processComp.NewProcessComponentsFactoryMidas(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work with indexAndReturnGenesisAccounts failing due to error on Encode", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		statusCompStub, ok := args.StatusComponents.(*testsMocks.StatusComponentsStub)
		require.True(t, ok)
		statusCompStub.Outport = &outport.OutportStub{
			HasDriversCalled: func() bool {
				return true
			},
		}
		realStateComp := args.State
		args.State = &factoryMocks.StateComponentsMock{
			Accounts: &testState.AccountsStub{
				GetAllLeavesCalled: func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeavesParser common.TrieLeafParser) error {
					leavesChannels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("invalid addr"), []byte("value"))
					close(leavesChannels.LeavesChan)
					leavesChannels.ErrChan.Close()
					return nil
				},
				CommitCalled:   realStateComp.AccountsAdapter().Commit,
				RootHashCalled: realStateComp.AccountsAdapter().RootHash,
			},
			PeersAcc:             realStateComp.PeerAccounts(),
			Tries:                realStateComp.TriesContainer(),
			AccountsAPI:          realStateComp.AccountsAdapterAPI(),
			StorageManagers:      realStateComp.TrieStorageManagers(),
			MissingNodesNotifier: realStateComp.MissingTrieNodesNotifier(),
		}
		coreCompStub := factoryMocks.NewCoreComponentsHolderStubFromRealComponent(args.CoreData)
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return &marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return nil
				},
			}
		}
		args.CoreData = coreCompStub

		pcf, _ := processComp.NewProcessComponentsFactoryMidas(args)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.Nil(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = args.State.Close()
	})
	t.Run("should work - shard", func(t *testing.T) {
		shardCoordinator := sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
		processArgs := GetSovereignProcessComponentsFactoryArgsMidas(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactoryMidas(processArgs)
		require.NotNil(t, pcf)

		instance, err := pcf.Create()
		require.NoError(t, err)
		require.NotNil(t, instance)

		err = instance.Close()
		require.NoError(t, err)
		_ = processArgs.State.Close()
	})
}

func testCreateWithArgsMidas(t *testing.T, args processComp.ProcessComponentsFactoryArgs, expectedErrSubstr string) {
	pcf, _ := processComp.NewProcessComponentsFactoryMidas(args)
	require.NotNil(t, pcf)

	instance, err := pcf.Create()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), expectedErrSubstr))
	require.Nil(t, instance)

	_ = args.State.Close()
}

func TestProcessComponentsFactoryMidas_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("creating process components factory in sovereign chain should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
		processArgs := GetSovereignProcessComponentsFactoryArgsMidas(shardCoordinator)
		pcf, _ := processComp.NewProcessComponentsFactoryMidas(processArgs)

		require.NotNil(t, pcf)

		pc, err := pcf.Create()

		assert.NotNil(t, pc)
		assert.Nil(t, err)
	})
}

func GetSovereignProcessComponentsFactoryArgsMidas(shardCoordinator sharding.Coordinator) processComp.ProcessComponentsFactoryArgs {
	coreComponents := components.GetSovereignCoreComponents()
	cryptoComponents := components.GetCryptoComponents(coreComponents)
	networkComponents := components.GetNetworkComponents(cryptoComponents)
	dataComponents := components.GetSovereignDataComponents(coreComponents, shardCoordinator)
	stateComponents := components.GetSovereignStateComponents(coreComponents, components.GetSovereignStatusCoreComponents())
	processArgs := GetSovereignProcessArgsMidas(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	return processArgs
}

func GetSovereignProcessArgsMidas(
	shardCoordinator sharding.Coordinator,
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
) processComp.ProcessComponentsFactoryArgs {
	processArgs := components.GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)

	initialAccounts := createSovereignAccountsMidas()
	runTypeComponents := components.GetRunTypeComponentsStub(GetSovereignRunTypeComponentsMidas()) // TODO:
	runTypeComponents.AccountParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return initialAccounts
		},
		GenerateInitialTransactionsCalled: func(shardCoordinator sharding.Coordinator, initialIndexingData map[uint32]*genesis.IndexingData) ([]*dataBlock.MiniBlock, map[uint32]*outportCore.TransactionPool, error) {
			txsPool := make(map[uint32]*outportCore.TransactionPool)
			for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
				txsPool[i] = &outportCore.TransactionPool{}
			}

			return make([]*dataBlock.MiniBlock, 4), txsPool, nil
		},
		InitialAccountsSplitOnAddressesShardsCalled: func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialAccountHandler, error) {
			return map[uint32][]genesis.InitialAccountHandler{
				0: initialAccounts,
			}, nil
		},
	}

	bootstrapComponentsFactoryArgs := components.GetBootStrapFactoryArgs()
	bootstrapComponentsFactoryArgs.RunTypeComponents = runTypeComponents
	bootstrapComponentsFactory, _ := bootstrapComp.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	bootstrapComponents, _ := bootstrapComp.NewTestManagedBootstrapComponents(bootstrapComponentsFactory)
	_ = bootstrapComponents.Create()
	_ = bootstrapComponents.SetShardCoordinator(shardCoordinator)

	statusCoreComponents := components.GetSovereignStatusCoreComponents()

	processArgs.BootstrapComponents = bootstrapComponents
	processArgs.StatusCoreComponents = statusCoreComponents
	processArgs.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{}
	processArgs.RunTypeComponents = runTypeComponents

	return processArgs
}

func createSovereignAccountsMidas() []genesis.InitialAccountHandler {
	addrConverter, _ := commonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
		Hrp:             "erd",
	})
	acc1 := data.InitialAccount{
		Address:      "erd1whq0zspt6ktnv37gqj303da0vygyqwf5q52m7erftd0rl7laygfs6rhpct",
		Supply:       big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
	acc2 := data.InitialAccount{
		Address:      "erd129ppuuvtylghsx7muf29xnzw5lm9v2v8h4942ynymjpu2ftycgtq0rgq3h",
		Supply:       big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(2500000000), big.NewInt(1000000000000)),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	acc1Bytes, _ := addrConverter.Decode(acc1.Address)
	acc1.SetAddressBytes(acc1Bytes)

	acc2Bytes, _ := addrConverter.Decode(acc2.Address)
	acc2.SetAddressBytes(acc2Bytes)
	return []genesis.InitialAccountHandler{&acc1, &acc2}
}

var log = logger.GetOrCreate("componentsMock")

func GetSovereignRunTypeComponentsMidas() factory.RunTypeComponentsHolder {
	sovereignComponentsFactory, _ := runType.NewSovereignRunTypeComponentsFactoryMidas(createSovRunTypeArgs())
	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(sovereignComponentsFactory)
	if err != nil {
		log.Error("getRunTypeComponents NewManagedRunTypeComponents", "error", err.Error())
		return nil
	}
	err = managedRunTypeComponents.Create()
	if err != nil {
		log.Error("getRunTypeComponents Create", "error", err.Error())
		return nil
	}
	return managedRunTypeComponents
}

func createSovRunTypeArgs() runType.ArgsSovereignRunTypeComponents {
	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(createArgsRunTypeComponents())

	return runType.ArgsSovereignRunTypeComponents{
		RunTypeComponentsFactory: runTypeComponentsFactory,
		Config: config.SovereignConfig{
			GenesisConfig: config.GenesisConfig{
				NativeESDT: "WEGLD-ab47da",
			},
		},
		DataCodec:     &sovereign.DataCodecMock{},
		TopicsChecker: &sovereign.TopicsCheckerMock{},
	}
}

func createArgsRunTypeComponents() runType.ArgsRunTypeComponents {
	return runType.ArgsRunTypeComponents{
		CoreComponents: &mockCoreComp.CoreComponentsStub{
			HasherField:                 &hashingMocks.HasherMock{},
			InternalMarshalizerField:    &marshallerMock.MarshalizerMock{},
			EnableEpochsHandlerField:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			AddressPubKeyConverterField: &testscommon.PubkeyConverterStub{},
		},
		CryptoComponents: &mockCoreComp.CryptoComponentsStub{
			TxKeyGen: &mockCoreComp.KeyGenMock{},
			BlockSig: &mockConsensus.SingleSignerMock{},
		},
		Configs: config.Configs{
			EconomicsConfig: &config.EconomicsConfig{
				GlobalSettings: config.GlobalSettings{
					GenesisTotalSupply:          "20000000000000000000000000",
					GenesisMintingSenderAddress: "erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3",
				},
			},
		},
		InitialAccounts: createAccounts(),
	}
}

func createAccounts() []genesis.InitialAccountHandler {
	addrConverter, _ := commonFactory.NewPubkeyConverter(config.PubkeyConfig{
		Length:          32,
		Type:            "bech32",
		SignatureLength: 0,
		Hrp:             "erd",
	})
	acc1 := data.InitialAccount{
		Address:      "erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp",
		Supply:       big.NewInt(0).Mul(big.NewInt(5000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(4997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
	acc2 := data.InitialAccount{
		Address:      "erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4",
		Supply:       big.NewInt(0).Mul(big.NewInt(5000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(4997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
	acc3 := data.InitialAccount{
		Address:      "erd10d2gufxesrp8g409tzxljlaefhs0rsgjle3l7nq38de59txxt8csj54cd3",
		Supply:       big.NewInt(0).Mul(big.NewInt(10000000), big.NewInt(1000000000000000000)),
		Balance:      big.NewInt(0).Mul(big.NewInt(9997500), big.NewInt(1000000000000000000)),
		StakingValue: big.NewInt(0).Mul(big.NewInt(2500), big.NewInt(1000000000000000000)),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}

	acc1Bytes, _ := addrConverter.Decode(acc1.Address)
	acc1.SetAddressBytes(acc1Bytes)
	acc2Bytes, _ := addrConverter.Decode(acc2.Address)
	acc2.SetAddressBytes(acc2Bytes)
	acc3Bytes, _ := addrConverter.Decode(acc3.Address)
	acc3.SetAddressBytes(acc3Bytes)

	return []genesis.InitialAccountHandler{&acc1, &acc2, &acc3}
}
