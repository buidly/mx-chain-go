package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var AbstractStakingSCAddress = []byte("abstractStaking")

func createMockArgumentsForValidatorSCMidasMidasWithSystemScAddresses(
	validatorScAddress []byte,
	stakingScAddress []byte,
	endOfEpochAddress []byte,
	abstractStakingAddr []byte,
) ArgsValidatorSmartContractMidas {
	args := ArgsValidatorSmartContractMidas{
		ArgsValidatorSmartContract: ArgsValidatorSmartContract{
			Eei:                &mock.SystemEIStub{},
			SigVerifier:        &mock.MessageSignVerifierMock{},
			ValidatorSCAddress: validatorScAddress,
			StakingSCAddress:   stakingScAddress,
			EndOfEpochAddress:  endOfEpochAddress,
			StakingSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "10000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				UnBondPeriodInEpochs:                 1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
			},
			Marshalizer:            &mock.MarshalizerMock{},
			GenesisTotalSupply:     big.NewInt(100000000),
			MinDeposit:             "0",
			DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
			GovernanceSCAddress:    vm.GovernanceSCAddress,
			ShardCoordinator:       &mock.ShardCoordinatorStub{},
			EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(
				common.StakeFlag,
				common.UnBondTokensV2Flag,
				common.ValidatorToDelegationFlag,
				common.DoubleKeyProtectionFlag,
				common.MultiClaimOnDelegationFlag,
			),
		},
		AbstractStakingAddr: abstractStakingAddr,
	}

	return args
}

func createMockArgumentsForValidatorSCMidas() ArgsValidatorSmartContractMidas {
	return createMockArgumentsForValidatorSCMidasMidasWithSystemScAddresses(
		[]byte("validator"),
		[]byte("staking"),
		[]byte("endOfEpoch"),
		AbstractStakingSCAddress,
	)
}

func createMockArgumentsForValidatorSCMidasWithRealAddresses() ArgsValidatorSmartContractMidas {
	return createMockArgumentsForValidatorSCMidasMidasWithSystemScAddresses(
		vm.ValidatorSCAddress,
		vm.StakingSCAddress,
		vm.EndOfEpochAddress,
		AbstractStakingSCAddress,
	)
}

func TestNewStakingValidatorSmartContractMidas_InvalidUnJailValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()

	arguments.StakingSCConfig.UnJailValue = ""
	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidUnJailCost))

	arguments.StakingSCConfig.UnJailValue = "0"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidUnJailCost))

	arguments.StakingSCConfig.UnJailValue = "-5"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidUnJailCost))
}

func TestNewStakingValidatorSmartContractMidas_InvalidMinStakeValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()

	arguments.StakingSCConfig.MinStakeValue = ""
	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStakeValue))

	arguments.StakingSCConfig.MinStakeValue = "0"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStakeValue))

	arguments.StakingSCConfig.MinStakeValue = "-5"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStakeValue))
}

func TestNewStakingValidatorSmartContractMidas_InvalidGenesisNodePrice(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()

	arguments.StakingSCConfig.GenesisNodePrice = ""
	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidNodePrice))

	arguments.StakingSCConfig.GenesisNodePrice = "0"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidNodePrice))

	arguments.StakingSCConfig.GenesisNodePrice = "-5"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidNodePrice))
}

func TestNewStakingValidatorSmartContractMidas_InvalidMinStepValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()

	arguments.StakingSCConfig.MinStepValue = ""
	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStepValue))

	arguments.StakingSCConfig.MinStepValue = "0"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStepValue))

	arguments.StakingSCConfig.MinStepValue = "-5"
	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidMinStepValue))
}

func TestNewStakingValidatorSmartContractMidas_NilSystemEnvironmentInterface(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.Eei = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrNilSystemEnvironmentInterface))
}

func TestNewStakingValidatorSmartContractMidas_NilStakingSmartContractAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.StakingSCAddress = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrNilStakingSmartContractAddress))
}

func TestNewStakingValidatorSmartContractMidas_NilValidatorSmartContractAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.ValidatorSCAddress = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrNilValidatorSmartContractAddress))
}

func TestNewStakingValidatorSmartContractMidas_NilSigVerifier(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.SigVerifier = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrNilMessageSignVerifier))
}

func TestNewStakingValidatorSmartContractMidas_NilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.Marshalizer = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrNilMarshalizer))
}

func TestNewStakingValidatorSmartContractMidas_InvalidGenesisTotalSupply(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.GenesisTotalSupply = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidGenesisTotalSupply))

	arguments.GenesisTotalSupply = big.NewInt(0)

	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidGenesisTotalSupply))

	arguments.GenesisTotalSupply = big.NewInt(-2)

	asc, err = NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidGenesisTotalSupply))
}

func TestNewStakingValidatorSmartContractMidas_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.EnableEpochsHandler = nil

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrNilEnableEpochsHandler))
}

func TestNewStakingValidatorSmartContractMidas_InvalidEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewStakingValidatorSmartContractMidas_EmptyEndOfEpochAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.EndOfEpochAddress = make([]byte, 0)

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	assert.True(t, errors.Is(err, vm.ErrInvalidEndOfEpochAccessAddress))
}

func TestNewStakingValidatorSmartContractMidas_EmptyDelegationManagerAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.DelegationMgrSCAddress = make([]byte, 0)

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestNewStakingValidatorSmartContractMidas_EmptyGovernanceAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForValidatorSCMidas()
	arguments.GovernanceSCAddress = make([]byte, 0)

	asc, err := NewValidatorSmartContractMidas(arguments)
	require.Nil(t, asc)
	require.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

// TODO: Not sure why this panics with runtime error...
//func TestNewStakingValidatorSmartContractMidas_NilShardCoordinator(t *testing.T) {
//	t.Parallel()
//
//	arguments := createMockArgumentsForValidatorSCMidas()
//	arguments.ShardCoordinator = nil
//
//	asc, err := NewValidatorSmartContractMidas(arguments)
//	require.True(t, check.IfNil(asc))
//	require.True(t, errors.Is(err, vm.ErrNilShardCoordinator))
//}

func TestStakingValidatorSCMidas_ExecuteStakeWithoutBlsKeysShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()
	validatorData := createAValidatorData(25000, 2, 12500)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	validatorAddr := []byte("validatorAddr")

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(key, validatorAddr) {
			return validatorDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Equal(key, validatorAddr) {
			var validatorDataRecovered ValidatorDataV2
			_ = json.Unmarshal(value, &validatorDataRecovered)
			assert.Equal(t, big.NewInt(35000), validatorDataRecovered.TotalStakeValue)
		}
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(0).Bytes(), validatorAddr, big.NewInt(35000).Bytes()}
	arguments.CallValue = big.NewInt(0)

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingValidatorSCMidas_ExecuteStakeAddedNewPubKeysShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()
	validatorData := createAValidatorData(25000, 2, 12500)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	key1 := []byte("Key1")
	key2 := []byte("Key2")
	validatorAddr := []byte("dummyAddress2")

	args := createMockArgumentsForValidatorSCMidas()

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("validator"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	eei.SetStorage(validatorAddr, validatorDataBytes)
	args.StakingSCConfig = argsStaking.StakingSCConfig

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key2, []byte("msg2"), validatorAddr, big.NewInt(25000 + 20000).Bytes()}

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingValidatorSCMidas_ExecuteStakeDoubleKeyAndCleanup(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()

	key1 := []byte("Key1")
	validatorAddr := []byte("validatorAddr")
	args := createMockArgumentsForValidatorSCMidas()

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("auction"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.DoubleKeyProtectionFlag)
	validatorSc, _ := NewValidatorSmartContractMidas(args)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key1, []byte("msg1"), validatorAddr, big.NewInt(20000).Bytes()}

	errCode := validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)

	registeredData := &ValidatorDataV2{}
	_ = validatorSc.marshalizer.Unmarshal(registeredData, eei.GetStorage(validatorAddr))
	assert.Equal(t, 2, len(registeredData.BlsPubKeys))

	enableEpochsHandler.AddActiveFlags(common.DoubleKeyProtectionFlag)
	arguments.Function = "cleanRegisteredData"
	arguments.CallValue = big.NewInt(0)
	arguments.Arguments = [][]byte{}
	arguments.CallerAddr = validatorAddr

	errCode = validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)

	_ = validatorSc.marshalizer.Unmarshal(registeredData, eei.GetStorage(validatorAddr))
	assert.Equal(t, 1, len(registeredData.BlsPubKeys))
}

