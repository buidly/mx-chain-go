package systemSmartContracts

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

var configChangeAddressMidas = []byte("config change address")

func createMockArgumentsForDelegationManagerMidas() ArgsNewDelegationManager {
	return ArgsNewDelegationManager{
		DelegationSCConfig: config.DelegationSystemSCConfig{
			MinServiceFee: 5,
			MaxServiceFee: 150,
		},
		DelegationMgrSCConfig: config.DelegationManagerSystemSCConfig{
			MinCreationDeposit: "10",
			MinStakeAmount:     "10",
		},
		Eei:                    &mock.SystemEIStub{},
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		StakingSCAddress:       vm.StakingSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		ConfigChangeAddress:    configChangeAddressMidas,
		GasCost:                vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
		Marshalizer:            &mock.MarshalizerMock{},
		EnableEpochsHandler:    enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.DelegationManagerFlag, common.ValidatorToDelegationFlag, common.MultiClaimOnDelegationFlag),
	}
}

func getDefaultVmInputForDelegationManagerMidas(funcName string, args [][]byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
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

func TestNewDelegationManagerSystemSCMidas_NilEeiShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.Eei = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewDelegationManagerSystemSCMidas_InvalidStakingSCAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.StakingSCAddress = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	expectedErr := fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationManagerSystemSCMidas_InvalidValidatorSCAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.ValidatorSCAddress = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	expectedErr := fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationManagerSystemSCMidas_InvalidDelegationManagerSCAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.DelegationMgrSCAddress = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	expectedErr := fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationManagerSystemSCMidas_InvalidConfigChangeAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.ConfigChangeAddress = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	expectedErr := fmt.Errorf("%w for config change address", vm.ErrInvalidAddress)
	assert.Equal(t, expectedErr, err)
}

func TestNewDelegationManagerSystemSCMidas_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.Marshalizer = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	assert.Equal(t, vm.ErrNilMarshalizer, err)
}

func TestNewDelegationManagerSystemSCMidas_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.EnableEpochsHandler = nil

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	assert.Equal(t, vm.ErrNilEnableEpochsHandler, err)
}

func TestNewDelegationManagerSystemSCMidas_InvalidEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestNewDelegationManagerSystemSCMidas_InvalidMinCreationDepositShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.DelegationMgrSCConfig.MinCreationDeposit = "-10"

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	assert.Equal(t, vm.ErrInvalidMinCreationDeposit, err)
}

func TestNewDelegationManagerSystemSCMidas_InvalidMinStakeAmountShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	args.DelegationMgrSCConfig.MinStakeAmount = "-10"

	dm, err := NewDelegationManagerSystemSCMidas(args)
	assert.Nil(t, dm)
	assert.Equal(t, vm.ErrInvalidMinStakeValue, err)
}

func TestNewDelegationManagerSystemSCMidas(t *testing.T) {
	t.Parallel()

	dm, err := NewDelegationManagerSystemSCMidas(createMockArgumentsForDelegationManagerMidas())
	assert.Nil(t, err)
	assert.NotNil(t, dm)
}

func TestDelegationManagerSystemSCMidas_ExecuteWithNilArgsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	dm, _ := NewDelegationManagerSystemSCMidas(args)

	output := dm.Execute(nil)
	assert.Equal(t, vmcommon.UserError, output)
}

func TestDelegationManagerSystemSCMidas_ExecuteWithDelegationManagerDisabled(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	enableEpochsHandler.RemoveActiveFlags(common.DelegationManagerFlag)
	vmInput := getDefaultVmInputForDelegationManagerMidas("createNewDelegationContract", [][]byte{})

	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "delegation manager contract is not enabled"))
}

func TestDelegationManagerSystemSCMidas_ExecuteInvalidFunction(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	vmInput := getDefaultVmInputForDelegationManagerMidas("func", [][]byte{})

	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "invalid function to call"))
}

