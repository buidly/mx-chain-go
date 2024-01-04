package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
	"github.com/multiversx/mx-chain-go/node/trieIterators/disabled"
)

type directStakedListHandlerFactory struct {
}

func NewDirectStakedListHandlerFactory() *directStakedListHandlerFactory {
	return &directStakedListHandlerFactory{}
}

// CreateDirectStakedListHandler will create a new instance of DirectStakedListHandler
func (ds *directStakedListHandlerFactory) CreateDirectStakedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error) {
	//TODO add unit tests
	if args.ShardID != core.MetachainShardId {
		return disabled.NewDisabledDirectStakedListProcessor(), nil
	}

	return trieIterators.NewDirectStakedListProcessor(args)
}

func (ds *directStakedListHandlerFactory) IsInterfaceNil() bool {
	return ds == nil
}
