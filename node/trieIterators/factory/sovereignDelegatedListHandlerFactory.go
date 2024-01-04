package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

type sovereignDelegatedListHandlerFactory struct {
}

func NewSovereignDelegatedListHandlerFactory() *sovereignDelegatedListHandlerFactory {
	return &sovereignDelegatedListHandlerFactory{}
}

// CreateDelegatedListHandler will create a new instance of DirectStakedListHandler
func (sd *sovereignDelegatedListHandlerFactory) CreateDelegatedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error) {
	return trieIterators.NewDelegatedListProcessor(args)
}

func (sd *sovereignDelegatedListHandlerFactory) IsInterfaceNil() bool {
	return sd == nil
}