func TestStakingValidatorSCMidas_ExecuteStakeUnJail(t *testing.T) {
	t.Parallel()

	validatorAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	args := createMockArgumentsForValidatorSCMidas()

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.StakingSCConfig.UnJailValue = "100"

	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed"), validatorAddress, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unJail"
	arguments.Arguments = [][]byte{stakerPubKey, validatorAddress}
	arguments.CallValue = big.NewInt(100)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("otherAddress")
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteUnjailOtherCallerShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unJail"
	arguments.CallerAddr = []byte("other")

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "unJail function not allowed to be called by address " + string(arguments.CallerAddr), eei.ReturnMessage)
}

func TestStakingValidatorSCMidas_ExecuteStakeUnStakeOneBlsPubKey(t *testing.T) {
	t.Parallel()

	validatorAddr := []byte("validatorAddr")
	arguments := CreateVmContractCallInputMidas()
	validatorData := createAValidatorData(25000, 2, 12500)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 1,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: []byte("address"),
		StakeValue:    big.NewInt(0),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(key, validatorAddr) {
			return validatorDataBytes
		}
		if bytes.Equal(key, validatorData.BlsPubKeys[0]) {
			return stakedDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		var stakedDataRecovered StakedDataV2_0
		_ = json.Unmarshal(value, &stakedDataRecovered)

		assert.Equal(t, false, stakedDataRecovered.Staked)
	}

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{validatorData.BlsPubKeys[0], validatorAddr, big.NewInt(12500).Bytes()}
	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingValidatorSCMidas_ExecuteStakeStakeTokensUnBondRestakeUnstake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1
	stubStaking, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stubStaking.AddActiveFlags(common.StakingV2Flag)
	argsStaking.MinNumNodes = 0
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(0).Bytes(), stakerAddress.Bytes(), big.NewInt(10100).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeTokens"
	arguments.Arguments = [][]byte{stakerAddress.Bytes(), big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 1
	}

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes(), stakerAddress.Bytes(), big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 100
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 10
	}

	arguments.Function = "unBondTokens"
	arguments.Arguments = [][]byte{}
	arguments.CallerAddr = stakerAddress.Bytes()
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.CallerAddr = AbstractStakingSCAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	marshaledData := eei.GetStorageFromAddress(args.StakingSCAddress, stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(0).Bytes(), stakerAddress.Bytes(), big.NewInt(10100).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeTokens"
	arguments.Arguments = [][]byte{stakerAddress.Bytes(), big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes(), stakerAddress.Bytes(), big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingValidatorSCMidas_ExecuteStakeUnStake1Stake1More(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.MinNumNodes = 0
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stubStaking, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stubStaking.AddActiveFlags(common.StakingV2Flag)
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 10
	}

	staker := []byte("staker")
	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey1"), []byte("signed"), staker, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("blsKey1"), staker, big.NewInt(0).Bytes()}

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey2"), []byte("signed"), staker, big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("blsKey2"), staker, big.NewInt(0).Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), []byte("blsKey3"), []byte("blsKey4"), []byte("signed"), staker, big.NewInt(15000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey3"), []byte("signed"), staker, big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("blsKey4"), []byte("signed"), staker, big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	marshaledData := eei.GetStorageFromAddress(args.StakingSCAddress, []byte("blsKey3"))
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, []byte("blsKey2"))
	stakedData = &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)

	marshaledData = eei.GetStorageFromAddress(args.StakingSCAddress, []byte("blsKey1"))
	stakedData = &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)
}

func TestStakingValidatorSCMidas_ExecuteStakeUnStakeUnBondUnStakeUnBondOneBlsPubKey(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("staker1")
	stakerPubKey := []byte("bls1")

	unBondPeriod := uint64(5)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unBondPeriod
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed"), stakerAddress, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("anotherKey"), []byte("signed"), []byte("anotherAddress"), big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 100
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 10
	}

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress, big.NewInt(0).Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 103
	}
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 120
	}
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 220
	}
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 320
	}
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	assert.Equal(t, 0, len(marshaledData))
}

func TestStakingValidatorSCMidas_StakeUnStake3XUnBond2xWaitingList(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey1 := []byte("blsKey1")
	stakerPubKey2 := []byte("blsKey2")
	stakerPubKey3 := []byte("blsKey3")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey1, []byte("signed"), stakerAddress, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey1, stakerAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey2, []byte("signed"), stakerAddress, big.NewInt(20000).Bytes()}

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey2, stakerAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey3, []byte("signed"), stakerAddress, big.NewInt(30000).Bytes()}

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey3, stakerAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey2, stakerAddress}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey3, stakerAddress}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingValidatorSCMidas_StakeShouldSetOwnerIfStakingV2IsEnabled(t *testing.T) {
	t.Parallel()

	ownerAddress := []byte("owner")
	blsKey := []byte("blsKey")

	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.MaxNumberOfNodesForStake = 1
	atArgParser := parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	eei := createDefaultEei()
	eei.inputParser = atArgParser
	argsStaking.Eei = eei
	eei.SetSCAddress(args.ValidatorSCAddress)
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stubStaking, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stubStaking.AddActiveFlags(common.StakingV2Flag)
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey, []byte("signed"), ownerAddress, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	registeredData, err := sc.getStakedData(blsKey)
	require.Nil(t, err)
	assert.Equal(t, ownerAddress, registeredData.OwnerAddress)
}

