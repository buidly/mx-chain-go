package factory

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/watchdog"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	stateFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/hashing"
	hasherFactory "github.com/ElrondNetwork/elrond-go/hashing/factory"
	"github.com/ElrondNetwork/elrond-go/marshal"
	marshalizerFactory "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
)

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config              config.Config
	RatingsConfig       config.RatingsConfig
	EconomicsConfig     config.EconomicsConfig
	NodesFilename       string
	WorkingDirectory    string
	ChanStopNodeProcess chan endProcess.ArgEndProcess
}

// coreComponentsFactory is responsible for creating the core components
type coreComponentsFactory struct {
	config              config.Config
	ratingsConfig       config.RatingsConfig
	economicsConfig     config.EconomicsConfig
	nodesFilename       string
	workingDir          string
	chanStopNodeProcess chan endProcess.ArgEndProcess
}

// coreComponents is the DTO used for core components
type coreComponents struct {
	hasher                   hashing.Hasher
	internalMarshalizer      marshal.Marshalizer
	vmMarshalizer            marshal.Marshalizer
	txSignMarshalizer        marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	addressPubKeyConverter   core.PubkeyConverter
	validatorPubKeyConverter core.PubkeyConverter
	statusHandler            core.AppStatusHandler
	pathHandler              storage.PathManagerHandler
	syncTimer                ntp.SyncTimer
	rounder                  consensus.Rounder
	alarmScheduler           core.TimersScheduler
	watchdog                 core.WatchdogTimer
	nodesSetupHandler        NodesSetupHandler
	economicsData            process.EconomicsHandler
	ratingsData              process.RatingsInfoHandler
	rater                    sharding.PeerAccountListAndRatingHandler
	genesisTime              time.Time
	chainID                  string
	minTransactionVersion    uint32
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) (*coreComponentsFactory, error) {
	return &coreComponentsFactory{
		config:              args.Config,
		ratingsConfig:       args.RatingsConfig,
		economicsConfig:     args.EconomicsConfig,
		workingDir:          args.WorkingDirectory,
		chanStopNodeProcess: args.ChanStopNodeProcess,
		nodesFilename:       args.NodesFilename,
	}, nil
}

