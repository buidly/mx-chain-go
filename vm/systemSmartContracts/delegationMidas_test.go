package systemSmartContracts

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsForDelegationMidas() ArgsNewDelegationMidas {
	return ArgsNewDelegationMidas{
		ArgsNewDelegation: ArgsNewDelegation{
			DelegationSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 10,
				MaxServiceFee: 200,
			},
			StakingSCConfig: config.StakingSystemSCConfig{
				MinStakeValue:    "10",
				UnJailValue:      "15",
				GenesisNodePrice: "100",
			},
			Eei:                    &mock.SystemEIStub{},
			SigVerifier:            &mock.MessageSignVerifierMock{},
			DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
			StakingSCAddress:       vm.StakingSCAddress,
			ValidatorSCAddress:     vm.ValidatorSCAddress,
			GasCost:                vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
			Marshalizer:            &mock.MarshalizerMock{},
			EndOfEpochAddress:      vm.EndOfEpochAddress,
			GovernanceSCAddress:    vm.GovernanceSCAddress,
			AddTokensAddress:       bytes.Repeat([]byte{1}, 32),
			EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(
				common.DelegationSmartContractFlag,
				common.StakingV2FlagAfterEpoch,
				common.AddTokensToDelegationFlag,
				common.DeleteDelegatorAfterClaimRewardsFlag,
				common.ComputeRewardCheckpointFlag,
				common.ValidatorToDelegationFlag,
				common.ReDelegateBelowMinCheckFlag,
				common.MultiClaimOnDelegationFlag,
			),
		},
		AbstractStakingAddr: AbstractStakingSCAddress,
	}
}

func addValidatorAndStakingScToVmContextMidas(eei *vmContext) {
	validatorArgs := createMockArgumentsForValidatorSCMidas()
	validatorArgs.Eei = eei
	validatorArgs.StakingSCConfig.GenesisNodePrice = "100"
	validatorArgs.StakingSCAddress = vm.StakingSCAddress
	enableEpochsHandler, _ := validatorArgs.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	validatorSc, _ := NewValidatorSmartContractMidas(validatorArgs)

	stakingArgs := createMockStakingScArguments()
	stakingArgs.Eei = eei
	stakingSc, _ := NewStakingSmartContract(stakingArgs)

	eei.inputParser = parsers.NewCallArgsParser()

	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		if bytes.Equal(key, vm.StakingSCAddress) {
			return stakingSc, nil
		}

		if bytes.Equal(key, vm.ValidatorSCAddress) {
			enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
			_ = validatorSc.saveRegistrationData([]byte("addr"), &ValidatorDataV2{
				RewardAddress:   []byte("rewardAddr"),
				TotalStakeValue: big.NewInt(1000),
				LockedStake:     big.NewInt(500),
				BlsPubKeys:      [][]byte{[]byte("blsKey1"), []byte("blsKey2")},
				TotalUnstaked:   big.NewInt(150),
				UnstakedInfo: []*UnstakedValue{
					{
						UnstakedEpoch: 10,
						UnstakedValue: big.NewInt(60),
					},
					{
						UnstakedEpoch: 50,
						UnstakedValue: big.NewInt(80),
					},
				},
				NumRegistered: 2,
			})
			validatorSc.unBondPeriod = 50
			return validatorSc, nil
		}

		return nil, nil
	}})
}

func getDefaultVmInputForFuncMidas(funcName string, args [][]byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     AbstractStakingSCAddress,
			Arguments:      args,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      funcName,
	}
}

func createDelegationManagerConfigMidas(eei *vmContext, marshalizer marshal.Marshalizer, minDelegationAmount *big.Int) {
	cfg := &DelegationManagement{
		MinDelegationAmount: minDelegationAmount,
	}

	marshaledData, _ := marshalizer.Marshal(cfg)
	eei.SetStorageForAddress(vm.DelegationManagerSCAddress, []byte(delegationManagementKey), marshaledData)
}

func createDelegationContractAndEEIMidas() (*delegationMidas, *vmContext) {
	args := createMockArgumentsForDelegationMidas()
	eei, _ := NewVMContext(VMContextArgs{
		BlockChainHook: &mock.BlockChainHookStub{
			CurrentEpochCalled: func() uint32 {
				return 2
			},
		},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB:      &stateMock.AccountsStub{},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	})
	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}}
	_ = eei.SetSystemSCContainer(systemSCContainerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSCMidas(args)
	return d, eei
}

func TestNewDelegationSystemSCMidas_NilSystemEnvironmentShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.Eei = nil

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewDelegationSystemSCMidas_InvalidStakingSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegationMidas()
	args.StakingSCAddress = []byte{}

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSCMidas_InvalidValidatorSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegationMidas()
	args.ValidatorSCAddress = []byte{}

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSCMidas_InvalidDelegationMgrSCAddrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	args := createMockArgumentsForDelegationMidas()
	args.DelegationMgrSCAddress = []byte{}

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSCMidas_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.Marshalizer = nil

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewDelegationSystemSCMidas_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.EnableEpochsHandler = nil

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
}

func TestNewDelegationSystemSCMidas_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewDelegationSystemSCMidas_NilSigVerifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.SigVerifier = nil

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, vm.ErrNilMessageSignVerifier, err)
}

func TestNewDelegationSystemSCMidas_InvalidUnJailValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.StakingSCConfig.UnJailValue = "-1"
	expectedErr := fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSCMidas_InvalidMinStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.StakingSCConfig.MinStakeValue = "-1"
	expectedErr := fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSCMidas_InvalidGenesisNodePriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	args.StakingSCConfig.GenesisNodePrice = "-1"
	expectedErr := fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)

	d, err := NewDelegationSystemSCMidas(args)
	assert.Nil(t, d)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationSystemSCMidas_OkParamsShouldWork(t *testing.T) {
	t.Parallel()

	d, err := NewDelegationSystemSCMidas(createMockArgumentsForDelegationMidas())
	assert.Nil(t, err)
	assert.False(t, check.IfNil(d))
}

func TestDelegationSystemSCMidas_ExecuteNilArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(nil)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInputArgsIsNil.Error()))
}

func TestDelegationSystemSCMidas_ExecuteDelegationDisabledShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	d, _ := NewDelegationSystemSCMidas(args)
	enableEpochsHandler.RemoveActiveFlags(common.DelegationSmartContractFlag)
	vmInput := getDefaultVmInputForFuncMidas("addNodes", [][]byte{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "delegation contract is not enabled"))
}

func TestDelegationSystemSCMidas_ExecuteInitScAlreadyPresentShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas(core.SCDeployInitFunctionName, [][]byte{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "smart contract was already initialized"))
}

func TestDelegationSystemSCMidas_ExecuteInitWrongNumOfArgs(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas(core.SCDeployInitFunctionName, [][]byte{[]byte("maxDelegationCap")})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments to init delegation contract"))
}

func TestDelegationSystemSCMidas_ExecuteInitShouldWork(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddr")
	maxDelegationCap := []byte{250}
	serviceFee := []byte{10}
	createdEpoch := uint32(150)
	callValue := big.NewInt(0)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return createdEpoch
		},
		CurrentNonceCalled: func() uint64 {
			return uint64(createdEpoch)
		},
	}
	createDelegationManagerConfigMidas(eei, args.Marshalizer, callValue)
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = 20
	args.StakingSCConfig.UnBondPeriodInEpochs = 20
	_ = eei.SetSystemSCContainer(
		createSystemSCContainer(eei),
	)

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas(core.SCDeployInitFunctionName, [][]byte{maxDelegationCap, serviceFee})
	vmInput.CallValue = callValue
	vmInput.RecipientAddr = createNewAddress(vm.FirstDelegationSCAddress)
	vmInput.CallerAddr = ownerAddr
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	retrievedOwnerAddress := d.eei.GetStorage([]byte(ownerKey))
	retrievedServiceFee := d.eei.GetStorage([]byte(serviceFeeKey))

	dConf, err := d.getDelegationContractConfig()
	assert.Nil(t, err)
	assert.Equal(t, ownerAddr, retrievedOwnerAddress)
	assert.Equal(t, big.NewInt(250), dConf.MaxDelegationCap)
	assert.Equal(t, []byte{10}, retrievedServiceFee)
	assert.Equal(t, uint64(createdEpoch), dConf.CreatedNonce)
	assert.Equal(t, big.NewInt(20).Uint64(), uint64(dConf.UnBondPeriodInEpochs))

	_, err = d.getDelegationStatus()
	assert.NotNil(t, err)

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	_, err = d.getFund(fundKey)
	assert.NotNil(t, err)

	dGlobalFund, err := d.getGlobalFundData()
	assert.Nil(t, err)
	assert.Equal(t, callValue, dGlobalFund.TotalActive)
	assert.Equal(t, big.NewInt(0), dGlobalFund.TotalUnStaked)

	isNew, _, err := d.getOrCreateDelegatorData(ownerAddr)
	assert.Nil(t, err)
	assert.True(t, isNew)
}

func TestDelegationSystemSCMidas_ExecuteAddNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	blsKey3 := []byte("blsKey3")
	signature := []byte("sig1")
	callValue := big.NewInt(130)
	vmInputArgs := make([][]byte, 0)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	sigVerifier := &mock.MessageSignVerifierMock{}
	sigVerifier.VerifyCalled = func(message []byte, signedMessage []byte, pubKey []byte) error {
		return errors.New("verify error")
	}
	args.SigVerifier = sigVerifier

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("addNodes", vmInputArgs)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey1, blsKey1}...)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey1, blsKey2, blsKey3}...)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "arguments must be of pair length - BLSKey and signedMessage"))

	vmInput.Arguments = append(vmInputArgs, [][]byte{blsKey1, signature}...)
	eei.gasRemaining = 10
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidBLSKeys.Error()))
	assert.Equal(t, blsKey1, eei.output[0])
	assert.Equal(t, []byte{invalidKey}, eei.output[1])
}

