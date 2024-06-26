package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// TxProcessorStub -
type TxProcessorStub struct {
	ProcessTransactionCalled           func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	VerifyTransactionCalled            func(tx *transaction.Transaction) error
	GetSenderAndReceiverAccountsCalled func(tx *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler, error)
}

// ProcessTransaction -
func (tps *TxProcessorStub) ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
	if tps.ProcessTransactionCalled != nil {
		return tps.ProcessTransactionCalled(transaction)
	}

	return 0, nil
}

// VerifyTransaction -
func (tps *TxProcessorStub) VerifyTransaction(tx *transaction.Transaction) error {
	if tps.VerifyTransactionCalled != nil {
		return tps.VerifyTransactionCalled(tx)
	}

	return nil
}

// GetSenderAndReceiverAccounts -
func (tps *TxProcessorStub) GetSenderAndReceiverAccounts(tx *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler, error) {
	if tps.GetSenderAndReceiverAccountsCalled != nil {
		return tps.GetSenderAndReceiverAccountsCalled(tx)
	}

	return nil, nil, nil
}

// IsInterfaceNil -
func (tps *TxProcessorStub) IsInterfaceNil() bool {
	return tps == nil
}
