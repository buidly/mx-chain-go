//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. validator.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"math/big"
)

type validatorSCMidas struct {
	validatorSC
	abstractStakingAddr []byte
}

type ArgsValidatorSmartContractMidas struct {
	ArgsValidatorSmartContract
	AbstractStakingAddr []byte
}

// NewValidatorSmartContract creates an validator smart contract
func NewValidatorSmartContractMidas(
	args ArgsValidatorSmartContractMidas,
) (*validatorSCMidas, error) {
	if check.IfNil(args.Eei) {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilSystemEnvironmentInterface)
	}
	if len(args.StakingSCAddress) == 0 {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilStakingSmartContractAddress)
	}
	if len(args.ValidatorSCAddress) == 0 {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilValidatorSmartContractAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilMarshalizer)
	}
	if check.IfNil(args.SigVerifier) {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilMessageSignVerifier)
	}
	if args.GenesisTotalSupply == nil || args.GenesisTotalSupply.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v in validatorSC", vm.ErrInvalidGenesisTotalSupply, args.GenesisTotalSupply)
	}
	if len(args.EndOfEpochAddress) < 1 {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrInvalidEndOfEpochAccessAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address in validatorSC", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilShardCoordinator)
	}
	if len(args.GovernanceSCAddress) < 1 {
		return nil, fmt.Errorf("%w for governance sc address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, fmt.Errorf("%w in validatorSC", vm.ErrNilEnableEpochsHandler)
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.StakingV2Flag,
		common.StakeFlag,
		common.ValidatorToDelegationFlag,
		common.DoubleKeyProtectionFlag,
		common.MultiClaimOnDelegationFlag,
		common.DelegationManagerFlag,
		common.UnBondTokensV2Flag,
	})
	if err != nil {
		return nil, err
	}

	baseConfig := ValidatorConfig{
		TotalSupply: big.NewInt(0).Set(args.GenesisTotalSupply),
	}

	var okValue bool
	baseConfig.UnJailPrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.UnJailValue, conversionBase)
	if !okValue || baseConfig.UnJailPrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)
	}
	baseConfig.MinStakeValue, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !okValue || baseConfig.MinStakeValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)
	}
	baseConfig.NodePrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !okValue || baseConfig.NodePrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)
	}
	baseConfig.MinStep, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStepValue, conversionBase)
	if !okValue || baseConfig.MinStep.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStepValue, args.StakingSCConfig.MinStepValue)
	}
	minUnstakeTokensValue, okValue := big.NewInt(0).SetString(args.StakingSCConfig.MinUnstakeTokensValue, conversionBase)
	if !okValue || minUnstakeTokensValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinUnstakeTokensValue, args.StakingSCConfig.MinUnstakeTokensValue)
	}
	minDeposit, okConvert := big.NewInt(0).SetString(args.MinDeposit, conversionBase)
	if !okConvert || minDeposit.Cmp(zero) < 0 {
		return nil, vm.ErrInvalidMinCreationDeposit
	}

	return &validatorSCMidas{
		validatorSC: validatorSC{
			eei:                    args.Eei,
			unBondPeriod:           args.StakingSCConfig.UnBondPeriod,
			unBondPeriodInEpochs:   args.StakingSCConfig.UnBondPeriodInEpochs,
			sigVerifier:            args.SigVerifier,
			baseConfig:             baseConfig,
			stakingSCAddress:       args.StakingSCAddress,
			validatorSCAddress:     args.ValidatorSCAddress,
			gasCost:                args.GasCost,
			marshalizer:            args.Marshalizer,
			minUnstakeTokensValue:  minUnstakeTokensValue,
			walletAddressLen:       len(args.ValidatorSCAddress),
			endOfEpochAddress:      args.EndOfEpochAddress,
			minDeposit:             minDeposit,
			delegationMgrSCAddress: args.DelegationMgrSCAddress,
			governanceSCAddress:    args.GovernanceSCAddress,
			shardCoordinator:       args.ShardCoordinator,
			enableEpochsHandler:    args.EnableEpochsHandler,
		},
		abstractStakingAddr: args.AbstractStakingAddr,
	}, nil
}

