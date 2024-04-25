package vm_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewVmContainerMetaCreatorFactoryMidas(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		bhhc := &factory.BlockChainHookHandlerFactoryMock{}
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(bhhc)
		require.Nil(t, err)
		require.False(t, vmContainerMetaFactory.IsInterfaceNil())
	})

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(nil)
		require.ErrorIs(t, err, process.ErrNilBlockChainHook)
		require.True(t, vmContainerMetaFactory.IsInterfaceNil())
	})
}

func TestVmContainerMetaFactoryMidas_CreateVmContainerFactoryMeta(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(bhhc)
	require.Nil(t, err)
	require.False(t, vmContainerMetaFactory.IsInterfaceNil())

	argsBlockchain := createMockBlockChainHookArgs()
	gasSchedule := makeGasSchedule()
	argsMeta := createVmContainerMockArgument(gasSchedule)
	args := vm.ArgsVmContainerFactory{
		Economics:           argsMeta.Economics,
		MessageSignVerifier: argsMeta.MessageSignVerifier,
		GasSchedule:         argsMeta.GasSchedule,
		NodesConfigProvider: argsMeta.NodesConfigProvider,
		Hasher:              argsMeta.Hasher,
		Marshalizer:         argsMeta.Marshalizer,
		SystemSCConfig:      argsMeta.SystemSCConfig,
		ValidatorAccountsDB: argsMeta.ValidatorAccountsDB,
		UserAccountsDB:      argsMeta.UserAccountsDB,
		ChanceComputer:      argsMeta.ChanceComputer,
		ShardCoordinator:    argsMeta.ShardCoordinator,
		PubkeyConv:          argsMeta.PubkeyConv,
		EnableEpochsHandler: argsMeta.EnableEpochsHandler,
		NodesCoordinator:    argsMeta.NodesCoordinator,
	}

	vmContainer, vmFactory, err := vmContainerMetaFactory.CreateVmContainerFactory(argsBlockchain, args)
	require.Nil(t, err)
	require.Equal(t, "*containers.virtualMachinesContainer", fmt.Sprintf("%T", vmContainer))
	require.Equal(t, "*metachain.vmContainerFactoryMidas", fmt.Sprintf("%T", vmFactory))
}
