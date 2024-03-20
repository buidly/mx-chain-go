package systemSmartContracts

import (
	"bytes"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
)

func createMockArgumentsForDelegationMidas() ArgsNewDelegationMidas {
	return ArgsNewDelegationMidas{
		ArgsNewDelegation: ArgsNewDelegation{
			DelegationSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 10,
				MaxServiceFee: 200,
			},
			StakingSCConfig: config.StakingSystemSCConfig{
				MinStakeValue:    "10",
				UnJailValue:      "15",
				GenesisNodePrice: "100",
			},
			Eei:                    &mock.SystemEIStub{},
			SigVerifier:            &mock.MessageSignVerifierMock{},
			DelegationMgrSCAddress: vm.DelegationManagerSCAddress,
			StakingSCAddress:       vm.StakingSCAddress,
			ValidatorSCAddress:     vm.ValidatorSCAddress,
			GasCost:                vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{ESDTIssue: 10}},
			Marshalizer:            &mock.MarshalizerMock{},
			EndOfEpochAddress:      vm.EndOfEpochAddress,
			GovernanceSCAddress:    vm.GovernanceSCAddress,
			AddTokensAddress:       bytes.Repeat([]byte{1}, 32),
			EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(
				common.DelegationSmartContractFlag,
				common.StakingV2FlagAfterEpoch,
				common.AddTokensToDelegationFlag,
				common.DeleteDelegatorAfterClaimRewardsFlag,
				common.ComputeRewardCheckpointFlag,
				common.ValidatorToDelegationFlag,
				common.ReDelegateBelowMinCheckFlag,
				common.MultiClaimOnDelegationFlag,
			),
		},
		AbstractStakingAddr: AbstractStakingSCAddress,
	}
}