func TestStakingValidatorSCMidas_ExecuteStakeUnStakeUnBondBlsPubKeyAndRestake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)
	nonce := uint64(1)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSCMidas()

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1000
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("anotherKey"), []byte("signed"), []byte("anotherCaller"), big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	nonce += 1
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes(), stakerAddress.Bytes(), big.NewInt(0).Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	nonce += args.StakingSCConfig.UnBondPeriod + 1
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes(), stakerAddress.Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedDataV2_0{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingValidatorSCMidas_ExecuteUnBond(t *testing.T) {
	t.Parallel()

	validatorAddr := []byte("dummyAddress1")
	arguments := CreateVmContractCallInputMidas()
	totalStake := uint64(25000)

	validatorData := createAValidatorData(totalStake, 2, 12500)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 1,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: validatorAddr,
		StakeValue:    big.NewInt(12500),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(validatorAddr, key) {
			return validatorDataBytes
		}
		if bytes.Equal(key, validatorData.BlsPubKeys[0]) {
			return stakedDataBytes
		}

		return nil
	}

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{validatorData.BlsPubKeys[0], validatorAddr}
	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestValidatorStakingSCMidas_ExecuteInit(t *testing.T) {
	t.Parallel()

	eei := createDefaultEei()
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	ownerAddr := stakingSmartContract.eei.GetStorage([]byte(ownerKey))
	assert.Equal(t, arguments.CallerAddr, ownerAddr)

	ownerBalanceBytes := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	ownerBalance := big.NewInt(0).SetBytes(ownerBalanceBytes)
	assert.Equal(t, big.NewInt(0), ownerBalance)

}

func TestValidatorStakingSCMidas_ExecuteInitTwoTimeShouldReturnUserError(t *testing.T) {
	t.Parallel()

	eei := createDefaultEei()
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteStakeOutOfGasShouldErr(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 2
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = parsers.NewCallArgsParser()
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.Stake = 10
	stakingSmartContract, err := NewValidatorSmartContractMidas(args)
	require.Nil(t, err)
	arguments := CreateVmContractCallInputMidas()
	arguments.GasProvided = 5
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfGas, retCode)

	assert.Equal(t, vm.InsufficientGasLimit, eei.returnMessage)
}

func TestValidatorStakingSCMidas_ExecuteStakeWrongStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return nil, state.ErrAccNotFound
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = parsers.NewCallArgsParser()
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = "10"
	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()

	validatorAddr := []byte("validatorAddr")

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(0).Bytes(), validatorAddr, big.NewInt(2).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, fmt.Sprintf("insufficient stake value: expected %s, got %v",
		args.StakingSCConfig.GenesisNodePrice, big.NewInt(2)))

	balance := eei.GetBalance(validatorAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestValidatorStakingSCMidas_ExecuteStakeWrongUnmarshalDataShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteStakeAlreadyStakedShouldNotErr(t *testing.T) {
	t.Parallel()

	stakerBlsKey1 := big.NewInt(101)
	expectedCallerAddress := []byte("caller")
	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei := createDefaultEei()
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{
		big.NewInt(1).Bytes(),
		stakerBlsKey1.Bytes(), []byte("signed"),
		expectedCallerAddress,
		nodePrice.Bytes(),
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(stakerBlsKey1.Bytes(), marshalizedExpectedRegData)

	validatorData := ValidatorDataV2{
		RewardAddress:   expectedCallerAddress,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{stakerBlsKey1.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(expectedCallerAddress, marshaledRegistrationData)
	arguments.Arguments = [][]byte{
		big.NewInt(1).Bytes(),
		stakerBlsKey1.Bytes(), []byte("signed"),
		expectedCallerAddress,
		big.NewInt(0).Mul(nodePrice, big.NewInt(2)).Bytes(),
	}
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, retCode)
	var registrationData ValidatorDataV2
	data := stakingSmartContract.eei.GetStorage(expectedCallerAddress)
	_ = json.Unmarshal(data, &registrationData)

	validatorData.TotalStakeValue = big.NewInt(0).Mul(nodePrice, big.NewInt(2))
	validatorData.MaxStakePerNode = big.NewInt(0).Mul(nodePrice, big.NewInt(2))

	assert.Equal(t, validatorData, registrationData)
}

func TestValidatorStakingSCMidas_ExecuteStakeStakedInStakingButNotInValidatorShouldErr(t *testing.T) {
	t.Parallel()

	stakerBlsKey1 := big.NewInt(101)
	expectedCallerAddress := []byte("caller")
	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei := createDefaultEei()
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{
		big.NewInt(1).Bytes(),
		stakerBlsKey1.Bytes(),
		[]byte("signed"),
		expectedCallerAddress,
		nodePrice.Bytes(),
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(stakerBlsKey1.Bytes(), marshalizedExpectedRegData)

	validatorData := ValidatorDataV2{
		RewardAddress:   expectedCallerAddress,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0),
		LockedStake:     big.NewInt(0),
		MaxStakePerNode: big.NewInt(0),
		NumRegistered:   0,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(expectedCallerAddress, marshaledRegistrationData)
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "bls key already registered"))
	var registrationData ValidatorDataV2
	data := stakingSmartContract.eei.GetStorage(expectedCallerAddress)
	_ = json.Unmarshal(data, &registrationData)

	assert.Equal(t, validatorData, registrationData)
}

func TestValidatorStakingSCMidas_ExecuteStakeNotEnoughArgsShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&StakedDataV2_0{})
		return registrationDataMarshalized
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteStakeNotEnoughFundsForMultipleNodesShouldErr(t *testing.T) {
	t.Parallel()

	validatorAddr := []byte("validator")
	stakerPubKey1 := big.NewInt(101)
	stakerPubKey2 := big.NewInt(102)
	args := createMockArgumentsForValidatorSCMidas()

	eei := createDefaultEei()
	eei.inputParser = parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = "10"
	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	eei.SetGasProvided(arguments.GasProvided)
	arguments.Arguments = [][]byte{
		big.NewInt(2).Bytes(),
		stakerPubKey1.Bytes(), []byte("signed"),
		stakerPubKey2.Bytes(), []byte("signed"),
		validatorAddr,
		big.NewInt(15).Bytes(),
	}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfFunds, retCode)

	balance := eei.GetBalance(validatorAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestValidatorStakingSCMidas_ExecuteStakeNotEnoughGasForMultipleNodesShouldErr(t *testing.T) {
	t.Parallel()
	stakerPubKey1 := big.NewInt(101)
	stakerPubKey2 := big.NewInt(102)
	args := createMockArgumentsForValidatorSCMidas()

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return nil, state.ErrAccNotFound
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = "10"
	args.GasCost.MetaChainSystemSCsCost.Stake = 10
	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.GasProvided = 15
	eei.SetGasProvided(arguments.GasProvided)
	arguments.Arguments = [][]byte{
		big.NewInt(2).Bytes(),
		stakerPubKey1.Bytes(), []byte("signed"),
		stakerPubKey2.Bytes(), []byte("signed"),
		[]byte("validatorAddr"),
		big.NewInt(15).Bytes(),
	}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.OutOfGas, retCode)

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestValidatorStakingSCMidas_ExecuteStakeOneKeyFailsOneRegisterStakeSCShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForValidatorSCMidas()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)

	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	args.Eei = eei
	executeSecond := true
	stakingSc := &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
		if args.Function != "register" {
			return vmcommon.Ok
		}

		if executeSecond {
			executeSecond = false
			return vmcommon.Ok
		}
		return vmcommon.UserError
	}}

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{
		big.NewInt(2).Bytes(),
		stakerPubKey.Bytes(), []byte("signed"),
		stakerPubKey.Bytes(), []byte("signed"),
		stakerAddress.Bytes(),
		big.NewInt(nodePrice.Int64() * 2).Bytes(),
	}

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	sc, _ := NewValidatorSmartContractMidas(args)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot register bls key"))

	data := sc.eei.GetStorage(arguments.CallerAddr)
	assert.Nil(t, data)
}

func TestValidatorStakingSCMidas_ExecuteStakeBeforeValidatorEnableNonce(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 99
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := ValidatorDataV2{
		RewardAddress:   stakerAddress.Bytes(),
		RegisterNonce:   0,
		Epoch:           99,
		BlsPubKeys:      [][]byte{stakerPubKey.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}

	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), nodePrice.Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData ValidatorDataV2
	data := sc.eei.GetStorage(stakerAddress.Bytes())
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestValidatorStakingSCMidas_ExecuteStake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := ValidatorDataV2{
		RewardAddress:   stakerAddress.Bytes(),
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{stakerPubKey.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}

	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed"), stakerAddress.Bytes(), big.NewInt(0).Set(nodePrice).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData ValidatorDataV2
	data := sc.eei.GetStorage(stakerAddress.Bytes())
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestValidatorStakingSCMidas_ExecuteStakeOtherCallerShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.CallerAddr = []byte("other")

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "stake function not allowed to be called by address " + string(arguments.CallerAddr), eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteUnStakeValueNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallValue = big.NewInt(1)
	arguments.Function = "unStake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.TransactionValueMustBeZero, eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteUnStakeAddressNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	notFoundkey := []byte("abc")
	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{notFoundkey, []byte("validatorAddr"), big.NewInt(0).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.CannotGetAllBlsKeysFromRegistrationData+
		fmt.Errorf("%w, key %s not found", vm.ErrBLSPublicKeyMismatch, hex.EncodeToString(notFoundkey)).Error(), eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteUnStakeUnmarshalErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	args.Marshalizer = &mock.MarshalizerMock{Fail: true}

	validatorAddr := []byte("validatorAddr")

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc"), validatorAddr, big.NewInt(0).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.CannotGetOrCreateRegistrationData+mock.ErrMockMarshalizer.Error(), eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteUnStakeAlreadyUnStakedAddrShouldNotErr(t *testing.T) {
	t.Parallel()

	expectedCallerAddress := []byte("caller")
	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei := createDefaultEei()
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc"), expectedCallerAddress, big.NewInt(0).Bytes()}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	validatorData := ValidatorDataV2{
		RewardAddress:   expectedCallerAddress,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)

	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(expectedCallerAddress, marshaledRegistrationData)
	retCode := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot unStake node which was already unStaked"))
}

func TestValidatorStakingSCMidas_ExecuteUnStakeFailsWithWrongCaller(t *testing.T) {
	t.Parallel()

	expectedCallerAddress := []byte("caller")
	wrongCallerAddress := []byte("wrongCaller")

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei := createDefaultEei()
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.CallerAddr = wrongCallerAddress
	arguments.Arguments = [][]byte{[]byte("abc")}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteUnstake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig.UnBondPeriod = 10
	callerAddress := []byte("caller")
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nodePrice,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
	}

	eei := createDefaultEei()
	eei.inputParser = parsers.NewCallArgsParser()

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args.Eei = eei
	eei.SetSCAddress(args.ValidatorSCAddress)

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc"), callerAddress, big.NewInt(0).Bytes()}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	validatorData := ValidatorDataV2{
		RewardAddress:   callerAddress,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)
	eei.SetStorage(callerAddress, marshaledRegistrationData)

	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: callerAddress,
		StakeValue:    nodePrice,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}
	marshaledStakedData, _ := json.Marshal(stakedData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshaledStakedData)
	stakingSc.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})
	eei.SetSCAddress(args.ValidatorSCAddress)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestValidatorStakingSCMidas_ExecuteUnstakeOtherCallerShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.CallerAddr = []byte("other")

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "unStake function not allowed to be called by address " + string(arguments.CallerAddr), eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteUnBondUnmarshalErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(200).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteUnBondValidatorNotUnStakeShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(
				&StakedDataV2_0{
					UnStakedNonce: 0,
				})
			return registrationDataMarshalized
		}
	}
	eei.BlockChainHookCalled = func() vm.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 10000
		}}
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteStakeUnStakeReturnsErrAsNotEnabled(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.BlockChainHookCalled = func() vm.BlockchainHook {
		return &mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 100
			},
			CurrentNonceCalled: func() uint64 {
				return 100
			}}
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.StakeFlag)
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("data"), big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.UnBondNotEnabled, eei.ReturnMessage)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{big.NewInt(0).Bytes(), []byte("data"), big.NewInt(0).Bytes()}
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.UnStakeNotEnabled, eei.ReturnMessage)

	arguments.Function = "stake"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, vm.StakeNotEnabled, eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteUnBondBeforePeriodEnds(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(100)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 10
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig.UnBondPeriod = 1000
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	validatorSc, _ := NewValidatorSmartContractMidas(args)
	eei.SetSCAddress([]byte("staking"))

	blsPubKey := []byte("pubkey")
	_ = validatorSc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RewardAddress: caller,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: 1,
					UnstakedValue: big.NewInt(1000),
				},
			},
			BlsPubKeys:      [][]byte{blsPubKey},
			TotalStakeValue: big.NewInt(1000), //in v1 this was still set to the unstaked value
			LockedStake:     big.NewInt(0),
			TotalUnstaked:   big.NewInt(1000),
		},
	)

	systemSc, _ := eei.GetContract(nil)
	stakingSc := systemSc.(*stakingSC)
	registrationData := &StakedDataV2_0{
		RegisterNonce: 0,
		UnStakedNonce: uint64(1),
		RewardAddress: caller,
		StakeValue:    big.NewInt(100),
	}
	_ = stakingSc.saveAsStakingDataV1P1(blsPubKey, registrationData)

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{blsPubKey, caller}

	retCode := validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "unBond is not possible"))
	assert.True(t, strings.Contains(eei.returnMessage, "unBond period did not pass"))
	assert.True(t, len(eei.GetStorage(caller)) != 0) //should have not removed the account data
}