func TestDelegationSystemSCMidas_ExecuteAddNodesStakedKeyAlreadyExistsInStakedKeysShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	sig := []byte("sig1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	key := &NodesData{
		BLSKey: blsKey,
	}
	dStatus := &DelegationContractStatus{
		StakedKeys: []*NodesData{key},
	}
	_ = d.saveDelegationStatus(dStatus)

	vmInput := getDefaultVmInputForFuncMidas("addNodes", [][]byte{blsKey, sig})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteAddNodesStakedKeyAlreadyExistsInUnStakedKeysShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	sig := []byte("sig1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	key := &NodesData{
		BLSKey: blsKey,
	}
	dStatus := &DelegationContractStatus{
		UnStakedKeys: []*NodesData{key},
	}
	_ = d.saveDelegationStatus(dStatus)

	vmInput := getDefaultVmInputForFuncMidas("addNodes", [][]byte{blsKey, sig})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteAddNodesStakedKeyAlreadyExistsInNotStakedKeysShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	sig := []byte("sig1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	key := &NodesData{
		BLSKey: blsKey,
	}
	dStatus := &DelegationContractStatus{
		NotStakedKeys: []*NodesData{key},
	}
	_ = d.saveDelegationStatus(dStatus)

	vmInput := getDefaultVmInputForFuncMidas("addNodes", [][]byte{blsKey, sig})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteAddNodesShouldSaveAddedKeysAsNotStakedKeys(t *testing.T) {
	t.Parallel()

	blsKeys := [][]byte{[]byte("blsKey1"), []byte("blsKey2")}
	signatures := [][]byte{[]byte("sig1"), []byte("sig2")}
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei
	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("addNodes", [][]byte{blsKeys[0], signatures[0], blsKeys[1], signatures[1]})
	vmInput.CallerAddr = []byte("owner")
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	delegStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 0, len(delegStatus.StakedKeys))
	assert.Equal(t, 2, len(delegStatus.NotStakedKeys))
	assert.Equal(t, blsKeys[0], delegStatus.NotStakedKeys[0].BLSKey)
	assert.Equal(t, signatures[0], delegStatus.NotStakedKeys[0].SignedMsg)
	assert.Equal(t, blsKeys[1], delegStatus.NotStakedKeys[1].BLSKey)
	assert.Equal(t, signatures[1], delegStatus.NotStakedKeys[1].SignedMsg)
}

func TestDelegationSystemSCMidas_ExecuteAddNodesWithNoArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei
	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("addNodes", [][]byte{})
	vmInput.CallerAddr = []byte("owner")
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))
}

func TestDelegationSystemSCMidas_ExecuteRemoveNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("removeNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 10
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteRemoveNodesNotPresentInNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFuncMidas("removeNodes", [][]byte{blsKey})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteRemoveNodesShouldRemoveKeyFromNotStakedKeys(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NotStakedKeys: []*NodesData{key1, key2},
	})

	vmInput := getDefaultVmInputForFuncMidas("removeNodes", [][]byte{blsKey1})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	delegStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 0, len(delegStatus.StakedKeys))
	assert.Equal(t, 1, len(delegStatus.NotStakedKeys))
	assert.Equal(t, blsKey2, delegStatus.NotStakedKeys[0].BLSKey)
}

func TestDelegationSystemSCMidas_ExecuteRemoveNodesWithNoArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei
	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("removeNodes", [][]byte{})
	vmInput.CallerAddr = []byte("owner")
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))
}

func TestDelegationSystemSCMidas_ExecuteStakeNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("stakeNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteStakeNodesNotPresentInNotStakedOrUnStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFuncMidas("stakeNodes", [][]byte{blsKey})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteStakeNodesVerifiesBothUnStakedAndNotStaked(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NotStakedKeys: []*NodesData{key1},
		UnStakedKeys:  []*NodesData{key2},
	})

	vmInput := getDefaultVmInputForFuncMidas("stakeNodes", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	globalFund := &GlobalFundData{
		TotalActive: big.NewInt(10),
	}
	_ = d.saveGlobalFundData(globalFund)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough in total active to stake"))

	globalFund = &GlobalFundData{
		TotalActive: big.NewInt(200),
	}
	_ = d.saveGlobalFundData(globalFund)
	addValidatorAndStakingScToVmContextMidas(eei)

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(200), globalFund.TotalActive)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))
}

func TestDelegationSystemSCMidas_ExecuteDelegateStakeNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	args.GasCost.MetaChainSystemSCsCost.Stake = 1
	eei := createDefaultEei()

	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))
	eei.SetSCAddress(vm.FirstDelegationSCAddress)
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("delegator")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NotStakedKeys: []*NodesData{key1},
		UnStakedKeys:  []*NodesData{key2},
	})
	globalFund := &GlobalFundData{
		TotalActive: big.NewInt(0),
	}
	_ = d.saveGlobalFundData(globalFund)
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:    big.NewInt(1000),
		InitialOwnerFunds:   big.NewInt(100),
		AutomaticActivation: true,
	})
	addValidatorAndStakingScToVmContextMidas(eei)

	delegatorAddr := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("delegate", [][]byte{delegatorAddr, big.NewInt(500).Bytes()})
	vmInput.GasProvided = 10000
	eei.gasRemaining = vmInput.GasProvided

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(500), globalFund.TotalActive)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.StakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.NotStakedKeys))

	vmOutput := eei.CreateVMOutput()
	assert.Equal(t, 6, len(vmOutput.OutputAccounts))

	vmInput.Arguments = [][]byte{delegatorAddr, big.NewInt(1000).Bytes()}
	output = d.Execute(vmInput)
	eei.gasRemaining = vmInput.GasProvided
	assert.Equal(t, vmcommon.Ok, output)

	globalFund, _ = d.getGlobalFundData()
	assert.Equal(t, big.NewInt(1000), globalFund.TotalActive)

	_, delegator, _ := d.getOrCreateDelegatorData(delegatorAddr)
	fund, _ := d.getFund(delegator.ActiveFund)
	assert.Equal(t, fund.Value, big.NewInt(1000))
}

func TestDelegationSystemSCMidas_ExecuteUnStakeNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("unStakeNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteUnStakeNodesNotPresentInStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFuncMidas("unStakeNodes", [][]byte{blsKey})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteUnStakeNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys: []*NodesData{key1, key2},
	})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(100)})
	addValidatorAndStakingScToVmContextMidas(eei)

	validatorMap := map[string][]byte{}
	registrationDataValidator := &ValidatorDataV2{BlsPubKeys: [][]byte{blsKey1, blsKey2}, RewardAddress: []byte("rewardAddr")}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	validatorMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: true}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: true}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	stakingMap[nodesConfigKey] = stkNodes

	eei.storageUpdate[string(args.ValidatorSCAddress)] = validatorMap
	eei.storageUpdate[string(args.StakingSCAddress)] = stakingMap

	vmInput := getDefaultVmInputForFuncMidas("unStakeNodes", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.StakedKeys))
}

func TestDelegationSystemSCMidas_ExecuteUnStakeNodesAtEndOfEpoch(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys: []*NodesData{key1, key2},
	})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(100)})
	validatorArgs := createMockArgumentsForValidatorSC()
	validatorArgs.Eei = eei
	validatorArgs.StakingSCConfig.GenesisNodePrice = "100"
	enableEpochsHandler, _ := validatorArgs.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.AddActiveFlags(common.StakingV2Flag)
	validatorArgs.StakingSCAddress = vm.StakingSCAddress
	validatorSc, _ := NewValidatorSmartContract(validatorArgs)

	stakingArgs := createMockStakingScArguments()
	stakingArgs.Eei = eei
	stakingSc, _ := NewStakingSmartContract(stakingArgs)

	eei.inputParser = parsers.NewCallArgsParser()
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		if bytes.Equal(key, vm.StakingSCAddress) {
			return stakingSc, nil
		}

		if bytes.Equal(key, vm.ValidatorSCAddress) {
			return validatorSc, nil
		}

		return nil, vm.ErrUnknownSystemSmartContract
	}})

	validatorMap := map[string][]byte{}
	registrationDataValidator := &ValidatorDataV2{
		BlsPubKeys:      [][]byte{blsKey1, blsKey2},
		RewardAddress:   []byte("rewardAddr"),
		TotalStakeValue: big.NewInt(1000000),
		NumRegistered:   2,
	}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	validatorMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: false, UnStakedNonce: 5, StakeValue: big.NewInt(0)}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr"), Staked: false, UnStakedNonce: 5, StakeValue: big.NewInt(0)}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	stakingMap[nodesConfigKey] = stkNodes

	eei.storageUpdate[string(args.ValidatorSCAddress)] = validatorMap
	eei.storageUpdate[string(args.StakingSCAddress)] = stakingMap

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 10
	}

	vmInput := getDefaultVmInputForFuncMidas("unStakeAtEndOfEpoch", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallerAddr = args.EndOfEpochAddress
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.UnStakedKeys))
	assert.Equal(t, 0, len(dStatus.StakedKeys))

	vmInput = getDefaultVmInputForFuncMidas("reStakeUnStakedNodes", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = []byte("owner")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ = d.getDelegationStatus()
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
	assert.Equal(t, 2, len(dStatus.StakedKeys))
}

func TestDelegationSystemSCMidas_ExecuteUnBondNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	callValue := big.NewInt(130)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("unBondNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{blsKey, blsKey}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey}
	eei.gasRemaining = 100
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteUnBondNodesNotPresentInUnStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFuncMidas("unBondNodes", [][]byte{blsKey})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteUnBondNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		UnStakedKeys: []*NodesData{key1, key2},
	})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(100)})
	addValidatorAndStakingScToVmContextMidas(eei)

	registrationDataValidator := &ValidatorDataV2{
		BlsPubKeys:      [][]byte{blsKey1, blsKey2},
		RewardAddress:   []byte("rewardAddr"),
		LockedStake:     big.NewInt(300),
		TotalStakeValue: big.NewInt(1000),
		NumRegistered:   2,
	}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	eei.SetStorageForAddress(vm.ValidatorSCAddress, []byte("addr"), regData)

	registrationDataStaking := &StakedDataV2_0{RewardAddress: []byte("rewardAddr")}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	eei.SetStorageForAddress(vm.StakingSCAddress, []byte("blsKey1"), regData)

	registrationDataStaking2 := &StakedDataV2_0{RewardAddress: []byte("rewardAddr")}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	eei.SetStorageForAddress(vm.StakingSCAddress, []byte("blsKey2"), regData)

	stakingNodesConfig := &StakingNodesConfig{StakedNodes: 5}
	stkNodes, _ := d.marshalizer.Marshal(stakingNodesConfig)
	eei.SetStorageForAddress(vm.StakingSCAddress, []byte(nodesConfigKey), stkNodes)

	vmInput := getDefaultVmInputForFuncMidas("unBondNodes", [][]byte{blsKey1, blsKey2})
	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 2, len(dStatus.NotStakedKeys))
	assert.Equal(t, 0, len(dStatus.UnStakedKeys))
}

func TestDelegationSystemSCMidas_ExecuteUnJailNodesUserErrors(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas("unJailNodes", [][]byte{})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not enough arguments"))

	vmInput.Arguments = [][]byte{blsKey, []byte("ownerAddr")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	vmInput.Arguments = [][]byte{blsKey, blsKey, []byte("ownerAddr")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.CallerAddr = []byte("caller")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "unJailNodes function not allowed to be called by address "+string(vmInput.CallerAddr)))
}

