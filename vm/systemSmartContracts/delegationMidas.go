//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"bytes"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"math/big"
)

type delegationMidas struct {
	delegation
	abstractStakingAddr []byte
}

type ArgsNewDelegationMidas struct {
	ArgsNewDelegation
	AbstractStakingAddr    []byte
}

// NewDelegationSystemSC creates a new delegation system SC
func NewDelegationSystemSCMidas(args ArgsNewDelegationMidas) (*delegationMidas, error) {
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
	if len(args.GovernanceSCAddress) < 1 {
		return nil, fmt.Errorf("%w for governance sc address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.SigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}
	if args.DelegationSCConfig.MinServiceFee > args.DelegationSCConfig.MaxServiceFee {
		return nil, fmt.Errorf("%w minServiceFee bigger than maxServiceFee", vm.ErrInvalidDelegationSCConfig)
	}
	if args.DelegationSCConfig.MaxServiceFee < 1 {
		return nil, fmt.Errorf("%w maxServiceFee must be more than 0", vm.ErrInvalidDelegationSCConfig)
	}
	if len(args.AddTokensAddress) < 1 {
		return nil, fmt.Errorf("%w for add tokens address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.AddTokensToDelegationFlag,
		common.DelegationSmartContractFlag,
		common.ChangeDelegationOwnerFlag,
		common.ReDelegateBelowMinCheckFlag,
		common.ValidatorToDelegationFlag,
		common.DeleteDelegatorAfterClaimRewardsFlag,
		common.ComputeRewardCheckpointFlag,
		common.StakingV2FlagAfterEpoch,
		common.FixDelegationChangeOwnerOnAccountFlag,
		common.MultiClaimOnDelegationFlag,
	})
	if err != nil {
		return nil, err
	}

	d := &delegationMidas{
		delegation: delegation{
			eei:                    args.Eei,
			stakingSCAddr:          args.StakingSCAddress,
			validatorSCAddr:        args.ValidatorSCAddress,
			delegationMgrSCAddress: args.DelegationMgrSCAddress,
			gasCost:                args.GasCost,
			marshalizer:            args.Marshalizer,
			minServiceFee:          args.DelegationSCConfig.MinServiceFee,
			maxServiceFee:          args.DelegationSCConfig.MaxServiceFee,
			sigVerifier:            args.SigVerifier,
			unBondPeriodInEpochs:   args.StakingSCConfig.UnBondPeriodInEpochs,
			endOfEpochAddr:         args.EndOfEpochAddress,
			governanceSCAddr:       args.GovernanceSCAddress,
			addTokensAddr:          args.AddTokensAddress,
			enableEpochsHandler:    args.EnableEpochsHandler,
		},
		abstractStakingAddr: args.AbstractStakingAddr,
	}

	var okValue bool

	d.unJailPrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.UnJailValue, conversionBase)
	if !okValue || d.unJailPrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)
	}
	d.minStakeValue, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !okValue || d.minStakeValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)
	}
	d.nodePrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !okValue || d.nodePrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)
	}

	return d, nil
}