func TestDelegationManagerSystemSCMidas_ExecuteInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	eei.SetSCAddress(dm.delegationMgrSCAddress)
	vmInput := getDefaultVmInputForDelegationManagerMidas(core.SCDeployInitFunctionName, [][]byte{})
	vmInput.CallValue = big.NewInt(15)

	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrCallValueMustBeZero.Error()))

	vmInput.CallValue = big.NewInt(0)
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dManagementData, _ := dm.getDelegationManagementData()
	assert.Equal(t, uint32(0), dManagementData.NumOfContracts)
	assert.Equal(t, vm.FirstDelegationSCAddress, dManagementData.LastAddress)
	assert.Equal(t, dm.minFee, dManagementData.MinServiceFee)
	assert.Equal(t, dm.maxFee, dManagementData.MaxServiceFee)
	assert.Equal(t, dm.minCreationDeposit, dManagementData.MinDeposit)

	dContractList, _ := getDelegationContractList(dm.eei, dm.marshalizer, dm.delegationMgrSCAddress)
	assert.Equal(t, 1, len(dContractList.Addresses))
}

func TestDelegationManagerSystemSCMidas_ExecuteCreateNewDelegationContractUserErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	eei.SetSCAddress(dm.delegationMgrSCAddress)
	vmInput := getDefaultVmInputForDelegationManagerMidas("createNewDelegationContract", [][]byte{})
	dm.gasCost.MetaChainSystemSCsCost.DelegationMgrOps = 10

	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)
	assert.True(t, strings.Contains(eei.returnMessage, "wrong number of arguments"))

	vmInput.Arguments = [][]byte{{10}, {150}}
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrNotEnoughGas.Error()))

	dm.gasCost.MetaChainSystemSCsCost.DelegationMgrOps = 0
	delegationsMap := map[string][]byte{}
	delegationsMap[string(vmInput.CallerAddr)] = []byte("deployed contract")
	eei.storageUpdate[string(eei.scAddress)] = delegationsMap
	vmInput.CallValue = big.NewInt(0)
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, "caller already deployed a delegation sc"))

	delete(delegationsMap, string(vmInput.CallerAddr))
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getDelegationManagementData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = saveDelegationManagementData(dm.eei, dm.marshalizer, dm.delegationMgrSCAddress, &DelegationManagement{
		MinDeposit: big.NewInt(10),
	})
	vmInput.CallValue = big.NewInt(9)
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.TransactionValueMustBeZero))
}

func createSystemSCContainerMidas(eei *vmContext) vm.SystemSCContainer {
	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	argsStaking.StakingAccessAddr = vm.ValidatorSCAddress
	stakingSc, _ := NewStakingSmartContract(argsStaking)

	argsValidator := createMockArgumentsForValidatorSCMidas()
	argsValidator.Eei = eei
	argsValidator.StakingSCAddress = vm.StakingSCAddress
	argsValidator.ValidatorSCAddress = vm.ValidatorSCAddress
	validatorScMidas, _ := NewValidatorSmartContractMidas(argsValidator)

	delegationSCArgs := createMockArgumentsForDelegationMidas()
	delegationSCArgs.Eei = eei
	delegationScMidas, _ := NewDelegationSystemSCMidas(delegationSCArgs)

	systemSCContainer := &mock.SystemSCContainerStub{
		GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
			switch string(key) {
			case string(vm.StakingSCAddress):
				return stakingSc, nil
			case string(vm.ValidatorSCAddress):
				return validatorScMidas, nil
			case string(vm.FirstDelegationSCAddress):
				return delegationScMidas, nil
			}
			return nil, vm.ErrUnknownSystemSmartContract
		},
	}

	return systemSCContainer
}

