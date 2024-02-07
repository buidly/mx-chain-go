package process

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/update"
)

// GenesisBlockCreatorHandler defines genesis block creator behavior
type GenesisBlockCreatorHandler interface {
	ImportHandler() update.ImportHandler
	CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error)
	GetIndexingData() map[uint32]*genesis.IndexingData
}

// GenesisBlockCreatorFactory defines a genesis block creator factory behavior
type GenesisBlockCreatorFactory interface {
	CreateGenesisBlockCreator(args ArgsGenesisBlockCreator) (GenesisBlockCreatorHandler, error)
	IsInterfaceNil() bool
}