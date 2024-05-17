package runType_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/runType"
	"github.com/stretchr/testify/require"
)


func TestNewSovereignRunTypeComponentsFactoryMidas(t *testing.T) {
	t.Parallel()

	t.Run("nil runType components factory", func(t *testing.T) {
		sovArgs := createSovRunTypeArgs()
		sovArgs.RunTypeComponentsFactory = nil
		srcf, err := runType.NewSovereignRunTypeComponentsFactoryMidas(sovArgs)
		require.Nil(t, srcf)
		require.ErrorIs(t, errors.ErrNilRunTypeComponentsFactory, err)
	})
	t.Run("nil data codec", func(t *testing.T) {
		sovArgs := createSovRunTypeArgs()
		sovArgs.DataCodec = nil
		srcf, err := runType.NewSovereignRunTypeComponentsFactoryMidas(sovArgs)
		require.Nil(t, srcf)
		require.ErrorIs(t, errors.ErrNilDataCodec, err)
	})
	t.Run("nil topics checker", func(t *testing.T) {
		sovArgs := createSovRunTypeArgs()
		sovArgs.TopicsChecker = nil
		srcf, err := runType.NewSovereignRunTypeComponentsFactoryMidas(sovArgs)
		require.Nil(t, srcf)
		require.ErrorIs(t, errors.ErrNilTopicsChecker, err)
	})
	t.Run("should work", func(t *testing.T) {
		srcf, err := runType.NewSovereignRunTypeComponentsFactoryMidas(createSovRunTypeArgs())
		require.NotNil(t, srcf)
		require.NoError(t, err)
	})
}

func TestSovereignRunTypeComponentsFactoryMidas_Create(t *testing.T) {
	t.Parallel()

	srcf, _ := runType.NewSovereignRunTypeComponentsFactoryMidas(createSovRunTypeArgs())

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)
}

func TestSovereignRunTypeComponentsFactoryMidas_Close(t *testing.T) {
	t.Parallel()

	srcf, _ := runType.NewSovereignRunTypeComponentsFactoryMidas(createSovRunTypeArgs())

	rc, err := srcf.Create()
	require.NoError(t, err)
	require.NotNil(t, rc)

	require.NoError(t, rc.Close())
}