func TestValidatorSCMidas_ExecuteUnBondBeforePeriodEndsForV2(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(100)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 10
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.UnBondPeriod = 1000
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	validatorSc, _ := NewValidatorSmartContractMidas(args)
	eei.SetSCAddress([]byte("staking"))

	blsPubKey := []byte("pubkey")
	_ = validatorSc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RewardAddress: caller,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: 1,
					UnstakedValue: big.NewInt(1000),
				},
			},
			BlsPubKeys:      [][]byte{blsPubKey},
			TotalStakeValue: big.NewInt(0),
			LockedStake:     big.NewInt(0),
			TotalUnstaked:   big.NewInt(1000),
		},
	)

	systemSc, _ := eei.GetContract(nil)
	stakingSc := systemSc.(*stakingSC)
	registrationData := &StakedDataV2_0{
		RegisterNonce: 0,
		UnStakedNonce: uint64(1),
		RewardAddress: caller,
		StakeValue:    big.NewInt(100),
	}
	_ = stakingSc.saveAsStakingDataV1P1(blsPubKey, registrationData)

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{blsPubKey, caller}

	retCode := validatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "unBond is not possible"))
	assert.True(t, strings.Contains(eei.returnMessage, "unBond period did not pass"))
	assert.True(t, len(eei.GetStorage(caller)) != 0) //should have not removed the account data
}

func TestValidatorStakingSCMidas_ExecuteUnBond(t *testing.T) {
	t.Parallel()

	unBondPeriod := uint64(100)
	unStakedNonce := uint64(10)
	stakeValue := big.NewInt(100)
	stakedData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakedNonce,
		RewardAddress: []byte("reward"),
		StakeValue:    big.NewInt(0).Set(stakeValue),
		JailedRound:   math.MaxUint64,
	}

	marshalizedStakedData, _ := json.Marshal(&stakedData)
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unStakedNonce + unBondPeriod + 1
		},
	}
	eei.inputParser = atArgParser

	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(ownerKey), scAddress)

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	args.StakingSCConfig.UnBondPeriod = unBondPeriod

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig = args.StakingSCConfig
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)

	validatorAddr := []byte("address")

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("abc"), validatorAddr}
	arguments.RecipientAddr = scAddress

	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedStakedData)
	stakingSc.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})
	eei.SetSCAddress(args.ValidatorSCAddress)

	validatorData := ValidatorDataV2{
		RewardAddress:   validatorAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: stakeValue,
		LockedStake:     stakeValue,
		MaxStakePerNode: stakeValue,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)
	eei.SetStorage(validatorAddr, marshaledRegistrationData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	assert.Equal(t, 0, len(data))
}

func TestValidatorStakingSCMidas_ExecuteUnbondOtherCallerShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unBond"
	arguments.CallerAddr = []byte("other")

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "unBond function not allowed to be called by address " + string(arguments.CallerAddr), eei.ReturnMessage)
}

func TestValidatorStakingSCMidas_ExecuteSlashOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "slash"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestValidatorStakingSCMidas_ExecuteUnStakeAndUnBondstake(t *testing.T) {
	t.Parallel()

	// Preparation
	unBondPeriod := uint64(100)
	valueStakedByTheCaller := big.NewInt(100)
	stakerAddress := []byte("address")
	stakerPubKey := []byte("pubKey")
	blockChainHook := &mock.BlockChainHookStub{}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	smartcontractAddress := "validator"
	eei.SetSCAddress([]byte(smartcontractAddress))

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = unBondPeriod
	args.StakingSCConfig.GenesisNodePrice = valueStakedByTheCaller.Text(10)
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)

	arguments := CreateVmContractCallInputMidas()
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress, big.NewInt(0).Bytes()}
	arguments.RecipientAddr = []byte(smartcontractAddress)

	stakedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)
	stakingSc.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	validatorData := ValidatorDataV2{
		RewardAddress:   stakerAddress,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: valueStakedByTheCaller,
		LockedStake:     valueStakedByTheCaller,
		MaxStakePerNode: valueStakedByTheCaller,
		NumRegistered:   1,
		TotalUnstaked:   big.NewInt(0),
	}
	marshaledRegistrationData, _ := json.Marshal(validatorData)
	eei.SetSCAddress(args.ValidatorSCAddress)
	eei.SetStorage(stakerAddress, marshaledRegistrationData)

	arguments.Function = "unStake"

	unStakeNonce := uint64(10)
	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return uint32(unStakeNonce)
	}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedDataV2_0
	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)

	expectedRegistrationData := StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakeNonce,
		UnStakedEpoch: uint32(unStakeNonce),
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}
	assert.Equal(t, expectedRegistrationData, registrationData)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress}

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce + unBondPeriod + 1
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return uint32(unStakeNonce + unBondPeriod + 1)
	}
	eei.SetSCAddress(args.ValidatorSCAddress)
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestValidatorStakingSCMidas_ExecuteGetShouldReturnUserErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "get"
	eei := createDefaultEei()
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, err)
}

func TestValidatorStakingSCMidas_ExecuteGetShouldOk(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "get"
	arguments.Arguments = [][]byte{arguments.CallerAddr}
	eei := createDefaultEei()
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	stakingSmartContract, _ := NewValidatorSmartContractMidas(args)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, err)
}

// Test scenario
// 1 -- call setConfig with wrong owner address should return error
// 2 -- call validator smart contract init and after that call setConfig with wrong number of arguments should return error
// 3 -- call setConfig after init was done successfully should work and config should be set correctly
func TestValidatorStakingSCMidas_SetConfig(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddress")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewValidatorSmartContractMidas(args)

	// call setConfig should return error -> wrong owner address
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "setConfig"
	retCode := sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)

	// call validator smart contract init
	arguments.Function = core.SCDeployInitFunctionName
	arguments.CallerAddr = ownerAddr
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.Ok, retCode)

	// call setConfig return error -> wrong number of arguments
	arguments.Function = "setConfig"
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)

	// call setConfig
	totalSupply := big.NewInt(10000000)
	minStep := big.NewInt(100)
	nodPrice := big.NewInt(20000)
	epoch := big.NewInt(1)
	unjailPrice := big.NewInt(100)
	arguments.Function = "setConfig"
	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), minStep.Bytes(),
		nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.Ok, retCode)

	validatorConfig := sc.getConfig(1)
	require.NotNil(t, validatorConfig)
	require.Equal(t, totalSupply, validatorConfig.TotalSupply)
	require.Equal(t, minStep, validatorConfig.MinStep)
	require.Equal(t, nodPrice, validatorConfig.NodePrice)
	require.Equal(t, unjailPrice, validatorConfig.UnJailPrice)
	require.Equal(t, minStakeValue, validatorConfig.MinStakeValue)
}