// Execute calls one of the functions from the delegation contract and runs the code according to the input
func (d *delegationMidas) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	d.mutExecution.RLock()
	defer d.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !d.enableEpochsHandler.IsFlagEnabled(common.DelegationSmartContractFlag) {
		d.eei.AddReturnMessage("delegation contract is not enabled")
		return vmcommon.UserError
	}
	if bytes.Equal(args.RecipientAddr, vm.FirstDelegationSCAddress) {
		d.eei.AddReturnMessage("first delegation sc address cannot be called")
		return vmcommon.UserError
	}

	if len(args.ESDTTransfers) > 0 {
		d.eei.AddReturnMessage("cannot transfer ESDT to system SCs")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return d.init(args)
	case initFromValidatorData:
		return d.initFromValidatorData(args)
	case mergeValidatorDataToDelegation:
		return d.mergeValidatorDataToDelegation(args)
	case "whitelistForMerge":
		return d.whitelistForMerge(args)
	case deleteWhitelistForMerge:
		return d.deleteWhitelistForMerge(args)
	case "getWhitelistForMerge":
		return d.getWhitelistForMerge(args)
	case "addNodes":
		return d.addNodes(args)
	case "removeNodes":
		return d.removeNodes(args)
	case "stakeNodes":
		return d.stakeNodes(args)
	case "unStakeNodes":
		return d.unStakeNodes(args)
	case "unBondNodes":
		return d.unBondNodes(args)
	case "unJailNodes":
		return d.unJailNodes(args)
	case delegate:
		return d.delegate(args)
	case "unDelegate":
		return d.unDelegate(args)
	case withdraw:
		return d.withdraw(args)
	case "changeServiceFee":
		return d.changeServiceFee(args)
	case "setAutomaticActivation":
		return d.setAutomaticActivation(args)
	case "modifyTotalDelegationCap":
		return d.modifyTotalDelegationCap(args)
	case "updateRewards":
		return d.updateRewards(args)
	case claimRewards:
		return d.claimRewards(args)
	case "getRewardData":
		return d.getRewardData(args)
	case "getClaimableRewards":
		return d.getClaimableRewards(args)
	case "getTotalCumulatedRewards":
		return d.getTotalCumulatedRewards(args)
	case "getNumUsers":
		return d.getNumUsers(args)
	case "getTotalUnStaked":
		return d.getTotalUnStaked(args)
	case "getTotalActiveStake":
		return d.getTotalActiveStake(args)
	case "getUserActiveStake":
		return d.getUserActiveStake(args)
	case "getUserUnStakedValue":
		return d.getUserUnStakedValue(args)
	case "getUserUnBondable":
		return d.getUserUnBondable(args)
	case "getUserUnDelegatedList":
		return d.getUserUnDelegatedList(args)
	case "getNumNodes":
		return d.getNumNodes(args)
	case "getAllNodeStates":
		return d.getAllNodeStates(args)
	case "getContractConfig":
		return d.getContractConfig(args)
	case "unStakeAtEndOfEpoch":
		return d.unStakeAtEndOfEpoch(args)
	case "reStakeUnStakedNodes":
		return d.reStakeUnStakedNodes(args)
	case "isDelegator":
		return d.isDelegator(args)
	case "getDelegatorFundsData":
		return d.getDelegatorFundsData(args)
	case "getTotalCumulatedRewardsForUser":
		return d.getTotalCumulatedRewardsForUser(args)
	case "setMetaData":
		return d.setMetaData(args)
	case "getMetaData":
		return d.getMetaData(args)
	case "addTokens":
		return d.addTokens(args)
	case "correctNodesStatus":
		return d.correctNodesStatus(args)
	case changeOwner:
		return d.changeOwner(args)
	case "synchronizeOwner":
		return d.synchronizeOwner(args)
	}

	d.eei.AddReturnMessage(args.Function + " is an unknown function")
	return vmcommon.UserError
}

func (d *delegationMidas) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := d.eei.GetStorage([]byte(ownerKey))
	if len(ownerAddress) != 0 {
		d.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("invalid number of arguments to init delegation contract")
		return vmcommon.UserError
	}
	serviceFee := big.NewInt(0).SetBytes(args.Arguments[1]).Uint64()
	if serviceFee < d.minServiceFee || serviceFee > d.maxServiceFee {
		d.eei.AddReturnMessage("service fee out of bounds")
		return vmcommon.UserError
	}
	maxDelegationCap := big.NewInt(0).SetBytes(args.Arguments[0])
	if maxDelegationCap.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("invalid max delegation cap")
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage("invalid call value")
		return vmcommon.UserError
	}

	initialOwnerFunds := big.NewInt(0).Set(args.CallValue)
	ownerAddress = args.CallerAddr
	returnCode := d.initDelegationStructures(initialOwnerFunds, args.CallerAddr, serviceFee, maxDelegationCap)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (d *delegationMidas) delegate(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.abstractStakingAddr) {
		d.eei.AddReturnMessage("delegate function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage("invalid call value")
		return vmcommon.UserError
	}

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)
	if err != nil {
		d.eei.AddReturnMessage("error getting minimum delegation amount " + err.Error())
		return vmcommon.UserError
	}

	delegatorAddress := args.Arguments[0]
	totalPower := big.NewInt(0).SetBytes(args.Arguments[1])

	_, delegator, err := d.getOrCreateDelegatorData(delegatorAddress)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	activeFund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalPowerAdded := big.NewInt(0).Sub(totalPower, activeFund.Value)
	if totalPowerAdded.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("invalid value to delegate")
		return vmcommon.UserError
	}

	minDelegationAmount := delegationManagement.MinDelegationAmount
	if totalPowerAdded.Cmp(minDelegationAmount) < 0 {
		d.eei.AddReturnMessage("delegate value must be higher than minDelegationAmount " + minDelegationAmount.String())
		return vmcommon.UserError
	}
	err = d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	dStatus, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return d.delegateUser(args, totalPowerAdded, totalPowerAdded, delegatorAddress, dStatus)
}

