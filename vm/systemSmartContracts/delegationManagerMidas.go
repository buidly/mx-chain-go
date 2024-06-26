//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"math/big"
)

type delegationManagerMidas struct {
	delegationManager
}

// NewDelegationManagerSystemSC creates a new delegation manager system SC
func NewDelegationManagerSystemSCMidas(args ArgsNewDelegationManager) (*delegationManagerMidas, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingSCAddress) < 1 {
		return nil, fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	}
	if len(args.ValidatorSCAddress) < 1 {
		return nil, fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	}
	if len(args.ConfigChangeAddress) < 1 {
		return nil, fmt.Errorf("%w for config change address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.DelegationManagerFlag,
		common.ValidatorToDelegationFlag,
		common.FixDelegationChangeOwnerOnAccountFlag,
		common.MultiClaimOnDelegationFlag,
	})
	if err != nil {
		return nil, err
	}

	minCreationDeposit, okConvert := big.NewInt(0).SetString(args.DelegationMgrSCConfig.MinCreationDeposit, conversionBase)
	if !okConvert || minCreationDeposit.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidMinCreationDeposit
	}

	minDelegationAmount, okConvert := big.NewInt(0).SetString(args.DelegationMgrSCConfig.MinStakeAmount, conversionBase)
	if !okConvert || minDelegationAmount.Cmp(zero) <= 0 {
		return nil, vm.ErrInvalidMinStakeValue
	}

	d := &delegationManagerMidas{
		delegationManager{
			eei:                    args.Eei,
			stakingSCAddr:          args.StakingSCAddress,
			validatorSCAddr:        args.ValidatorSCAddress,
			delegationMgrSCAddress: args.DelegationMgrSCAddress,
			configChangeAddr:       args.ConfigChangeAddress,
			gasCost:                args.GasCost,
			marshalizer:            args.Marshalizer,
			minCreationDeposit:     minCreationDeposit,
			minDelegationAmount:    minDelegationAmount,
			minFee:                 args.DelegationSCConfig.MinServiceFee,
			maxFee:                 args.DelegationSCConfig.MaxServiceFee,
			enableEpochsHandler:    args.EnableEpochsHandler,
		},
	}

	return d, nil
}

// Execute calls one of the functions from the delegation manager contract and runs the code according to the input
func (d *delegationManagerMidas) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	d.mutExecution.RLock()
	defer d.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if !d.enableEpochsHandler.IsFlagEnabled(common.DelegationManagerFlag) {
		d.eei.AddReturnMessage("delegation manager contract is not enabled")
		return vmcommon.UserError
	}

	if len(args.ESDTTransfers) > 0 {
		d.eei.AddReturnMessage("cannot transfer ESDT to system SCs")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return d.init(args) // TODO: delegationManagementData doesn't seem to be properly saved?
	case "createNewDelegationContract":
		return d.createNewDelegationContract(args)
	case "getAllContractAddresses":
		return d.getAllContractAddresses(args)
	case "getContractConfig":
		return d.getContractConfig(args)
	}

	d.eei.AddReturnMessage("invalid function to call")
	return vmcommon.UserError
}

func (d *delegationManagerMidas) createNewDelegationContract(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}

	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationMgrOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if d.callerAlreadyDeployed(args.CallerAddr) {
		d.eei.AddReturnMessage("caller already deployed a delegation sc")
		return vmcommon.UserError
	}

	_, returnCode := d.deployNewContract(args, false, core.SCDeployInitFunctionName, args.CallerAddr, big.NewInt(0), args.Arguments)

	return returnCode
}