func TestDelegationSystemSCMidas_ExecuteUnJailNodesNotPresentInStakedOrUnStakedShouldErr(t *testing.T) {
	t.Parallel()

	blsKey := []byte("blsKey1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{})

	vmInput := getDefaultVmInputForFuncMidas("unJailNodes", [][]byte{blsKey, []byte("owner")})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrBLSPublicKeyMismatch.Error()))
}

func TestDelegationSystemSCMidas_ExecuteUnJailNodesNotDelegatorShouldErr(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei
	vmInput := getDefaultVmInputForFuncMidas("unJailNodes", [][]byte{blsKey1, blsKey2, []byte("notDelegator")})
	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:   []*NodesData{key1},
		UnStakedKeys: []*NodesData{key2},
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "not a delegator"))
}

func TestDelegationSystemSCMidas_ExecuteUnJailNodes(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("unJailNodes", [][]byte{blsKey1, blsKey2, []byte("delegator")})

	key1 := &NodesData{BLSKey: blsKey1}
	key2 := &NodesData{BLSKey: blsKey2}
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:   []*NodesData{key1},
		UnStakedKeys: []*NodesData{key2},
		NumUsers:     1,
	})
	addValidatorAndStakingScToVmContextMidas(eei)

	_ = d.saveDelegatorData([]byte("delegator"), &DelegatorData{ActiveFund: []byte("someFund"), UnClaimedRewards: big.NewInt(0), TotalCumulatedRewards: big.NewInt(0)})

	validatorMap := map[string][]byte{}
	registrationDataValidator := &ValidatorDataV2{
		BlsPubKeys:    [][]byte{blsKey1, blsKey2},
		RewardAddress: []byte("rewardAddr"),
	}
	regData, _ := d.marshalizer.Marshal(registrationDataValidator)
	validatorMap["addr"] = regData

	stakingMap := map[string][]byte{}
	registrationDataStaking := &StakedDataV2_0{
		RewardAddress: []byte("rewardAddr"),
		Jailed:        true,
	}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking)
	stakingMap["blsKey1"] = regData

	registrationDataStaking2 := &StakedDataV2_0{
		RewardAddress: []byte("rewardAddr"),
		Jailed:        true,
	}
	regData, _ = d.marshalizer.Marshal(registrationDataStaking2)
	stakingMap["blsKey2"] = regData

	eei.storageUpdate[string(args.ValidatorSCAddress)] = validatorMap
	eei.storageUpdate[string(args.StakingSCAddress)] = stakingMap

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDelegationSystemSCMidas_ExecuteDelegateUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("delegate", [][]byte{delegator, big.NewInt(0).Bytes()})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "delegate value must be higher than minDelegationAmount"))

	vmInput.Arguments = [][]byte{delegator, big.NewInt(100).Bytes()}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.CallValue = big.NewInt(15)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid call value"))

	vmInput.CallerAddr = delegator
	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.True(t, strings.Contains(eei.returnMessage, "delegate function not allowed to be called by address "+string(delegator)))
}

func TestDelegationSystemSCMidas_ExecuteDelegateWrongInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFuncMidas("delegate", [][]byte{[]byte("delegator"), big.NewInt(15).Bytes()})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteDelegate(t *testing.T) {
	t.Parallel()

	delegator1 := []byte("delegator1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	addValidatorAndStakingScToVmContextMidas(eei)
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFuncMidas("delegate", [][]byte{delegator1, big.NewInt(15).Bytes()})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:  big.NewInt(100),
		InitialOwnerFunds: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(100),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "total delegation cap reached"))

	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(0),
	})

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	dFund, _ := d.getFund(fundKey)
	assert.Equal(t, big.NewInt(15), dFund.Value)
	assert.Equal(t, delegator1, dFund.Address)
	assert.Equal(t, active, dFund.Type)

	dGlobalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(15), dGlobalFund.TotalActive)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, uint64(1), dStatus.NumUsers)

	_, dData, _ := d.getOrCreateDelegatorData(delegator1)
	assert.Equal(t, fundKey, dData.ActiveFund)
}

func TestDelegationSystemSCMidas_ExecuteDelegateFailsWhenGettingDelegationManagement(t *testing.T) {
	t.Parallel()

	delegator1 := []byte("delegator1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	addValidatorAndStakingScToVmContextMidas(eei)

	vmInput := getDefaultVmInputForFuncMidas("delegate", [][]byte{delegator1, big.NewInt(15).Bytes()})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "error getting minimum delegation amount data was not found under requested key"))
}

func TestDelegationSystemSCMidas_ExecuteDelegateOtherCallerShouldErr(t *testing.T) {
	t.Parallel()

	delegator1 := []byte("delegator1")
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	addValidatorAndStakingScToVmContextMidas(eei)
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFuncMidas("delegate", [][]byte{})
	vmInput.CallValue = big.NewInt(15)
	vmInput.CallerAddr = delegator1
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:  big.NewInt(100),
		InitialOwnerFunds: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(100),
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Equal(t, "delegate function not allowed to be called by address "+string(delegator1), eei.returnMessage)
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))

	vmInput.CallerAddr = []byte("owner")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "unDelegate function not allowed to be called by address "+string(vmInput.CallerAddr)))
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateUserErrorsWhenAnInvalidValueToDelegateWasProvided(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegator := []byte("delegator")
	negativeValueToUndelegate := big.NewInt(100)
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, negativeValueToUndelegate.Bytes()})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate"))
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateUserErrorsWhenGettingMinimumDelegationAmount(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, {20}})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "error getting minimum delegation amount"))
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateUserNotDelegatorOrNoActiveFundShouldErr(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, {100}})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller is not a delegator"))

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund: fundKey,
	})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getFund %s", vm.ErrDataNotFoundUnderKey, string(fundKey))
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(50),
	})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate"))

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(11),
	})
	vmInput.Arguments = [][]byte{delegator, {1}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid value to undelegate - need to undelegate all - do not leave dust behind"))
}

func TestDelegationSystemSCMidas_ExecuteUnDelegatePartOfFunds(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	args.Eei = eei
	addValidatorAndStakingScToVmContextMidas(eei)
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, {20}})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund(fundKey)
	assert.Equal(t, big.NewInt(20), dFund.Value)
	assert.Equal(t, active, dFund.Type)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(80), dFund.Value)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, delegator, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(20), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(80), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 50,
	})

	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 100
	}

	vmInput.Arguments = [][]byte{delegator, {0}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	eei.output = make([][]byte, 0)
	vmInput = getDefaultVmInputForFuncMidas("getUserUnDelegatedList", [][]byte{})
	vmInput.Arguments = [][]byte{delegator}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	assert.Equal(t, 4, len(eei.output))
	assert.Equal(t, eei.output[0], []byte{80})
	assert.Equal(t, eei.output[1], []byte{})
	assert.Equal(t, eei.output[2], []byte{20})
	assert.Equal(t, eei.output[3], []byte{50})
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateAllFunds(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	addValidatorAndStakingScToVmContextMidas(eei)
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, {0}})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund(fundKey)
	assert.Nil(t, dFund)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(100), dFund.Value)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, delegator, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateAllFundsAsOwner(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei
	addValidatorAndStakingScToVmContextMidas(eei)
	minDelegationAmount := big.NewInt(10)
	createDelegationManagerConfigMidas(eei, args.Marshalizer, minDelegationAmount)

	delegator := []byte("ownerAsDelegator")
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, {0}})
	d, _ := NewDelegationSystemSCMidas(args)

	d.eei.SetStorage([]byte(ownerKey), delegator)
	_ = d.saveDelegationContractConfig(&DelegationConfig{InitialOwnerFunds: big.NewInt(100), MaxDelegationCap: big.NewInt(0)})
	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	_ = d.saveDelegationStatus(&DelegationContractStatus{StakedKeys: []*NodesData{{BLSKey: []byte("blsKey"), SignedMsg: []byte("someMsg")}}})
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	_ = d.saveDelegationStatus(&DelegationContractStatus{})
	vmInput.Arguments = [][]byte{delegator, {50}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	vmInput.Arguments = [][]byte{delegator, {0}}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dFund, _ := d.getFund(fundKey)
	assert.Nil(t, dFund)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(100), dFund.Value)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, delegator, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])

	managementData := &DelegationManagement{
		MinDeposit:          big.NewInt(10),
		MinDelegationAmount: minDelegationAmount,
	}
	marshaledData, _ := d.marshalizer.Marshal(managementData)
	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)

	vmInput.Function = "delegate"
	vmInput.Arguments = [][]byte{delegator, big.NewInt(1000).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDelegationSystemSCMidas_ExecuteUnDelegateMultipleTimesSameAndDiffEpochAndWithdraw(t *testing.T) {
	t.Parallel()

	fundKey := append([]byte(fundKeyPrefix), []byte{1}...)
	nextFundKey := append([]byte(fundKeyPrefix), []byte{2}...)
	thirdFundKey := append([]byte(fundKeyPrefix), []byte{3}...)
	args := createMockArgumentsForDelegationMidas()
	currentEpoch := uint32(10)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriodInEpochs = 10
	addValidatorAndStakingScToVmContextMidas(eei)
	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("unDelegate", [][]byte{delegator, {100}})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		UnStakedFunds:         [][]byte{},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(100),
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive:   big.NewInt(100),
		TotalUnStaked: big.NewInt(0),
	})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:            big.NewInt(0),
		InitialOwnerFunds:           big.NewInt(0),
		AutomaticActivation:         false,
		ChangeableServiceFee:        false,
		CreatedNonce:                0,
		UnBondPeriodInEpochs:        args.StakingSCConfig.UnBondPeriodInEpochs,
		CheckCapOnReDelegateRewards: false,
	})
	_ = d.saveDelegationStatus(&DelegationContractStatus{NumUsers: 10})
	d.eei.SetStorage([]byte(lastFundKey), fundKey)

	for i := 0; i < 5; i++ {
		vmInput.Arguments = [][]byte{delegator, big.NewInt(int64(90 - i*10)).Bytes()}
		output := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, output)
	}
	currentEpoch += 1
	for i := 0; i < 5; i++ {
		vmInput.Arguments = [][]byte{delegator, big.NewInt(int64(40 - i*10)).Bytes()}
		output := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, output)
	}

	dFund, _ := d.getFund(fundKey)
	assert.Nil(t, dFund)

	dFund, _ = d.getFund(nextFundKey)
	assert.Equal(t, big.NewInt(50), dFund.Value)
	assert.Equal(t, currentEpoch-1, dFund.Epoch)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, delegator, dFund.Address)

	dFund, _ = d.getFund(thirdFundKey)
	assert.Equal(t, big.NewInt(50), dFund.Value)
	assert.Equal(t, currentEpoch, dFund.Epoch)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, delegator, dFund.Address)

	globalFund, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(0), globalFund.TotalActive)
	assert.Equal(t, big.NewInt(100), globalFund.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, 2, len(dData.UnStakedFunds))
	assert.Equal(t, nextFundKey, dData.UnStakedFunds[0])
	assert.Equal(t, thirdFundKey, dData.UnStakedFunds[1])

	currentEpoch += d.unBondPeriodInEpochs - 1
	vmInput.Function = "withdraw"
	vmInput.Arguments = [][]byte{delegator}
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, dData, _ = d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, thirdFundKey, dData.UnStakedFunds[0])

	dFund, _ = d.getFund(thirdFundKey)
	assert.Equal(t, big.NewInt(50), dFund.Value)
	assert.Equal(t, currentEpoch-d.unBondPeriodInEpochs+1, dFund.Epoch)
	assert.Equal(t, unStaked, dFund.Type)
	assert.Equal(t, delegator, dFund.Address)

	currentEpoch += 1
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	isNew, _, _ := d.getOrCreateDelegatorData(delegator)
	assert.True(t, isNew)
}

