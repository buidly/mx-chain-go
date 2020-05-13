package mock

import "math/big"

// TxExecutionProcessorStub -
type TxExecutionProcessorStub struct {
	ExecuteTransactionCalled func(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error
	GetNonceCalled           func(senderBytes []byte) (uint64, error)
	AddBalanceCalled         func(senderBytes []byte, value *big.Int) error
	AddNonceCalled           func(senderBytes []byte, nonce uint64) error
}

// ExecuteTransaction -
func (teps *TxExecutionProcessorStub) ExecuteTransaction(nonce uint64, sndAddr []byte, rcvAddress []byte, value *big.Int, data []byte) error {
	if teps.ExecuteTransactionCalled != nil {
		return teps.ExecuteTransactionCalled(nonce, sndAddr, rcvAddress, value, data)
	}

	return nil
}

// GetNonce -
func (teps *TxExecutionProcessorStub) GetNonce(senderBytes []byte) (uint64, error) {
	if teps.GetNonceCalled != nil {
		return teps.GetNonceCalled(senderBytes)
	}

	return 0, nil
}

// AddBalance -
func (teps *TxExecutionProcessorStub) AddBalance(senderBytes []byte, value *big.Int) error {
	if teps.AddBalanceCalled != nil {
		return teps.AddBalanceCalled(senderBytes, value)
	}

	return nil
}

// AddNonce -
func (teps *TxExecutionProcessorStub) AddNonce(senderBytes []byte, nonce uint64) error {
	if teps.AddNonceCalled != nil {
		return teps.AddNonceCalled(senderBytes, nonce)
	}

	return nil
}

// IsInterfaceNil -
func (teps *TxExecutionProcessorStub) IsInterfaceNil() bool {
	return teps == nil
}