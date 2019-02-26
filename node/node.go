package node

import (
	"context"
	"fmt"
	"math/big"
	gosync "sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/pkg/errors"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(2000 * time.Millisecond)

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic topicName = "consensus"

// SendTransactionsPipe is the pipe used for sending new transactions
const SendTransactionsPipe = "send transactions pipe"

type topicName string

var log = logger.NewDefaultLogger()

// Option represents a functional configuration parameter that can operate
//  over the None struct.
type Option func(*Node) error

// Node is a structure that passes the configuration parameters and initializes
//  required services as requested
type Node struct {
	marshalizer                  marshal.Marshalizer
	ctx                          context.Context
	hasher                       hashing.Hasher
	initialNodesPubkeys          []string
	initialNodesBalances         map[string]*big.Int
	roundDuration                uint64
	consensusGroupSize           int
	messenger                    p2p.Messenger
	syncer                       ntp.SyncTimer
	blockProcessor               process.BlockProcessor
	genesisTime                  time.Time
	elasticSubrounds             bool
	accounts                     state.AccountsAdapter
	addrConverter                state.AddressConverter
	uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
	interceptorsResolversCreator process.InterceptorsResolversFactory

	privateKey       crypto.PrivateKey
	publicKey        crypto.PublicKey
	singleSignKeyGen crypto.KeyGenerator
	singlesig        crypto.SingleSigner
	multisig         crypto.MultiSigner
	forkDetector     process.ForkDetector

	blkc             *blockchain.BlockChain
	dataPool         data.TransientDataHolder
	shardCoordinator sharding.ShardCoordinator

	isRunning bool
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	if n.IsRunning() {
		return errors.New("cannot apply options while node is running")
	}
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
	}
	return nil
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) (*Node, error) {
	node := &Node{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, errors.New("error applying option: " + err.Error())
		}
	}
	return node, nil
}

// IsRunning will return the current state of the node
func (n *Node) IsRunning() bool {
	return n.isRunning
}

// Start will create a new messenger and and set up the Node state as running
func (n *Node) Start() error {
	err := n.P2PBootstrap()
	if err == nil {
		n.isRunning = true
	}
	return err
}

// Stop closes the messenger and undos everything done in Start
func (n *Node) Stop() error {
	if !n.IsRunning() {
		return nil
	}
	err := n.messenger.Close()
	if err != nil {
		return err
	}

	n.messenger = nil
	return nil
}

// P2PBootstrap will try to connect to many peers as possible
func (n *Node) P2PBootstrap() error {
	if n.messenger == nil {
		return ErrNilMessenger
	}

	return n.messenger.Bootstrap()
}

// CreateShardedStores instantiate sharded cachers for Transactions and Headers
func (n *Node) CreateShardedStores() error {
	if n.shardCoordinator == nil {
		return ErrNilShardCoordinator
	}

	if n.dataPool == nil {
		return ErrNilDataPool
	}

	transactionsDataStore := n.dataPool.Transactions()
	headersDataStore := n.dataPool.Headers()

	if transactionsDataStore == nil {
		return errors.New("nil transaction sharded data store")
	}

	if headersDataStore == nil {
		return errors.New("nil header sharded data store")
	}

	shards := n.shardCoordinator.NoShards()

	for i := uint32(0); i < shards; i++ {
		transactionsDataStore.CreateShardStore(i)
		headersDataStore.CreateShardStore(i)
	}

	return nil
}

