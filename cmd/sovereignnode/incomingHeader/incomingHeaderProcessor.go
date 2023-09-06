package incomingHeader

import (
	"encoding/hex"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("headerSubscriber")

// ArgsIncomingHeaderProcessor is a struct placeholder for args needed to create a new incoming header processor
type ArgsIncomingHeaderProcessor struct {
	HeadersPool HeadersPool
	TxPool      TransactionPool
	Marshaller  marshal.Marshalizer
	Hasher      hashing.Hasher
}

type incomingHeaderProcessor struct {
	scrProc            *scrProcessor
	extendedHeaderProc *extendedHeaderProcessor
	mapHashes          map[string]struct{}
	mutex              sync.RWMutex
}

// NewIncomingHeaderProcessor creates an incoming header processor which should be able to receive incoming headers and events
// from a chain to local sovereign chain. This handler will validate the events(using proofs in the future) and create
// incoming miniblocks and transaction(which will be added in pool) to be executed in sovereign shard.
func NewIncomingHeaderProcessor(args ArgsIncomingHeaderProcessor) (*incomingHeaderProcessor, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, errNilHeadersPool
	}
	if check.IfNil(args.TxPool) {
		return nil, errNilTxPool
	}
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, core.ErrNilHasher
	}

	scrProc := &scrProcessor{
		txPool:     args.TxPool,
		marshaller: args.Marshaller,
		hasher:     args.Hasher,
	}

	extendedHearProc := &extendedHeaderProcessor{
		headersPool: args.HeadersPool,
		marshaller:  args.Marshaller,
		hasher:      args.Hasher,
	}

	return &incomingHeaderProcessor{
		scrProc:            scrProc,
		extendedHeaderProc: extendedHearProc,
		mapHashes:          make(map[string]struct{}),
	}, nil
}

// AddHeader will receive the incoming header, validate it, create incoming mbs and transactions and add them to pool
func (ihp *incomingHeaderProcessor) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	log.Info("received incoming header", "hash", hex.EncodeToString(headerHash), "nonce", header.GetHeaderHandler().GetNonce())

	ihp.mutex.Lock()
	defer ihp.mutex.Unlock()

	if _, found := ihp.mapHashes[string(headerHash)]; found {
		log.Error("incomingHeaderProcessor.AddHeader already exists")
		return nil
	}

	incomingSCRs, err := ihp.scrProc.createIncomingSCRs(header.GetIncomingEventHandlers())
	if err != nil {
		return err
	}

	extendedHeader, err := createExtendedHeader(header, incomingSCRs)
	if err != nil {
		return err
	}

	err = ihp.extendedHeaderProc.addExtendedHeaderToPool(extendedHeader)
	if err != nil {
		return err
	}

	ihp.scrProc.addSCRsToPool(incomingSCRs)
	ihp.mapHashes[string(headerHash)] = struct{}{}
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihp *incomingHeaderProcessor) IsInterfaceNil() bool {
	return ihp == nil
}
