package factory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockNewSystemScFactoryMidasArgs() ArgsNewSystemSCFactory {
	gasMap := wasmConfig.MakeGasMapForTests()
	gasMap = defaults.FillGasMapInternal(gasMap, 1)
	gasSchedule := testscommon.NewGasScheduleNotifierMock(gasMap)
	return ArgsNewSystemSCFactory{
		SystemEI:            &mock.SystemEIStub{},
		Economics:           &mock.EconomicsHandlerStub{},
		SigVerifier:         &mock.MessageSignVerifierMock{},
		GasSchedule:         gasSchedule,
		NodesConfigProvider: &mock.NodesConfigProviderStub{},
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &hashingMocks.HasherMock{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "100000000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					NumNodes:         3,
					MinPassThreshold: 1,
					MinQuorum:        2,
					MinVetoThreshold: 2,
					ProposalCost:     "100",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             100,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				StakeLimitPercentage:                 100.0,
				NodeLimitPercentage:                  100.0,
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 10000,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "10",
				MinStakeAmount:      "10",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			SoftAuctionConfig: config.SoftAuctionConfig{
				TopUpStep:             "10",
				MinTopUp:              "1",
				MaxTopUp:              "32000000",
				MaxNumberOfIterations: 100000,
			},
		},
		AddressPubKeyConverter: &testscommon.PubkeyConverterMock{},
		ShardCoordinator:       &mock.ShardCoordinatorStub{},
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		NodesCoordinator:       &mock.NodesCoordinatorStub{},
	}
}

func TestNewSystemSCFactoryMidas_NilSystemEI(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.SystemEI = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilSystemEnvironmentInterface))
}

func TestNewSystemSCFactoryMidas_NilNodesCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.NodesCoordinator = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilNodesCoordinator))
}

func TestNewSystemSCFactoryMidas_NilSigVerifier(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.SigVerifier = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilMessageSignVerifier))
}

func TestNewSystemSCFactoryMidas_NilNodesConfigProvider(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.NodesConfigProvider = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilNodesConfigProvider))
}

func TestNewSystemSCFactoryMidas_NilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.Marshalizer = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilMarshalizer))
}

func TestNewSystemSCFactoryMidas_NilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.Hasher = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilHasher))
}

func TestNewSystemSCFactoryMidas_NilEconomicsData(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.Economics = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilEconomicsData))
}

func TestNewSystemSCFactoryMidas_NilSystemScConfig(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.SystemSCConfig = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilSystemSCConfig))
}

func TestNewSystemSCFactoryMidas_NilPubKeyConverter(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.AddressPubKeyConverter = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilAddressPubKeyConverter))
}

func TestNewSystemSCFactoryMidas_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.ShardCoordinator = nil
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	require.Nil(t, scFactory)
	require.True(t, errors.Is(err, vm.ErrNilShardCoordinator))
}

func TestNewSystemSCFactoryMidas_Ok(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	scFactory, err := NewSystemSCFactoryMidas(arguments)

	assert.Nil(t, err)
	assert.NotNil(t, scFactory)
}

func TestNewSystemSCFactoryMidas_GasScheduleChangeMissingElementsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	arguments := createMockNewSystemScFactoryMidasArgs()
	scFactory, _ := NewSystemSCFactoryMidas(arguments)

	gasSchedule, err := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV3.toml")
	delete(gasSchedule["MetaChainSystemSCsCost"], "UnstakeTokens")
	require.Nil(t, err)

	scFactory.GasScheduleChange(gasSchedule)

	assert.Equal(t, uint64(1), scFactory.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
}

func TestNewSystemSCFactoryMidas_GasScheduleChangeShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	arguments := createMockNewSystemScFactoryMidasArgs()
	scFactory, _ := NewSystemSCFactoryMidas(arguments)

	gasSchedule, err := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV3.toml")
	require.Nil(t, err)

	scFactory.GasScheduleChange(gasSchedule)

	assert.Equal(t, uint64(5000000), scFactory.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
}

func TestSystemSCFactoryMidas_CreateWithBadDelegationManagerConfigChangeAddressShouldError(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	arguments.SystemSCConfig.DelegationManagerSystemSCConfig.ConfigChangeAddress = "not a hex string"
	scFactory, _ := NewSystemSCFactoryMidas(arguments)

	container, err := scFactory.Create()

	assert.True(t, check.IfNil(container))
	assert.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestSystemSCFactoryMidas_Create(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	scFactory, _ := NewSystemSCFactoryMidas(arguments)

	container, err := scFactory.Create()
	assert.Nil(t, err)
	require.NotNil(t, container)
	assert.Equal(t, 6, container.Len())
}

func TestSystemSCFactoryMidas_CreateForGenesis(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	scFactory, _ := NewSystemSCFactoryMidas(arguments)

	container, err := scFactory.CreateForGenesis()
	assert.Nil(t, err)
	assert.Equal(t, 4, container.Len())
}

func TestSystemSCFactoryMidas_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryMidasArgs()
	scFactory, _ := NewSystemSCFactoryMidas(arguments)
	assert.False(t, scFactory.IsInterfaceNil())

	scFactory = nil
	require.Nil(t, scFactory)
}
