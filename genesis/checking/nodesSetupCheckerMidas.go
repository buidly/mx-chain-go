package checking

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/genesis"
)

type nodeSetupCheckerMidas struct {
	nodeSetupChecker
}

// NewNodesSetupChecker will create a node setup checker able to check the initial nodes against the provided genesis values
func NewNodesSetupCheckerMidas(args ArgsNodesSetupChecker) (*nodeSetupCheckerMidas, error) {
	if check.IfNil(args.AccountsParser) {
		return nil, genesis.ErrNilAccountsParser
	}
	if args.InitialNodePrice == nil {
		return nil, genesis.ErrNilInitialNodePrice
	}
	if args.InitialNodePrice.Cmp(big.NewInt(minimumAcceptedNodePrice)) < 0 {
		return nil, fmt.Errorf("%w, minimum accepted is %d",
			genesis.ErrInvalidInitialNodePrice, minimumAcceptedNodePrice)
	}
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if check.IfNil(args.KeyGenerator) {
		return nil, genesis.ErrNilKeyGenerator
	}

	return &nodeSetupCheckerMidas{
		nodeSetupChecker: nodeSetupChecker{
			accountsParser:           args.AccountsParser,
			initialNodePrice:         args.InitialNodePrice,
			validatorPubkeyConverter: args.ValidatorPubKeyConverter,
			keyGenerator:             args.KeyGenerator,
		},
	}, nil
}

func (nsc *nodeSetupCheckerMidas) Check(initialNodes []nodesCoordinator.GenesisNodeInfoHandler) error {
	err := nsc.checkGenesisNodes(initialNodes)
	if err != nil {
		return err
	}

	initialAccounts := nsc.getClonedInitialAccounts()
	delegated := nsc.createDelegatedValues(initialAccounts)
	err = nsc.traverseInitialNodesSubtractingStakedValue(initialAccounts, initialNodes, delegated)
	if err != nil {
		return err
	}

	return nsc.checkRemainderInitialAccounts(initialAccounts, delegated)
}

func (nsc *nodeSetupCheckerMidas) traverseInitialNodesSubtractingStakedValue(
	initialAccounts []genesis.InitialAccountHandler,
	initialNodes []nodesCoordinator.GenesisNodeInfoHandler,
	delegated map[string]*delegationAddress,
) error {
	for _, initialNode := range initialNodes {
		err := nsc.subtractStakedValue(initialNode.AddressBytes(), initialAccounts, delegated)
		if err != nil {
			validatorPubkeyEncoded := nsc.validatorPubkeyConverter.SilentEncode(initialNode.PubKeyBytes(), log)

			return fmt.Errorf("'%w' while processing node pubkey %s",
				err, validatorPubkeyEncoded)
		}
	}

	return nil
}

func (nsc *nodeSetupCheckerMidas) subtractStakedValue(
	addressBytes []byte,
	initialAccounts []genesis.InitialAccountHandler,
	delegated map[string]*delegationAddress,
) error {

	for _, ia := range initialAccounts {
		if bytes.Equal(ia.AddressBytes(), addressBytes) {
			// Changed this for Midas since no staking value should be set at genesis
			if ia.GetStakingValue().Cmp(zero) != 0 {
				return errors.New("staking value should be zero")
			}

			return nil
		}

		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			return genesis.ErrNilDelegationHandler
		}
		if !bytes.Equal(dh.AddressBytes(), addressBytes) {
			continue
		}

		addr, ok := delegated[string(dh.AddressBytes())]
		if !ok {
			continue
		}

		addr.value.Sub(addr.value, nsc.initialNodePrice)
		if addr.value.Cmp(zero) < 0 {
			return genesis.ErrDelegationValueIsNotEnough
		}

		return nil
	}

	return genesis.ErrNodeNotStaked
}