func TestValidatorStakingSCMidas_SetConfig_InvalidParameters(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddress")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)

	// call setConfig should return error -> wrong owner address
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "setConfig"

	// call validator smart contract init
	arguments.Function = core.SCDeployInitFunctionName
	arguments.CallerAddr = ownerAddr
	_ = sc.Execute(arguments)

	totalSupply := big.NewInt(10000000)
	minStep := big.NewInt(100)
	nodPrice := big.NewInt(20000)
	epoch := big.NewInt(1)
	unjailPrice := big.NewInt(100)
	arguments.Function = "setConfig"

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), zero.Bytes(), minStep.Bytes(),
		nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode := sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidGenesisTotalSupply.Error()))

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), zero.Bytes(),
		nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidMinStepValue.Error()))

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), minStep.Bytes(),
		zero.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNodePrice.Error()))

	arguments.Arguments = [][]byte{minStakeValue.Bytes(), totalSupply.Bytes(), minStep.Bytes(),
		nodPrice.Bytes(), zero.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)
	require.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidUnJailCost.Error()))
}

func TestValidatorStakingSCMidas_getBlsStatusWrongCaller(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "getBlsKeysStatus"
	arguments.CallerAddr = []byte("wrong caller")

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "this is only a view function"))
}

func TestValidatorStakingSCMidas_getBlsStatusWrongNumOfArguments(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("validator")
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "number of arguments must be equal to 1"))
}

func TestValidatorStakingSCMidas_getBlsStatusWrongRegistrationData(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)

	wrongStorageEntry := make(map[string][]byte)
	wrongStorageEntry["erdKey"] = []byte("entry val")
	eei.storageUpdate["addr"] = wrongStorageEntry
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("validator")
	arguments.Arguments = append(arguments.Arguments, []byte("erdKey"))
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot get or create registration data: error "))
}

func TestValidatorStakingSCMidas_getBlsStatusNoBlsKeys(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("validator")
	arguments.Function = "getBlsKeysStatus"
	arguments.Arguments = append(arguments.Arguments, []byte("erd key"))

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, strings.Contains(eei.returnMessage, "no bls keys"))
}

func TestValidatorStakingSCMidas_getBlsStatusShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)

	firstAddr := "addr 1"
	secondAddr := "addr 2"
	validatorData := ValidatorDataV2{
		BlsPubKeys: [][]byte{[]byte(firstAddr), []byte(secondAddr)},
	}
	serializedValidatorData, _ := args.Marshalizer.Marshal(validatorData)

	registrationData1 := &StakedDataV2_0{
		Staked:        true,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: []byte("rewards addr"),
		JailedRound:   math.MaxUint64,
		StakedNonce:   math.MaxUint64,
	}
	serializedRegistrationData1, _ := args.Marshalizer.Marshal(registrationData1)

	registrationData2 := &StakedDataV2_0{
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: []byte("rewards addr"),
		JailedRound:   math.MaxUint64,
		StakedNonce:   math.MaxUint64,
	}
	serializedRegistrationData2, _ := args.Marshalizer.Marshal(registrationData2)

	storageEntry := make(map[string][]byte)
	storageEntry["erdKey"] = serializedValidatorData
	eei.storageUpdate["addr"] = storageEntry

	stakingEntry := make(map[string][]byte)
	stakingEntry[firstAddr] = serializedRegistrationData1
	stakingEntry[secondAddr] = serializedRegistrationData2
	eei.storageUpdate["staking"] = stakingEntry
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("validator")
	arguments.Arguments = append(arguments.Arguments, []byte("erdKey"))
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, returnCode)

	output := eei.CreateVMOutput()
	assert.Equal(t, 4, len(output.ReturnData))
	assert.Equal(t, []byte(firstAddr), output.ReturnData[0])
	assert.Equal(t, []byte("staked"), output.ReturnData[1])
	assert.Equal(t, []byte(secondAddr), output.ReturnData[2])
	assert.Equal(t, []byte("unStaked"), output.ReturnData[3])
}

func TestValidatorStakingSCMidas_getBlsStatusShouldWorkEvenIfAnErrorOccursForOneOfTheBlsKeys(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)

	firstAddr := "addr 1"
	secondAddr := "addr 2"
	validatorData := ValidatorDataV2{
		BlsPubKeys: [][]byte{[]byte(firstAddr), []byte(secondAddr)},
	}
	serializedValidatorData, _ := args.Marshalizer.Marshal(validatorData)

	registrationData := &StakedDataV2_0{
		Staked:        true,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: []byte("rewards addr"),
		JailedRound:   math.MaxUint64,
		StakedNonce:   math.MaxUint64,
	}
	serializedRegistrationData, _ := args.Marshalizer.Marshal(registrationData)

	storageEntry := make(map[string][]byte)
	storageEntry["erdKey"] = serializedValidatorData
	eei.storageUpdate["addr"] = storageEntry

	stakingEntry := make(map[string][]byte)
	stakingEntry[firstAddr] = []byte("wrong data for first bls key")
	stakingEntry[secondAddr] = serializedRegistrationData
	eei.storageUpdate["staking"] = stakingEntry
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = []byte("validator")
	arguments.Arguments = append(arguments.Arguments, []byte("erdKey"))
	arguments.Function = "getBlsKeysStatus"

	returnCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, returnCode)

	output := eei.CreateVMOutput()
	assert.Equal(t, 2, len(output.ReturnData))
	assert.Equal(t, []byte(secondAddr), output.ReturnData[0])
	assert.Equal(t, []byte("staked"), output.ReturnData[1])
}

func TestStakingValidatorSCMidas_UnstakeTokensNotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1).Bytes()}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_UnstakeTokensInvalidArgumentsShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	registrationData := &ValidatorDataV2{RewardAddress: caller}
	marshaledData, _ := args.Marshalizer.Marshal(registrationData)
	eei.SetStorage(caller, marshaledData)
	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{[]byte("a")}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid number of arguments: expected 2, got 1", vmOutput.ReturnMessage)

	eei = createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller = []byte("caller")
	sc, _ = NewValidatorSmartContractMidas(args)

	eei.SetStorage(caller, marshaledData)
	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{[]byte("a"), []byte("b"), []byte("c")}, zero, vmcommon.UserError)
	vmOutput = eei.CreateVMOutput()
	assert.Equal(t, "invalid number of arguments: expected 2, got 3", vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_UnstakeTokensWithCallValueShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1).Bytes()}, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.TransactionValueMustBeZero, vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_UnstakeTokensOtherCallerShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, caller, [][]byte{big.NewInt(1).Bytes()}, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "unStakeTokens function not allowed to be called by address "+string(caller), vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_UnstakeTokensOverMaxShouldUnstake(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 11).Bytes()}, zero, vmcommon.Ok)

	registrationData, _ := sc.getOrCreateRegistrationData(caller)
	assert.Equal(t, 1, len(registrationData.UnstakedInfo))
	assert.True(t, big.NewInt(999).Cmp(registrationData.TotalStakeValue) == 0)
	assert.True(t, registrationData.UnstakedInfo[0].UnstakedValue.Cmp(big.NewInt(11)) == 0)
}

func TestStakingValidatorSCMidas_UnstakeTokensUnderMinimumAllowedShouldErr(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.MinUnstakeTokensValue = "2"
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 1).Bytes()}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.True(t, strings.Contains(vmOutput.ReturnMessage, "can not unstake the provided value either because is under the minimum threshold"))
}

func TestStakingValidatorSCMidas_UnstakeAllTokensWithActiveNodesShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startNonce := uint64(56)
	nonce := startNonce
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			nonce++
			return nonce
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.MinDeposit = "1000"
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 11).Bytes()}, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.True(t, strings.Contains(vmOutput.ReturnMessage, "cannot unStake tokens, the validator would remain without min deposit, nodes are still active"))
}

func TestStakingValidatorSCMidas_UnstakeTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startEpoch := uint32(56)
	epoch := startEpoch
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			epoch++
			return epoch
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 1).Bytes()}, zero, vmcommon.Ok)
	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 1 - 2).Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1007),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(1),
			},
			{
				UnstakedEpoch: startEpoch + 2,
				UnstakedValue: big.NewInt(2),
			},
		},
		TotalUnstaked: big.NewInt(3),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnstakeTokensHavingUnstakedShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	startEpoch := uint32(32)
	epoch := startEpoch
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			epoch++
			return epoch
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: 1,
					UnstakedValue: big.NewInt(5),
				},
			},
			TotalUnstaked: big.NewInt(5),
		},
	)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 6).Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1004),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: 1,
				UnstakedValue: big.NewInt(5),
			},
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(6),
			},
		},
		TotalUnstaked: big.NewInt(11),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnstakeAllTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1010),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResultMidas(t, "unStakeTokens", sc, AbstractStakingSCAddress, [][]byte{caller, big.NewInt(1010 - 10).Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch + unbondPeriod,
				UnstakedValue: big.NewInt(10),
			},
		},
		TotalUnstaked: big.NewInt(10),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnbondTokensNotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_UnbondTokensOneArgument(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedEpoch: startEpoch - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(3),
				},
				{
					UnstakedEpoch: startEpoch + 1,
					UnstakedValue: big.NewInt(4),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	unBondRequest := big.NewInt(4)
	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, [][]byte{unBondRequest.Bytes()}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch,
				UnstakedValue: big.NewInt(2),
			},
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(4),
			},
		},
		TotalUnstaked: big.NewInt(6),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)
	assert.Equal(t, expected, recovered)

	//outTransferValue := eei.outputAccounts[string(caller)].OutputTransfers[0].Value
	//assert.True(t, unBondRequest.Cmp(outTransferValue) == 0)
}

