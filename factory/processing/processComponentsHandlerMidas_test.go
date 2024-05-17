package processing_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/process/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManagedProcessComponentsMidas(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedProcessComponents, err := processComp.NewManagedProcessComponentsMidas(nil)
		require.Equal(t, errorsMx.ErrNilProcessComponentsFactory, err)
		require.Nil(t, managedProcessComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(createMockProcessComponentsFactoryArgs())
		managedProcessComponents, err := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedProcessComponents)
	})
}

func TestManagedProcessComponentsMidas_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testManagedProcessComponentsCreateShouldWork(t, common.MetachainShardId, getRunTypeComponentsMock())
	testManagedProcessComponentsCreateShouldWork(t, 0, getRunTypeComponentsMock())
	testManagedProcessComponentsCreateShouldWork(t,  core.SovereignChainShardId, getSovereignRunTypeComponentsMock())
}

func TestManagedProcessComponentsMidas_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid params should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PublicKeyPeerId.Type = "invalid"
		processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(args)
		managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
		require.NotNil(t, managedProcessComponents)

		err := managedProcessComponents.Create()
		require.Error(t, err)
	})
	t.Run("meta should create meta components", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
		shardCoordinator.CurrentShard = core.MetachainShardId
		shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
			if core.IsSmartContractOnMetachain(address[len(address)-1:], address) {
				return core.MetachainShardId
			}
			return 0
		}

		args := createMockProcessComponentsFactoryArgs()
		componentsMock.SetShardCoordinator(t, args.BootstrapComponents, shardCoordinator)
		processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(args)
		managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
		_ = managedProcessComponents.Create()

		assert.Equal(t, "*sync.metaForkDetector", fmt.Sprintf("%T", managedProcessComponents.ForkDetector()))
		assert.Equal(t, "*track.metaBlockTrack", fmt.Sprintf("%T", managedProcessComponents.BlockTracker()))
	})
	t.Run("shard should create shard components", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
		shardCoordinator.CurrentShard = 0
		args := createMockProcessComponentsFactoryArgs()
		componentsMock.SetShardCoordinator(t, args.BootstrapComponents, shardCoordinator)
		processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(args)
		managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
		_ = managedProcessComponents.Create()

		assert.Equal(t, "*sync.shardForkDetector", fmt.Sprintf("%T", managedProcessComponents.ForkDetector()))
		assert.Equal(t, "*track.shardBlockTrack", fmt.Sprintf("%T", managedProcessComponents.BlockTracker()))
	})
	t.Run("sovereign should create sovereign components", func(t *testing.T) {
		t.Parallel()

		args := createMockSovereignProcessComponentsFactoryArgs()
		processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(args)
		managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
		_ = managedProcessComponents.Create()

		assert.Equal(t, "*sync.sovereignChainShardForkDetector", fmt.Sprintf("%T", managedProcessComponents.ForkDetector()))
		assert.Equal(t, "*track.sovereignChainShardBlockTrack", fmt.Sprintf("%T", managedProcessComponents.BlockTracker()))
	})

}

func TestManagedProcessComponentsMidas_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(createMockProcessComponentsFactoryArgs())
	managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
	require.NotNil(t, managedProcessComponents)
	require.Equal(t, errorsMx.ErrNilProcessComponents, managedProcessComponents.CheckSubcomponents())

	err := managedProcessComponents.Create()
	require.NoError(t, err)

	require.Nil(t, managedProcessComponents.CheckSubcomponents())
}

func TestManagedProcessComponentsMidas_Close(t *testing.T) {
	t.Parallel()

	processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(createMockProcessComponentsFactoryArgs())
	managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
	err := managedProcessComponents.Create()
	require.NoError(t, err)

	err = managedProcessComponents.Close()
	require.NoError(t, err)

	err = managedProcessComponents.Close()
	require.NoError(t, err)
}

func TestManagedProcessComponentsMidas_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedProcessComponents, _ := processComp.NewManagedProcessComponentsMidas(nil)
	require.True(t, managedProcessComponents.IsInterfaceNil())

	processComponentsFactory, _ := processComp.NewProcessComponentsFactoryMidas(createMockProcessComponentsFactoryArgs())
	managedProcessComponents, _ = processComp.NewManagedProcessComponentsMidas(processComponentsFactory)
	require.False(t, managedProcessComponents.IsInterfaceNil())
}
