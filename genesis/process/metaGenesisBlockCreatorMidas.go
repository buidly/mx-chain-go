package process

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	disabledCommon "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	disabledGuardian "github.com/multiversx/mx-chain-go/process/guardian/disabled"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/process/smartContract/processProxy"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	syncDisabled "github.com/multiversx/mx-chain-go/process/sync/disabled"
	processTransaction "github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"sync"
)

func createProcessorsForMetaGenesisBlockMidas(arg ArgsGenesisBlockCreator, enableEpochsConfig config.EnableEpochs, roundConfig config.RoundConfig) (*genesisProcessors, error) {
	epochNotifier := forking.NewGenericEpochNotifier()
	temporaryMetaHeader := &block.MetaBlock{
		Epoch:     arg.StartEpochNum,
		TimeStamp: arg.GenesisTime,
	}
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifier)
	if err != nil {
		return nil, err
	}
	epochNotifier.CheckEpoch(temporaryMetaHeader)

	roundNotifier := forking.NewGenericRoundNotifier()
	enableRoundsHandler, err := enablers.NewEnableRoundsHandler(roundConfig, roundNotifier)
	if err != nil {
		return nil, err
	}

	builtInFuncs := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:                 arg.Accounts,
		PubkeyConv:               arg.Core.AddressPubKeyConverter(),
		StorageService:           arg.Data.StorageService(),
		BlockChain:               arg.Data.Blockchain(),
		ShardCoordinator:         arg.ShardCoordinator,
		Marshalizer:              arg.Core.InternalMarshalizer(),
		Uint64Converter:          arg.Core.Uint64ByteSliceConverter(),
		BuiltInFunctions:         builtInFuncs,
		NFTStorageHandler:        &disabled.SimpleNFTStorage{},
		GlobalSettingsHandler:    &disabled.ESDTGlobalSettingsHandler{},
		DataPool:                 arg.Data.Datapool(),
		CompiledSCPool:           arg.Data.Datapool().SmartContracts(),
		EpochNotifier:            epochNotifier,
		EnableEpochsHandler:      enableEpochsHandler,
		NilCompiledSCStore:       true,
		GasSchedule:              arg.GasSchedule,
		Counter:                  counters.NewDisabledCounter(),
		MissingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
	}

	pubKeyVerifier, err := disabled.NewMessageSignVerifier(arg.BlockSignKeyGen)
	if err != nil {
		return nil, err
	}

	blockChainHookImpl, err := arg.RunTypeComponents.BlockChainHookHandlerCreator().CreateBlockChainHookHandler(argsHook)
	if err != nil {
		return nil, err
	}

	argsNewVMContainerFactory := metachain.ArgsNewVMContainerFactory{
		BlockChainHook:          blockChainHookImpl,
		PubkeyConv:              argsHook.PubkeyConv,
		Economics:               arg.Economics,
		MessageSignVerifier:     pubKeyVerifier,
		GasSchedule:             arg.GasSchedule,
		NodesConfigProvider:     arg.InitialNodesSetup,
		Hasher:                  arg.Core.Hasher(),
		Marshalizer:             arg.Core.InternalMarshalizer(),
		SystemSCConfig:          &arg.SystemSCConfig,
		ValidatorAccountsDB:     arg.ValidatorAccounts,
		UserAccountsDB:          arg.Accounts,
		ChanceComputer:          &disabled.Rater{},
		ShardCoordinator:        arg.ShardCoordinator,
		EnableEpochsHandler:     enableEpochsHandler,
		NodesCoordinator:        &disabled.NodesCoordinator{},
		VMContextCreatorHandler: arg.RunTypeComponents.VMContextCreator(),
	}
	virtualMachineFactory, err := metachain.NewVMContainerFactoryMidas(argsNewVMContainerFactory)
	if err != nil {
		return nil, err
	}

	vmContainer, err := virtualMachineFactory.CreateForGenesis()
	if err != nil {
		return nil, err
	}

	err = blockChainHookImpl.SetVMContainer(vmContainer)
	if err != nil {
		return nil, err
	}

	genesisFeeHandler := &disabled.FeeHandler{}
	argsFactory := metachain.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator:        arg.ShardCoordinator,
		Marshalizer:             arg.Core.InternalMarshalizer(),
		Hasher:                  arg.Core.Hasher(),
		PubkeyConverter:         arg.Core.AddressPubKeyConverter(),
		Store:                   arg.Data.StorageService(),
		PoolsHolder:             arg.Data.Datapool(),
		EconomicsFee:            genesisFeeHandler,
		EnableEpochsHandler:     enableEpochsHandler,
		TxExecutionOrderHandler: arg.TxExecutionOrderHandler,
	}
	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(argsFactory)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(block.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	badTxForwarder, err := interimProcContainer.Get(block.InvalidBlock)
	if err != nil {
		return nil, err
	}

	esdtTransferParser, err := parsers.NewESDTTransferParser(arg.Core.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:    arg.ShardCoordinator,
		BuiltInFunctions:    builtInFuncs,
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: enableEpochsHandler,
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(arg.Economics, txTypeHandler, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	argsParser := smartContract.NewArgumentParser()
	argsNewSCProcessor := scrCommon.ArgsNewSmartContractProcessor{
		VmContainer:         vmContainer,
		ArgsParser:          argsParser,
		Hasher:              arg.Core.Hasher(),
		Marshalizer:         arg.Core.InternalMarshalizer(),
		AccountsDB:          arg.Accounts,
		BlockChainHook:      virtualMachineFactory.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncs,
		PubkeyConv:          arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:    arg.ShardCoordinator,
		ScrForwarder:        scForwarder,
		TxFeeHandler:        genesisFeeHandler,
		EconomicsFee:        genesisFeeHandler,
		TxTypeHandler:       txTypeHandler,
		GasHandler:          gasHandler,
		GasSchedule:         arg.GasSchedule,
		TxLogsProcessor:     arg.TxLogsProcessor,
		BadTxForwarder:      badTxForwarder,
		EnableRoundsHandler: enableRoundsHandler,
		EnableEpochsHandler: enableEpochsHandler,
		IsGenesisProcessing: true,
		WasmVMChangeLocker:  &sync.RWMutex{}, // local Locker as to not interfere with the rest of the components
		VMOutputCacher:      txcache.NewDisabledCache(),
		EpochNotifier:       epochNotifier,
	}

	scProcessorProxy, err := processProxy.NewSmartContractProcessorProxy(argsNewSCProcessor)
	if err != nil {
		return nil, err
	}

	argsNewMetaTxProcessor := processTransaction.ArgsNewMetaTxProcessor{
		Hasher:              arg.Core.Hasher(),
		Marshalizer:         arg.Core.InternalMarshalizer(),
		Accounts:            arg.Accounts,
		PubkeyConv:          arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:    arg.ShardCoordinator,
		ScProcessor:         scProcessorProxy,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        genesisFeeHandler,
		EnableEpochsHandler: enableEpochsHandler,
		TxVersionChecker:    disabled.NewDisabledTxVersionChecker(),
		GuardianChecker:     disabledGuardian.NewDisabledGuardedAccountHandler(),
	}
	txProcessor, err := processTransaction.NewMetaTxProcessor(argsNewMetaTxProcessor)
	if err != nil {
		return nil, process.ErrNilTxProcessor
	}

	disabledRequestHandler := &disabled.RequestHandler{}
	disabledBlockTracker := &disabled.BlockTracker{}
	disabledBlockSizeComputationHandler := &disabled.BlockSizeComputationHandler{}
	disabledBalanceComputationHandler := &disabled.BalanceComputationHandler{}
	disabledScheduledTxsExecutionHandler := &disabled.ScheduledTxsExecutionHandler{}
	disabledProcessedMiniBlocksTracker := &disabled.ProcessedMiniBlocksTracker{}

	argsPreProc := metachain.ArgPreProcessorsContainerFactory{
		ShardCoordinator:             arg.ShardCoordinator,
		Store:                        arg.Data.StorageService(),
		Marshaller:                   arg.Core.InternalMarshalizer(),
		Hasher:                       arg.Core.Hasher(),
		DataPool:                     arg.Data.Datapool(),
		Accounts:                     arg.Accounts,
		RequestHandler:               disabledRequestHandler,
		TxProcessor:                  txProcessor,
		ScResultProcessor:            scProcessorProxy,
		EconomicsFee:                 arg.Economics,
		GasHandler:                   gasHandler,
		BlockTracker:                 disabledBlockTracker,
		PubkeyConverter:              arg.Core.AddressPubKeyConverter(),
		BlockSizeComputation:         disabledBlockSizeComputationHandler,
		BalanceComputation:           disabledBalanceComputationHandler,
		EnableEpochsHandler:          enableEpochsHandler,
		TxTypeHandler:                txTypeHandler,
		ScheduledTxsExecutionHandler: disabledScheduledTxsExecutionHandler,
		ProcessedMiniBlocksTracker:   disabledProcessedMiniBlocksTracker,
		TxExecutionOrderHandler:      arg.TxExecutionOrderHandler,
		TxPreProcessorCreator:        arg.RunTypeComponents.TxPreProcessorCreator(),
	}
	preProcFactory, err := metachain.NewPreProcessorsContainerFactory(argsPreProc)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	argsDetector := coordinator.ArgsPrintDoubleTransactionsDetector{
		Marshaller:          arg.Core.InternalMarshalizer(),
		Hasher:              arg.Core.Hasher(),
		EnableEpochsHandler: enableEpochsHandler,
	}
	doubleTransactionsDetector, err := coordinator.NewPrintDoubleTransactionsDetector(argsDetector)
	if err != nil {
		return nil, err
	}

	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:                       arg.Core.Hasher(),
		Marshalizer:                  arg.Core.InternalMarshalizer(),
		ShardCoordinator:             arg.ShardCoordinator,
		Accounts:                     arg.Accounts,
		MiniBlockPool:                arg.Data.Datapool().MiniBlocks(),
		RequestHandler:               disabledRequestHandler,
		PreProcessors:                preProcContainer,
		InterProcessors:              interimProcContainer,
		GasHandler:                   gasHandler,
		FeeHandler:                   genesisFeeHandler,
		BlockSizeComputation:         disabledBlockSizeComputationHandler,
		BalanceComputation:           disabledBalanceComputationHandler,
		EconomicsFee:                 genesisFeeHandler,
		TxTypeHandler:                txTypeHandler,
		TransactionsLogProcessor:     arg.TxLogsProcessor,
		EnableEpochsHandler:          enableEpochsHandler,
		ScheduledTxsExecutionHandler: disabledScheduledTxsExecutionHandler,
		DoubleTransactionsDetector:   doubleTransactionsDetector,
		ProcessedMiniBlocksTracker:   disabledProcessedMiniBlocksTracker,
		TxExecutionOrderHandler:      arg.TxExecutionOrderHandler,
	}
	txCoordinator, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	if err != nil {
		return nil, err
	}

	apiBlockchain, err := blockchain.NewMetaChain(disabledCommon.NewAppStatusHandler())
	if err != nil {
		return nil, err
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             arg.Economics,
		BlockChainHook:           virtualMachineFactory.BlockChainHookImpl(),
		MainBlockChain:           arg.Data.Blockchain(),
		APIBlockChain:            apiBlockchain,
		WasmVMChangeLocker:       &sync.RWMutex{},
		Bootstrapper:             syncDisabled.NewDisabledBootstrapper(),
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
		HistoryRepository:        arg.HistoryRepository,
		ShardCoordinator:         arg.ShardCoordinator,
		StorageService:           arg.Data.StorageService(),
		Marshaller:               arg.Core.InternalMarshalizer(),
		Hasher:                   arg.Core.Hasher(),
		Uint64ByteSliceConverter: arg.Core.Uint64ByteSliceConverter(),
	}
	queryService, err := smartContract.NewSCQueryService(argsNewSCQueryService)
	if err != nil {
		return nil, err
	}

	return &genesisProcessors{
		txCoordinator:  txCoordinator,
		systemSCs:      virtualMachineFactory.SystemSmartContractContainer(),
		blockchainHook: virtualMachineFactory.BlockChainHookImpl(),
		txProcessor:    txProcessor,
		scProcessor:    scProcessorProxy,
		scrProcessor:   scProcessorProxy,
		rwdProcessor:   nil,
		queryService:   queryService,
		vmContainer:    vmContainer,
	}, nil
}