func TestStakingValidatorSCMidas_UnbondTokensWithCallValueShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, nil, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.TransactionValueMustBeZero, vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_UnBondTokensV1ShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	enableEpochsHandler.RemoveActiveFlags(common.UnBondTokensV2Flag)
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedEpoch: startEpoch - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(3),
				},
				{
					UnstakedEpoch: startEpoch + 1,
					UnstakedValue: big.NewInt(4),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, nil, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(4),
			},
		},
		TotalUnstaked: big.NewInt(4),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnBondTokensV2ShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch + 1,
					UnstakedValue: big.NewInt(4),
				},
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedEpoch: startEpoch - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(3),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, nil, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(4),
			},
		},
		TotalUnstaked: big.NewInt(4),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnBondTokensV2WithTooMuchToUnbondShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch + 1,
					UnstakedValue: big.NewInt(4),
				},
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedEpoch: startEpoch - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(3),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	unbondValueBytes := big.NewInt(7).Bytes()
	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, [][]byte{unbondValueBytes}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(4),
			},
		},
		TotalUnstaked: big.NewInt(4),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnBondTokensV2WithSplitShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch + 1,
					UnstakedValue: big.NewInt(4),
				},
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedEpoch: startEpoch - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(3),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	unbondValueBytes := big.NewInt(2).Bytes()
	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, [][]byte{unbondValueBytes}, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo: []*UnstakedValue{
			{
				UnstakedEpoch: startEpoch + 1,
				UnstakedValue: big.NewInt(4),
			},
			{
				UnstakedEpoch: startEpoch - 1,
				UnstakedValue: big.NewInt(1),
			},
			{
				UnstakedEpoch: startEpoch,
				UnstakedValue: big.NewInt(3),
			},
		},
		TotalUnstaked: big.NewInt(8),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UnBondAllTokensWithMinDepositShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.MinDeposit = "1000"
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(999),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.True(t, strings.Contains(vmOutput.ReturnMessage, "cannot unBond tokens, the validator would remain without min deposit, nodes are still active"))
}

func TestStakingValidatorSCMidas_UnBondAllTokensShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint32(10)
	startEpoch := uint32(56)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return startEpoch + unbondPeriod
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(startEpoch + unbondPeriod)
		},
	}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.StakingSCConfig.UnBondPeriodInEpochs = unbondPeriod
	eei := createVmContextWithStakingScMidas(minStakeValue, uint64(unbondPeriod), blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: big.NewInt(1000),
			LockedStake:     big.NewInt(1000),
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      [][]byte{[]byte("key")},
			NumRegistered:   1,
			UnstakedInfo: []*UnstakedValue{
				{
					UnstakedEpoch: startEpoch - 2,
					UnstakedValue: big.NewInt(1),
				},
				{
					UnstakedEpoch: startEpoch - 1,
					UnstakedValue: big.NewInt(2),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(3),
				},
				{
					UnstakedEpoch: startEpoch,
					UnstakedValue: big.NewInt(4),
				},
			},
			TotalUnstaked: big.NewInt(10),
		},
	)

	callFunctionAndCheckResultMidas(t, "unBondTokens", sc, caller, nil, zero, vmcommon.Ok)

	expected := &ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   caller,
		TotalStakeValue: big.NewInt(1000),
		LockedStake:     big.NewInt(1000),
		MaxStakePerNode: big.NewInt(0),
		BlsPubKeys:      [][]byte{[]byte("key")},
		NumRegistered:   1,
		UnstakedInfo:    make([]*UnstakedValue, 0),
		TotalUnstaked:   big.NewInt(0),
	}

	recovered, err := sc.getOrCreateRegistrationData(caller)
	require.Nil(t, err)

	assert.Equal(t, expected, recovered)
}

func TestStakingValidatorSCMidas_UpdateStakingV2NotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "updateStakingV2", sc, args.ValidatorSCAddress, make([][]byte, 0), zero, vmcommon.UserError)

	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_GetTopUpNotEnabledShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "getTopUp", sc, caller, nil, zero, vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "invalid method to call", vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_GetTopUpTotalStakedWithValueShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(1), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.TransactionValueMustBeZero, vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_GetTopUpTotalStakedInsufficientGasShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.Get = 1
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.OutOfGas)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, vm.InsufficientGasLimit, vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_GetTopUpTotalStakedCallerDoesNotExistShouldError(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	callFunctionAndCheckResultMidas(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.UserError)
	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, "caller not registered in staking/validator sc", vmOutput.ReturnMessage)
}

func TestStakingValidatorSCMidas_GetTopUpTotalStakedShouldWork(t *testing.T) {
	t.Parallel()

	minStakeValue := big.NewInt(1000)
	unbondPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unbondPeriod, blockChainHook)
	args.Eei = eei
	caller := []byte("caller")
	sc, _ := NewValidatorSmartContractMidas(args)

	totalStake := big.NewInt(33827)
	lockedStake := big.NewInt(4564)
	_ = sc.saveRegistrationData(
		caller,
		&ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   caller,
			TotalStakeValue: totalStake,
			LockedStake:     lockedStake,
			MaxStakePerNode: big.NewInt(0),
			BlsPubKeys:      make([][]byte, 0),
			NumRegistered:   1,
			UnstakedInfo:    nil,
			TotalUnstaked:   nil,
		},
	)

	callFunctionAndCheckResultMidas(t, "getTotalStakedTopUpStakedBlsKeys", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.Ok)
	vmOutput := eei.CreateVMOutput()

	assert.Equal(t, totalStake.Bytes(), vmOutput.ReturnData[0])
	assert.Equal(t, totalStake.Bytes(), vmOutput.ReturnData[1])

	eei.output = make([][]byte, 0)
	callFunctionAndCheckResultMidas(t, "getTotalStaked", sc, caller, [][]byte{caller}, big.NewInt(0), vmcommon.Ok)
	vmOutput = eei.CreateVMOutput()
	assert.Equal(t, totalStake.String(), string(vmOutput.ReturnData[0]))
}

func TestStakingValidatorSCMidas_UnStakeUnBondFromWaitingList(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey1 := []byte("blsKey1")
	stakerPubKey2 := []byte("blsKey2")
	stakerPubKey3 := []byte("blsKey3")

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100000
		},
	}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stubStaking, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stubStaking.AddActiveFlags(common.StakingV2Flag)
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey1, []byte("signed"), stakerAddress, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey2, []byte("signed"), stakerAddress, big.NewInt(20000).Bytes()}

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey3, []byte("signed"), stakerAddress, big.NewInt(30000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey2, stakerPubKey3, stakerAddress, big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey2, stakerPubKey3, stakerAddress}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	registrationData, _ := sc.getOrCreateRegistrationData(stakerAddress)
	assert.Equal(t, len(registrationData.UnstakedInfo), 0)

	eei.returnMessage = ""
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestStakingValidatorSCMidas_StakeUnStakeUnBondTokensNoNodes(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100000
		},
	}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stubStaking, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	stubStaking.AddActiveFlags(common.StakingV2Flag)
	argsStaking.StakingSCConfig.MaxNumberOfNodesForStake = 1
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig = argsStaking.StakingSCConfig
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(0).Bytes(), vm.DelegationManagerSCAddress, big.NewInt(10000).Bytes()}

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStakeTokens"
	arguments.Arguments = [][]byte{vm.DelegationManagerSCAddress, big.NewInt(0).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unBondTokens"
	arguments.CallerAddr = vm.DelegationManagerSCAddress
	arguments.Arguments = [][]byte{big.NewInt(10000).Bytes()}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestValidatorStakingSCMidas_UnStakeUnBondPaused(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPubKey")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()

	nonce := uint64(0)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			defer func() {
				nonce++
			}()
			return nonce
		},
	}

	args := createMockArgumentsForValidatorSCMidas()
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	eei := createVmContextWithStakingScMidas(minStakeValue, unboundPeriod, blockChainHook)
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)

	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stakeMidas(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	togglePauseUnStakeUnBondMidas(t, sc, true)

	expectedReturnMessage := "unStake/unBond is paused as not enough total staked in protocol"
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey, stakerAddress, big.NewInt(0).Bytes()}

	eei.returnMessage = ""
	errCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unStakeTokens"
	arguments.CallerAddr = AbstractStakingSCAddress
	arguments.Arguments = [][]byte{stakerAddress, big.NewInt(0).Bytes()}
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerAddress}
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unBondTokens"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{stakerPubKey}
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unStakeNodes"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	arguments.Function = "unBondNodes"
	eei.returnMessage = ""
	errCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)
	assert.Equal(t, eei.returnMessage, expectedReturnMessage)

	togglePauseUnStakeUnBondMidas(t, sc, false)
}