// StartConsensus will start the consesus service for the current node
func (n *Node) StartConsensus() error {

	genesisHeader, genesisHeaderHash, err := n.createGenesisBlock()

	if err != nil {
		return err
	}

	n.blkc.GenesisBlock = genesisHeader
	n.blkc.GenesisHeaderHash = genesisHeaderHash

	rounder, err := n.createRounder()

	if err != nil {
		return err
	}

	chronologyHandler, err := n.createChronologyHandler(rounder)

	if err != nil {
		return err
	}

	bootstraper, err := n.createBootstraper(rounder)

	if err != nil {
		return err
	}

	consensusState, err := n.createConsensusState()

	if err != nil {
		return err
	}

	worker, err := bn.NewWorker(
		bootstraper,
		consensusState,
		n.singleSignKeyGen,
		n.marshalizer,
		n.privateKey,
		rounder,
		n.shardCoordinator,
		n.singlesig,
	)
	if err != nil {
		return err
	}

	err = n.createConsensusTopic(worker)
	if err != nil {
		return err
	}

	worker.SendMessage = n.sendMessage
	worker.BroadcastTxBlockBody = n.broadcastBlockBody
	worker.BroadcastHeader = n.broadcastHeader

	validatorGroupSelector, err := n.createValidatorGroupSelector()

	if err != nil {
		return err
	}

	fct, err := bn.NewFactory(
		n.blkc,
		n.blockProcessor,
		bootstraper,
		chronologyHandler,
		consensusState,
		n.hasher,
		n.marshalizer,
		n.multisig,
		rounder,
		n.shardCoordinator,
		n.syncer,
		validatorGroupSelector,
		worker,
	)

	if err != nil {
		return err
	}

	err = fct.GenerateSubrounds()

	if err != nil {
		return err
	}

	go chronologyHandler.StartRounds()

	return nil
}

// GetBalance gets the balance for a specific address
func (n *Node) GetBalance(addressHex string) (*big.Int, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	address, err := n.addrConverter.CreateAddressFromHex(addressHex)
	if err != nil {
		return nil, errors.New("invalid address, could not decode from hex: " + err.Error())
	}
	account, err := n.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param: " + err.Error())
	}

	if account == nil {
		return big.NewInt(0), nil
	}

	return account.BaseAccount().Balance, nil
}

// GenerateAndSendBulkTransactions is a method for generating and propagating a set
// of transactions to be processed. It is mainly used for demo purposes
func (n *Node) GenerateAndSendBulkTransactions(receiverHex string, value *big.Int, noOfTx uint64) error {
	if noOfTx == 0 {
		return errors.New("can not generate and broadcast 0 transactions")
	}

	if n.publicKey == nil {
		return ErrNilPublicKey
	}

	if n.singlesig == nil {
		return ErrNilSingleSig
	}

	senderAddressBytes, err := n.publicKey.ToByteArray()
	if err != nil {
		return err
	}

	if n.addrConverter == nil {
		return ErrNilAddressConverter
	}
	senderAddress, err := n.addrConverter.CreateAddressFromPublicKeyBytes(senderAddressBytes)
	if err != nil {
		return err
	}

	receiverAddress, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return errors.New("could not create receiver address from provided param: " + err.Error())
	}

	if n.accounts == nil {
		return ErrNilAccountsAdapter
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return errors.New("could not fetch sender account from provided param: " + err.Error())
	}
	newNonce := uint64(0)
	if senderAccount != nil {
		newNonce = senderAccount.BaseAccount().Nonce
	}

	wg := gosync.WaitGroup{}
	wg.Add(int(noOfTx))

	mutTransactions := gosync.RWMutex{}
	transactions := make([][]byte, 0)

	mutErrFound := gosync.Mutex{}
	var errFound error

	for nonce := newNonce; nonce < newNonce+noOfTx; nonce++ {
		go func(crtNonce uint64) {
			_, signedTxBuff, err := n.generateAndSignTx(
				crtNonce,
				value,
				receiverAddress.Bytes(),
				senderAddressBytes,
				nil,
			)

			if err != nil {
				mutErrFound.Lock()
				errFound = errors.New(fmt.Sprintf("failure generating transaction %d: %s", crtNonce, err.Error()))
				mutErrFound.Unlock()

				wg.Done()
				return
			}

			mutTransactions.Lock()
			transactions = append(transactions, signedTxBuff)
			mutTransactions.Unlock()
			wg.Done()
		}(nonce)
	}

	wg.Wait()

	if errFound != nil {
		return errFound
	}

	if len(transactions) != int(noOfTx) {
		return errors.New(fmt.Sprintf("generated only %d from required %d transactions", len(transactions), noOfTx))
	}

	for i := 0; i < len(transactions); i++ {
		n.messenger.BroadcastOnPipe(
			SendTransactionsPipe,
			string(factory.TransactionTopic),
			transactions[i],
		)

		if err != nil {
			return errors.New("could not broadcast transaction: " + err.Error())
		}
	}

	return nil
}

