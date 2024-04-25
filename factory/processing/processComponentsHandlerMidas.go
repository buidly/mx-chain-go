package processing

import (
	"github.com/multiversx/mx-chain-go/errors"
)

type managedProcessComponentsMidas struct {
	managedProcessComponents
	factory              *processComponentsFactoryMidas
}

// NewManagedProcessComponents returns a news instance of managedProcessComponents
func NewManagedProcessComponentsMidas(pcf *processComponentsFactoryMidas) (*managedProcessComponentsMidas, error) {
	if pcf == nil {
		return nil, errors.ErrNilProcessComponentsFactory
	}

	return &managedProcessComponentsMidas{
		managedProcessComponents: managedProcessComponents{
			processComponents: nil,
			factory:           &pcf.processComponentsFactory, // not actually used, since it was only used in Create
		},
		factory: pcf,
	}, nil
}

// Create will create the managed components
func (mpc *managedProcessComponentsMidas) Create() error {
	pc, err := mpc.factory.Create()
	if err != nil {
		return err
	}

	mpc.mutProcessComponents.Lock()
	mpc.processComponents = pc
	mpc.mutProcessComponents.Unlock()

	return nil
}

// IsInterfaceNil returns true if the interface is nil
func (mpc *managedProcessComponentsMidas) IsInterfaceNil() bool {
	return mpc == nil
}