// Create creates the core components
func (ccf *coreComponentsFactory) Create() (*coreComponents, error) {
	hasher, err := hasherFactory.NewHasher(ccf.config.Hasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrHasherCreation, err.Error())
	}

	internalMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (internal): %s", errors.ErrMarshalizerCreation, err.Error())
	}

	vmMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (vm): %s", errors.ErrMarshalizerCreation, err.Error())
	}

	txSignMarshalizer, err := marshalizerFactory.NewMarshalizer(ccf.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (tx sign): %s", errors.ErrMarshalizerCreation, err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	addressPubkeyConverter, err := stateFactory.NewPubkeyConverter(ccf.config.AddressPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}
	validatorPubkeyConverter, err := stateFactory.NewPubkeyConverter(ccf.config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}

	pruningStorerPathTemplate, staticStorerPathTemplate := ccf.createStorerTemplatePaths()
	pathHandler, err := pathmanager.NewPathManager(pruningStorerPathTemplate, staticStorerPathTemplate)
	if err != nil {
		return nil, err
	}

	syncer := ntp.NewSyncTime(ccf.config.NTPConfig, nil)
	syncer.StartSyncingTime()
	log.Debug("NTP average clock offset", "value", syncer.ClockOffset())

	genesisNodesConfig, err := sharding.NewNodesSetup(
		ccf.nodesFilename,
		addressPubkeyConverter,
		validatorPubkeyConverter,
	)
	if err != nil {
		return nil, err
	}

	startRound := int64(0)
	if ccf.config.Hardfork.AfterHardFork {
		log.Debug("changed genesis time after hardfork",
			"old genesis time", genesisNodesConfig.StartTime,
			"new genesis time", ccf.config.Hardfork.GenesisTime)
		genesisNodesConfig.StartTime = ccf.config.Hardfork.GenesisTime
		startRound = int64(ccf.config.Hardfork.StartRound)
	}

	if ccf.config.GeneralSettings.StartInEpochEnabled {
		delayedStartInterval := 2 * time.Second
		time.Sleep(delayedStartInterval)
	}

	if genesisNodesConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)
		ntpTime := syncer.CurrentTime()
		genesisNodesConfig.StartTime = (ntpTime.Unix()/60 + 1) * 60
	}

	startTime := time.Unix(genesisNodesConfig.StartTime, 0)

	log.Info("start time",
		"formatted", startTime.Format("Mon Jan 2 15:04:05 MST 2006"),
		"seconds", startTime.Unix())

	log.Debug("config", "file", ccf.nodesFilename)

	genesisTime := time.Unix(genesisNodesConfig.StartTime, 0)
	rounder, err := round.NewRound(
		genesisTime,
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(genesisNodesConfig.RoundDuration),
		syncer,
		startRound,
	)
	if err != nil {
		return nil, err
	}

	alarmScheduler := alarm.NewAlarmScheduler()
	watchdogTimer, err := watchdog.NewWatchdog(alarmScheduler, ccf.chanStopNodeProcess)
	if err != nil {
		return nil, err
	}

	log.Trace("creating economics data components")
	economicsData, err := economics.NewEconomicsData(&ccf.economicsConfig)
	if err != nil {
		return nil, err
	}

	log.Trace("creating ratings data")
	ratingDataArgs := rating.RatingsDataArg{
		Config:                   ccf.ratingsConfig,
		ShardConsensusSize:       genesisNodesConfig.ConsensusGroupSize,
		MetaConsensusSize:        genesisNodesConfig.MetaChainConsensusGroupSize,
		ShardMinNodes:            genesisNodesConfig.MinNodesPerShard,
		MetaMinNodes:             genesisNodesConfig.MetaChainMinNodes,
		RoundDurationMiliseconds: genesisNodesConfig.RoundDuration,
	}
	ratingsData, err := rating.NewRatingsData(ratingDataArgs)
	if err != nil {
		return nil, err
	}

	rater, err := rating.NewBlockSigningRater(ratingsData)
	if err != nil {
		return nil, err
	}

	return &coreComponents{
		hasher:                   hasher,
		internalMarshalizer:      internalMarshalizer,
		vmMarshalizer:            vmMarshalizer,
		txSignMarshalizer:        txSignMarshalizer,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		addressPubKeyConverter:   addressPubkeyConverter,
		validatorPubKeyConverter: validatorPubkeyConverter,
		statusHandler:            statusHandler.NewNilStatusHandler(),
		pathHandler:              pathHandler,
		syncTimer:                syncer,
		rounder:                  rounder,
		alarmScheduler:           alarmScheduler,
		watchdog:                 watchdogTimer,
		nodesSetupHandler:        genesisNodesConfig,
		economicsData:            economicsData,
		ratingsData:              ratingsData,
		rater:                    rater,
		genesisTime:              genesisTime,
		chainID:                  ccf.config.GeneralSettings.ChainID,
		minTransactionVersion:    ccf.config.GeneralSettings.MinTransactionVersion,
	}, nil
}

func (ccf *coreComponentsFactory) createStorerTemplatePaths() (string, string) {
	pathTemplateForPruningStorer := filepath.Join(
		ccf.workingDir,
		core.DefaultDBPath,
		ccf.config.GeneralSettings.ChainID,
		fmt.Sprintf("%s_%s", core.DefaultEpochString, core.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		ccf.workingDir,
		core.DefaultDBPath,
		ccf.config.GeneralSettings.ChainID,
		core.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	return pathTemplateForPruningStorer, pathTemplateForStaticStorer
}

// Close closes all underlying components
func (cc *coreComponents) Close() error {
	if cc.statusHandler != nil {
		cc.statusHandler.Close()
	}

	return nil
}