func TestDelegationManagerSystemSCMidas_ExecuteCreateNewDelegationContract(t *testing.T) {
	t.Parallel()

	maxDelegationCap := []byte{250}
	serviceFee := []byte{10}
	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	_ = eei.SetSystemSCContainer(
		createSystemSCContainerMidas(eei),
	)

	args.Eei = eei
	createDelegationManagerConfig(eei, args.Marshalizer, big.NewInt(20))

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	eei.SetSCAddress(dm.delegationMgrSCAddress)
	vmInput := getDefaultVmInputForDelegationManagerMidas("createNewDelegationContract", [][]byte{maxDelegationCap, serviceFee})

	_ = dm.saveDelegationContractList(&DelegationContractList{Addresses: make([][]byte, 0)})
	_ = saveDelegationManagementData(dm.eei, dm.marshalizer, dm.delegationMgrSCAddress, &DelegationManagement{
		MinDeposit:          big.NewInt(10),
		LastAddress:         vm.FirstDelegationSCAddress,
		MinDelegationAmount: big.NewInt(1),
	})

	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	dManagement, _ := dm.getDelegationManagementData()
	assert.Equal(t, uint32(1), dManagement.NumOfContracts)
	expectedAddress := createNewAddress(vm.FirstDelegationSCAddress)
	assert.Equal(t, expectedAddress, dManagement.LastAddress)

	dList, _ := getDelegationContractList(dm.eei, dm.marshalizer, dm.delegationMgrSCAddress)
	assert.Equal(t, 1, len(dList.Addresses))
	assert.Equal(t, expectedAddress, dList.Addresses[0])

	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, expectedAddress, eei.output[0])

	outAcc := eei.outputAccounts[string(expectedAddress)]
	assert.Equal(t, vm.FirstDelegationSCAddress, outAcc.Code)
	assert.Equal(t, vmInput.CallerAddr, outAcc.CodeDeployerAddress)

	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}
	expectedMetaData := codeMetaData.ToBytes()
	assert.Equal(t, expectedMetaData, outAcc.CodeMetadata)

	systemSc, _ := eei.systemContracts.Get(vm.FirstDelegationSCAddress)
	delegationSc := systemSc.(*delegationMidas)
	eei.scAddress = createNewAddress(vm.FirstDelegationSCAddress)
	dContractConfig, _ := delegationSc.getDelegationContractConfig()
	retrievedOwnerAddress := eei.GetStorage([]byte(ownerKey))
	retrievedServiceFee := eei.GetStorage([]byte(serviceFeeKey))
	assert.Equal(t, vmInput.CallerAddr, retrievedOwnerAddress)
	assert.Equal(t, []byte{10}, retrievedServiceFee)
	assert.Equal(t, big.NewInt(250), dContractConfig.MaxDelegationCap)
}

func TestDelegationManagerSystemSCMidas_ExecuteGetAllContractAddresses(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	eei.SetSCAddress(dm.delegationMgrSCAddress)
	vmInput := getDefaultVmInputForDelegationManagerMidas("getAllContractAddresses", [][]byte{})

	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidCaller.Error()))

	vmInput.CallerAddr = dm.delegationMgrSCAddress
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getDelegationContractList", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))

	_ = dm.saveDelegationContractList(&DelegationContractList{Addresses: [][]byte{addr1, addr2}})
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
	assert.Equal(t, 1, len(eei.output))
	assert.Equal(t, addr2, eei.output[0])
}

func TestCreateNewAddressMidas_NextAddressShouldWork(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		lastAddress         []byte
		expectedNextAddress []byte
	}

	tests := []*testStruct{
		{
			lastAddress:         []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 255, 255, 255},
			expectedNextAddress: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 255, 255, 255},
		},
		{
			lastAddress:         []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 255},
			expectedNextAddress: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255},
		},
		{
			lastAddress:         []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255},
			expectedNextAddress: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 255, 255, 255},
		},
		{
			lastAddress:         []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255},
			expectedNextAddress: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 255, 255, 255},
		},
		{
			lastAddress:         []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 34, 23, 255, 255, 255, 255, 255, 255, 255, 255},
			expectedNextAddress: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 34, 24, 0, 0, 0, 0, 0, 255, 255, 255},
		},
	}

	for _, test := range tests {
		nextAddress := createNewAddress(test.lastAddress)
		assert.Equal(t, test.expectedNextAddress, nextAddress,
			fmt.Sprintf("expected: %v, got %d", test.expectedNextAddress, nextAddress))
	}
}