func TestDelegationSystemSCMidas_ExecuteWithdrawUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("withdraw", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))

	vmInput.Arguments = [][]byte{[]byte("delegator")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller is not a delegator"))

	vmInput.CallerAddr = []byte("caller")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "withdraw function not allowed to be called by address "+string(vmInput.CallerAddr)))
}

func TestDelegationSystemSCMidas_ExecuteWithdrawWrongInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("withdraw", [][]byte{delegator})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteWithdraw(t *testing.T) {
	t.Parallel()

	fundKey1 := []byte{1}
	fundKey2 := []byte{2}
	currentNonce := uint64(60)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return currentNonce
		},
		CurrentEpochCalled: func() uint32 {
			return uint32(currentNonce)
		},
	}
	args.Eei = eei
	addValidatorAndStakingScToVmContext(eei)

	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("withdraw", [][]byte{delegator})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		UnStakedFunds:         [][]byte{fundKey1, fundKey2},
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})
	_ = d.saveFund(fundKey1, &Fund{
		Value:   big.NewInt(60),
		Address: delegator,
		Epoch:   10,
		Type:    unStaked,
	})
	_ = d.saveFund(fundKey2, &Fund{
		Value:   big.NewInt(80),
		Address: delegator,
		Epoch:   50,
		Type:    unStaked,
	})
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 50,
	})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalUnStaked: big.NewInt(140),
		TotalActive:   big.NewInt(0),
	})

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NumUsers: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.Equal(t, eei.returnMessage, "nothing to unBond")

	gFundData, _ := d.getGlobalFundData()
	assert.Equal(t, big.NewInt(80), gFundData.TotalUnStaked)

	_, dData, _ := d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, 1, len(dData.UnStakedFunds))
	assert.Equal(t, fundKey2, dData.UnStakedFunds[0])

	fundKey, _ := d.getFund(fundKey1)
	assert.Nil(t, fundKey)

	_ = d.saveDelegationStatus(&DelegationContractStatus{NumUsers: 2})
	currentNonce = 150
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	isNew, _, _ := d.getOrCreateDelegatorData(delegator)
	assert.True(t, isNew)

	dStatus, _ := d.getDelegationStatus()
	assert.Equal(t, uint64(1), dStatus.NumUsers)
}

func TestDelegationSystemSCMidas_ExecuteChangeServiceFeeUserErrors(t *testing.T) {
	t.Parallel()

	newServiceFee := []byte{50}
	callValue := big.NewInt(15)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("changeServiceFee", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallerAddr = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{newServiceFee, newServiceFee}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{newServiceFee, []byte("wrong arg")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments"))

	vmInput.Arguments = [][]byte{big.NewInt(5).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "new service fee out of bounds"))

	vmInput.Arguments = [][]byte{[]byte("210")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "new service fee out of bounds"))
}

func TestDelegationSystemSCMidas_ExecuteChangeServiceFee(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("changeServiceFee", [][]byte{big.NewInt(70).Bytes()})
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	_ = d.saveGlobalFundData(&GlobalFundData{TotalActive: big.NewInt(0)})
	vmInput.CallerAddr = []byte("owner")

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	retrievedServiceFee := d.eei.GetStorage([]byte(serviceFeeKey))
	assert.Equal(t, []byte{70}, retrievedServiceFee)
}

func TestDelegationSystemSCMidas_ExecuteModifyTotalDelegationCapUserErrors(t *testing.T) {
	t.Parallel()

	newServiceFee := []byte{50}
	callValue := big.NewInt(15)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("ownerAddr")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("modifyTotalDelegationCap", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallerAddr = []byte("owner")
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only owner can call this method"))

	delegationsMap[ownerKey] = []byte("owner")
	vmInput.CallValue = callValue
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.Arguments = [][]byte{newServiceFee, newServiceFee}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrDuplicatesFoundInArguments.Error()))

	vmInput.Arguments = [][]byte{newServiceFee, []byte("wrong arg")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid number of arguments"))

	vmInput.Arguments = [][]byte{newServiceFee}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	vmInput.Arguments = [][]byte{big.NewInt(70).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr = fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationSystemSCMidas_ExecuteModifyTotalDelegationCap(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()

	delegationsMap := map[string][]byte{}
	delegationsMap[ownerKey] = []byte("owner")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("modifyTotalDelegationCap", [][]byte{big.NewInt(500).Bytes()})
	vmInput.CallerAddr = []byte("owner")
	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: big.NewInt(1000),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot make total delegation cap smaller than active"))

	vmInput.Arguments = [][]byte{big.NewInt(1500).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConfig, _ := d.getDelegationContractConfig()
	assert.Equal(t, big.NewInt(1500), dConfig.MaxDelegationCap)

	vmInput.Arguments = [][]byte{big.NewInt(0).Bytes()}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, big.NewInt(0), dConfig.MaxDelegationCap)
}

func TestDelegationMidas_getSuccessAndUnSuccessKeysAllUnSuccess(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("bls1")
	blsKey2 := []byte("bls2")
	returnData := [][]byte{blsKey1, {failed}, blsKey2, {failed}}
	blsKeys := [][]byte{blsKey1, blsKey2}

	okKeys, failedKeys := getSuccessAndUnSuccessKeys(returnData, blsKeys)
	assert.Nil(t, okKeys)
	assert.Equal(t, 2, len(failedKeys))
	assert.Equal(t, blsKey1, failedKeys[0])
	assert.Equal(t, blsKey2, failedKeys[1])
}

func TestDelegationMidas_getSuccessAndUnSuccessKeysAllSuccess(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("bls1")
	blsKey2 := []byte("bls2")
	returnData := [][]byte{blsKey1, {ok}, blsKey2, {ok}}
	blsKeys := [][]byte{blsKey1, blsKey2}

	okKeys, failedKeys := getSuccessAndUnSuccessKeys(returnData, blsKeys)
	assert.Equal(t, 0, len(failedKeys))
	assert.Equal(t, 2, len(okKeys))
	assert.Equal(t, blsKey1, okKeys[0])
	assert.Equal(t, blsKey2, okKeys[1])
}

func TestDelegationMidas_getSuccessAndUnSuccessKeys(t *testing.T) {
	t.Parallel()

	blsKey1 := []byte("bls1")
	blsKey2 := []byte("bls2")
	blsKey3 := []byte("bls3")
	returnData := [][]byte{blsKey1, {ok}, blsKey2, {failed}, blsKey3, {waiting}}
	blsKeys := [][]byte{blsKey1, blsKey2, blsKey3}

	okKeys, failedKeys := getSuccessAndUnSuccessKeys(returnData, blsKeys)
	assert.Equal(t, 2, len(okKeys))
	assert.Equal(t, blsKey1, okKeys[0])
	assert.Equal(t, blsKey3, okKeys[1])

	assert.Equal(t, 1, len(failedKeys))
	assert.Equal(t, blsKey2, failedKeys[0])
}

func TestDelegationMidas_ExecuteUpdateRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("updateRewards", [][]byte{})
	vmInput.CallerAddr = []byte("eoeAddress")
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "only end of epoch address can call this function"))

	vmInput.CallerAddr = vm.EndOfEpochAddress
	vmInput.Arguments = [][]byte{[]byte("arg")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "must call without arguments"))

	vmInput.Arguments = [][]byte{}
	vmInput.CallValue = big.NewInt(-10)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "cannot call with negative value"))
}

func TestDelegationMidas_ExecuteUpdateRewards(t *testing.T) {
	t.Parallel()

	currentEpoch := uint32(15)
	callValue := big.NewInt(20)
	totalActive := big.NewInt(200)
	serviceFee := big.NewInt(100)
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return currentEpoch
		},
	}
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("updateRewards", [][]byte{})
	vmInput.CallValue = callValue
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSCMidas(args)

	d.eei.SetStorage([]byte(totalActiveKey), totalActive.Bytes())
	d.eei.SetStorage([]byte(serviceFeeKey), serviceFee.Bytes())

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	wasPresent, rewardData, err := d.getRewardComputationData(currentEpoch)
	assert.True(t, wasPresent)
	assert.Nil(t, err)
	assert.Equal(t, serviceFee.Uint64(), rewardData.ServiceFee)
	assert.Equal(t, totalActive, rewardData.TotalActive)
	assert.Equal(t, callValue, rewardData.RewardsToDistribute)
}

func TestDelegationMidas_ExecuteClaimRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("claimRewards", [][]byte{{10}})
	d, _ := NewDelegationSystemSCMidas(args)

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	vmInput.CallerAddr = []byte("caller")
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "claimRewards function not allowed to be called by address "+string(vmInput.CallerAddr)))

	vmInput.CallerAddr = AbstractStakingSCAddress
	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))

	vmInput.Arguments = [][]byte{[]byte("test")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller is not a delegator"))
}

