package metachain

import (
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/vm"
	systemVMFactory "github.com/multiversx/mx-chain-go/vm/factory"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

type vmContainerFactoryMidas struct {
	vmContainerFactory
}

func NewVMContainerFactoryMidas(args ArgsNewVMContainerFactory) (*vmContainerFactoryMidas, error) {
	if check.IfNil(args.Economics) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", process.ErrNilEconomicsData)
	}
	if check.IfNil(args.MessageSignVerifier) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilMessageSignVerifier)
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", process.ErrNilNodesConfigProvider)
	}
	if check.IfNil(args.Hasher) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", process.ErrNilHasher)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", process.ErrNilMarshalizer)
	}
	if args.SystemSCConfig == nil {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", process.ErrNilSystemSCConfig)
	}
	if check.IfNil(args.ValidatorAccountsDB) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilValidatorAccountsDB)
	}
	if check.IfNil(args.UserAccountsDB) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilUserAccountsDB)
	}
	if check.IfNil(args.ChanceComputer) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilChanceComputer)
	}
	if check.IfNil(args.GasSchedule) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilGasSchedule)
	}
	if check.IfNil(args.PubkeyConv) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilAddressPubKeyConverter)
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", vm.ErrNilShardCoordinator)
	}
	if check.IfNil(args.BlockChainHook) {
		return nil, process.ErrNilBlockChainHook
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, fmt.Errorf("%w in NewVMContainerFactory", process.ErrNilNodesCoordinator)
	}

	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactoryMidas{
		vmContainerFactory: vmContainerFactory{
			blockChainHook:         args.BlockChainHook,
			cryptoHook:             cryptoHook,
			economics:              args.Economics,
			messageSigVerifier:     args.MessageSignVerifier,
			gasSchedule:            args.GasSchedule,
			nodesConfigProvider:    args.NodesConfigProvider,
			hasher:                 args.Hasher,
			marshalizer:            args.Marshalizer,
			systemSCConfig:         args.SystemSCConfig,
			validatorAccountsDB:    args.ValidatorAccountsDB,
			userAccountsDB:         args.UserAccountsDB,
			chanceComputer:         args.ChanceComputer,
			addressPubKeyConverter: args.PubkeyConv,
			shardCoordinator:       args.ShardCoordinator,
			enableEpochsHandler:    args.EnableEpochsHandler,
			nodesCoordinator:       args.NodesCoordinator,
		},
	}, nil
}

func (vmf *vmContainerFactoryMidas) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createSystemVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.SystemVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactoryMidas) CreateForGenesis() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createSystemVMForGenesis()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.SystemVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactoryMidas) createSystemVM() (vmcommon.VMExecutionHandler, error) {
	scFactory, systemEI, err := vmf.createSystemVMFactoryAndEEI()
	if err != nil {
		return nil, err
	}

	vmf.systemContracts, err = scFactory.Create()
	if err != nil {
		return nil, err
	}

	vmf.scFactory = scFactory

	return vmf.finalizeSystemVMCreation(systemEI)
}

// createSystemVMForGenesis will create the same VMExecutionHandler structure used when the mainnet genesis was created
func (vmf *vmContainerFactoryMidas) createSystemVMForGenesis() (vmcommon.VMExecutionHandler, error) {
	scFactory, systemEI, err := vmf.createSystemVMFactoryAndEEI()
	if err != nil {
		return nil, err
	}

	vmf.systemContracts, err = scFactory.CreateForGenesis()
	if err != nil {
		return nil, err
	}

	return vmf.finalizeSystemVMCreation(systemEI)
}

func (vmf *vmContainerFactoryMidas) createSystemVMFactoryAndEEI() (vm.SystemSCContainerFactory, vm.ContextHandler, error) {
	atArgumentParser := parsers.NewCallArgsParser()
	vmContextArgs := systemSmartContracts.VMContextArgs{
		BlockChainHook:      vmf.blockChainHook,
		CryptoHook:          vmf.cryptoHook,
		InputParser:         atArgumentParser,
		ValidatorAccountsDB: vmf.validatorAccountsDB,
		UserAccountsDB:      vmf.userAccountsDB,
		ChanceComputer:      vmf.chanceComputer,
		EnableEpochsHandler: vmf.enableEpochsHandler,
		ShardCoordinator:    vmf.shardCoordinator,
	}
	systemEI, err := systemSmartContracts.NewVMContext(vmContextArgs)
	if err != nil {
		return nil, nil, err
	}

	argsNewSystemScFactory := systemVMFactory.ArgsNewSystemSCFactory{
		SystemEI:               systemEI,
		SigVerifier:            vmf.messageSigVerifier,
		GasSchedule:            vmf.gasSchedule,
		NodesConfigProvider:    vmf.nodesConfigProvider,
		Hasher:                 vmf.hasher,
		Marshalizer:            vmf.marshalizer,
		SystemSCConfig:         vmf.systemSCConfig,
		Economics:              vmf.economics,
		AddressPubKeyConverter: vmf.addressPubKeyConverter,
		ShardCoordinator:       vmf.shardCoordinator,
		EnableEpochsHandler:    vmf.enableEpochsHandler,
		NodesCoordinator:       vmf.nodesCoordinator,
	}
	scFactory, err := systemVMFactory.NewSystemSCFactoryMidas(argsNewSystemScFactory)
	if err != nil {
		return nil, nil, err
	}

	return scFactory, systemEI, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactoryMidas) IsInterfaceNil() bool {
	return vmf == nil
}
