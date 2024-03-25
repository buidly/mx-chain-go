package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
)

// sovereignEpochStartBootstrapperFactory defines the epoch start bootstrapper factory for chain run sovereign
type sovereignEpochStartBootstrapperFactory struct {
	epochStartBootstrapperFactory EpochStartBootstrapperCreator
}

// NewSovereignEpochStartBootstrapperFactory creates a new epoch start bootstrapper factory for chain run sovereign
func NewSovereignEpochStartBootstrapperFactory(esbf EpochStartBootstrapperCreator) (EpochStartBootstrapperCreator, error) {
	if check.IfNil(esbf) {
		return nil, errors.ErrNilEpochStartBootstrapperFactory
	}
	return &sovereignEpochStartBootstrapperFactory{
		epochStartBootstrapperFactory: esbf,
	}, nil
}

// CreateEpochStartBootstrapper creates a new epoch start bootstrapper for chain run sovereign
func (bcf *sovereignEpochStartBootstrapperFactory) CreateEpochStartBootstrapper(epochStartBootstrapArgs ArgsEpochStartBootstrap) (EpochStartBootstrapper, error) {
	epochStartBootstrapper, err := NewEpochStartBootstrap(epochStartBootstrapArgs)
	if err != nil {
		return nil, err
	}

	sesb, err := NewSovereignChainEpochStartBootstrap(epochStartBootstrapper)
	if err != nil {
		return nil, err
	}

	return sesb, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bcf *sovereignEpochStartBootstrapperFactory) IsInterfaceNil() bool {
	return bcf == nil
}