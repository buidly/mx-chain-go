package factory

import (
	"path"
	"path/filepath"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type trieCreator struct {
	snapshotDbCfg            config.DBConfig
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	pathManager              storage.PathManagerHandler
	trieStorageManagerConfig config.TrieStorageManagerConfig
}

var log = logger.GetOrCreate("trie")

// NewTrieFactory creates a new trie factory
func NewTrieFactory(
	args TrieFactoryArgs,
) (*trieCreator, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, trie.ErrNilHasher
	}
	if check.IfNil(args.PathManager) {
		return nil, trie.ErrNilPathManager
	}

	return &trieCreator{
		snapshotDbCfg:            args.SnapshotDbCfg,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		pathManager:              args.PathManager,
		trieStorageManagerConfig: args.TrieStorageManagerConfig,
	}, nil
}

// Create creates a new trie
func (tc *trieCreator) Create(
	trieStorageCfg config.StorageConfig,
	shardID string,
	pruningEnabled bool,
	checkpointsEnabled bool,
	maxTrieLevelInMem uint,
) (data.StorageManager, data.Trie, error) {
	trieStoragePath, mainDb := path.Split(tc.pathManager.PathForStatic(shardID, trieStorageCfg.DB.FilePath))

	dbConfig := factory.GetDBFromConfig(trieStorageCfg.DB)
	dbConfig.FilePath = path.Join(trieStoragePath, mainDb)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(trieStorageCfg.Cache),
		dbConfig,
		factory.GetBloomFromConfig(trieStorageCfg.Bloom),
	)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("trie pruning status", "enabled", pruningEnabled)
	if !pruningEnabled {
		return tc.newTrieAndTrieStorageWithoutPruning(accountsTrieStorage, maxTrieLevelInMem)
	}

	snapshotDbCfg := config.DBConfig{
		FilePath:          filepath.Join(trieStoragePath, tc.snapshotDbCfg.FilePath),
		Type:              tc.snapshotDbCfg.Type,
		BatchDelaySeconds: tc.snapshotDbCfg.BatchDelaySeconds,
		MaxBatchSize:      tc.snapshotDbCfg.MaxBatchSize,
		MaxOpenFiles:      tc.snapshotDbCfg.MaxOpenFiles,
	}

	checkpointHashesHolder := hashesHolder.NewCheckpointHashesHolder(
		tc.trieStorageManagerConfig.CheckpointHashesHolderMaxSize,
		uint64(tc.hasher.Size()),
	)
	args := trie.NewTrieStorageManagerArgs{
		DB:                     accountsTrieStorage,
		Marshalizer:            tc.marshalizer,
		Hasher:                 tc.hasher,
		SnapshotDbConfig:       snapshotDbCfg,
		GeneralConfig:          tc.trieStorageManagerConfig,
		CheckpointHashesHolder: checkpointHashesHolder,
	}

	log.Debug("trie checkpoints status", "enabled", checkpointsEnabled)
	if !checkpointsEnabled {
		return tc.newTrieAndTrieStorageWithoutCheckpoints(args, maxTrieLevelInMem)
	}

	return tc.newTrieAndTrieStorage(args, maxTrieLevelInMem)
}

func (tc *trieCreator) newTrieAndTrieStorage(
	args trie.NewTrieStorageManagerArgs,
	maxTrieLevelInMem uint,
) (data.StorageManager, data.Trie, error) {
	trieStorage, err := trie.NewTrieStorageManager(args)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

func (tc *trieCreator) newTrieAndTrieStorageWithoutCheckpoints(
	args trie.NewTrieStorageManagerArgs,
	maxTrieLevelInMem uint,
) (data.StorageManager, data.Trie, error) {
	trieStorage, err := trie.NewTrieStorageManagerWithoutCheckpoints(args)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

func (tc *trieCreator) newTrieAndTrieStorageWithoutPruning(
	accountsTrieStorage data.DBWriteCacher,
	maxTrieLevelInMem uint,
) (data.StorageManager, data.Trie, error) {
	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *trieCreator) IsInterfaceNil() bool {
	return tc == nil
}