func TestDelegationMidas_ExecuteClaimRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	eei.inputParser = &mock.ArgumentParserMock{}
	args.Eei = eei

	args.DelegationSCConfig.MaxServiceFee = 10000
	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("claimRewards", [][]byte{delegator})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{1}
	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(1000),
	})

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          1000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	destAcc, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.True(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	assert.Equal(t, 1, len(destAcc.OutputTransfers))
	outputTransfer := destAcc.OutputTransfers[0]
	assert.Equal(t, big.NewInt(135), outputTransfer.Value)

	_, delegatorData, _ := d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, uint32(3), delegatorData.RewardsCheckpoint)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, uint64(135), delegatorData.TotalCumulatedRewards.Uint64())

	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 3
	}

	_ = d.saveRewardData(3, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})

	vmInput = getDefaultVmInputForFuncMidas("getTotalCumulatedRewardsForUser", [][]byte{delegator})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	_, delegatorData, _ = d.getOrCreateDelegatorData(delegator)
	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
	assert.Equal(t, uint64(135), delegatorData.TotalCumulatedRewards.Uint64())
	lastValue := eei.output[len(eei.output)-1]
	assert.Equal(t, big.NewInt(0).SetBytes(lastValue).Uint64(), uint64(180))
}

func TestDelegationMidas_ExecuteClaimRewardsShouldDeleteDelegator(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	blockChainHook := &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 10
		},
	}
	eei := createDefaultEei()
	eei.blockChainHook = blockChainHook
	args.Eei = eei

	args.DelegationSCConfig.MaxServiceFee = 10000
	delegator := []byte("delegator")
	vmInput := getDefaultVmInputForFuncMidas("claimRewards", [][]byte{delegator})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegator, &DelegatorData{
		ActiveFund:            nil,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(135),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NumUsers: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	destAcc, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
	assert.True(t, exists)
	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
	assert.True(t, exists)

	assert.Equal(t, 1, len(destAcc.OutputTransfers))
	outputTransfer := destAcc.OutputTransfers[0]
	assert.Equal(t, big.NewInt(135), outputTransfer.Value)

	vmInput = getDefaultVmInputForFuncMidas("getTotalCumulatedRewardsForUser", [][]byte{vmInput.CallerAddr})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	res := d.eei.GetStorage(vmInput.CallerAddr)
	require.Len(t, res, 0)
}

func TestDelegationMidas_ExecuteGetRewardDataUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getRewardData", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "must call with 1 arguments"))

	vmInput.Arguments = [][]byte{{2}}
	vmInput.CallValue = big.NewInt(-10)
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "reward not found"))
}

func TestDelegationMidas_ExecuteGetRewardData(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getRewardData", [][]byte{{2}})
	d, _ := NewDelegationSystemSCMidas(args)

	rewardsToDistribute := big.NewInt(100)
	totalActive := big.NewInt(2000)
	serviceFee := uint64(10000)
	_ = d.saveRewardData(2, &RewardComputationData{
		RewardsToDistribute: rewardsToDistribute,
		TotalActive:         totalActive,
		ServiceFee:          serviceFee,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 3, len(eei.output))
	assert.Equal(t, rewardsToDistribute, big.NewInt(0).SetBytes(eei.output[0]))
	assert.Equal(t, totalActive, big.NewInt(0).SetBytes(eei.output[1]))
	assert.Equal(t, uint16(serviceFee), binary.BigEndian.Uint16(eei.output[2]))
}

func TestDelegationMidas_ExecuteGetClaimableRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getClaimableRewards", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))
}

func TestDelegationMidas_ExecuteGetClaimableRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	delegatorAddr := []byte("address")
	vmInput := getDefaultVmInputForFuncMidas("getClaimableRewards", [][]byte{delegatorAddr})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{1}
	_ = d.saveDelegatorData(delegatorAddr, &DelegatorData{
		ActiveFund:            fundKey,
		RewardsCheckpoint:     0,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: big.NewInt(1000),
	})

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          1000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          1000,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(135), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetTotalCumulatedRewardsUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getTotalCumulatedRewards", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "this is a view function only"))
}

func TestDelegationMidas_ExecuteGetTotalCumulatedRewards(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getTotalCumulatedRewards", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveRewardData(0, &RewardComputationData{
		RewardsToDistribute: big.NewInt(100),
		TotalActive:         big.NewInt(1000),
		ServiceFee:          10000,
	})

	_ = d.saveRewardData(1, &RewardComputationData{
		RewardsToDistribute: big.NewInt(200),
		TotalActive:         big.NewInt(2000),
		ServiceFee:          10000,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(300), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetNumUsersUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getNumUsers", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetNumUsers(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getNumUsers", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		NumUsers: 3,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, []byte{3}, eei.output[0])
}

func TestDelegationMidas_ExecuteGetTotalUnStakedUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getTotalUnStaked", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetTotalUnStaked(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getTotalUnStaked", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSCMidas(args)

	totalUnstaked := big.NewInt(1100)
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalUnStaked: totalUnstaked,
		TotalActive:   big.NewInt(0),
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, totalUnstaked, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetTotalActiveStakeUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getTotalActiveStake", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getGlobalFundData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetTotalActiveStake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getTotalActiveStake", [][]byte{})
	vmInput.CallerAddr = vm.EndOfEpochAddress
	d, _ := NewDelegationSystemSCMidas(args)

	totalActive := big.NewInt(5000)
	_ = d.saveGlobalFundData(&GlobalFundData{
		TotalActive: totalActive,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, totalActive, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetUserActiveStakeUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getUserActiveStake", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))
}

func TestDelegationMidas_ExecuteGetUserActiveStakeNoActiveFund(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getUserActiveStake", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund: nil,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetUserActiveStake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getUserActiveStake", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund: fundKey,
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: fundValue,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, fundValue, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetUserUnStakedValueUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getUserUnStakedValue", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))
}

func TestDelegationMidas_ExecuteGetUserUnStakedValueNoUnStakedFund(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getUserUnStakedValue", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: nil,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetUserUnStakedValue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getUserUnStakedValue", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey1 := []byte{2}
	fundKey2 := []byte{3}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: [][]byte{fundKey1, fundKey2},
	})

	_ = d.saveFund(fundKey1, &Fund{
		Value: fundValue,
	})

	_ = d.saveFund(fundKey2, &Fund{
		Value: fundValue,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(300), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetUserUnBondableUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getUserUnBondable", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "view function works only for existing delegators"))

	_ = d.saveDelegatorData([]byte("address"), &DelegatorData{})
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetUserUnBondableNoUnStakedFund(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getUserUnBondable", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: nil,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, big.NewInt(0), big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetUserUnBondable(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return 500
		},
		CurrentEpochCalled: func() uint32 {
			return 500
		},
	}
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getUserUnBondable", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey1 := []byte{2}
	fundKey2 := []byte{3}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		UnStakedFunds: [][]byte{fundKey1, fundKey2},
	})

	_ = d.saveFund(fundKey1, &Fund{
		Value: fundValue,
		Epoch: 400,
	})

	_ = d.saveFund(fundKey2, &Fund{
		Value: fundValue,
		Epoch: 495,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		UnBondPeriodInEpochs: 10,
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, fundValue, big.NewInt(0).SetBytes(eei.output[0]))
}

func TestDelegationMidas_ExecuteGetNumNodesUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getNumNodes", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetNumNodes(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getNumNodes", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:    []*NodesData{{}},
		NotStakedKeys: []*NodesData{{}, {}},
		UnStakedKeys:  []*NodesData{{}, {}, {}},
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, []byte{6}, eei.output[0])
}

func TestDelegationMidas_ExecuteGetAllNodeStatesUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getAllNodeStates", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation status", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetAllNodeStates(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getAllNodeStates", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	blsKey1 := []byte("blsKey1")
	blsKey2 := []byte("blsKey2")
	blsKey3 := []byte("blsKey3")
	blsKey4 := []byte("blsKey4")
	_ = d.saveDelegationStatus(&DelegationContractStatus{
		StakedKeys:    []*NodesData{{BLSKey: blsKey1}},
		NotStakedKeys: []*NodesData{{BLSKey: blsKey2}, {BLSKey: blsKey3}},
		UnStakedKeys:  []*NodesData{{BLSKey: blsKey4}},
	})

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 7, len(eei.output))
	assert.Equal(t, []byte("staked"), eei.output[0])
	assert.Equal(t, blsKey1, eei.output[1])
	assert.Equal(t, []byte("notStaked"), eei.output[2])
	assert.Equal(t, blsKey2, eei.output[3])
	assert.Equal(t, blsKey3, eei.output[4])
	assert.Equal(t, []byte("unStaked"), eei.output[5])
	assert.Equal(t, blsKey4, eei.output[6])
}

func TestDelegationMidas_ExecuteGetContractConfigUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getContractConfig", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	vmInput.CallValue = big.NewInt(10)
	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 10
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	vmInput.Arguments = [][]byte{[]byte("address")}
	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 0
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidNumOfArguments.Error()))

	vmInput.Arguments = [][]byte{}
	output = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w delegation contract config", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationMidas_ExecuteGetContractConfig(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	vmInput := getDefaultVmInputForFuncMidas("getContractConfig", [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	ownerAddress := []byte("owner")
	maxDelegationCap := big.NewInt(200)
	serviceFee := uint64(10000)
	initialOwnerFunds := big.NewInt(500)
	createdNonce := uint64(100)
	unBondPeriodInEpoch := uint32(144000)
	_ = d.saveDelegationContractConfig(&DelegationConfig{
		MaxDelegationCap:            maxDelegationCap,
		InitialOwnerFunds:           initialOwnerFunds,
		AutomaticActivation:         true,
		ChangeableServiceFee:        true,
		CheckCapOnReDelegateRewards: true,
		CreatedNonce:                createdNonce,
		UnBondPeriodInEpochs:        unBondPeriodInEpoch,
	})
	eei.SetStorage([]byte(ownerKey), ownerAddress)
	eei.SetStorage([]byte(serviceFeeKey), big.NewInt(0).SetUint64(serviceFee).Bytes())

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	require.Equal(t, 10, len(eei.output))
	assert.Equal(t, ownerAddress, eei.output[0])
	assert.Equal(t, big.NewInt(0).SetUint64(serviceFee), big.NewInt(0).SetBytes(eei.output[1]))
	assert.Equal(t, maxDelegationCap, big.NewInt(0).SetBytes(eei.output[2]))
	assert.Equal(t, initialOwnerFunds, big.NewInt(0).SetBytes(eei.output[3]))
	assert.Equal(t, []byte("true"), eei.output[4])
	assert.Equal(t, []byte("true"), eei.output[5])
	assert.Equal(t, []byte("true"), eei.output[6])
	assert.Equal(t, []byte("true"), eei.output[7])
	assert.Equal(t, big.NewInt(0).SetUint64(createdNonce), big.NewInt(0).SetBytes(eei.output[8]))
	assert.Equal(t, big.NewInt(0).SetUint64(uint64(unBondPeriodInEpoch)), big.NewInt(0).SetBytes(eei.output[9]))
}