// createRounder method creates a round object
func (n *Node) createRounder() (consensus.Rounder, error) {
	rnd, err := round.NewRound(
		n.genesisTime,
		n.syncer.CurrentTime(),
		time.Millisecond*time.Duration(n.roundDuration),
		n.syncer)

	return rnd, err
}

// createChronologyHandler method creates a chronology object
func (n *Node) createChronologyHandler(rounder consensus.Rounder) (consensus.ChronologyHandler, error) {
	chr, err := chronology.NewChronology(
		n.genesisTime,
		rounder,
		n.syncer)

	if err != nil {
		return nil, err
	}

	return chr, nil
}

func (n *Node) createBootstraper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	bootstrap, err := sync.NewBootstrap(
		n.dataPool,
		n.blkc,
		rounder,
		n.blockProcessor,
		WaitTime,
		n.hasher,
		n.marshalizer,
		n.forkDetector,
		n.interceptorsResolversCreator.ResolverContainer(),
	)

	if err != nil {
		return nil, err
	}

	bootstrap.StartSync()

	return bootstrap, nil
}

// createConsensusState method creates a consensusState object
func (n *Node) createConsensusState() (*spos.ConsensusState, error) {
	selfId, err := n.publicKey.ToByteArray()

	if err != nil {
		return nil, err
	}

	roundConsensus := spos.NewRoundConsensus(
		n.initialNodesPubkeys,
		n.consensusGroupSize,
		string(selfId))

	roundConsensus.ResetRoundState()

	roundThreshold := spos.NewRoundThreshold()

	roundStatus := spos.NewRoundStatus()
	roundStatus.ResetRoundStatus()

	consensusState := spos.NewConsensusState(
		roundConsensus,
		roundThreshold,
		roundStatus)

	return consensusState, nil
}

// createValidatorGroupSelector creates a index hashed group selector object
func (n *Node) createValidatorGroupSelector() (consensus.ValidatorGroupSelector, error) {
	validatorGroupSelector, err := groupSelectors.NewIndexHashedGroupSelector(n.consensusGroupSize, n.hasher)

	if err != nil {
		return nil, err
	}

	validatorsList := make([]consensus.Validator, 0)

	for i := 0; i < len(n.initialNodesPubkeys); i++ {
		validator, err := validators.NewValidator(big.NewInt(0), 0, []byte(n.initialNodesPubkeys[i]))

		if err != nil {
			return nil, err
		}

		validatorsList = append(validatorsList, validator)
	}

	err = validatorGroupSelector.LoadEligibleList(validatorsList)

	if err != nil {
		return nil, err
	}

	return validatorGroupSelector, nil
}

// createConsensusTopic creates a consensus topic for node
func (n *Node) createConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	if n.messenger.HasTopicValidator(string(ConsensusTopic)) {
		return ErrValidatorAlreadySet
	}

	if !n.messenger.HasTopic(string(ConsensusTopic)) {
		err := n.messenger.CreateTopic(string(ConsensusTopic), true)
		if err != nil {
			return err
		}
	}

	return n.messenger.RegisterMessageProcessor(string(ConsensusTopic), messageProcessor)
}

func (n *Node) generateAndSignTx(
	nonce uint64,
	value *big.Int,
	rcvAddrBytes []byte,
	sndAddrBytes []byte,
	dataBytes []byte,
) (*transaction.Transaction, []byte, error) {

	tx := transaction.Transaction{
		Nonce:   nonce,
		Value:   value,
		RcvAddr: rcvAddrBytes,
		SndAddr: sndAddrBytes,
		Data:    dataBytes,
	}

	if n.marshalizer == nil {
		return nil, nil, ErrNilMarshalizer
	}

	if n.privateKey == nil {
		return nil, nil, ErrNilPrivateKey
	}

	marshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, nil, errors.New("could not marshal transaction")
	}

	sig, err := n.singlesig.Sign(n.privateKey, marshalizedTx)
	if err != nil {
		return nil, nil, errors.New("could not sign the transaction")
	}
	tx.Signature = sig

	signedMarshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, nil, errors.New("could not marshal signed transaction")
	}

	return &tx, signedMarshalizedTx, nil
}