func TestValidatorSCMidas_getUnStakedTokensList_InvalidArgumentsCountShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()
	args := createMockArgumentsForValidatorSCMidas()

	retMessage := ""

	eei := &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)

	arguments.Function = "getUnStakedTokensList"
	arguments.CallValue = big.NewInt(10)
	arguments.Arguments = make([][]byte, 0)

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)

	assert.Contains(t, retMessage, "number of arguments")
}

func TestValidatorSCMidas_getUnStakedTokensList_CallValueNotZeroShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInputMidas()
	args := createMockArgumentsForValidatorSCMidas()

	retMessage := ""

	eei := &mock.SystemEIStub{
		AddReturnMessageCalled: func(msg string) {
			retMessage = msg
		},
	}
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)

	arguments.Function = "getUnStakedTokensList"
	arguments.CallValue = big.NewInt(10)
	arguments.Arguments = append(arguments.Arguments, []byte("pubKey"))

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, errCode)

	assert.Equal(t, vm.TransactionValueMustBeZero, retMessage)
}

func TestValidatorSCMidas_getUnStakedTokensList(t *testing.T) {
	t.Parallel()

	currentNonce := uint64(12)
	unBondPeriod := uint32(5)

	eeiFinishedValues := make([][]byte, 0)
	arguments := CreateVmContractCallInputMidas()
	validatorData := createAValidatorData(25000000, 2, 12500000)
	validatorData.UnstakedInfo = append(validatorData.UnstakedInfo,
		&UnstakedValue{
			UnstakedEpoch: 6,
			UnstakedValue: big.NewInt(10),
		},
		&UnstakedValue{
			UnstakedEpoch: 10,
			UnstakedValue: big.NewInt(11),
		},
	)
	validatorDataBytes, _ := json.Marshal(&validatorData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return validatorDataBytes
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Equal(key, arguments.CallerAddr) {
			var validatorDataRecovered ValidatorDataV2
			_ = json.Unmarshal(value, &validatorDataRecovered)
			assert.Equal(t, big.NewInt(26000000), validatorDataRecovered.TotalStakeValue)
		}
	}
	eei.BlockChainHookCalled = func() vm.BlockchainHook {
		return &mock.BlockChainHookStub{
			CurrentNonceCalled: func() uint64 {
				return currentNonce
			},
			CurrentEpochCalled: func() uint32 {
				return uint32(currentNonce)
			},
		}
	}
	eei.FinishCalled = func(value []byte) {
		eeiFinishedValues = append(eeiFinishedValues, value)
	}

	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)

	stakingValidatorSc.unBondPeriodInEpochs = unBondPeriod

	arguments.Function = "getUnStakedTokensList"
	arguments.CallValue = big.NewInt(0)
	arguments.Arguments = append(arguments.Arguments, []byte("pubKey"))

	errCode := stakingValidatorSc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)

	assert.Equal(t, 4, len(eeiFinishedValues))

	expectedValues := [][]byte{
		{10}, // value 10
		{},   // elapsed nonces > unbond period
		{11}, // value 11
		{3},  // number of nonces remaining
	}

	assert.Equal(t, expectedValues, eeiFinishedValues)
}

func TestValidatorSCMidas_getMinUnStakeTokensValueDelegationManagerNotActive(t *testing.T) {
	t.Parallel()

	minUnstakeTokens := 12345
	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	args.StakingSCConfig.MinUnstakeTokensValue = fmt.Sprintf("%d", minUnstakeTokens)

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)
	currentMinUnstakeTokens, err := stakingValidatorSc.getMinUnStakeTokensValue()
	require.Nil(t, err)
	assert.Equal(t, big.NewInt(int64(minUnstakeTokens)), currentMinUnstakeTokens)
}

func TestValidatorSCMidas_getMinUnStakeTokensValueFromDelegationManager(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForValidatorSCMidas()
	storedDelegationManagementData := &DelegationManagement{
		MinDelegationAmount: big.NewInt(457293),
	}
	buff, _ := args.Marshalizer.Marshal(storedDelegationManagementData)

	minUnstakeTokens := 12345
	eei := &mock.SystemEIStub{
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			return buff
		},
	}

	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.DelegationManagerFlag)
	args.StakingSCConfig.MinUnstakeTokensValue = fmt.Sprintf("%d", minUnstakeTokens)

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)
	currentMinUnstakeTokens, err := stakingValidatorSc.getMinUnStakeTokensValue()
	require.Nil(t, err)
	assert.Equal(t, storedDelegationManagementData.MinDelegationAmount, currentMinUnstakeTokens)
}

func TestStakingValidatorSCMidas_checkInputArgsForValidatorToDelegationErrors(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	sc, _ := NewValidatorSmartContractMidas(args)

	arguments := CreateVmContractCallInputMidas()
	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
	returnCode := sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid method to call")

	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
	eei.returnMessage = ""
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid caller address")

	eei.returnMessage = ""
	arguments.CallerAddr = sc.delegationMgrSCAddress
	arguments.CallValue.SetUint64(10)
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "callValue must be 0")

	eei.returnMessage = ""
	arguments.CallValue.SetUint64(0)
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid number of arguments")

	eei.returnMessage = ""
	arguments.Arguments = [][]byte{[]byte("arg1"), []byte("arg2")}
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid argument, wanted an address for the first and second argument")

	eei.returnMessage = ""
	randomAddress := bytes.Repeat([]byte{1}, len(arguments.CallerAddr))
	arguments.Arguments = [][]byte{randomAddress, randomAddress}
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "destination address must be a delegation smart contract")

	eei.returnMessage = ""
	randomAddress = bytes.Repeat([]byte{0}, len(arguments.CallerAddr))
	arguments.Arguments = [][]byte{randomAddress, randomAddress}
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "sender and destination addresses are equal")

	eei.returnMessage = ""
	arguments.Arguments = [][]byte{vm.StakingSCAddress, randomAddress}
	returnCode = sc.checkInputArgsForValidatorToDelegation(arguments)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "sender address must not be a smart contract")
}

func TestStakingValidatorSCMidas_getAndValidateRegistrationDataErrors(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser
	args := createMockArgumentsForValidatorSCMidas()
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	addr := []byte("address")

	eei.returnMessage = ""
	_, returnCode := sc.getAndValidateRegistrationData(addr)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "address does not contain any staked nodes")

	validatorData := &ValidatorDataV2{
		RewardAddress: []byte("a"),
		UnstakedInfo:  []*UnstakedValue{{UnstakedValue: big.NewInt(10)}},
		TotalUnstaked: big.NewInt(10),
		TotalSlashed:  big.NewInt(10),
		BlsPubKeys:    [][]byte{[]byte("first key"), []byte("second key")},
	}
	marshaledData, _ := sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(addr, marshaledData)
	eei.returnMessage = ""
	_, returnCode = sc.getAndValidateRegistrationData(addr)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "reward address mismatch")

	eei.SetStorage(addr, []byte("randomValue"))
	eei.returnMessage = ""
	_, returnCode = sc.getAndValidateRegistrationData(addr)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid character 'r' looking for beginning of value")

	validatorData = &ValidatorDataV2{
		RewardAddress: addr,
		UnstakedInfo:  []*UnstakedValue{{UnstakedValue: big.NewInt(10)}},
		TotalUnstaked: big.NewInt(10),
		TotalSlashed:  big.NewInt(10),
		BlsPubKeys:    [][]byte{[]byte("first key"), []byte("second key")},
	}
	marshaledData, _ = sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(addr, marshaledData)
	eei.returnMessage = ""
	_, returnCode = sc.getAndValidateRegistrationData(addr)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "clean unstaked info before merge")

	validatorData.UnstakedInfo = nil
	marshaledData, _ = sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(addr, marshaledData)
	eei.returnMessage = ""
	_, returnCode = sc.getAndValidateRegistrationData(addr)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "cannot merge with validator who was slashed")

	validatorData.TotalSlashed.SetUint64(0)
	marshaledData, _ = sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(addr, marshaledData)
	eei.returnMessage = ""
	_, returnCode = sc.getAndValidateRegistrationData(addr)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "cannot merge with validator who has unStaked tokens")
}