func TestDelegationMidas_ExecuteUnknownFunc(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	invalidFunc := "invalid func"
	vmInput := getDefaultVmInputForFuncMidas(invalidFunc, [][]byte{})
	d, _ := NewDelegationSystemSCMidas(args)

	output := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := invalidFunc + " is an unknown function"
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr))
}

func TestDelegationMidas_computeAndUpdateRewardsWithTotalActiveZeroDoesNotPanic(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 1
		},
	}
	args.Eei = eei
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}
	dData := &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}

	rewards := big.NewInt(1000)
	_ = d.saveFund(fundKey, &Fund{Value: big.NewInt(1)})
	_ = d.saveRewardData(1, &RewardComputationData{
		TotalActive:         big.NewInt(0),
		RewardsToDistribute: rewards,
	})

	ownerAddr := []byte("ownerAddress")
	eei.SetStorage([]byte(ownerKey), ownerAddr)

	err := d.computeAndUpdateRewards([]byte("other address"), dData)
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0), dData.UnClaimedRewards)
}

func TestDelegationMidas_computeAndUpdateRewardsWithTotalActiveZeroSendsAllRewardsToOwner(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 1
		},
	}
	args.Eei = eei
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}
	dData := &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	}

	rewards := big.NewInt(1000)
	_ = d.saveFund(fundKey, &Fund{Value: big.NewInt(1)})
	_ = d.saveRewardData(1, &RewardComputationData{
		TotalActive:         big.NewInt(0),
		RewardsToDistribute: rewards,
	})

	ownerAddr := []byte("ownerAddress")
	eei.SetStorage([]byte(ownerKey), ownerAddr)

	err := d.computeAndUpdateRewards(ownerAddr, dData)
	assert.Nil(t, err)
	assert.Equal(t, rewards, dData.UnClaimedRewards)
}

func TestDelegationMidas_isDelegatorShouldErrBecauseAddressIsNotFound(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("address which didn't delegate")
	vmInput := getDefaultVmInputForFuncMidas("isDelegator", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestDelegationMidas_isDelegatorShouldWork(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("isDelegator", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}

	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund: fundKey,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func TestDelegationMidas_getDelegatorFundsDataDelegatorNotFoundShouldErr(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, "existing delegators")
}

func TestDelegationMidas_getDelegatorFundsDataCannotLoadFundsShouldErr(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		AutomaticActivation:  false,
		ChangeableServiceFee: true,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, vm.ErrDataNotFoundUnderKey.Error())
}

func TestDelegationMidas_getDelegatorFundsDataCannotFindConfigShouldErr(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: fundValue,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)
	assert.Contains(t, eei.returnMessage, vm.ErrDataNotFoundUnderKey.Error())
}

func TestDelegationMidas_getDelegatorFundsDataShouldWork(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	delegatorAddress := []byte("delegatorAddress")
	vmInput := getDefaultVmInputForFuncMidas("getDelegatorFundsData", [][]byte{delegatorAddress})
	d, _ := NewDelegationSystemSCMidas(args)

	fundKey := []byte{2}
	fundValue := big.NewInt(150)
	_ = d.saveDelegatorData(delegatorAddress, &DelegatorData{
		ActiveFund:            fundKey,
		UnClaimedRewards:      big.NewInt(0),
		TotalCumulatedRewards: big.NewInt(0),
	})

	_ = d.saveFund(fundKey, &Fund{
		Value: fundValue,
	})

	_ = d.saveDelegationContractConfig(&DelegationConfig{
		AutomaticActivation:  false,
		ChangeableServiceFee: true,
	})

	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	assert.Equal(t, fundValue.Bytes(), eei.output[0])
}

func TestDelegationMidas_setAndGetDelegationMetadata(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)

	vmInput := getDefaultVmInputForFuncMidas("setMetaData", [][]byte{[]byte("name"), []byte("website"), []byte("identifier")})
	d.eei.SetStorage([]byte(ownerKey), vmInput.CallerAddr)
	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmInputErr := getDefaultVmInputForFuncMidas("setMetaData", [][]byte{[]byte("one")})
	retCode = d.Execute(vmInputErr)
	assert.Equal(t, vmcommon.UserError, retCode)

	vmInputGet := getDefaultVmInputForFuncMidas("getMetaData", [][]byte{})
	retCode = d.Execute(vmInputGet)
	assert.Equal(t, vmcommon.Ok, retCode)

	assert.Equal(t, eei.output[0], vmInput.Arguments[0])
	assert.Equal(t, eei.output[1], vmInput.Arguments[1])
	assert.Equal(t, eei.output[2], vmInput.Arguments[2])
}

func TestDelegationMidas_setAutomaticActivation(t *testing.T) {
	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	args.Eei = eei

	d, _ := NewDelegationSystemSCMidas(args)
	_ = d.saveDelegationContractConfig(&DelegationConfig{})

	vmInput := getDefaultVmInputForFuncMidas("setAutomaticActivation", [][]byte{[]byte("true")})
	d.eei.SetStorage([]byte(ownerKey), vmInput.CallerAddr)
	retCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	dConfig, _ := d.getDelegationContractConfig()
	assert.Equal(t, dConfig.AutomaticActivation, true)

	vmInput = getDefaultVmInputForFuncMidas("setAutomaticActivation", [][]byte{[]byte("abcd")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, retCode)

	vmInput = getDefaultVmInputForFuncMidas("setAutomaticActivation", [][]byte{[]byte("false")})
	retCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, retCode)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, dConfig.AutomaticActivation, false)

	dConfig, _ = d.getDelegationContractConfig()
	assert.Equal(t, dConfig.CheckCapOnReDelegateRewards, false)
}

func TestDelegationMidas_GetDelegationManagementNoDataShouldError(t *testing.T) {
	t.Parallel()

	d := &delegation{
		eei: &mock.SystemEIStub{
			GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
				return nil
			},
		},
	}

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)

	assert.Nil(t, delegationManagement)
	assert.True(t, errors.Is(err, vm.ErrDataNotFoundUnderKey))
}

func TestDelegationMidas_GetDelegationManagementMarshalizerFailsShouldError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	d := &delegation{
		eei: &mock.SystemEIStub{
			GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
				return make([]byte, 1)
			},
		},
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		},
	}

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)

	assert.Nil(t, delegationManagement)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestDelegationMidas_GetDelegationManagementShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	minDelegationAmount := big.NewInt(45)
	minDeposit := big.NewInt(2232)
	cfg := &DelegationManagement{
		MinDelegationAmount: minDelegationAmount,
		MinDeposit:          minDeposit,
	}

	buff, err := marshalizer.Marshal(cfg)
	require.Nil(t, err)

	d := &delegation{
		eei: &mock.SystemEIStub{
			GetStorageFromAddressCalled: func(address []byte, key []byte) []byte {
				return buff
			},
		},
		marshalizer: marshalizer,
	}

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)

	assert.Nil(t, err)
	require.NotNil(t, delegationManagement)
	assert.Equal(t, minDeposit, delegationManagement.MinDeposit)
	assert.Equal(t, minDelegationAmount, delegationManagement.MinDelegationAmount)
}

// TODO:
//func TestDelegationMidas_ExecuteInitFromValidatorData(t *testing.T) {
//	t.Parallel()
//
//	args := createMockArgumentsForDelegationMidas()
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return 2
//		},
//	}
//	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}})
//
//	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(1000))
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	vmInput := getDefaultVmInputForFuncMidas(core.SCDeployInitFunctionName, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})
//	vmInput.CallValue = big.NewInt(1000)
//	vmInput.RecipientAddr = createNewAddress(vm.FirstDelegationSCAddress)
//	vmInput.CallerAddr = []byte("stakingProvider")
//	output := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, output)
//}

func TestDelegationMidas_checkArgumentsForValidatorToDelegation(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSCMidas(args)
	vmInput := getDefaultVmInputForFuncMidas(initFromValidatorData, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})

	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
	returnCode := d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, initFromValidatorData+" is an unknown function")

	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
	eei.returnMessage = ""
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "only delegation manager sc can call this function")

	eei.returnMessage = ""
	vmInput.CallerAddr = d.delegationMgrSCAddress
	vmInput.CallValue.SetUint64(10)
	vmInput.Arguments = [][]byte{}
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "call value must be 0")

	eei.returnMessage = ""
	vmInput.CallValue.SetUint64(0)
	vmInput.Arguments = [][]byte{}
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "not enough arguments")

	eei.returnMessage = ""
	vmInput.Arguments = [][]byte{[]byte("key")}
	returnCode = d.checkArgumentsForValidatorToDelegation(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid arguments, first must be an address")
}

func TestDelegationMidas_getAndVerifyValidatorData(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	eei := createDefaultEei()
	eei.blockChainHook = &mock.BlockChainHookStub{
		CurrentEpochCalled: func() uint32 {
			return 2
		},
	}
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
			return vmcommon.Ok
		}}, nil
	}})

	args.Eei = eei
	args.DelegationSCConfig.MaxServiceFee = 10000
	args.DelegationSCConfig.MinServiceFee = 0
	d, _ := NewDelegationSystemSCMidas(args)

	addr := []byte("address")
	_, returnCode := d.getAndVerifyValidatorData(addr)
	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())
	assert.Equal(t, returnCode, vmcommon.UserError)

	eei.SetStorageForAddress(d.validatorSCAddr, addr, addr)
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)

	validatorData := &ValidatorDataV2{
		RewardAddress:   []byte("randomAddress"),
		TotalSlashed:    big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		TotalStakeValue: big.NewInt(0),
		UnstakedInfo:    []*UnstakedValue{{UnstakedValue: big.NewInt(10)}},
		NumRegistered:   3,
		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
	}
	marshaledData, _ := d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, addr, marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "invalid reward address on validator data")

	validatorData.RewardAddress = addr
	marshaledData, _ = d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, addr, marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)

	managementData := &DelegationManagement{
		NumOfContracts:      0,
		LastAddress:         vm.FirstDelegationSCAddress,
		MinServiceFee:       0,
		MaxServiceFee:       100,
		MinDeposit:          big.NewInt(100),
		MinDelegationAmount: big.NewInt(100),
	}
	marshaledData, _ = d.marshalizer.Marshal(managementData)
	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "not enough stake to make delegation contract")

	validatorData.TotalStakeValue.SetUint64(10000)
	marshaledData, _ = d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, addr, marshaledData)

	eei.returnMessage = ""
	_, returnCode = d.getAndVerifyValidatorData(addr)
	assert.Equal(t, returnCode, vmcommon.UserError)
	assert.Equal(t, eei.returnMessage, "clean unStaked info before changing validator to delegation contract")
}