// Execute calls one of the functions from the validator smart contract and runs the code according to the input
func (v *validatorSCMidas) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	v.mutExecution.RLock()
	defer v.mutExecution.RUnlock()

	err := CheckIfNil(args)
	if err != nil {
		v.eei.AddReturnMessage("nil arguments: " + err.Error())
		return vmcommon.UserError
	}

	if len(args.ESDTTransfers) > 0 {
		v.eei.AddReturnMessage("cannot transfer ESDT to system SCs")
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return v.init(args)
	case "stake":
		return v.stake(args)
	case "stakeNodes":
		return v.stakeNodes(args)
	case "unStake":
		return v.unStake(args)
	case "unStakeNodes":
		return v.unStakeNodes(args)
	case "unStakeTokens":
		return v.unStakeTokens(args)
	case "unBond":
		return v.unBond(args)
	case "unBondNodes":
		return v.unBondNodes(args)
	case "unBondTokens":
		return v.unBondTokens(args)
	case "get":
		return v.get(args)
	case "setConfig":
		return v.setConfig(args)
	case "unJail":
		return v.unJail(args)
	case "getTotalStaked":
		return v.getTotalStaked(args)
	case "getTotalStakedTopUpStakedBlsKeys":
		return v.getTotalStakedTopUpStakedBlsKeys(args)
	case "getBlsKeysStatus":
		return v.getBlsKeysStatus(args)
	case "cleanRegisteredData":
		return v.cleanRegisteredData(args)
	case "pauseUnStakeUnBond":
		return v.pauseUnStakeUnBond(args)
	case "unPauseUnStakeUnBond":
		return v.unPauseStakeUnBond(args)
	case "getUnStakedTokensList":
		return v.getUnStakedTokensList(args)
	case "reStakeUnStakedNodes":
		return v.reStakeUnStakedNodes(args)
	//case "mergeValidatorData":
	//	return v.mergeValidatorData(args) // TODO: These should also be overwritten or we should disable this functionality?
	//case "changeOwnerOfValidatorData":
	//	return v.changeOwnerOfValidatorData(args)
	}

	v.eei.AddReturnMessage("invalid method to call")
	return vmcommon.UserError
}

