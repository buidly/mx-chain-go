package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignRunTypeComponentsFactoryMidas(t *testing.T) {
	t.Parallel()

	srcf, err := runType.NewSovereignRunTypeComponentsFactoryMidas(nil, createSovConfig())
	require.Nil(t, srcf)
	require.ErrorIs(t, errors.ErrNilRunTypeComponentsFactory, err)

	rcf, _ := runType.NewRunTypeComponentsFactory(createCoreComponents())
	srcf, err = runType.NewSovereignRunTypeComponentsFactoryMidas(rcf, createSovConfig())
	require.NotNil(t, srcf)
	require.NoError(t, err)
}

func TestSovereignRunTypeComponentsFactoryMidas_Create(t *testing.T) {
	t.Parallel()

	rcf, _ := runType.NewRunTypeComponentsFactory(createCoreComponents())
	srcf, _ := runType.NewSovereignRunTypeComponentsFactoryMidas(rcf, createSovConfig())

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)
}

func TestSovereignRunTypeComponentsFactoryMidas_Close(t *testing.T) {
	t.Parallel()

	rcf, _ := runType.NewRunTypeComponentsFactory(createCoreComponents())
	srcf, _ := runType.NewSovereignRunTypeComponentsFactoryMidas(rcf, createSovConfig())

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}
