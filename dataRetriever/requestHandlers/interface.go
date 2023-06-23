package requestHandlers

import (
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// HashSliceRequester can request multiple hashes at once
type HashSliceRequester interface {
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}

// ChunkRequester can request a chunk of a large data
type ChunkRequester interface {
	RequestDataFromReferenceAndChunk(reference []byte, chunkIndex uint32) error
}

// NonceRequester can request data for a specific nonce
type NonceRequester interface {
	RequestDataFromNonce(nonce uint64, epoch uint32) error
}

// EpochRequester can request data for a specific epoch
type EpochRequester interface {
	RequestDataFromEpoch(identifier []byte) error
}

// HeaderRequester defines what a block header requester can do
type HeaderRequester interface {
	NonceRequester
	EpochRequester
}

// ResolverRequestFactoryHandler defines the resolver requester factory handler
type ResolverRequestFactoryHandler interface {
	CreateResolverRequestHandler(resolverRequestArgs ResolverRequestArgs) (process.RequestHandler, error)
	IsInterfaceNil() bool
}

// ResolverRequestArgs holds all dependencies required by the process data factory to create components
type ResolverRequestArgs struct {
	RequestersFinder      dataRetriever.RequestersFinder
	RequestedItemsHandler dataRetriever.RequestedItemsHandler
	WhiteListHandler      process.WhiteListHandler
	MaxTxsToRequest       int
	ShardID               uint32
	RequestInterval       time.Duration
}