func (v *validatorSCMidas) stake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, v.abstractStakingAddr) {
		v.eei.AddReturnMessage("stake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isGenesis := v.eei.BlockChainHook().CurrentNonce() == 0
	stakeEnabled := isGenesis || v.enableEpochsHandler.IsFlagEnabled(common.StakeFlag)
	if !stakeEnabled {
		v.eei.AddReturnMessage(vm.StakeNotEnabled)
		return vmcommon.UserError
	}

	if len(args.Arguments) < 3 {
		v.eei.AddReturnMessage("invalid number of arguments to call stake")
		return vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	validatorAddress := args.Arguments[lenArgs-2]
	totalPower := big.NewInt(0).SetBytes(args.Arguments[lenArgs-1])
	lenAndBlsKeys := args.Arguments[:lenArgs-2]

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	registrationData, err := v.getOrCreateRegistrationData(validatorAddress)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	registrationData.TotalStakeValue.Set(totalPower)
	if registrationData.TotalStakeValue.Cmp(validatorConfig.NodePrice) < 0 &&
		!core.IsSmartContractAddress(validatorAddress) {
		v.eei.AddReturnMessage(
			fmt.Sprintf("insufficient stake value: expected %s, got %s",
				validatorConfig.NodePrice.String(),
				registrationData.TotalStakeValue.String(),
			),
		)
		return vmcommon.UserError
	}

	if lenArgs == 3 {
		return v.updateStakeValue(registrationData, validatorAddress)
	}

	if !isNumBlsKeysCorrectToStake(lenAndBlsKeys) {
		v.eei.AddReturnMessage("invalid number of bls keys")
		return vmcommon.UserError
	}

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	if maxNodesToRun == 0 {
		v.eei.AddReturnMessage("number of nodes argument must be greater than zero")
		return vmcommon.UserError
	}

	err = v.eei.UseGas((maxNodesToRun - 1) * v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isAlreadyRegistered := len(registrationData.RewardAddress) > 0
	if !isAlreadyRegistered {
		registrationData.RewardAddress = validatorAddress
	}

	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = v.eei.BlockChainHook().CurrentEpoch()

	blsKeys, newKeys, err := v.registerBLSKeys(registrationData, validatorAddress, validatorAddress, lenAndBlsKeys)
	if err != nil {
		v.eei.AddReturnMessage("cannot register bls key: error " + err.Error())
		return vmcommon.UserError
	}
	if v.enableEpochsHandler.IsFlagEnabled(common.DoubleKeyProtectionFlag) && checkDoubleBLSKeys(blsKeys) {
		v.eei.AddReturnMessage("invalid arguments, found same bls key twice")
		return vmcommon.UserError
	}

	numQualified := big.NewInt(0).Div(registrationData.TotalStakeValue, validatorConfig.NodePrice)
	if uint64(len(registrationData.BlsPubKeys)) > numQualified.Uint64() {
		if !v.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
			// backward compatibility
			v.eei.AddReturnMessage("insufficient funds")
			return vmcommon.OutOfFunds
		}

		if uint64(len(newKeys)) > numQualified.Uint64() {
			totalNeeded := big.NewInt(0).Mul(big.NewInt(int64(len(newKeys))), validatorConfig.NodePrice)
			v.eei.AddReturnMessage("not enough total stake to activate nodes," +
				" totalStake: " + registrationData.TotalStakeValue.String() + ", needed: " + totalNeeded.String())
			return vmcommon.UserError
		}

		numStakedJailedWaiting, _, errGet := v.getNumStakedAndWaitingNodes(registrationData, make(map[string]struct{}), false)
		if errGet != nil {
			v.eei.AddReturnMessage(errGet.Error())
			return vmcommon.UserError
		}

		numTotalNodes := uint64(len(newKeys)) + numStakedJailedWaiting
		if numTotalNodes > numQualified.Uint64() {
			totalNeeded := big.NewInt(0).Mul(big.NewInt(0).SetUint64(numTotalNodes), validatorConfig.NodePrice)
			v.eei.AddReturnMessage("not enough total stake to activate nodes," +
				" totalStake: " + registrationData.TotalStakeValue.String() + ", needed: " + totalNeeded.String())
			return vmcommon.UserError
		}
	}

	v.activateNewBLSKeysMidas(registrationData, blsKeys, newKeys, &validatorConfig, args, validatorAddress)

	err = v.saveRegistrationData(validatorAddress, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

// TODO: Test this
func (v *validatorSCMidas) stakeNodes(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isGenesis := v.eei.BlockChainHook().CurrentNonce() == 0
	stakeEnabled := isGenesis || v.enableEpochsHandler.IsFlagEnabled(common.StakeFlag)
	if !stakeEnabled {
		v.eei.AddReturnMessage(vm.StakeNotEnabled)
		return vmcommon.UserError
	}

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	registrationData, err := v.getOrCreateRegistrationData(args.CallerAddr)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	if registrationData.TotalStakeValue.Cmp(validatorConfig.NodePrice) < 0 &&
		!core.IsSmartContractAddress(args.CallerAddr) {
		v.eei.AddReturnMessage(
			fmt.Sprintf("insufficient stake value: expected %s, got %s",
				validatorConfig.NodePrice.String(),
				registrationData.TotalStakeValue.String(),
			),
		)
		return vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	if lenArgs == 0 {
		return v.updateStakeValue(registrationData, args.CallerAddr)
	}

	if !isNumBlsKeysCorrectToStake(args.Arguments) {
		v.eei.AddReturnMessage("invalid number of bls keys")
		return vmcommon.UserError
	}

	maxNodesToRun := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	if maxNodesToRun == 0 {
		v.eei.AddReturnMessage("number of nodes argument must be greater than zero")
		return vmcommon.UserError
	}

	err = v.eei.UseGas((maxNodesToRun - 1) * v.gasCost.MetaChainSystemSCsCost.Stake)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	isAlreadyRegistered := len(registrationData.RewardAddress) > 0
	if !isAlreadyRegistered {
		registrationData.RewardAddress = args.CallerAddr
	}

	registrationData.MaxStakePerNode = big.NewInt(0).Set(registrationData.TotalStakeValue)
	registrationData.Epoch = v.eei.BlockChainHook().CurrentEpoch()

	blsKeys, newKeys, err := v.registerBLSKeys(registrationData, args.CallerAddr, args.CallerAddr, args.Arguments)
	if err != nil {
		v.eei.AddReturnMessage("cannot register bls key: error " + err.Error())
		return vmcommon.UserError
	}
	if v.enableEpochsHandler.IsFlagEnabled(common.DoubleKeyProtectionFlag) && checkDoubleBLSKeys(blsKeys) {
		v.eei.AddReturnMessage("invalid arguments, found same bls key twice")
		return vmcommon.UserError
	}

	numQualified := big.NewInt(0).Div(registrationData.TotalStakeValue, validatorConfig.NodePrice)
	if uint64(len(registrationData.BlsPubKeys)) > numQualified.Uint64() {
		if !v.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
			// backward compatibility
			v.eei.AddReturnMessage("insufficient funds")
			return vmcommon.OutOfFunds
		}

		if uint64(len(newKeys)) > numQualified.Uint64() {
			totalNeeded := big.NewInt(0).Mul(big.NewInt(int64(len(newKeys))), validatorConfig.NodePrice)
			v.eei.AddReturnMessage("not enough total stake to activate nodes," +
				" totalStake: " + registrationData.TotalStakeValue.String() + ", needed: " + totalNeeded.String())
			return vmcommon.UserError
		}

		numStakedJailedWaiting, _, errGet := v.getNumStakedAndWaitingNodes(registrationData, make(map[string]struct{}), false)
		if errGet != nil {
			v.eei.AddReturnMessage(errGet.Error())
			return vmcommon.UserError
		}

		numTotalNodes := uint64(len(newKeys)) + numStakedJailedWaiting
		if numTotalNodes > numQualified.Uint64() {
			totalNeeded := big.NewInt(0).Mul(big.NewInt(0).SetUint64(numTotalNodes), validatorConfig.NodePrice)
			v.eei.AddReturnMessage("not enough total stake to activate nodes," +
				" totalStake: " + registrationData.TotalStakeValue.String() + ", needed: " + totalNeeded.String())
			return vmcommon.UserError
		}
	}

	v.activateNewBLSKeysMidas(registrationData, blsKeys, newKeys, &validatorConfig, args, args.CallerAddr)

	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}


func (v *validatorSCMidas) activateNewBLSKeysMidas(
	registrationData *ValidatorDataV2,
	blsKeys [][]byte,
	newKeys [][]byte,
	validatorConfig *ValidatorConfig,
	args *vmcommon.ContractCallInput,
	ownerAddress []byte,
) {
	numRegisteredBlsKeys := len(registrationData.BlsPubKeys)
	allNodesActivated := v.activateStakingFor(
		blsKeys,
		newKeys,
		registrationData,
		validatorConfig.NodePrice,
		registrationData.RewardAddress,
		ownerAddress,
	)

	if !allNodesActivated && len(blsKeys) > 0 {
		nodeLimit := int64(v.computeNodeLimit())
		entry := &vmcommon.LogEntry{
			Identifier: []byte(args.Function),
			Address:    args.RecipientAddr,
			Topics: [][]byte{
				[]byte(numberOfNodesTooHigh),
				big.NewInt(int64(numRegisteredBlsKeys)).Bytes(),
				big.NewInt(nodeLimit).Bytes(),
			},
		}
		v.eei.AddLogEntry(entry)
	}

}

// This is the complete unStake - which after enabling economics V2 will create unStakedFunds on the registration data
func (v *validatorSCMidas) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, v.abstractStakingAddr) {
		v.eei.AddReturnMessage("unStake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}

	registrationData, validatorAddress, totalPower, blsKeys, returnCode := v.basicChecksForUnStakeNodes(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	numSuccessFromActive, numSuccessFromWaiting := v.unStakeNodesFromStakingSC(blsKeys, registrationData)
	if !v.enableEpochsHandler.IsFlagEnabled(common.StakingV2Flag) {
		// unStakeV1 returns from this point
		return vmcommon.Ok
	}

	if totalPower.Cmp(registrationData.TotalStakeValue) > 0 {
		v.eei.AddReturnMessage("New total power after unstake can not be greater than old total power")
		return vmcommon.UserError
	}

	difference := big.NewInt(0).Sub(registrationData.TotalStakeValue, totalPower)

	unStakedEpoch := v.eei.BlockChainHook().CurrentEpoch()
	if registrationData.NumRegistered == 0 {
		unStakedEpoch = 0
	}

	if numSuccessFromActive+numSuccessFromWaiting > 0 {
		returnCode = v.processUnStakeValue(registrationData, difference, unStakedEpoch)
		if returnCode != vmcommon.Ok {
			return returnCode
		}
	}

	if registrationData.NumRegistered > 0 && registrationData.TotalStakeValue.Cmp(v.minDeposit) < 0 {
		v.eei.AddReturnMessage("cannot unStake tokens, the validator would remain without min deposit, nodes are still active")
		return vmcommon.UserError
	}

	err := v.saveRegistrationData(validatorAddress, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSCMidas) unStakeTokens(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, v.abstractStakingAddr) {
		v.eei.AddReturnMessage("unStakeTokens function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected %d, got %d", 2, len(args.Arguments)))
		return vmcommon.UserError
	}

	validatorAddress := args.Arguments[0]
	totalPower := big.NewInt(0).SetBytes(args.Arguments[1])

	registrationData, returnCode := v.basicCheckForUnStakeUnBond(args, validatorAddress)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	unStakeValue := big.NewInt(0).Sub(registrationData.TotalStakeValue, totalPower)
	if unStakeValue.Cmp(zero) < 0 {
		v.eei.AddReturnMessage("invalid value to unstake")
		return vmcommon.UserError
	}

	unStakedEpoch := v.eei.BlockChainHook().CurrentEpoch()
	if registrationData.NumRegistered == 0 {
		unStakedEpoch = 0
	}
	returnCode = v.processUnStakeValue(registrationData, unStakeValue, unStakedEpoch)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	if registrationData.NumRegistered > 0 && registrationData.TotalStakeValue.Cmp(v.minDeposit) < 0 {
		v.eei.AddReturnMessage("cannot unStake tokens, the validator would remain without min deposit, nodes are still active")
		return vmcommon.UserError
	}

	err = v.saveRegistrationData(validatorAddress, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSCMidas) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, v.abstractStakingAddr) {
		v.eei.AddReturnMessage("unBond function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}
	registrationData, validatorAddress, blsKeys, returnCode := v.checkUnBondArguments(args)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	unBondedKeys := v.unBondNodesFromStakingSC(blsKeys)

	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	totalUnBond := big.NewInt(0).Mul(validatorConfig.NodePrice, big.NewInt(int64(len(unBondedKeys))))
	if len(unBondedKeys) > 0 {
		totalUnBond, returnCode = v.unBondTokensFromRegistrationData(registrationData, totalUnBond)
		if returnCode != vmcommon.Ok {
			return returnCode
		}
	}

	returnCode = v.updateRegistrationDataAfterUnBond(registrationData, unBondedKeys, validatorAddress)
	if returnCode != vmcommon.Ok {
		return returnCode
	}

	return vmcommon.Ok
}

func (v *validatorSCMidas) unBondTokens(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, returnCode := v.basicCheckForUnStakeUnBond(args, args.CallerAddr)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if v.isUnStakeUnBondPaused() {
		v.eei.AddReturnMessage("unStake/unBond is paused as not enough total staked in protocol")
		return vmcommon.UserError
	}
	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnBondTokens)
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	valueToUnBond := big.NewInt(0)
	if len(args.Arguments) > 1 {
		v.eei.AddReturnMessage("too many arguments")
		return vmcommon.UserError
	}
	if len(args.Arguments) == 1 {
		valueToUnBond = big.NewInt(0).SetBytes(args.Arguments[0])
		if valueToUnBond.Cmp(zero) <= 0 {
			v.eei.AddReturnMessage("cannot unBond negative value or zero value")
			return vmcommon.UserError
		}
	}

	totalUnBond, returnCode := v.unBondTokensFromRegistrationData(registrationData, valueToUnBond)
	if returnCode != vmcommon.Ok {
		return returnCode
	}
	if totalUnBond.Cmp(zero) == 0 {
		v.eei.AddReturnMessage("no tokens that can be unbond at this time")
		if v.enableEpochsHandler.IsFlagEnabled(common.MultiClaimOnDelegationFlag) {
			return vmcommon.UserError
		}
		return vmcommon.Ok
	}

	if registrationData.NumRegistered > 0 && registrationData.TotalStakeValue.Cmp(v.minDeposit) < 0 {
		v.eei.AddReturnMessage("cannot unBond tokens, the validator would remain without min deposit, nodes are still active")
		return vmcommon.UserError
	}

	err = v.saveRegistrationData(args.CallerAddr, registrationData)
	if err != nil {
		v.eei.AddReturnMessage("cannot save registration data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (v *validatorSCMidas) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, v.abstractStakingAddr) {
		v.eei.AddReturnMessage("unJail function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	if !v.enableEpochsHandler.IsFlagEnabled(common.StakeFlag) {
		return v.unJailV1(args)
	}

	if len(args.Arguments) < 2 {
		v.eei.AddReturnMessage("invalid number of arguments: expected at least 2")
		return vmcommon.UserError
	}

	numBLSKeys := len(args.Arguments) - 1
	validatorAddress := args.Arguments[numBLSKeys]
	blsKeys := args.Arguments[:numBLSKeys]
	validatorConfig := v.getConfig(v.eei.BlockChainHook().CurrentEpoch())
	totalUnJailPrice := big.NewInt(0).Mul(validatorConfig.UnJailPrice, big.NewInt(int64(numBLSKeys)))

	// TODO: Add support for ESDT?
	if totalUnJailPrice.Cmp(args.CallValue) != 0 {
		v.eei.AddReturnMessage("wanted exact unjail price * numNodes")
		return vmcommon.UserError
	}

	err := v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnJail * uint64(numBLSKeys))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return vmcommon.OutOfGas
	}

	registrationData, err := v.getOrCreateRegistrationData(validatorAddress)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return vmcommon.UserError
	}

	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return vmcommon.UserError
	}

	transferBack := big.NewInt(0)
	for _, blsKey := range blsKeys {
		vmOutput, errExec := v.executeOnStakingSC([]byte("unJail@" + hex.EncodeToString(blsKey)))
		if errExec != nil || vmOutput.ReturnCode != vmcommon.Ok {
			transferBack.Add(transferBack, validatorConfig.UnJailPrice)
			v.eei.Finish(blsKey)
			v.eei.Finish([]byte{failed})
			continue
		}
	}

	// TODO: Add support for ESDT?
	if transferBack.Cmp(zero) > 0 {
		v.eei.Transfer(validatorAddress, args.RecipientAddr, transferBack, nil, 0)
	}

	finalUnJailFunds := big.NewInt(0).Sub(args.CallValue, transferBack)
	v.addToUnJailFunds(finalUnJailFunds)

	return vmcommon.Ok
}

func (v *validatorSCMidas) basicChecksForUnStakeNodes(args *vmcommon.ContractCallInput) (*ValidatorDataV2, []byte, *big.Int, [][]byte, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return nil, nil, nil, nil, vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 2, 0))
		return nil, nil, nil, nil, vmcommon.UserError
	}
	if !v.enableEpochsHandler.IsFlagEnabled(common.StakeFlag) {
		v.eei.AddReturnMessage(vm.UnStakeNotEnabled)
		return nil, nil, nil, nil, vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	validatorAddress := args.Arguments[lenArgs-2]
	totalPower := big.NewInt(0).SetBytes(args.Arguments[lenArgs-1])

	registrationData, err := v.getOrCreateRegistrationData(validatorAddress)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return nil, nil, nil, nil, vmcommon.UserError
	}

	err = v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnStake * uint64(lenArgs-2))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return nil, nil, nil, nil, vmcommon.OutOfGas
	}

	blsKeys := args.Arguments[:lenArgs-2]
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return nil, nil, nil, nil, vmcommon.UserError
	}

	return registrationData, validatorAddress, totalPower, blsKeys, vmcommon.Ok
}