func TestStakingValidatorSCMidas_ChangeOwnerOfValidatorData(t *testing.T) {
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100000
		},
	}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	enableEpochsHandler, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = vm.ESDTSCAddress
	arguments.Function = "changeOwnerOfValidatorData"
	arguments.Arguments = [][]byte{}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "invalid caller address")

	arguments.CallValue.SetUint64(0)
	arguments.CallerAddr = sc.delegationMgrSCAddress
	randomAddress := bytes.Repeat([]byte{1}, len(arguments.CallerAddr))
	arguments.Arguments = [][]byte{randomAddress, vm.FirstDelegationSCAddress}

	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "address does not contain any staked nodes")

	validatorData := &ValidatorDataV2{
		RewardAddress:   []byte("not a valid address"),
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(0),
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ := sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(randomAddress, marshaledData)
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "reward address mismatch")

	validatorData = &ValidatorDataV2{
		RewardAddress:   randomAddress,
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(0),
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ = sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(randomAddress, marshaledData)

	eei.SetStorage(vm.FirstDelegationSCAddress, []byte("something"))
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "there is already a validator data under the new address")

	eei.SetStorage(vm.FirstDelegationSCAddress, nil)
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "cannot change owner and reward address for a key which is not registered")

	eei.SetStorage(randomAddress, nil)
	stakeMidas(t, sc, stakingSc.stakeValue, randomAddress, randomAddress, []byte("firsstKey"), big.NewInt(1).Bytes())
	stakeMidas(t, sc, stakingSc.stakeValue.Mul(stakingSc.stakeValue, big.NewInt(2)), randomAddress, randomAddress, []byte("secondKey"), big.NewInt(1).Bytes())
	stakeMidas(t, sc, stakingSc.stakeValue.Mul(stakingSc.stakeValue,  big.NewInt(3)), randomAddress, randomAddress, []byte("thirddKey"), big.NewInt(1).Bytes())

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.Equal(t, len(eei.GetStorage(randomAddress)), 0)

	stakedData, _ := sc.getStakedData([]byte("secondKey"))
	assert.Equal(t, stakedData.OwnerAddress, vm.FirstDelegationSCAddress)
	assert.Equal(t, stakedData.RewardAddress, vm.FirstDelegationSCAddress)
}

func TestStakingValidatorSCMidas_MergeValidatorData(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 100000
		},
	}
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	enableEpochsHandler, _ := argsStaking.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	stakingSc, _ := NewStakingSmartContract(argsStaking)
	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	args := createMockArgumentsForValidatorSCMidas()
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewValidatorSmartContractMidas(args)
	arguments := CreateVmContractCallInputMidas()
	arguments.CallerAddr = vm.ESDTSCAddress
	arguments.Function = "mergeValidatorData"
	arguments.Arguments = [][]byte{}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "invalid caller address")

	arguments.CallValue.SetUint64(0)
	arguments.CallerAddr = sc.delegationMgrSCAddress
	randomAddress := bytes.Repeat([]byte{1}, len(arguments.CallerAddr))
	arguments.Arguments = [][]byte{randomAddress, vm.FirstDelegationSCAddress}

	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "address does not contain any staked nodes")

	validatorData := &ValidatorDataV2{
		RewardAddress:   []byte("not a valid address"),
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(0),
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ := sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(randomAddress, marshaledData)

	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "reward address mismatch")

	validatorData = &ValidatorDataV2{
		RewardAddress:   randomAddress,
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(0),
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ = sc.marshalizer.Marshal(validatorData)
	eei.SetStorage(randomAddress, marshaledData)

	eei.SetStorage(vm.FirstDelegationSCAddress, []byte("something"))
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "invalid character 's' looking for beginning of value")

	eei.SetStorage(vm.FirstDelegationSCAddress, nil)
	eei.returnMessage = ""
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, eei.returnMessage, "cannot merge with an empty state")

	eei.SetStorage(randomAddress, nil)
	stakeMidas(t, sc, stakingSc.stakeValue, randomAddress, randomAddress, []byte("firsstKey"), big.NewInt(1).Bytes())
	stakeMidas(t, sc, stakingSc.stakeValue.Mul(stakingSc.stakeValue, big.NewInt(2)), randomAddress, randomAddress, []byte("secondKey"), big.NewInt(1).Bytes())
	stakeMidas(t, sc, stakingSc.stakeValue.Mul(stakingSc.stakeValue, big.NewInt(3)), randomAddress, randomAddress, []byte("thirddKey"), big.NewInt(1).Bytes())

	stakeMidas(t, sc, stakingSc.stakeValue, vm.FirstDelegationSCAddress, vm.FirstDelegationSCAddress, []byte("fourthKey"), big.NewInt(1).Bytes())
	stakeMidas(t, sc, stakingSc.stakeValue.Mul(stakingSc.stakeValue, big.NewInt(2)), vm.FirstDelegationSCAddress, vm.FirstDelegationSCAddress, []byte("fifthhKey"), big.NewInt(1).Bytes())
	stakeMidas(t, sc, stakingSc.stakeValue.Mul(stakingSc.stakeValue, big.NewInt(3)), vm.FirstDelegationSCAddress, vm.FirstDelegationSCAddress, []byte("sixthhKey"), big.NewInt(1).Bytes())

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
	assert.Equal(t, len(eei.GetStorage(randomAddress)), 0)

	stakedData, _ := sc.getStakedData([]byte("secondKey"))
	assert.Equal(t, stakedData.OwnerAddress, vm.FirstDelegationSCAddress)
	assert.Equal(t, stakedData.RewardAddress, vm.FirstDelegationSCAddress)
}

func TestValidatorSCMidas_getMinUnStakeTokensValueFromDelegationManagerMarshalizerFail(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	args := createMockArgumentsForValidatorSCMidas()
	args.Marshalizer = &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return expectedErr
		},
	}

	minUnstakeTokens := 12345
	eei := &mock.SystemEIStub{
		GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
			return make([]byte, 1)
		},
	}

	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.DelegationManagerFlag)
	args.StakingSCConfig.MinUnstakeTokensValue = fmt.Sprintf("%d", minUnstakeTokens)

	stakingValidatorSc, _ := NewValidatorSmartContractMidas(args)
	currentMinUnstakeTokens, err := stakingValidatorSc.getMinUnStakeTokensValue()
	require.Nil(t, currentMinUnstakeTokens)
	assert.Equal(t, expectedErr, err)
}

func createVmContextWithStakingScMidas(stakeValue *big.Int, unboundPeriod uint64, blockChainHook vm.BlockchainHook) *vmContext {
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unboundPeriod
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	return eei
}

func createVmContextWithStakingScMidasWithRealAddresses(stakeValue *big.Int, unboundPeriod uint64, blockChainHook vm.BlockchainHook) *vmContext {
	atArgParser := parsers.NewCallArgsParser()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = atArgParser

	argsStaking := createMockStakingScArgumentsWithRealSystemScAddresses()
	argsStaking.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unboundPeriod
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSc, nil
	}})

	return eei
}

func doClaimMidas(t *testing.T, asc *validatorSCMidas, stakerAddr, receiverAdd []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "claim"
	arguments.RecipientAddr = receiverAdd
	arguments.CallerAddr = stakerAddr

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func stakeMidas(t *testing.T, asc *validatorSCMidas, totalPower *big.Int, receiverAdd, stakerAddr, stakerPubKey, nodesToRunBytes []byte) {
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "stake"
	arguments.RecipientAddr = receiverAdd
	arguments.CallerAddr = AbstractStakingSCAddress
	arguments.Arguments = [][]byte{nodesToRunBytes, stakerPubKey, []byte("signed"), stakerAddr, totalPower.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode := asc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func togglePauseUnStakeUnBondMidas(t *testing.T, v *validatorSCMidas, value bool) {
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = "unPauseUnStakeUnBond"
	arguments.CallerAddr = v.endOfEpochAddress
	arguments.CallValue = big.NewInt(0)

	if value {
		arguments.Function = "pauseUnStakeUnBond"
	}

	retCode := v.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	assert.Equal(t, value, v.isUnStakeUnBondPaused())
}

func callFunctionAndCheckResultMidas(
	t *testing.T,
	function string,
	asc *validatorSCMidas,
	callerAddr []byte,
	args [][]byte,
	callValue *big.Int,
	expectedCode vmcommon.ReturnCode,
) {
	arguments := CreateVmContractCallInputMidas()
	arguments.Function = function
	arguments.CallerAddr = callerAddr
	arguments.Arguments = args
	arguments.CallValue = callValue

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func CreateVmContractCallInputMidas() *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  AbstractStakingSCAddress,
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("rcpntaddr"),
		Function:      "something",
	}
}