// TODO:
//func TestDelegationMidas_initFromValidatorData(t *testing.T) {
//	t.Parallel()
//
//	args := createMockArgumentsForDelegationMidas()
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return 2
//		},
//	}
//	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}}
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//
//	_ = eei.SetSystemSCContainer(systemSCContainerStub)
//
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	vmInput := getDefaultVmInputForFuncMidas(initFromValidatorData, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, initFromValidatorData+" is an unknown function")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//
//	eei.returnMessage = ""
//	vmInput.CallerAddr = d.delegationMgrSCAddress
//	vmInput.CallValue.SetUint64(0)
//	oldAddress := bytes.Repeat([]byte{1}, len(vmInput.CallerAddr))
//	vmInput.Arguments = [][]byte{oldAddress}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).Bytes(), big.NewInt(0).SetUint64(d.maxServiceFee + 1).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "service fee out of bounds")
//
//	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.UserError
//		}}, vm.ErrEmptyStorage
//	}
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "storage is nil for given key@storage is nil for given key")
//
//	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.UserError
//		}}, nil
//	}
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//
//	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())
//
//	validatorData := &ValidatorDataV2{
//		RewardAddress:   vmInput.RecipientAddr,
//		TotalSlashed:    big.NewInt(0),
//		TotalUnstaked:   big.NewInt(0),
//		TotalStakeValue: big.NewInt(1000000),
//		NumRegistered:   3,
//		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
//	}
//	marshaledData, _ := d.marshalizer.Marshal(validatorData)
//	eei.SetStorageForAddress(d.validatorSCAddr, vmInput.RecipientAddr, marshaledData)
//
//	managementData := &DelegationManagement{
//		NumOfContracts:      0,
//		LastAddress:         vm.FirstDelegationSCAddress,
//		MinServiceFee:       0,
//		MaxServiceFee:       100,
//		MinDeposit:          big.NewInt(100),
//		MinDelegationAmount: big.NewInt(100),
//	}
//	marshaledData, _ = d.marshalizer.Marshal(managementData)
//	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(0).SetUint64(d.maxServiceFee).Bytes(), big.NewInt(0).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())
//
//	for i, blsKey := range validatorData.BlsPubKeys {
//		stakedData := &StakedDataV2_0{
//			Staked: true,
//		}
//		if i == 0 {
//			stakedData.Staked = false
//		}
//		marshaledData, _ = d.marshalizer.Marshal(stakedData)
//		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
//	}
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, big.NewInt(1).Bytes(), big.NewInt(0).SetUint64(d.maxServiceFee).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "total delegation cap reached")
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress, validatorData.TotalStakeValue.Bytes(), big.NewInt(0).SetUint64(d.maxServiceFee).Bytes()}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//}

//func TestDelegationMidas_mergeValidatorDataToDelegation(t *testing.T) {
//	args := createMockArgumentsForDelegationMidas()
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return 2
//		},
//	}
//	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}}
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//
//	_ = eei.SetSystemSCContainer(systemSCContainerStub)
//
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	vmInput := getDefaultVmInputForFuncMidas(mergeValidatorDataToDelegation, [][]byte{big.NewInt(0).Bytes(), big.NewInt(0).Bytes()})
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, mergeValidatorDataToDelegation+" is an unknown function")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//
//	eei.returnMessage = ""
//	vmInput.CallerAddr = d.delegationMgrSCAddress
//	vmInput.CallValue.SetUint64(0)
//	oldAddress := bytes.Repeat([]byte{1}, len(vmInput.CallerAddr))
//	vmInput.Arguments = [][]byte{oldAddress, oldAddress}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{oldAddress}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())
//
//	validatorData := &ValidatorDataV2{
//		RewardAddress:   oldAddress,
//		TotalSlashed:    big.NewInt(0),
//		TotalUnstaked:   big.NewInt(0),
//		TotalStakeValue: big.NewInt(1000000),
//		NumRegistered:   3,
//		BlsPubKeys:      [][]byte{[]byte("firsstKey"), []byte("secondKey"), []byte("thirddKey")},
//	}
//	marshaledData, _ := d.marshalizer.Marshal(validatorData)
//	eei.SetStorageForAddress(d.validatorSCAddr, oldAddress, marshaledData)
//
//	managementData := &DelegationManagement{
//		NumOfContracts:      0,
//		LastAddress:         vm.FirstDelegationSCAddress,
//		MinServiceFee:       0,
//		MaxServiceFee:       100,
//		MinDeposit:          big.NewInt(100),
//		MinDelegationAmount: big.NewInt(100),
//	}
//	marshaledData, _ = d.marshalizer.Marshal(managementData)
//	eei.SetStorageForAddress(d.delegationMgrSCAddress, []byte(delegationManagementKey), marshaledData)
//
//	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.UserError
//		}}, vm.ErrEmptyStorage
//	}
//
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "storage is nil for given key@storage is nil for given key")
//
//	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.UserError
//		}}, nil
//	}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//
//	systemSCContainerStub.GetCalled = func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "data was not found under requested key delegation status")
//
//	_ = d.saveDelegationStatus(createNewDelegationContractStatus())
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, vm.ErrEmptyStorage.Error())
//
//	for i, blsKey := range validatorData.BlsPubKeys {
//		stakedData := &StakedDataV2_0{
//			Staked: true,
//		}
//		if i == 2 {
//			stakedData.Staked = false
//		}
//		marshaledData, _ = d.marshalizer.Marshal(stakedData)
//		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
//	}
//
//	createNewContractInput := getDefaultVmInputForFuncMidas(core.SCDeployInitFunctionName, [][]byte{big.NewInt(1000000).Bytes(), big.NewInt(0).Bytes()})
//	//createNewContractInput.CallValue = big.NewInt(1000000)
//	createNewContractInput.CallerAddr = d.delegationMgrSCAddress
//	returnCode = d.Execute(createNewContractInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "total delegation cap reached")
//
//	dConfig, _ := d.getDelegationContractConfig()
//	dConfig.MaxDelegationCap.SetUint64(0)
//	_ = d.saveDelegationContractConfig(dConfig)
//
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//
//	dStatus, err := d.getDelegationStatus()
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(dStatus.UnStakedKeys))
//	assert.Equal(t, 2, len(dStatus.StakedKeys))
//}

//func TestDelegationMidas_whitelistForMerge(t *testing.T) {
//	args := createMockArgumentsForDelegationMidas()
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return 2
//		},
//	}
//	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}}
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//
//	_ = eei.SetSystemSCContainer(systemSCContainerStub)
//
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))
//
//	vmInput := getDefaultVmInputForFuncMidas("whitelistForMerge", [][]byte{[]byte("address")})
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "whitelistForMerge"+" is an unknown function")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "can be called by owner or the delegation manager")
//
//	vmInput.CallerAddr = []byte("address0")
//	vmInput.GasProvided = 0
//	eei.gasRemaining = 0
//	d.gasCost.MetaChainSystemSCsCost.DelegationOps = 1
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.OutOfGas, returnCode)
//
//	vmInput.GasProvided = 1000
//	eei.gasRemaining = vmInput.GasProvided
//	vmInput.Arguments = [][]byte{}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//
//	vmInput.Arguments = [][]byte{[]byte("a")}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid argument, wanted an address")
//
//	vmInput.Arguments = [][]byte{[]byte("address0")}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "cannot whitelist own address")
//
//	vmInput.Arguments = [][]byte{[]byte("address1")}
//	vmInput.CallValue = big.NewInt(10)
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "non-payable function")
//
//	vmInput.CallValue = big.NewInt(0)
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//	assert.Equal(t, []byte("address1"), d.eei.GetStorage([]byte(whitelistedAddress)))
//}
//
//func TestDelegationMidas_deleteWhitelistForMerge(t *testing.T) {
//	args := createMockArgumentsForDelegationMidas()
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return 2
//		},
//	}
//	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}}
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//
//	_ = eei.SetSystemSCContainer(systemSCContainerStub)
//
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))
//
//	vmInput := getDefaultVmInputForFuncMidas("deleteWhitelistForMerge", [][]byte{[]byte("address")})
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "deleteWhitelistForMerge"+" is an unknown function")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))
//	vmInput.CallerAddr = []byte("address0")
//
//	vmInput.GasProvided = 1000
//	eei.gasRemaining = vmInput.GasProvided
//	vmInput.Arguments = [][]byte{[]byte("a")}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//
//	d.eei.SetStorage([]byte(whitelistedAddress), []byte("address"))
//	vmInput.Arguments = [][]byte{}
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//	assert.Equal(t, 0, len(d.eei.GetStorage([]byte(whitelistedAddress))))
//
//	d.eei.SetStorage([]byte(whitelistedAddress), []byte("address"))
//	vmInput.Arguments = [][]byte{}
//	vmInput.CallerAddr = vm.DelegationManagerSCAddress
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//	assert.Equal(t, 0, len(d.eei.GetStorage([]byte(whitelistedAddress))))
//}
//
//func TestDelegationMidas_GetWhitelistForMerge(t *testing.T) {
//	args := createMockArgumentsForDelegationMidas()
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return 2
//		},
//	}
//	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}}
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//
//	_ = eei.SetSystemSCContainer(systemSCContainerStub)
//
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))
//
//	vmInput := getDefaultVmInputForFuncMidas("getWhitelistForMerge", make([][]byte, 0))
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "getWhitelistForMerge"+" is an unknown function")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//
//	addr := []byte("address1")
//	vmInput = getDefaultVmInputForFuncMidas("whitelistForMerge", [][]byte{addr})
//	vmInput.CallValue = big.NewInt(0)
//	vmInput.CallerAddr = []byte("address0")
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//
//	vmInput = getDefaultVmInputForFuncMidas("getWhitelistForMerge", make([][]byte, 0))
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//	require.Equal(t, 1, len(eei.output))
//	assert.Equal(t, addr, eei.output[0])
//}