func TestDelegationManagerMidas_GetContractConfigErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	vmInput := getDefaultVmInputForDelegationManagerMidas("getContractConfig", [][]byte{})
	vmInput.CallerAddr = []byte("not the correct caller")
	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	assert.True(t, strings.Contains(eei.returnMessage, vm.ErrInvalidCaller.Error()))

	//missing data
	vmInput.CallerAddr = vm.DelegationManagerSCAddress
	output = dm.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)
	expectedErr := fmt.Errorf("%w getDelegationManagementData", vm.ErrDataNotFoundUnderKey)
	assert.True(t, strings.Contains(eei.returnMessage, expectedErr.Error()))
}

func TestDelegationManagerMidas_GetContractConfigShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	args.Eei = eei

	dm, _ := NewDelegationManagerSystemSCMidas(args)
	eei.SetSCAddress(dm.delegationMgrSCAddress)
	vmInput := getDefaultVmInputForDelegationManagerMidas("getContractConfig", [][]byte{})

	delegationManagement := &DelegationManagement{
		NumOfContracts:      123,
		LastAddress:         []byte("last address"),
		MinServiceFee:       456,
		MaxServiceFee:       789,
		MinDeposit:          big.NewInt(112233),
		MinDelegationAmount: big.NewInt(445566),
	}

	_ = saveDelegationManagementData(dm.eei, dm.marshalizer, dm.delegationMgrSCAddress, delegationManagement)

	vmInput.CallerAddr = vm.DelegationManagerSCAddress
	output := dm.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)

	results := eei.CreateVMOutput()
	//this test also verify the position in results.ReturnData
	assert.Equal(t, big.NewInt(int64(delegationManagement.NumOfContracts)).Bytes(), results.ReturnData[0])
	assert.Equal(t, delegationManagement.LastAddress, results.ReturnData[1])
	assert.Equal(t, big.NewInt(int64(delegationManagement.MinServiceFee)).Bytes(), results.ReturnData[2])
	assert.Equal(t, big.NewInt(int64(delegationManagement.MaxServiceFee)).Bytes(), results.ReturnData[3])
	assert.Equal(t, delegationManagement.MinDeposit.Bytes(), results.ReturnData[4])
	assert.Equal(t, delegationManagement.MinDelegationAmount.Bytes(), results.ReturnData[5])
}

func TestDelegationManagerSystemSCMidas_checkValidatorToDelegationInput(t *testing.T) {
	maxDelegationCap := []byte{250}
	serviceFee := []byte{10}
	args := createMockArgumentsForDelegationManagerMidas()
	eei := createDefaultEei()
	_ = eei.SetSystemSCContainer(
		createSystemSCContainerMidas(eei),
	)

	args.Eei = eei
	args.GasCost.MetaChainSystemSCsCost.ValidatorToDelegation = 100
	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
	d, _ := NewDelegationManagerSystemSCMidas(args)
	vmInput := getDefaultVmInputForDelegationManagerMidas("createNewDelegationContract", [][]byte{maxDelegationCap, serviceFee})

	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
	returnCode := d.checkValidatorToDelegationInput(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "invalid function to call")

	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
	eei.returnMessage = ""
	vmInput.CallValue.SetUint64(10)
	returnCode = d.checkValidatorToDelegationInput(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "callValue must be 0")

	eei.returnMessage = ""
	vmInput.CallValue.SetUint64(0)
	vmInput.GasProvided = 0
	returnCode = d.checkValidatorToDelegationInput(vmInput)
	assert.Equal(t, vmcommon.OutOfGas, returnCode)

	eei.returnMessage = ""
	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
	eei.gasRemaining = vmInput.GasProvided
	vmInput.CallerAddr = vm.ESDTSCAddress
	returnCode = d.checkValidatorToDelegationInput(vmInput)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.Equal(t, eei.returnMessage, "cannot change from validator to delegation contract for a smart contract")
}

