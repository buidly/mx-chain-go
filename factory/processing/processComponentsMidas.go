package processing

import (
	"errors"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/epochProviders"
	"github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/fallback"
	"github.com/multiversx/mx-chain-go/genesis/checking"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
	"github.com/multiversx/mx-chain-go/process/block/pendingMb"
	"github.com/multiversx/mx-chain-go/process/block/poolsCleaner"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/receipts"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/process/transactionLog"
	"github.com/multiversx/mx-chain-go/process/txsSender"
	"github.com/multiversx/mx-chain-go/redundancy"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"math/big"
	"time"
)

type processComponentsFactoryMidas struct {
	processComponentsFactory
}

func NewProcessComponentsFactoryMidas(args ProcessComponentsFactoryArgs) (*processComponentsFactoryMidas, error) {
	err := checkProcessComponentsArgs(args)
	if err != nil {
		return nil, err
	}

	return &processComponentsFactoryMidas{
		processComponentsFactory: processComponentsFactory{
			config:                                args.Config,
			epochConfig:                           args.EpochConfig,
			prefConfigs:                           args.PrefConfigs,
			importDBConfig:                        args.ImportDBConfig,
			economicsConfig:                       args.EconomicsConfig,
			accountsParser:                        args.AccountsParser,
			smartContractParser:                   args.SmartContractParser,
			gasSchedule:                           args.GasSchedule,
			nodesCoordinator:                      args.NodesCoordinator,
			data:                                  args.Data,
			coreData:                              args.CoreData,
			crypto:                                args.Crypto,
			state:                                 args.State,
			network:                               args.Network,
			bootstrapComponents:                   args.BootstrapComponents,
			statusComponents:                      args.StatusComponents,
			requestedItemsHandler:                 args.RequestedItemsHandler,
			whiteListHandler:                      args.WhiteListHandler,
			whiteListerVerifiedTxs:                args.WhiteListerVerifiedTxs,
			maxRating:                             args.MaxRating,
			systemSCConfig:                        args.SystemSCConfig,
			importStartHandler:                    args.ImportStartHandler,
			historyRepo:                           args.HistoryRepo,
			epochNotifier:                         args.CoreData.EpochNotifier(),
			statusCoreComponents:                  args.StatusCoreComponents,
			flagsConfig:                           args.FlagsConfig,
			txExecutionOrderHandler:               args.TxExecutionOrderHandler,
			genesisNonce:                          args.GenesisNonce,
			genesisRound:                          args.GenesisRound,
			roundConfig:                           args.RoundConfig,
			runTypeComponents:                     args.RunTypeComponents,
			shardCoordinatorFactory:               args.ShardCoordinatorFactory,
			genesisBlockCreatorFactory:            args.GenesisBlockCreatorFactory,
			genesisMetaBlockChecker:               args.GenesisMetaBlockChecker,
			requesterContainerFactoryCreator:      args.RequesterContainerFactoryCreator,
			incomingHeaderSubscriber:              args.IncomingHeaderSubscriber,
			interceptorsContainerFactoryCreator:   args.InterceptorsContainerFactoryCreator,
			shardResolversContainerFactoryCreator: args.ShardResolversContainerFactoryCreator,
			txPreprocessorCreator:                 args.TxPreProcessorCreator,
			extraHeaderSigVerifierHolder:          args.ExtraHeaderSigVerifierHolder,
			outGoingOperationsPool:                args.OutGoingOperationsPool,
			dataCodec:                             args.DataCodec,
			topicsChecker:                         args.TopicsChecker,
		},
	}, nil
}