func (d *delegationMidas) unDelegate(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if !bytes.Equal(args.CallerAddr, d.abstractStakingAddr) {
		d.eei.AddReturnMessage("unDelegate function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 2 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}

	delegatorAddress := args.Arguments[0]
	totalPower := big.NewInt(0).SetBytes(args.Arguments[1])

	isNew, delegator, err := d.getOrCreateDelegatorData(delegatorAddress)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("caller is not a delegator")
		return vmcommon.UserError
	}

	activeFund, err := d.getFund(delegator.ActiveFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalPowerSubstracted := big.NewInt(0).Sub(activeFund.Value, totalPower)
	if totalPowerSubstracted.Cmp(zero) < 0 {
		d.eei.AddReturnMessage("invalid value to undelegate")
		return vmcommon.UserError
	}

	delegationManagement, err := getDelegationManagement(d.eei, d.marshalizer, d.delegationMgrSCAddress)
	if err != nil {
		d.eei.AddReturnMessage("error getting minimum delegation amount " + err.Error())
		return vmcommon.UserError
	}

	minDelegationAmount := delegationManagement.MinDelegationAmount

	if totalPower.Cmp(zero) > 0 && totalPower.Cmp(minDelegationAmount) < 0 {
		d.eei.AddReturnMessage("invalid value to undelegate - need to undelegate all - do not leave dust behind")
		return vmcommon.UserError
	}
	err = d.checkOwnerCanUnDelegate(delegatorAddress, activeFund, totalPowerSubstracted)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = d.computeAndUpdateRewards(delegatorAddress, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	returnData, returnCode := d.executeOnValidatorSCWithValueInArgs(args.RecipientAddr, "unStakeTokens", totalPowerSubstracted)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	actualUserUnStake, err := d.resolveUnStakedUnBondResponse(returnData, totalPowerSubstracted)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	activeFund.Value.Sub(activeFund.Value, actualUserUnStake)
	err = d.saveFund(delegator.ActiveFund, activeFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.addNewUnStakedFund(delegatorAddress, delegator, actualUserUnStake)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	globalFund.TotalActive.Sub(globalFund.TotalActive, actualUserUnStake)
	globalFund.TotalUnStaked.Add(globalFund.TotalUnStaked, actualUserUnStake)

	if len(delegator.UnStakedFunds) > maxNumOfUnStakedFunds {
		d.eei.AddReturnMessage("number of unDelegate limit reached, withDraw required")
		return vmcommon.UserError
	}

	if activeFund.Value.Cmp(zero) == 0 {
		delegator.ActiveFund = nil
	}

	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	err = d.saveDelegatorData(delegatorAddress, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	zeroValueByteSlice := make([]byte, 0)
	unDelegateFundKey := delegator.UnStakedFunds[len(delegator.UnStakedFunds)-1]
	d.createAndAddLogEntry(args, totalPowerSubstracted.Bytes(), totalPower.Bytes(), zeroValueByteSlice, globalFund.TotalActive.Bytes(), unDelegateFundKey)

	return vmcommon.Ok
}

func (d *delegationMidas) withdraw(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.abstractStakingAddr) {
		d.eei.AddReturnMessage("withdraw function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(zero) != 0 {
		d.eei.AddReturnMessage(vm.ErrCallValueMustBeZero.Error())
		return vmcommon.UserError
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	delegatorAddress := args.Arguments[0]

	isNew, delegator, err := d.getOrCreateDelegatorData(delegatorAddress)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("caller is not a delegator")
		return vmcommon.UserError
	}

	dConfig, err := d.getDelegationContractConfig()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	globalFund, err := d.getGlobalFundData()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	totalUnBondable, err := d.getUnBondableTokens(delegator, dConfig.UnBondPeriodInEpochs)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if totalUnBondable.Cmp(zero) == 0 {
		d.eei.AddReturnMessage("nothing to unBond")
		if d.enableEpochsHandler.IsFlagEnabled(common.MultiClaimOnDelegationFlag) {
			return vmcommon.UserError
		}
		return vmcommon.Ok
	}

	if globalFund.TotalUnStaked.Cmp(totalUnBondable) < 0 {
		d.eei.AddReturnMessage("cannot unBond - contract error")
		return vmcommon.UserError
	}

	returnData, returnCode := d.executeOnValidatorSCWithValueInArgs(args.RecipientAddr, "unBondTokens", totalUnBondable)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	actualUserUnBond, err := d.resolveUnStakedUnBondResponse(returnData, totalUnBondable)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	currentEpoch := d.eei.BlockChainHook().CurrentEpoch()
	totalUnBonded := big.NewInt(0)
	tempUnStakedFunds := make([][]byte, 0)
	var fund *Fund
	withdrawFundKeys := make([][]byte, 0)
	for fundIndex, fundKey := range delegator.UnStakedFunds {
		fund, err = d.getFund(fundKey)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
		if currentEpoch-fund.Epoch < dConfig.UnBondPeriodInEpochs {
			tempUnStakedFunds = append(tempUnStakedFunds, delegator.UnStakedFunds[fundIndex])
			continue
		}

		totalUnBonded.Add(totalUnBonded, fund.Value)
		if totalUnBonded.Cmp(actualUserUnBond) > 0 {
			unBondedFromThisFund := big.NewInt(0).Sub(totalUnBonded, actualUserUnBond)
			fund.Value.Sub(fund.Value, unBondedFromThisFund)
			err = d.saveFund(fundKey, fund)
			if err != nil {
				d.eei.AddReturnMessage(err.Error())
				return vmcommon.UserError
			}
			break
		}

		withdrawFundKeys = append(withdrawFundKeys, fundKey)
		d.eei.SetStorage(fundKey, nil)
	}
	delegator.UnStakedFunds = tempUnStakedFunds

	globalFund.TotalUnStaked.Sub(globalFund.TotalUnStaked, actualUserUnBond)
	err = d.saveGlobalFundData(globalFund)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	err = d.saveDelegatorData(args.CallerAddr, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	var wasDeleted bool
	wasDeleted, err = d.deleteDelegatorOnWithdrawIfNeeded(delegatorAddress, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	d.createAndAddLogEntryForWithdraw(args.Function, delegatorAddress, actualUserUnBond, globalFund, delegator, d.numUsers(), wasDeleted, withdrawFundKeys)

	return vmcommon.Ok
}

func (d *delegationMidas) unJailNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, d.abstractStakingAddr) {
		d.eei.AddReturnMessage("unJailNodes function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) > 2 {
		d.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}

	lenArgs := len(args.Arguments)
	delegatorAddress := args.Arguments[lenArgs - 1]
	blsKeys := args.Arguments[:lenArgs - 1]

	duplicates := checkForDuplicates(blsKeys)
	if duplicates {
		d.eei.AddReturnMessage(vm.ErrDuplicatesFoundInArguments.Error())
		return vmcommon.UserError
	}
	status, err := d.getDelegationStatus()
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	listToCheck := append(status.StakedKeys, status.UnStakedKeys...)
	foundAll := verifyIfAllBLSPubKeysExist(listToCheck, blsKeys)
	if !foundAll {
		d.eei.AddReturnMessage(vm.ErrBLSPublicKeyMismatch.Error())
		return vmcommon.UserError
	}

	isNew, delegator, err := d.getOrCreateDelegatorData(delegatorAddress)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew || len(delegator.ActiveFund) == 0 {
		d.eei.AddReturnMessage("not a delegator")
		return vmcommon.UserError
	}

	// TODO: Add support for ESDT
	//vmOutput, err := d.executeOnValidatorSC(args.RecipientAddr, "unJail", blsKeys, args.CallValue)
	//if err != nil {
	//	d.eei.AddReturnMessage(err.Error())
	//	return vmcommon.UserError
	//}
	//if vmOutput.ReturnCode != vmcommon.Ok {
	//	return vmOutput.ReturnCode
	//}
	//
	//sendBackValue := getTransferBackFromVMOutput(vmOutput)
	//if sendBackValue.Cmp(zero) > 0 {
	//	err = d.eei.Transfer(args.CallerAddr, args.RecipientAddr, sendBackValue, nil, 0)
	//	if err != nil {
	//		d.eei.AddReturnMessage(err.Error())
	//		return vmcommon.UserError
	//	}
	//}

	d.createAndAddLogEntry(args, blsKeys...)

	return vmcommon.Ok
}

func (d *delegationMidas) claimRewards(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := d.eei.UseGas(d.gasCost.MetaChainSystemSCsCost.DelegationOps)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.OutOfGas
	}
	if !bytes.Equal(args.CallerAddr, d.abstractStakingAddr) {
		d.eei.AddReturnMessage("unJailNodes function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		d.eei.AddReturnMessage("wrong number of arguments")
		return vmcommon.FunctionWrongSignature
	}

	delegatorAddress := args.Arguments[0]

	isNew, delegator, err := d.getOrCreateDelegatorData(delegatorAddress)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if isNew {
		d.eei.AddReturnMessage("caller is not a delegator")
		return vmcommon.UserError
	}

	err = d.computeAndUpdateRewards(delegatorAddress, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	// Send rewards back to Abstract Staking
	err = d.eei.Transfer(args.CallerAddr, args.RecipientAddr, delegator.UnClaimedRewards, nil, 0)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	unclaimedRewardsBytes := delegator.UnClaimedRewards.Bytes()
	delegator.TotalCumulatedRewards.Add(delegator.TotalCumulatedRewards, delegator.UnClaimedRewards)
	delegator.UnClaimedRewards.SetUint64(0)
	err = d.saveDelegatorData(delegatorAddress, delegator)
	if err != nil {
		d.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	var wasDeleted bool
	if d.enableEpochsHandler.IsFlagEnabled(common.DeleteDelegatorAfterClaimRewardsFlag) {
		wasDeleted, err = d.deleteDelegatorOnClaimRewardsIfNeeded(delegatorAddress, delegator)
		if err != nil {
			d.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	d.createAndAddLogEntry(args, unclaimedRewardsBytes, boolToSlice(wasDeleted))

	return vmcommon.Ok
}