// TODO:
//func TestDelegationManagerSystemSCMidas_MakeNewContractFromValidatorData(t *testing.T) {
//	maxDelegationCap := []byte{250}
//	serviceFee := []byte{10}
//	args := createMockArgumentsForDelegationManagerMidas()
//	eei := createDefaultEei()
//	_ = eei.SetSystemSCContainer(
//		createSystemSCContainerMidas(eei),
//	)
//
//	args.Eei = eei
//	args.GasCost.MetaChainSystemSCsCost.ValidatorToDelegation = 100
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//	d, _ := NewDelegationManagerSystemSCMidas(args)
//	vmInput := getDefaultVmInputForDelegationManagerMidas("makeNewContractFromValidatorData", [][]byte{maxDelegationCap, serviceFee})
//	_ = d.init(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid function to call")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//
//	eei.returnMessage = ""
//	vmInput.CallValue.SetUint64(0)
//	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
//	eei.gasRemaining = vmInput.GasProvided
//	vmInput.Arguments = append(vmInput.Arguments, []byte("someotherarg"))
//
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//
//	eei.gasRemaining = vmInput.GasProvided
//	vmInput.Arguments = [][]byte{maxDelegationCap, serviceFee}
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationSameOwner(t *testing.T) {
//	maxDelegationCap := []byte{250}
//	serviceFee := []byte{10}
//	args := createMockArgumentsForDelegationManagerMidas()
//	eei := createDefaultEei()
//	_ = eei.SetSystemSCContainer(
//		createSystemSCContainerMidas(eei),
//	)
//
//	args.Eei = eei
//	args.GasCost.MetaChainSystemSCsCost.ValidatorToDelegation = 100
//	enableEpochsHandler, _ := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//	d, _ := NewDelegationManagerSystemSCMidas(args)
//	vmInput := getDefaultVmInputForDelegationManagerMidas("mergeValidatorToDelegationSameOwner", [][]byte{maxDelegationCap, serviceFee})
//	_ = d.init(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
//
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid function to call")
//
//	enableEpochsHandler.AddActiveFlags(common.ValidatorToDelegationFlag)
//
//	eei.returnMessage = ""
//	vmInput.CallValue.SetUint64(0)
//	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
//	eei.gasRemaining = vmInput.GasProvided
//
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{[]byte("somearg")}
//	eei.gasRemaining = vmInput.GasProvided
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid argument, wanted an address")
//
//	eei.returnMessage = ""
//	vmInput.Arguments = [][]byte{vmInput.CallerAddr}
//	eei.gasRemaining = vmInput.GasProvided
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "the caller does not own a delegation sc")
//
//	eei.returnMessage = ""
//	eei.gasRemaining = vmInput.GasProvided
//
//	eei.SetStorage(vmInput.CallerAddr, make([]byte, len(vmInput.CallerAddr)))
//
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "did not find delegation contract with given address for this caller")
//
//	eei.returnMessage = ""
//	eei.gasRemaining = vmInput.GasProvided
//	eei.SetStorage(vmInput.CallerAddr, vmInput.CallerAddr)
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, vm.ErrUnknownSystemSmartContract.Error())
//
//	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//		return &mock.SystemSCStub{ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//			return vmcommon.Ok
//		}}, nil
//	}})
//	eei.returnMessage = ""
//	eei.gasRemaining = vmInput.GasProvided
//
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//}