func (v *validatorSCMidas) checkUnBondArguments(args *vmcommon.ContractCallInput) (*ValidatorDataV2, []byte, [][]byte, vmcommon.ReturnCode) {
	if args.CallValue.Cmp(zero) != 0 {
		v.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return nil, nil, nil, vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		v.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 2, 0))
		return nil, nil, nil, vmcommon.UserError
	}
	if !v.enableEpochsHandler.IsFlagEnabled(common.StakeFlag) {
		v.eei.AddReturnMessage(vm.UnBondNotEnabled)
		return nil, nil, nil, vmcommon.UserError
	}

	lenArgs := len(args.Arguments)
	validatorAddress := args.Arguments[lenArgs-1]

	registrationData, err := v.getOrCreateRegistrationData(validatorAddress)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetOrCreateRegistrationData + err.Error())
		return nil, nil, nil, vmcommon.UserError
	}

	blsKeys := args.Arguments[:lenArgs-1]
	err = verifyBLSPublicKeys(registrationData, blsKeys)
	if err != nil {
		v.eei.AddReturnMessage(vm.CannotGetAllBlsKeysFromRegistrationData + err.Error())
		return nil, nil, nil, vmcommon.UserError
	}

	err = v.eei.UseGas(v.gasCost.MetaChainSystemSCsCost.UnBond * uint64(lenArgs-1))
	if err != nil {
		v.eei.AddReturnMessage(vm.InsufficientGasLimit)
		return nil, nil, nil, vmcommon.OutOfGas
	}

	return registrationData, validatorAddress, blsKeys, vmcommon.Ok
}

func isNumBlsKeysCorrectToStake(args [][]byte) bool {
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()
	return uint64(len(args)) == 2*maxNodesToRun+1
}
