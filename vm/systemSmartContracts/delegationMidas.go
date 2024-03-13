//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. delegation.proto
package systemSmartContracts

import (
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/vm"
	"math/big"
)

type delegationMidas struct {
	delegation
}

// NewDelegationSystemSC creates a new delegation system SC
func NewDelegationSystemSCMidas(args ArgsNewDelegation) (*delegationMidas, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingSCAddress) < 1 {
		return nil, fmt.Errorf("%w for staking sc address", vm.ErrInvalidAddress)
	}
	if len(args.ValidatorSCAddress) < 1 {
		return nil, fmt.Errorf("%w for validator sc address", vm.ErrInvalidAddress)
	}
	if len(args.DelegationMgrSCAddress) < 1 {
		return nil, fmt.Errorf("%w for delegation sc address", vm.ErrInvalidAddress)
	}
	if len(args.GovernanceSCAddress) < 1 {
		return nil, fmt.Errorf("%w for governance sc address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.SigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}
	if args.DelegationSCConfig.MinServiceFee > args.DelegationSCConfig.MaxServiceFee {
		return nil, fmt.Errorf("%w minServiceFee bigger than maxServiceFee", vm.ErrInvalidDelegationSCConfig)
	}
	if args.DelegationSCConfig.MaxServiceFee < 1 {
		return nil, fmt.Errorf("%w maxServiceFee must be more than 0", vm.ErrInvalidDelegationSCConfig)
	}
	if len(args.AddTokensAddress) < 1 {
		return nil, fmt.Errorf("%w for add tokens address", vm.ErrInvalidAddress)
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, vm.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.AddTokensToDelegationFlag,
		common.DelegationSmartContractFlag,
		common.ChangeDelegationOwnerFlag,
		common.ReDelegateBelowMinCheckFlag,
		common.ValidatorToDelegationFlag,
		common.DeleteDelegatorAfterClaimRewardsFlag,
		common.ComputeRewardCheckpointFlag,
		common.StakingV2FlagAfterEpoch,
		common.FixDelegationChangeOwnerOnAccountFlag,
		common.MultiClaimOnDelegationFlag,
	})
	if err != nil {
		return nil, err
	}

	d := &delegationMidas{
		delegation{
			eei:                    args.Eei,
			stakingSCAddr:          args.StakingSCAddress,
			validatorSCAddr:        args.ValidatorSCAddress,
			delegationMgrSCAddress: args.DelegationMgrSCAddress,
			gasCost:                args.GasCost,
			marshalizer:            args.Marshalizer,
			minServiceFee:          args.DelegationSCConfig.MinServiceFee,
			maxServiceFee:          args.DelegationSCConfig.MaxServiceFee,
			sigVerifier:            args.SigVerifier,
			unBondPeriodInEpochs:   args.StakingSCConfig.UnBondPeriodInEpochs,
			endOfEpochAddr:         args.EndOfEpochAddress,
			governanceSCAddr:       args.GovernanceSCAddress,
			addTokensAddr:          args.AddTokensAddress,
			enableEpochsHandler:    args.EnableEpochsHandler,
		},
	}

	var okValue bool

	d.unJailPrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.UnJailValue, conversionBase)
	if !okValue || d.unJailPrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidUnJailCost, args.StakingSCConfig.UnJailValue)
	}
	d.minStakeValue, okValue = big.NewInt(0).SetString(args.StakingSCConfig.MinStakeValue, conversionBase)
	if !okValue || d.minStakeValue.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidMinStakeValue, args.StakingSCConfig.MinStakeValue)
	}
	d.nodePrice, okValue = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !okValue || d.nodePrice.Cmp(zero) <= 0 {
		return nil, fmt.Errorf("%w, value is %v", vm.ErrInvalidNodePrice, args.StakingSCConfig.GenesisNodePrice)
	}

	return d, nil
}