func (pcf *processComponentsFactoryMidas) Create() (*processComponents, error) {
	currentEpochProvider, err := epochProviders.CreateCurrentEpochProvider(
		pcf.config,
		pcf.coreData.GenesisNodesSetup().GetRoundDuration(),
		pcf.coreData.GenesisTime().Unix(),
		pcf.prefConfigs.Preferences.FullArchive,
	)
	if err != nil {
		return nil, err
	}

	pcf.epochNotifier.RegisterNotifyHandler(currentEpochProvider)

	fallbackHeaderValidator, err := fallback.NewFallbackHeaderValidator(
		pcf.data.Datapool().Headers(),
		pcf.coreData.InternalMarshalizer(),
		pcf.data.StorageService(),
	)
	if err != nil {
		return nil, err
	}

	argsHeaderSig := &headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:                  pcf.coreData.InternalMarshalizer(),
		Hasher:                       pcf.coreData.Hasher(),
		NodesCoordinator:             pcf.nodesCoordinator,
		MultiSigContainer:            pcf.crypto.MultiSignerContainer(),
		SingleSigVerifier:            pcf.crypto.BlockSigner(),
		KeyGen:                       pcf.crypto.BlockSignKeyGen(),
		FallbackHeaderValidator:      fallbackHeaderValidator,
		ExtraHeaderSigVerifierHolder: pcf.extraHeaderSigVerifierHolder,
	}
	headerSigVerifier, err := headerCheck.NewHeaderSigVerifier(argsHeaderSig)
	if err != nil {
		return nil, err
	}

	mainPeerShardMapper, err := pcf.prepareNetworkShardingCollectorForMessenger(pcf.network.NetworkMessenger())
	if err != nil {
		return nil, err
	}
	fullArchivePeerShardMapper, err := pcf.prepareNetworkShardingCollectorForMessenger(pcf.network.FullArchiveNetworkMessenger())
	if err != nil {
		return nil, err
	}

	err = pcf.network.InputAntiFloodHandler().SetPeerValidatorMapper(mainPeerShardMapper)
	if err != nil {
		return nil, err
	}

	resolversContainerFactory, err := pcf.newResolverContainerFactory()
	if err != nil {
		return nil, err
	}

	resolversContainer, err := resolversContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	requestersContainerFactory, err := pcf.newRequestersContainerFactory(currentEpochProvider)
	if err != nil {
		return nil, err
	}

	requestersContainer, err := requestersContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	requestersFinder, err := containers.NewRequestersFinder(requestersContainer, pcf.bootstrapComponents.ShardCoordinator())
	if err != nil {
		return nil, err
	}

	requestHandler, err := pcf.createResolverRequestHandler(requestersFinder)
	if err != nil {
		return nil, err
	}

	txLogsStorage, err := pcf.data.StorageService().GetStorer(dataRetriever.TxLogsUnit)
	if err != nil {
		return nil, err
	}

	if !pcf.config.LogsAndEvents.SaveInStorageEnabled && pcf.config.DbLookupExtensions.Enabled {
		log.Warn("processComponentsFactory.Create() node will save logs in storage because DbLookupExtensions is enabled")
	}

	saveLogsInStorage := pcf.config.LogsAndEvents.SaveInStorageEnabled || pcf.config.DbLookupExtensions.Enabled
	txLogsProcessor, err := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:               txLogsStorage,
		Marshalizer:          pcf.coreData.InternalMarshalizer(),
		SaveInStorageEnabled: saveLogsInStorage,
	})
	if err != nil {
		return nil, err
	}

	pcf.txLogsProcessor = txLogsProcessor
	genesisBlocks, initialTxs, err := pcf.generateGenesisHeadersAndApplyInitialBalances()
	if err != nil {
		return nil, err
	}

	genesisAccounts, err := pcf.indexAndReturnGenesisAccounts()
	if err != nil {
		log.Warn("cannot index genesis accounts", "error", err)
	}

	err = pcf.setGenesisHeader(genesisBlocks)
	if err != nil {
		return nil, err
	}

	validatorStatisticsProcessor, err := pcf.newValidatorStatisticsProcessor()
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := validatorStatisticsProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	err = pcf.genesisMetaBlockChecker.SetValidatorRootHashOnGenesisMetaBlock(genesisBlocks[core.MetachainShardId], validatorStatsRootHash)
	if err != nil {
		return nil, err
	}

	epochStartTrigger, err := pcf.newEpochStartTrigger(requestHandler)
	if err != nil {
		return nil, err
	}

	requestHandler.SetEpoch(epochStartTrigger.Epoch())

	err = dataRetriever.SetEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return nil, err
	}
	err = dataRetriever.SetEpochHandlerToHdrRequester(requestersContainer, epochStartTrigger)
	if err != nil {
		return nil, err
	}

	log.Debug("Validator stats created", "validatorStatsRootHash", validatorStatsRootHash)

	err = pcf.prepareGenesisBlock(genesisBlocks)
	if err != nil {
		return nil, err
	}

	bootStr, err := pcf.data.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return nil, err
	}

	bootStorer, err := bootstrapStorage.NewBootstrapStorer(pcf.coreData.InternalMarshalizer(), bootStr)
	if err != nil {
		return nil, err
	}

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      pcf.coreData.Hasher(),
		Marshalizer: pcf.coreData.InternalMarshalizer(),
	}
	headerValidator, err := pcf.runTypeComponents.HeaderValidatorCreator().CreateHeaderValidator(argsHeaderValidator)
	if err != nil {
		return nil, err
	}

	blockTracker, err := pcf.newBlockTracker(
		headerValidator,
		requestHandler,
		genesisBlocks,
	)
	if err != nil {
		return nil, err
	}

	argsMiniBlocksPoolsCleaner := poolsCleaner.ArgMiniBlocksPoolsCleaner{
		ArgBasePoolsCleaner: poolsCleaner.ArgBasePoolsCleaner{
			RoundHandler:                   pcf.coreData.RoundHandler(),
			ShardCoordinator:               pcf.bootstrapComponents.ShardCoordinator(),
			MaxRoundsToKeepUnprocessedData: pcf.config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedMiniBlocks,
		},
		MiniblocksPool: pcf.data.Datapool().MiniBlocks(),
	}
	mbsPoolsCleaner, err := poolsCleaner.NewMiniBlocksPoolsCleaner(argsMiniBlocksPoolsCleaner)
	if err != nil {
		return nil, err
	}

	mbsPoolsCleaner.StartCleaning()

	argsBasePoolsCleaner := poolsCleaner.ArgTxsPoolsCleaner{
		ArgBasePoolsCleaner: poolsCleaner.ArgBasePoolsCleaner{
			RoundHandler:                   pcf.coreData.RoundHandler(),
			ShardCoordinator:               pcf.bootstrapComponents.ShardCoordinator(),
			MaxRoundsToKeepUnprocessedData: pcf.config.PoolsCleanersConfig.MaxRoundsToKeepUnprocessedTransactions,
		},
		AddressPubkeyConverter: pcf.coreData.AddressPubKeyConverter(),
		DataPool:               pcf.data.Datapool(),
	}
	txsPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(argsBasePoolsCleaner)
	if err != nil {
		return nil, err
	}

	txsPoolsCleaner.StartCleaning()

	_, err = track.NewMiniBlockTrack(
		pcf.data.Datapool(),
		pcf.bootstrapComponents.ShardCoordinator(),
		pcf.whiteListHandler,
	)
	if err != nil {
		return nil, err
	}

	hardforkTrigger, err := pcf.createHardforkTrigger(epochStartTrigger)
	if err != nil {
		return nil, err
	}

	interceptorContainerFactory, blackListHandler, err := pcf.newInterceptorContainerFactory(
		headerSigVerifier,
		pcf.bootstrapComponents.HeaderIntegrityVerifier(),
		blockTracker,
		epochStartTrigger,
		requestHandler,
		mainPeerShardMapper,
		fullArchivePeerShardMapper,
		hardforkTrigger,
	)
	if err != nil {
		return nil, err
	}

	// TODO refactor all these factory calls
	mainInterceptorsContainer, fullArchiveInterceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	exportFactoryHandler, err := pcf.createExportFactoryHandler(
		headerValidator,
		requestHandler,
		resolversContainer,
		requestersContainer,
		mainInterceptorsContainer,
		fullArchiveInterceptorsContainer,
		headerSigVerifier,
		blockTracker,
	)
	if err != nil {
		return nil, err
	}

	err = hardforkTrigger.SetExportFactoryHandler(exportFactoryHandler)
	if err != nil {
		return nil, err
	}

	var pendingMiniBlocksHandler process.PendingMiniBlocksHandler
	pendingMiniBlocksHandler, err = pendingMb.NewNilPendingMiniBlocks()
	if err != nil {
		return nil, err
	}
	if pcf.bootstrapComponents.ShardCoordinator().SelfId() == core.MetachainShardId {
		pendingMiniBlocksHandler, err = pendingMb.NewPendingMiniBlocks()
		if err != nil {
			return nil, err
		}
	}

	forkDetector, err := pcf.newForkDetector(blackListHandler, blockTracker)
	if err != nil {
		return nil, err
	}

	scheduledTxsExecutionHandler, err := pcf.createScheduledTxsExecutionHandler()
	if err != nil {
		return nil, err
	}

	esdtDataStorageArgs := vmcommonBuiltInFunctions.ArgsNewESDTDataStorage{
		Accounts:              pcf.state.AccountsAdapterAPI(),
		GlobalSettingsHandler: disabled.NewDisabledGlobalSettingHandler(),
		Marshalizer:           pcf.coreData.InternalMarshalizer(),
		ShardCoordinator:      pcf.bootstrapComponents.ShardCoordinator(),
		EnableEpochsHandler:   pcf.coreData.EnableEpochsHandler(),
	}
	pcf.esdtNftStorage, err = vmcommonBuiltInFunctions.NewESDTDataStorage(esdtDataStorageArgs)
	if err != nil {
		return nil, err
	}

	processedMiniBlocksTracker := processedMb.NewProcessedMiniBlocksTracker()

	receiptsRepository, err := receipts.NewReceiptsRepository(receipts.ArgsNewReceiptsRepository{
		Store:      pcf.data.StorageService(),
		Marshaller: pcf.coreData.InternalMarshalizer(),
		Hasher:     pcf.coreData.Hasher(),
	})
	if err != nil {
		return nil, err
	}

	blockCutoffProcessingHandler, err := cutoff.CreateBlockProcessingCutoffHandler(pcf.prefConfigs.BlockProcessingCutoff)
	if err != nil {
		return nil, err
	}

	sentSignaturesTracker, err := track.NewSentSignaturesTracker(pcf.crypto.KeysHandler())
	if err != nil {
		return nil, fmt.Errorf("%w when assembling components for the sent signatures tracker", err)
	}

	blockProcessorComponents, err := pcf.newBlockProcessor(
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
		pcf.coreData.WasmVMChangeLocker(),
		scheduledTxsExecutionHandler,
		processedMiniBlocksTracker,
		receiptsRepository,
		blockCutoffProcessingHandler,
		pcf.state.MissingTrieNodesNotifier(),
		sentSignaturesTracker,
	)
	if err != nil {
		return nil, err
	}

	startEpochNum := pcf.bootstrapComponents.EpochBootstrapParams().Epoch()
	if startEpochNum == 0 {
		err = pcf.indexGenesisBlocks(genesisBlocks, initialTxs, genesisAccounts)
		if err != nil {
			return nil, err
		}
	}

	cacheRefreshDuration := time.Duration(pcf.config.ValidatorStatistics.CacheRefreshIntervalInSec) * time.Second
	argVSP := peer.ArgValidatorsProvider{
		NodesCoordinator:                  pcf.nodesCoordinator,
		StartEpoch:                        startEpochNum,
		EpochStartEventNotifier:           pcf.coreData.EpochStartNotifierWithConfirm(),
		CacheRefreshIntervalDurationInSec: cacheRefreshDuration,
		ValidatorStatistics:               validatorStatisticsProcessor,
		MaxRating:                         pcf.maxRating,
		ValidatorPubKeyConverter:          pcf.coreData.ValidatorPubKeyConverter(),
		AddressPubKeyConverter:            pcf.coreData.AddressPubKeyConverter(),
		AuctionListSelector:               pcf.auctionListSelectorAPI,
		StakingDataProvider:               pcf.stakingDataProviderAPI,
	}

	validatorsProvider, err := peer.NewValidatorsProvider(argVSP)
	if err != nil {
		return nil, err
	}

	conversionBase := 10
	genesisNodePrice, ok := big.NewInt(0).SetString(pcf.systemSCConfig.StakingSystemSCConfig.GenesisNodePrice, conversionBase)
	if !ok {
		return nil, errors.New("invalid genesis node price")
	}

	nodesSetupChecker, err := checking.NewNodesSetupCheckerMidas(
		pcf.accountsParser,
		genesisNodePrice,
		pcf.coreData.ValidatorPubKeyConverter(),
		pcf.crypto.BlockSignKeyGen(),
	)
	if err != nil {
		return nil, err
	}

	err = nodesSetupChecker.Check(pcf.coreData.GenesisNodesSetup().AllInitialNodes())
	if err != nil {
		return nil, err
	}

	observerBLSPrivateKey, observerBLSPublicKey := pcf.crypto.BlockSignKeyGen().GeneratePair()
	observerBLSPublicKeyBuff, err := observerBLSPublicKey.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("error generating observerBLSPublicKeyBuff, %w", err)
	} else {
		log.Debug("generated BLS private key for redundancy handler. This key will be used on heartbeat messages "+
			"if the node is in backup mode and the main node is active", "hex public key", observerBLSPublicKeyBuff)
	}

	maxRoundsOfInactivity := int(pcf.prefConfigs.Preferences.RedundancyLevel) * pcf.config.Redundancy.MaxRoundsOfInactivityAccepted
	nodeRedundancyArg := redundancy.ArgNodeRedundancy{
		MaxRoundsOfInactivity: maxRoundsOfInactivity,
		Messenger:             pcf.network.NetworkMessenger(),
		ObserverPrivateKey:    observerBLSPrivateKey,
	}
	nodeRedundancyHandler, err := redundancy.NewNodeRedundancy(nodeRedundancyArg)
	if err != nil {
		return nil, err
	}

	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}
	args := txsSender.ArgsTxsSenderWithAccumulator{
		Marshaller:        pcf.coreData.InternalMarshalizer(),
		ShardCoordinator:  pcf.bootstrapComponents.ShardCoordinator(),
		NetworkMessenger:  pcf.network.NetworkMessenger(),
		AccumulatorConfig: pcf.config.Antiflood.TxAccumulator,
		DataPacker:        dataPacker,
	}
	txsSenderWithAccumulator, err := txsSender.NewTxsSenderWithAccumulator(args)
	if err != nil {
		return nil, err
	}

	apiTransactionEvaluator, vmFactoryForTxSimulate, err := pcf.createAPITransactionEvaluator()
	if err != nil {
		return nil, fmt.Errorf("%w when assembling components for the transactions simulator processor", err)
	}

	return &processComponents{
		nodesCoordinator:                 pcf.nodesCoordinator,
		shardCoordinator:                 pcf.bootstrapComponents.ShardCoordinator(),
		mainInterceptorsContainer:        mainInterceptorsContainer,
		fullArchiveInterceptorsContainer: fullArchiveInterceptorsContainer,
		resolversContainer:               resolversContainer,
		requestersFinder:                 requestersFinder,
		roundHandler:                     pcf.coreData.RoundHandler(),
		forkDetector:                     forkDetector,
		blockProcessor:                   blockProcessorComponents.blockProcessor,
		epochStartTrigger:                epochStartTrigger,
		epochStartNotifier:               pcf.coreData.EpochStartNotifierWithConfirm(),
		blackListHandler:                 blackListHandler,
		bootStorer:                       bootStorer,
		headerSigVerifier:                headerSigVerifier,
		validatorsStatistics:             validatorStatisticsProcessor,
		validatorsProvider:               validatorsProvider,
		blockTracker:                     blockTracker,
		pendingMiniBlocksHandler:         pendingMiniBlocksHandler,
		requestHandler:                   requestHandler,
		txLogsProcessor:                  txLogsProcessor,
		headerConstructionValidator:      headerValidator,
		headerIntegrityVerifier:          pcf.bootstrapComponents.HeaderIntegrityVerifier(),
		mainPeerShardMapper:              mainPeerShardMapper,
		fullArchivePeerShardMapper:       fullArchivePeerShardMapper,
		apiTransactionEvaluator:          apiTransactionEvaluator,
		miniBlocksPoolCleaner:            mbsPoolsCleaner,
		txsPoolCleaner:                   txsPoolsCleaner,
		fallbackHeaderValidator:          fallbackHeaderValidator,
		whiteListHandler:                 pcf.whiteListHandler,
		whiteListerVerifiedTxs:           pcf.whiteListerVerifiedTxs,
		historyRepository:                pcf.historyRepo,
		importStartHandler:               pcf.importStartHandler,
		requestedItemsHandler:            pcf.requestedItemsHandler,
		importHandler:                    pcf.importHandler,
		nodeRedundancyHandler:            nodeRedundancyHandler,
		currentEpochProvider:             currentEpochProvider,
		vmFactoryForTxSimulator:          vmFactoryForTxSimulate,
		vmFactoryForProcessing:           blockProcessorComponents.vmFactoryForProcessing,
		scheduledTxsExecutionHandler:     scheduledTxsExecutionHandler,
		txsSender:                        txsSenderWithAccumulator,
		hardforkTrigger:                  hardforkTrigger,
		processedMiniBlocksTracker:       processedMiniBlocksTracker,
		esdtDataStorageForApi:            pcf.esdtNftStorage,
		accountsParser:                   pcf.accountsParser,
		receiptsRepository:               receiptsRepository,
		sentSignaturesTracker:            sentSignaturesTracker,
	}, nil
}