//func createTestEEIAndDelegationFormMergeValidatorMidas() (*delegationManagerMidas, *vmContext) {
//	args := createMockArgumentsForDelegationManagerMidas()
//	eei := createDefaultEei()
//	_ = eei.SetSystemSCContainer(
//		createSystemSCContainerMidas(eei),
//	)
//
//	args.Eei = eei
//	args.GasCost.MetaChainSystemSCsCost.ValidatorToDelegation = 100
//	d, _ := NewDelegationManagerSystemSCMidas(args)
//	_ = d.init(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
//
//	return d, eei
//}
//
//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListInvalidFunctionCall(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//
//	maxDelegationCap := []byte{250}
//	serviceFee := []byte{10}
//	eei.returnMessage = ""
//	vmInput := getDefaultVmInputForDelegationManagerMidas("mergeValidatorToDelegationWithWhitelist", [][]byte{maxDelegationCap, serviceFee})
//	enableEpochsHandler, _ := d.enableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
//	enableEpochsHandler.RemoveActiveFlags(common.ValidatorToDelegationFlag)
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid function to call")
//}
//
//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListInvalidNumArgs(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//
//	maxDelegationCap := []byte{250}
//	serviceFee := []byte{10}
//	eei.returnMessage = ""
//	vmInput := getDefaultVmInputForDelegationManagerMidas("mergeValidatorToDelegationWithWhitelist", [][]byte{maxDelegationCap, serviceFee})
//	vmInput.CallValue.SetUint64(0)
//	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
//	eei.gasRemaining = vmInput.GasProvided
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid number of arguments")
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListInvalidArgument(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//
//	eei.returnMessage = ""
//	vmInput := getDefaultVmInputForDelegationManagerMidas("mergeValidatorToDelegationWithWhitelist", [][]byte{[]byte("somearg")})
//	vmInput.CallValue.SetUint64(0)
//	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
//	eei.gasRemaining = vmInput.GasProvided
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "invalid argument, wanted an address")
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListNotWhitelisted(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//	eei.returnMessage = ""
//	vmInput := getDefaultVmInputForDelegationManagerMidas("mergeValidatorToDelegationWithWhitelist", make([][]byte, 0))
//	vmInput.Arguments = [][]byte{vmInput.CallerAddr}
//	vmInput.CallValue.SetUint64(0)
//	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
//	eei.gasRemaining = vmInput.GasProvided
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "address is not whitelisted for merge")
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListMissingSmartContract(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//	vmInput := prepareVmInputContextAndDelegationManagerMidas(d, eei)
//	eei.SetStorage(vmInput.CallerAddr, vmInput.CallerAddr)
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, vm.ErrUnknownSystemSmartContract.Error())
//}

