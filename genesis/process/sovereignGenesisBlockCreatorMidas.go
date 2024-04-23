package process

import (
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/factory"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"math"
	"math/big"
)

type sovereignGenesisBlockCreatorMidas struct {
	sovereignGenesisBlockCreator
}

// NewSovereignGenesisBlockCreator creates a new sovereign genesis block creator instance
func NewSovereignGenesisBlockCreatorMidas(gbc *genesisBlockCreator) (*sovereignGenesisBlockCreatorMidas, error) {
	if gbc == nil {
		return nil, errNilGenesisBlockCreator
	}

	log.Debug("NewSovereignGenesisBlockCreatorMidas", "native esdt token", gbc.arg.Config.SovereignConfig.GenesisConfig.NativeESDT)

	return &sovereignGenesisBlockCreatorMidas{
		sovereignGenesisBlockCreator{
			genesisBlockCreator: gbc,
		},
	}, nil
}

func (gbc *sovereignGenesisBlockCreatorMidas) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	err := gbc.initGenesisAccounts()
	if err != nil {
		return nil, err
	}

	if !mustDoGenesisProcess(gbc.arg) {
		return gbc.createSovereignEmptyGenesisBlocks()
	}

	err = gbc.computeSovereignDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
	if err != nil {
		return nil, err
	}

	shardIDs := make([]uint32, 1)
	shardIDs[0] = core.SovereignChainShardId
	argsCreateBlock, err := gbc.createGenesisBlocksArgs(shardIDs)
	if err != nil {
		return nil, err
	}

	return gbc.createSovereignHeaders(argsCreateBlock)
}

func (gbc *sovereignGenesisBlockCreatorMidas) createSovereignHeaders(args *headerCreatorArgs) (map[uint32]data.HeaderHandler, error) {
	shardID := core.SovereignChainShardId
	log.Debug("sovereignGenesisBlockCreator.createSovereignHeaders", "shard", shardID)

	var genesisBlock data.HeaderHandler
	var scResults [][]byte
	var err error

	genesisBlock, scResults, gbc.initialIndexingData[shardID], err = createSovereignShardGenesisBlockMidas(
		args.mapArgsGenesisBlockCreator[shardID],
		args.nodesListSplitter,
	)

	if err != nil {
		return nil, fmt.Errorf("'%w' while generating genesis block for shard %d", err, shardID)
	}

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	allScAddresses := make([][]byte, 0)
	allScAddresses = append(allScAddresses, scResults...)
	genesisBlocks[shardID] = genesisBlock
	err = gbc.saveGenesisBlock(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("'%w' while saving genesis block for shard %d", err, shardID)
	}

	err = gbc.checkDelegationsAgainstDeployedSC(allScAddresses, gbc.arg)
	if err != nil {
		return nil, err
	}

	gb := genesisBlocks[shardID]
	log.Info("sovereignGenesisBlockCreator.createSovereignHeaders",
		"shard", gb.GetShardID(),
		"nonce", gb.GetNonce(),
		"round", gb.GetRound(),
		"root hash", gb.GetRootHash(),
	)

	return genesisBlocks, nil
}

func createSovereignShardGenesisBlockMidas(
	arg ArgsGenesisBlockCreator,
	nodesListSplitter genesis.NodesListSplitter,
) (data.HeaderHandler, [][]byte, *genesis.IndexingData, error) {
	sovereignGenesisConfig := createSovereignGenesisConfig(arg.EpochConfig.EnableEpochs)
	shardProcessors, err := createProcessorsForShardGenesisBlock(arg, sovereignGenesisConfig, createGenesisRoundConfig(arg.RoundConfig))
	if err != nil {
		return nil, nil, nil, err
	}

	genesisBlock, scAddresses, indexingData, err := baseCreateShardGenesisBlock(arg, nodesListSplitter, shardProcessors)
	if err != nil {
		return nil, nil, nil, err
	}

	metaProcessor, err := createProcessorsForMetaGenesisBlock(arg, sovereignGenesisConfig, createGenesisRoundConfig(arg.RoundConfig))
	if err != nil {
		return nil, nil, nil, err
	}

	deploySystemSCTxs, err := deploySystemSmartContracts(arg, metaProcessor.txProcessor, metaProcessor.systemSCs)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.DeploySystemScTxs = deploySystemSCTxs

	stakingTxs, err := setSovereignStakedDataMidas(arg, metaProcessor, nodesListSplitter)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.StakingTxs = stakingTxs

	metaScrsTxs := metaProcessor.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	indexingData.ScrsTxs = mergeScrs(indexingData.ScrsTxs, metaScrsTxs)

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating sovereign genesis block while commiting", err)
	}

	err = setRootHash(genesisBlock, rootHash)
	if err != nil {
		return nil, nil, nil, err
	}

	validatorRootHash, err := arg.ValidatorAccounts.RootHash()
	if err != nil {
		return nil, nil, nil, err
	}

	err = genesisBlock.SetValidatorStatsRootHash(validatorRootHash)
	if err != nil {
		return nil, nil, nil, err
	}

	err = metaProcessor.vmContainer.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	return genesisBlock, scAddresses, indexingData, nil
}

func setSovereignStakedDataMidas(
	arg ArgsGenesisBlockCreator,
	processors *genesisProcessors,
	nodesListSplitter genesis.NodesListSplitter,
) ([]data.TransactionHandler, error) {
	scQueryBlsKeys := &process.SCQuery{
		ScAddress: vm.StakingSCAddress,
		FuncName:  "isStaked",
	}

	stakingTxs := make([]data.TransactionHandler, 0)

	// create staking smart contract state for genesis - update fixed stake value from all
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	stakeValuePower := arg.GenesisNodePrice

	stakedNodes := nodesListSplitter.GetAllNodes()
	for _, nodeInfo := range stakedNodes {
		// TODO: Should we get the AbstractStakingSCAddress address from other part? Is this even possible to do?
		senderAcc, err := arg.Accounts.LoadAccount(factory.AbstractStakingSCAddress)
		if err != nil {
			return nil, err
		}

		// This was modified to support the new `stake` implementation for Midas
		tx := &transaction.Transaction{
			Nonce:    senderAcc.GetNonce(),
			Value:    big.NewInt(0),
			RcvAddr:  vm.ValidatorSCAddress,
			SndAddr:  factory.AbstractStakingSCAddress,
			GasPrice: 0,
			GasLimit: math.MaxUint64,
			Data: []byte(
				"stake@" + oneEncoded + "@" + hex.EncodeToString(nodeInfo.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis")) + "@" + hex.EncodeToString(nodeInfo.AddressBytes()) + "@" + hex.EncodeToString(stakeValuePower.Bytes()),
			),
			Signature: nil,
		}

		stakingTxs = append(stakingTxs, tx)

		_, err = processors.txProcessor.ProcessTransaction(tx)
		if err != nil {
			return nil, err
		}

		scQueryBlsKeys.Arguments = [][]byte{nodeInfo.PubKeyBytes()}
		vmOutput, _, err := processors.queryService.ExecuteQuery(scQueryBlsKeys)
		if err != nil {
			return nil, err
		}

		if vmOutput.ReturnCode != vmcommon.Ok {
			return nil, genesis.ErrBLSKeyNotStaked
		}
	}

	log.Debug("sovereign genesis block",
		"num nodes staked", len(stakedNodes),
	)

	return stakingTxs, nil
}
