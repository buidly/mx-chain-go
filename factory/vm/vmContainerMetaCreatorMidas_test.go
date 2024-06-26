package vm_test

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/vmContext"
	"testing"

	"github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewVmContainerMetaCreatorFactoryMidas(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(&factory.BlockChainHookHandlerFactoryMock{}, &vmContext.VMContextCreatorStub{})
		require.Nil(t, err)
		require.False(t, vmContainerMetaFactory.IsInterfaceNil())
	})

	t.Run("nil blockchain hook creator", func(t *testing.T) {
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(nil, &vmContext.VMContextCreatorStub{})
		require.ErrorIs(t, err, errors.ErrNilBlockChainHookCreator)
		require.True(t, vmContainerMetaFactory.IsInterfaceNil())
	})

	t.Run("nil vm context creator", func(t *testing.T) {
		vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(&factory.BlockChainHookHandlerFactoryMock{}, nil)
		require.ErrorIs(t, err, errors.ErrNilVMContextCreator)
		require.True(t, vmContainerMetaFactory.IsInterfaceNil())
	})
}

func TestVmContainerMetaFactoryMidas_CreateVmContainerFactoryMeta(t *testing.T) {
	t.Parallel()

	vmContainerMetaFactory, err := vm.NewVmContainerMetaFactoryMidas(&factory.BlockChainHookHandlerFactoryMock{}, &vmContext.VMContextCreatorStub{})
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