// TODO:
//func TestDelegationMidas_OptimizeRewardsComputation(t *testing.T) {
//	args := createMockArgumentsForDelegationMidas()
//	currentEpoch := uint32(2)
//	eei := createDefaultEei()
//	eei.blockChainHook = &mock.BlockChainHookStub{
//		CurrentEpochCalled: func() uint32 {
//			return currentEpoch
//		},
//	}
//	systemSCContainerStub := &mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}}
//
//	_ = eei.SetSystemSCContainer(systemSCContainerStub)
//	createDelegationManagerConfigMidas(eei, args.Marshalizer, big.NewInt(10))
//
//	args.Eei = eei
//	args.DelegationSCConfig.MaxServiceFee = 10000
//	args.DelegationSCConfig.MinServiceFee = 0
//	d, _ := NewDelegationSystemSCMidas(args)
//	_ = d.saveDelegationStatus(&DelegationContractStatus{})
//	_ = d.saveDelegationContractConfig(&DelegationConfig{
//		MaxDelegationCap:  big.NewInt(10000),
//		InitialOwnerFunds: big.NewInt(1000),
//	})
//	_ = d.saveGlobalFundData(&GlobalFundData{
//		TotalActive: big.NewInt(1000),
//	})
//
//	d.eei.SetStorage([]byte(ownerKey), []byte("address0"))
//
//	delegator := []byte("delegator")
//	_ = d.saveDelegatorData(delegator, &DelegatorData{
//		ActiveFund:            nil,
//		UnStakedFunds:         [][]byte{},
//		UnClaimedRewards:      big.NewInt(1000),
//		TotalCumulatedRewards: big.NewInt(0),
//		RewardsCheckpoint:     0,
//	})
//
//	vmInput := getDefaultVmInputForFuncMidas("updateRewards", [][]byte{})
//	vmInput.CallValue = big.NewInt(20)
//	vmInput.CallerAddr = vm.EndOfEpochAddress
//
//	for i := 0; i < 10; i++ {
//		currentEpoch++
//		output := d.Execute(vmInput)
//		assert.Equal(t, vmcommon.Ok, output)
//	}
//
//	vmInput = getDefaultVmInputForFuncMidas("delegate", [][]byte{})
//	vmInput.CallerAddr = AbstractStakingSCAddress
//	vmInput.Arguments = [][]byte{delegator, big.NewInt(1000).Bytes()}
//
//	output := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, output)
//
//	currentEpoch++
//	vmInput = getDefaultVmInputForFuncMidas("updateRewards", [][]byte{})
//	vmInput.CallValue = big.NewInt(20)
//	vmInput.CallerAddr = vm.EndOfEpochAddress
//	output = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, output)
//
//	vmInput = getDefaultVmInputForFuncMidas("claimRewards", [][]byte{})
//	vmInput.CallerAddr = delegator
//
//	output = d.Execute(vmInput)
//	fmt.Println(eei.returnMessage)
//	assert.Equal(t, vmcommon.Ok, output)
//
//	destAcc, exists := eei.outputAccounts[string(vmInput.CallerAddr)]
//	assert.True(t, exists)
//	_, exists = eei.outputAccounts[string(vmInput.RecipientAddr)]
//	assert.True(t, exists)
//
//	assert.Equal(t, 1, len(destAcc.OutputTransfers))
//	outputTransfer := destAcc.OutputTransfers[0]
//	assert.Equal(t, big.NewInt(1010), outputTransfer.Value)
//
//	_, delegatorData, _ := d.getOrCreateDelegatorData(vmInput.CallerAddr)
//	assert.Equal(t, uint32(14), delegatorData.RewardsCheckpoint)
//	assert.Equal(t, uint64(0), delegatorData.UnClaimedRewards.Uint64())
//	assert.Equal(t, 1010, int(delegatorData.TotalCumulatedRewards.Uint64()))
//}

func TestDelegationMidas_correctNodesStatus(t *testing.T) {
	d, eei := createDelegationContractAndEEIMidas()
	vmInput := getDefaultVmInputForFuncMidas("correctNodesStatus", nil)

	enableEpochsHandler, _ := d.enableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	enableEpochsHandler.RemoveActiveFlags(common.AddTokensToDelegationFlag)
	returnCode := d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "correctNodesStatus is an unknown function")

	enableEpochsHandler.AddActiveFlags(common.AddTokensToDelegationFlag)
	eei.returnMessage = ""
	vmInput.CallValue.SetUint64(10)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "call value must be zero")

	eei.returnMessage = ""
	eei.gasRemaining = 1
	d.gasCost.MetaChainSystemSCsCost.GetAllNodeStates = 10
	vmInput.CallValue.SetUint64(0)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, returnCode)

	eei.returnMessage = ""
	eei.gasRemaining = 11
	vmInput.CallValue.SetUint64(0)
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "data was not found under requested key delegation status")

	wrongStatus := &DelegationContractStatus{
		StakedKeys:    []*NodesData{{BLSKey: []byte("key1")}, {BLSKey: []byte("key2")}, {BLSKey: []byte("key3")}},
		NotStakedKeys: []*NodesData{{BLSKey: []byte("key4")}, {BLSKey: []byte("key5")}, {BLSKey: []byte("key3")}},
		UnStakedKeys:  []*NodesData{{BLSKey: []byte("key6")}, {BLSKey: []byte("key7")}, {BLSKey: []byte("key3")}},
		NumUsers:      0,
	}
	_ = d.saveDelegationStatus(wrongStatus)

	stakedKeys := [][]byte{[]byte("key1"), []byte("key4"), []byte("key7")}
	unStakedKeys := [][]byte{[]byte("key2"), []byte("key6")}
	for i, blsKey := range stakedKeys {
		stakedData := &StakedDataV2_0{
			Staked: true,
		}
		if i == 2 {
			stakedData.Staked = false
			stakedData.Jailed = true
		}
		marshaledData, _ := d.marshalizer.Marshal(stakedData)
		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
	}

	for _, blsKey := range unStakedKeys {
		stakedData := &StakedDataV2_0{
			Staked: false,
		}
		marshaledData, _ := d.marshalizer.Marshal(stakedData)
		eei.SetStorageForAddress(d.stakingSCAddr, blsKey, marshaledData)
	}

	eei.returnMessage = ""
	eei.gasRemaining = 11
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "storage is nil for given key")

	validatorData := &ValidatorDataV2{BlsPubKeys: [][]byte{[]byte("key8")}}
	marshaledData, _ := d.marshalizer.Marshal(validatorData)
	eei.SetStorageForAddress(d.validatorSCAddr, vmInput.RecipientAddr, marshaledData)

	stakedData := &StakedDataV2_0{
		Staked: false,
		Jailed: true,
	}
	marshaledData, _ = d.marshalizer.Marshal(stakedData)
	eei.SetStorageForAddress(d.stakingSCAddr, []byte("key8"), marshaledData)
	stakedKeys = append(stakedKeys, []byte("key8"))

	eei.returnMessage = ""
	eei.gasRemaining = 11
	returnCode = d.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, returnCode)

	correctedStatus, _ := d.getDelegationStatus()
	assert.Equal(t, 4, len(correctedStatus.StakedKeys))
	assert.Equal(t, 2, len(correctedStatus.UnStakedKeys))
	assert.Equal(t, 2, len(correctedStatus.NotStakedKeys))

	for _, stakedKey := range stakedKeys {
		found := false
		for _, stakedNode := range correctedStatus.StakedKeys {
			if bytes.Equal(stakedNode.BLSKey, stakedKey) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	for _, unStakedKey := range unStakedKeys {
		found := false
		for _, unStakedNode := range correctedStatus.UnStakedKeys {
			if bytes.Equal(unStakedNode.BLSKey, unStakedKey) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	notStakedKeys := [][]byte{[]byte("key3"), []byte("key5")}
	for _, notStakedKey := range notStakedKeys {
		found := false
		for _, notStakedNode := range correctedStatus.NotStakedKeys {
			if bytes.Equal(notStakedNode.BLSKey, notStakedKey) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}
}

func TestDelegationSystemSCMidas_SynchronizeOwner(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationMidas()
	epochHandler := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	account := &stateMock.AccountWrapMock{}

	argsVmContext := VMContextArgs{
		BlockChainHook:      &mock.BlockChainHookStub{},
		CryptoHook:          hooks.NewVMCryptoHook(),
		InputParser:         &mock.ArgumentParserMock{},
		ValidatorAccountsDB: &stateMock.AccountsStub{},
		UserAccountsDB: &stateMock.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return account, nil
			},
		},
		ChanceComputer:      &mock.RaterMock{},
		EnableEpochsHandler: args.EnableEpochsHandler,
		ShardCoordinator:    &mock.ShardCoordinatorStub{},
	}
	eei, err := NewVMContext(argsVmContext)
	require.Nil(t, err)

	delegationsMap := map[string][]byte{}
	ownerAddress := []byte("1111")
	scAddress := bytes.Repeat([]byte{1}, len(ownerAddress))
	eei.SetSCAddress(scAddress)
	delegationsMap[ownerKey] = ownerAddress
	marshalledData, _ := args.Marshalizer.Marshal(&DelegatorData{RewardsCheckpoint: 10})
	delegationsMap[string(ownerAddress)] = marshalledData
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	args.Eei = eei

	vmInputArgs := make([][]byte, 0)

	d, _ := NewDelegationSystemSCMidas(args)

	// do not run these tests in parallel
	t.Run("function is disabled", func(t *testing.T) {
		vmInput := getDefaultVmInputForFuncMidas("synchronizeOwner", vmInputArgs)
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, "synchronizeOwner is an unknown function", eei.GetReturnMessage())
	})

	epochHandler.AddActiveFlags(common.FixDelegationChangeOwnerOnAccountFlag)
	eei.ResetReturnMessage()

	t.Run("transfer value is not zero", func(t *testing.T) {
		vmInput := getDefaultVmInputForFuncMidas("synchronizeOwner", vmInputArgs)
		vmInput.CallValue = big.NewInt(1)
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, vm.ErrCallValueMustBeZero.Error(), eei.GetReturnMessage())
		eei.ResetReturnMessage()
	})
	t.Run("wrong arguments", func(t *testing.T) {
		vmInput := getDefaultVmInputForFuncMidas("synchronizeOwner", [][]byte{[]byte("argument")})
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, "invalid number of arguments, expected 0", eei.GetReturnMessage())
		eei.ResetReturnMessage()
	})
	t.Run("wrong stored address", func(t *testing.T) {
		vmInput := getDefaultVmInputForFuncMidas("synchronizeOwner", vmInputArgs)
		eei.SetSCAddress(scAddress[:1])
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.UserError, returnCode)
		assert.Equal(t, "wrong new owner address", eei.GetReturnMessage())
		assert.Equal(t, 0, len(account.Owner))
		eei.ResetReturnMessage()
	})
	t.Run("should work", func(t *testing.T) {
		vmInput := getDefaultVmInputForFuncMidas("synchronizeOwner", vmInputArgs)
		eei.SetSCAddress(scAddress)
		returnCode := d.Execute(vmInput)
		assert.Equal(t, vmcommon.Ok, returnCode)
		assert.Equal(t, "", eei.GetReturnMessage())
		assert.Equal(t, ownerAddress, account.Owner)
		eei.ResetReturnMessage()
	})
}
