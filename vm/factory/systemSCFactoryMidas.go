package factory

import (
	"fmt"
	systemSmartContracts "github.com/multiversx/mx-chain-go/vm/systemSmartContracts"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/vm"
)

type systemSCFactoryMidas struct {
	systemSCFactory
}

var AbstractStakingSCAddress = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 255, 255}

func NewSystemSCFactoryMidas(args ArgsNewSystemSCFactory) (*systemSCFactoryMidas, error) {
	if check.IfNil(args.SystemEI) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilSystemEnvironmentInterface)
	}
	if check.IfNil(args.SigVerifier) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilMessageSignVerifier)
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilNodesConfigProvider)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilMarshalizer)
	}
	if check.IfNil(args.Hasher) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilHasher)
	}
	if check.IfNil(args.Economics) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilEconomicsData)
	}
	if args.SystemSCConfig == nil {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilSystemSCConfig)
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilAddressPubKeyConverter)
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilShardCoordinator)
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, fmt.Errorf("%w in NewSystemSCFactory", vm.ErrNilEnableEpochsHandler)
	}

	scf := &systemSCFactoryMidas{
		systemSCFactory{
			systemEI:               args.SystemEI,
			sigVerifier:            args.SigVerifier,
			nodesConfigProvider:    args.NodesConfigProvider,
			marshalizer:            args.Marshalizer,
			hasher:                 args.Hasher,
			systemSCConfig:         args.SystemSCConfig,
			economics:              args.Economics,
			addressPubKeyConverter: args.AddressPubKeyConverter,
			shardCoordinator:       args.ShardCoordinator,
			enableEpochsHandler:    args.EnableEpochsHandler,
		},
	}

	err := scf.createGasConfig(args.GasSchedule.LatestGasSchedule())
	if err != nil {
		return nil, err
	}

	scf.systemSCsContainer = NewSystemSCContainer()
	args.GasSchedule.RegisterNotifyHandler(scf)

	return scf, nil
}

// CreateForGenesis instantiates all the system smart contracts and returns a container containing them to be used in the genesis process
func (scf *systemSCFactoryMidas) CreateForGenesis() (vm.SystemSCContainer, error) {
	staking, err := scf.createStakingContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.StakingSCAddress, staking)
	if err != nil {
		return nil, err
	}

	validatorSC, err := scf.createValidatorContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.ValidatorSCAddress, validatorSC)
	if err != nil {
		return nil, err
	}

	esdt, err := scf.createESDTContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.ESDTSCAddress, esdt)
	if err != nil {
		return nil, err
	}

	governance, err := scf.createGovernanceContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.GovernanceSCAddress, governance)
	if err != nil {
		return nil, err
	}

	err = scf.systemEI.SetSystemSCContainer(scf.systemSCsContainer)
	if err != nil {
		return nil, err
	}

	return scf.systemSCsContainer, nil
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactoryMidas) Create() (vm.SystemSCContainer, error) {
	_, err := scf.CreateForGenesis()
	if err != nil {
		return nil, err
	}

	delegationManager, err := scf.createDelegationManagerContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.DelegationManagerSCAddress, delegationManager)
	if err != nil {
		return nil, err
	}

	delegation, err := scf.createDelegationContract()
	if err != nil {
		return nil, err
	}

	err = scf.systemSCsContainer.Add(vm.FirstDelegationSCAddress, delegation)
	if err != nil {
		return nil, err
	}

	err = scf.systemEI.SetSystemSCContainer(scf.systemSCsContainer)
	if err != nil {
		return nil, err
	}

	return scf.systemSCsContainer, nil
}

func (scf *systemSCFactoryMidas) createValidatorContract() (vm.SystemSmartContract, error) {
	args := systemSmartContracts.ArgsValidatorSmartContractMidas{
		ArgsValidatorSmartContract: systemSmartContracts.ArgsValidatorSmartContract{
			Eei:                    scf.systemEI,
			SigVerifier:            scf.sigVerifier,
			StakingSCConfig:        scf.systemSCConfig.StakingSystemSCConfig,
			StakingSCAddress:       vm.StakingSCAddress,
			EndOfEpochAddress:      vm.EndOfEpochAddress,
			ValidatorSCAddress:     vm.ValidatorSCAddress,
			GasCost:                scf.gasCost,
			Marshalizer:            scf.marshalizer,
			GenesisTotalSupply:     scf.economics.GenesisTotalSupply(),
			MinDeposit:             scf.systemSCConfig.DelegationManagerSystemSCConfig.MinCreationDeposit,
			DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
			GovernanceSCAddress:    vm.GovernanceSCAddress,
			ShardCoordinator:       scf.shardCoordinator,
			EnableEpochsHandler:    scf.enableEpochsHandler,
		},
		AbstractStakingAddr:    AbstractStakingSCAddress,
	}
	validatorSC, err := systemSmartContracts.NewValidatorSmartContractMidas(args)
	return validatorSC, err
}

func (scf *systemSCFactoryMidas) createDelegationContract() (vm.SystemSmartContract, error) {
	addTokensAddress, err := scf.addressPubKeyConverter.Decode(scf.systemSCConfig.DelegationManagerSystemSCConfig.ConfigChangeAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for DelegationManagerSystemSCConfig.ConfigChangeAddress in systemSCFactory", vm.ErrInvalidAddress)
	}

	argsDelegation := systemSmartContracts.ArgsNewDelegation{
		DelegationSCConfig:     scf.systemSCConfig.DelegationSystemSCConfig,
		StakingSCConfig:        scf.systemSCConfig.StakingSystemSCConfig,
		Eei:                    scf.systemEI,
		SigVerifier:            scf.sigVerifier,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		StakingSCAddress:       vm.StakingSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		GasCost:                scf.gasCost,
		Marshalizer:            scf.marshalizer,
		EndOfEpochAddress:      vm.EndOfEpochAddress,
		GovernanceSCAddress:    vm.GovernanceSCAddress,
		AddTokensAddress:       addTokensAddress,
		EnableEpochsHandler:    scf.enableEpochsHandler,
	}
	delegation, err := systemSmartContracts.NewDelegationSystemSCMidas(argsDelegation)
	return delegation, err
}

func (scf *systemSCFactoryMidas) createDelegationManagerContract() (vm.SystemSmartContract, error) {
	configChangeAddres, err := scf.addressPubKeyConverter.Decode(scf.systemSCConfig.DelegationManagerSystemSCConfig.ConfigChangeAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for DelegationManagerSystemSCConfig.ConfigChangeAddress in systemSCFactory", vm.ErrInvalidAddress)
	}

	argsDelegationManager := systemSmartContracts.ArgsNewDelegationManager{
		DelegationMgrSCConfig:  scf.systemSCConfig.DelegationManagerSystemSCConfig,
		DelegationSCConfig:     scf.systemSCConfig.DelegationSystemSCConfig,
		Eei:                    scf.systemEI,
		DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
		StakingSCAddress:       vm.StakingSCAddress,
		ValidatorSCAddress:     vm.ValidatorSCAddress,
		ConfigChangeAddress:    configChangeAddres,
		GasCost:                scf.gasCost,
		Marshalizer:            scf.marshalizer,
		EnableEpochsHandler:    scf.enableEpochsHandler,
	}
	delegationManager, err := systemSmartContracts.NewDelegationManagerSystemSCMidas(argsDelegationManager)
	return delegationManager, err
}