//func prepareVmInputContextAndDelegationManagerMidas(d *delegationManagerMidas, eei *vmContext) *vmcommon.ContractCallInput {
//	eei.returnMessage = ""
//	vmInput := getDefaultVmInputForDelegationManagerMidas("mergeValidatorToDelegationWithWhitelist", make([][]byte, 0))
//	vmInput.Arguments = [][]byte{vmInput.CallerAddr}
//	vmInput.CallValue.SetUint64(0)
//	vmInput.GasProvided = d.gasCost.MetaChainSystemSCsCost.ValidatorToDelegation
//	eei.gasRemaining = vmInput.GasProvided
//	d.eei.SetStorageForAddress(vmInput.CallerAddr, []byte(whitelistedAddress), vmInput.CallerAddr)
//
//	return vmInput
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListMergeFailShouldErr(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//	vmInput := prepareVmInputContextAndDelegationManagerMidas(d, eei)
//
//	deleteWhiteListCalled := false
//	_ = eei.SetSystemSCContainer(
//		&mock.SystemSCContainerStub{
//			GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//				return &mock.SystemSCStub{
//					ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//						if args.Function == deleteWhitelistForMerge {
//							deleteWhiteListCalled = true
//						}
//						if args.Function == mergeValidatorDataToDelegation {
//							return vmcommon.UserError
//						}
//
//						return vmcommon.Ok
//					},
//				}, nil
//			}})
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.False(t, deleteWhiteListCalled)
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListDeleteWhitelistFailShouldErr(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//	vmInput := prepareVmInputContextAndDelegationManagerMidas(d, eei)
//
//	_ = eei.SetSystemSCContainer(
//		&mock.SystemSCContainerStub{
//			GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//				return &mock.SystemSCStub{
//					ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//						if args.Function == deleteWhitelistForMerge {
//							return vmcommon.UserError
//						}
//
//						return vmcommon.Ok
//					},
//				}, nil
//			}})
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//}

//func TestDelegationManagerSystemSCMidas_mergeValidatorToDelegationWithWhiteListShouldWork(t *testing.T) {
//	d, eei := createTestEEIAndDelegationFormMergeValidatorMidas()
//	vmInput := prepareVmInputContextAndDelegationManagerMidas(d, eei)
//
//	deleteWhiteListCalled := false
//	_ = eei.SetSystemSCContainer(
//		&mock.SystemSCContainerStub{
//			GetCalled: func(key []byte) (vm.SystemSmartContract, error) {
//				return &mock.SystemSCStub{
//					ExecuteCalled: func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
//						if args.Function == deleteWhitelistForMerge {
//							deleteWhiteListCalled = true
//						}
//
//						return vmcommon.Ok
//					},
//				}, nil
//			}})
//
//	returnCode := d.Execute(vmInput)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//	assert.True(t, deleteWhiteListCalled)
//}

//func TestDelegationManagerSystemSCMidas_MakeNewContractFromValidatorDataWithJailedNodes(t *testing.T) {
//	maxDelegationCap := []byte{0}
//	serviceFee := []byte{10}
//	args := createMockArgumentsForDelegationManagerMidas()
//	eei := createDefaultEei()
//	_ = eei.SetSystemSCContainer(
//		createSystemSCContainerMidas(eei),
//	)
//
//	args.Eei = eei
//	args.GasCost.MetaChainSystemSCsCost.ValidatorToDelegation = 100
//	d, _ := NewDelegationManagerSystemSCMidas(args)
//	vmInput := getDefaultVmInputForDelegationManagerMidas("makeNewContractFromValidatorData", [][]byte{maxDelegationCap, serviceFee})
//	vmInput.CallerAddr = bytes.Repeat([]byte{1}, 32)
//	eei.scAddress = vm.DelegationManagerSCAddress
//	_ = d.init(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
//
//	validator, _ := eei.systemContracts.Get(d.validatorSCAddr)
//	s, _ := eei.systemContracts.Get(d.stakingSCAddr)
//	staking := s.(*stakingSC)
//
//	key1 := []byte("Key1")
//	key2 := []byte("Key2")
//
//	arguments := &vmcommon.ContractCallInput{}
//	arguments.CallerAddr = AbstractStakingSCAddress
//	arguments.RecipientAddr = d.validatorSCAddr
//	arguments.Function = "stake"
//	arguments.CallValue = big.NewInt(0)
//	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key2, []byte("msg2"), vmInput.CallerAddr, big.NewInt(0).Mul(big.NewInt(2), big.NewInt(10000000)).Bytes()}
//
//	eei.scAddress = vm.ValidatorSCAddress
//	returnCode := validator.Execute(arguments)
//	assert.Equal(t, vmcommon.Ok, returnCode)
//
//	eei.scAddress = vm.StakingSCAddress
//	doJail(t, staking, staking.jailAccessAddr, key1, vmcommon.Ok)
//
//	eei.scAddress = vm.DelegationManagerSCAddress
//	vmInput.RecipientAddr = vm.DelegationManagerSCAddress
//	vmInput.Arguments = [][]byte{maxDelegationCap, serviceFee}
//	vmInput.GasProvided = 1000000
//	eei.gasRemaining = 1000000
//
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "can not migrate nodes while jailed nodes exists")
//}
//
//func TestDelegationManagerSystemSCMidas_MakeNewContractFromValidatorDataCallerAlreadyDeployedADelegationSC(t *testing.T) {
//	maxDelegationCap := []byte{0}
//	serviceFee := []byte{10}
//	args := createMockArgumentsForDelegationManagerMidas()
//	eei := createDefaultEei()
//	_ = eei.SetSystemSCContainer(
//		createSystemSCContainerMidas(eei),
//	)
//
//	caller := bytes.Repeat([]byte{1}, 32)
//
//	args.Eei = eei
//	args.GasCost.MetaChainSystemSCsCost.ValidatorToDelegation = 100
//	d, _ := NewDelegationManagerSystemSCMidas(args)
//	vmInput := getDefaultVmInputForDelegationManagerMidas("makeNewContractFromValidatorData", [][]byte{maxDelegationCap, serviceFee})
//	vmInput.CallerAddr = caller
//	eei.scAddress = vm.DelegationManagerSCAddress
//	_ = d.init(&vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
//
//	validator, _ := eei.systemContracts.Get(d.validatorSCAddr)
//	key1 := []byte("Key1")
//	key2 := []byte("Key2")
//
//	arguments := &vmcommon.ContractCallInput{}
//	arguments.CallerAddr = AbstractStakingSCAddress
//	arguments.RecipientAddr = d.validatorSCAddr
//	arguments.Function = "stake"
//	arguments.CallValue = big.NewInt(0)
//	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key2, []byte("msg2"), caller, big.NewInt(0).Mul(big.NewInt(2), big.NewInt(10000000)).Bytes()}
//
//	eei.scAddress = vm.ValidatorSCAddress
//	returnCode := validator.Execute(arguments)
//	require.Equal(t, returnCode, vmcommon.Ok)
//
//	vmInput = getDefaultVmInputForDelegationManagerMidas("makeNewContractFromValidatorData", [][]byte{maxDelegationCap, serviceFee})
//	vmInput.CallerAddr = caller
//	vmInput.RecipientAddr = vm.DelegationManagerSCAddress
//	vmInput.GasProvided = 1000000
//	eei.gasRemaining = 1000000
//	eei.returnMessage = ""
//	returnCode = d.Execute(vmInput)
//	assert.Equal(t, vmcommon.UserError, returnCode)
//	assert.Equal(t, eei.returnMessage, "caller already deployed a delegation sc")
//}

func TestDelegationManagerMidas_CorrectOwnerOnAccount(t *testing.T) {
	t.Parallel()

	delegationAddress := []byte("delegation address")
	caller := []byte("caller")
	t.Run("the fix is disabled, returns nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgumentsForDelegationManagerMidas()
		args.Eei = &mock.SystemEIStub{
			UpdateCodeDeployerAddressCalled: func(scAddress string, newOwner []byte) error {
				assert.Fail(t, "should have not called UpdateCodeDeployerAddress")
				return nil
			},
		}

		dm, _ := NewDelegationManagerSystemSCMidas(args)
		err := dm.correctOwnerOnAccount(delegationAddress, caller)
		assert.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgumentsForDelegationManagerMidas()
		epochsHandler := args.EnableEpochsHandler.(*enableEpochsHandlerMock.EnableEpochsHandlerStub)
		epochsHandler.AddActiveFlags(common.FixDelegationChangeOwnerOnAccountFlag)
		updateCalled := false
		args.Eei = &mock.SystemEIStub{
			UpdateCodeDeployerAddressCalled: func(scAddress string, newOwner []byte) error {
				assert.Equal(t, scAddress, string(delegationAddress))
				assert.Equal(t, caller, newOwner)
				updateCalled = true

				return nil
			},
		}

		dm, _ := NewDelegationManagerSystemSCMidas(args)
		err := dm.correctOwnerOnAccount(delegationAddress, caller)
		assert.Nil(t, err)
		assert.True(t, updateCalled)
	})
}
