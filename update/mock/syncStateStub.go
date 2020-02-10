package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type SyncStateStub struct {
	GetEpochStartMetaBlockCalled  func() (*block.MetaBlock, error)
	GetUnfinishedMetaBlocksCalled func() (map[string]*block.MetaBlock, error)
	SyncAllStateCalled            func(epoch uint32) error
	GetAllTriesCalled             func() (map[string]data.Trie, error)
	GetAllTransactionsCalled      func() (map[string]data.TransactionHandler, error)
	GetAllMiniBlocksCalled        func() (map[string]*block.MiniBlock, error)
}

func (sss *SyncStateStub) GetEpochStartMetaBlock() (*block.MetaBlock, error) {
	if sss.GetEpochStartMetaBlockCalled != nil {
		return sss.GetEpochStartMetaBlockCalled()
	}
	return nil, nil
}
func (sss *SyncStateStub) GetUnfinishedMetaBlocks() (map[string]*block.MetaBlock, error) {
	if sss.GetUnfinishedMetaBlocksCalled != nil {
		return sss.GetUnfinishedMetaBlocksCalled()
	}
	return nil, nil
}
func (sss *SyncStateStub) SyncAllState(epoch uint32) error {
	if sss.SyncAllStateCalled != nil {
		return sss.SyncAllStateCalled(epoch)
	}
	return nil
}
func (sss *SyncStateStub) GetAllTries() (map[string]data.Trie, error) {
	if sss.GetAllTriesCalled != nil {
		return sss.GetAllTriesCalled()
	}
	return nil, nil
}
func (sss *SyncStateStub) GetAllTransactions() (map[string]data.TransactionHandler, error) {
	if sss.GetAllTransactionsCalled != nil {
		return sss.GetAllTransactionsCalled()
	}
	return nil, nil
}
func (sss *SyncStateStub) GetAllMiniBlocks() (map[string]*block.MiniBlock, error) {
	if sss.GetAllMiniBlocksCalled != nil {
		return sss.GetAllMiniBlocksCalled()
	}
	return nil, nil
}
func (sss *SyncStateStub) IsInterfaceNil() bool {
	return sss == nil
}