//GenerateTransaction generates a new transaction with sender, receiver, amount and code
func (n *Node) GenerateTransaction(senderHex string, receiverHex string, value *big.Int, transactionData string) (*transaction.Transaction, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	if n.privateKey == nil {
		return nil, errors.New("initialize PrivateKey first")
	}

	receiverAddress, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return nil, errors.New("could not create receiver address from provided param")
	}
	senderAddress, err := n.addrConverter.CreateAddressFromHex(senderHex)
	if err != nil {
		return nil, errors.New("could not create sender address from provided param")
	}
	senderAccount, err := n.accounts.GetExistingAccount(senderAddress)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}
	newNonce := uint64(0)
	if senderAccount != nil {
		newNonce = senderAccount.BaseAccount().Nonce
	}

	tx, _, err := n.generateAndSignTx(
		newNonce,
		value,
		receiverAddress.Bytes(),
		senderAddress.Bytes(),
		[]byte(transactionData))

	return tx, err
}

// SendTransaction will send a new transaction on the topic channel
func (n *Node) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value *big.Int,
	transactionData string,
	signature []byte) (*transaction.Transaction, error) {

	sender, err := n.addrConverter.CreateAddressFromHex(senderHex)
	if err != nil {
		return nil, err
	}
	receiver, err := n.addrConverter.CreateAddressFromHex(receiverHex)
	if err != nil {
		return nil, err
	}

	tx := transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		RcvAddr:   receiver.Bytes(),
		SndAddr:   sender.Bytes(),
		Data:      []byte(transactionData),
		Signature: signature,
	}

	marshalizedTx, err := n.marshalizer.Marshal(&tx)
	if err != nil {
		return nil, errors.New("could not marshal transaction")
	}

	n.messenger.BroadcastOnPipe(
		SendTransactionsPipe,
		string(factory.TransactionTopic),
		marshalizedTx,
	)

	return &tx, nil
}

//GetTransaction gets the transaction
func (n *Node) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// GetCurrentPublicKey will return the current node's public key
func (n *Node) GetCurrentPublicKey() string {
	if n.publicKey != nil {
		pkey, _ := n.publicKey.ToByteArray()
		return fmt.Sprintf("%x", pkey)
	}
	return ""
}

// GetAccount will return acount details for a given address
func (n *Node) GetAccount(address string) (*state.Account, error) {
	if n.addrConverter == nil || n.accounts == nil {
		return nil, errors.New("initialize AccountsAdapter and AddressConverter first")
	}

	addr, err := n.addrConverter.CreateAddressFromHex(address)
	if err != nil {
		return nil, errors.New("could not create address object from provided string")
	}
	account, err := n.accounts.GetExistingAccount(addr)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param")
	}
	return account.BaseAccount(), nil
}

func (n *Node) createGenesisBlock() (*block.Header, []byte, error) {
	blockBody, err := n.blockProcessor.CreateGenesisBlockBody(n.initialNodesBalances, 0)
	if err != nil {
		return nil, nil, err
	}

	marshalizedBody, err := n.marshalizer.Marshal(blockBody)
	if err != nil {
		return nil, nil, err
	}
	blockBodyHash := n.hasher.Compute(string(marshalizedBody))
	header := &block.Header{
		Nonce:         0,
		ShardId:       blockBody.ShardID,
		TimeStamp:     uint64(n.genesisTime.Unix()),
		BlockBodyHash: blockBodyHash,
		BlockBodyType: block.StateBlock,
		Signature:     blockBodyHash,
	}

	marshalizedHeader, err := n.marshalizer.Marshal(header)

	if err != nil {
		return nil, nil, err
	}

	blockHeaderHash := n.hasher.Compute(string(marshalizedHeader))

	return header, blockHeaderHash, nil
}

func (n *Node) sendMessage(cnsDta *spos.ConsensusMessage) {
	cnsDtaBuff, err := n.marshalizer.Marshal(cnsDta)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	n.messenger.Broadcast(
		string(ConsensusTopic),
		cnsDtaBuff)
}

func (n *Node) broadcastBlockBody(msg []byte) {
	n.messenger.Broadcast(
		string(factory.TxBlockBodyTopic),
		msg)
}

func (n *Node) broadcastHeader(msg []byte) {
	n.messenger.Broadcast(
		string(factory.HeadersTopic),
		msg,
	)
}