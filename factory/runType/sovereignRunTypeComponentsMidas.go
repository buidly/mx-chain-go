package runType

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/errors"
	factoryVm "github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/state/factory"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
)

type sovereignRunTypeComponentsFactoryMidas struct {
	*runTypeComponentsFactory
	cfg config.SovereignConfig
}

// NewSovereignRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewSovereignRunTypeComponentsFactoryMidas(fact *runTypeComponentsFactory, cfg config.SovereignConfig) (*sovereignRunTypeComponentsFactoryMidas, error) {
	if check.IfNil(fact) {
		return nil, errors.ErrNilRunTypeComponentsFactory
	}

	return &sovereignRunTypeComponentsFactoryMidas{
		runTypeComponentsFactory: fact,
		cfg:                      cfg,
	}, nil
}

// Create creates the runType components
func (rcf *sovereignRunTypeComponentsFactoryMidas) Create() (*runTypeComponents, error) {
	rtc, err := rcf.runTypeComponentsFactory.Create()
	if err != nil {
		return nil, err
	}

	sovBlockChainHookHandlerFactory, err := hooks.NewSovereignBlockChainHookFactory(rtc.blockChainHookHandlerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockChainHookFactory failed: %w", err)
	}

	epochStartBootstrapperFactory, err := bootstrap.NewSovereignEpochStartBootstrapperFactory(rtc.epochStartBootstrapperCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignEpochStartBootstrapperFactory failed: %w", err)
	}

	bootstrapperFromStorageFactory, err := storageBootstrap.NewSovereignShardStorageBootstrapperFactory(rtc.bootstrapperFromStorageCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignShardStorageBootstrapperFactory failed: %w", err)
	}

	bootstrapperFactory, err := storageBootstrap.NewSovereignShardBootstrapFactory(rtc.bootstrapperCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignShardBootstrapFactory failed: %w", err)
	}

	blockProcessorFactory, err := block.NewSovereignBlockProcessorFactory(rtc.blockProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockProcessorFactory failed: %w", err)
	}

	forkDetectorFactory, err := sync.NewSovereignForkDetectorFactory(rtc.forkDetectorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignForkDetectorFactory failed: %w", err)
	}

	blockTrackerFactory, err := track.NewSovereignBlockTrackerFactory(rtc.blockTrackerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockTrackerFactory failed: %w", err)
	}

	requestHandlerFactory, err := requestHandlers.NewSovereignResolverRequestHandlerFactory(rtc.requestHandlerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignResolverRequestHandlerFactory failed: %w", err)
	}

	headerValidatorFactory, err := block.NewSovereignHeaderValidatorFactory(rtc.headerValidatorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignHeaderValidatorFactory failed: %w", err)
	}

	scheduledTxsExecutionFactory, err := preprocess.NewSovereignScheduledTxsExecutionFactory()
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignScheduledTxsExecutionFactory failed: %w", err)
	}

	transactionCoordinatorFactory, err := coordinator.NewSovereignTransactionCoordinatorFactory(rtc.transactionCoordinatorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignTransactionCoordinatorFactory failed: %w", err)
	}

	validatorStatisticsProcessorFactory, err := peer.NewSovereignValidatorStatisticsProcessorFactory(rtc.validatorStatisticsProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignValidatorStatisticsProcessorFactory failed: %w", err)
	}

	additionalStorageServiceCreator, err := storageFactory.NewSovereignAdditionalStorageServiceFactory()
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignAdditionalStorageServiceFactory failed: %w", err)
	}

	scProcessorCreator, err := processorV2.NewSovereignSCProcessFactory(rtc.scProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignSCProcessFactory failed: %w", err)
	}

	scResultPreProcessorCreator, err := preprocess.NewSovereignSmartContractResultPreProcessorFactory(rtc.scResultPreProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignSmartContractResultPreProcessorFactory failed: %w", err)
	}

	sovVMContextCreator := systemSmartContracts.NewOneShardSystemVMEEICreator()
	rtc.vmContainerMetaFactory, err = factoryVm.NewVmContainerMetaFactoryMidas(sovBlockChainHookHandlerFactory, sovVMContextCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewVmContainerMetaFactoryMidas failed: %w", err)
	}

	rtc.vmContainerShardFactory, err = factoryVm.NewVmContainerShardFactory(sovBlockChainHookHandlerFactory)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewVmContainerShardFactory failed: %w", err)
	}

	sovereignVmContainerShardCreator, err := factoryVm.NewSovereignVmContainerShardFactory(sovBlockChainHookHandlerFactory, rtc.vmContainerMetaFactory, rtc.vmContainerShardFactory)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignVmContainerShardFactory failed: %w", err)
	}

	accountsCreator, err := factory.NewSovereignAccountCreator(factory.ArgsSovereignAccountCreator{
		ArgsAccountCreator: factory.ArgsAccountCreator{
			Hasher:              rcf.coreComponents.Hasher(),
			Marshaller:          rcf.coreComponents.InternalMarshalizer(),
			EnableEpochsHandler: rcf.coreComponents.EnableEpochsHandler(),
		},
		BaseTokenID: rcf.cfg.GenesisConfig.NativeESDT,
	})
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignAccountCreator failed: %w", err)
	}

	return &runTypeComponents{
		blockChainHookHandlerCreator:        sovBlockChainHookHandlerFactory,
		epochStartBootstrapperCreator:       epochStartBootstrapperFactory,
		bootstrapperFromStorageCreator:      bootstrapperFromStorageFactory,
		bootstrapperCreator:                 bootstrapperFactory,
		blockProcessorCreator:               blockProcessorFactory,
		forkDetectorCreator:                 forkDetectorFactory,
		blockTrackerCreator:                 blockTrackerFactory,
		requestHandlerCreator:               requestHandlerFactory,
		headerValidatorCreator:              headerValidatorFactory,
		scheduledTxsExecutionCreator:        scheduledTxsExecutionFactory,
		transactionCoordinatorCreator:       transactionCoordinatorFactory,
		validatorStatisticsProcessorCreator: validatorStatisticsProcessorFactory,
		additionalStorageServiceCreator:     additionalStorageServiceCreator,
		scProcessorCreator:                  scProcessorCreator,
		scResultPreProcessorCreator:         scResultPreProcessorCreator,
		consensusModel:                      consensus.ConsensusModelV2,
		vmContainerMetaFactory:              rtc.vmContainerMetaFactory,
		vmContainerShardFactory:             sovereignVmContainerShardCreator,
		accountsCreator:                     accountsCreator,
		vmContextCreator:                    sovVMContextCreator,
	}, nil
}
