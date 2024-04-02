package staking

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

const (
	minGasPrice     = 1000000000
	txVersion       = 1
	mockTxSignature = "sig"

	OkReturnCode                           = "ok"
	UnStakedStatus                         = "unStaked"
	MockBLSSignature                       = "010101"
	GasLimitForStakeOperation              = 50_000_000
	GasLimitForUnBond                      = 12_000_000
	MaxNumOfBlockToGenerateWhenExecutingTx = 7

	QueuedStatus    = "queued"
	StakedStatus    = "staked"
	NotStakedStatus = "notStaked"
	AuctionStatus   = "auction"
)

var (
	ZeroValue              = big.NewInt(0)
	InitialDelegationValue = big.NewInt(0).Mul(OneEGLD, big.NewInt(1250))
	MinimumStakeValue      = big.NewInt(0).Mul(OneEGLD, big.NewInt(2500))
	OneEGLD                = big.NewInt(1000000000000000000)
)

// GetNonce will return the nonce of the provided address
func GetNonce(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, address dtos.WalletAddress) uint64 {
	account, err := cs.GetAccount(address)
	require.Nil(t, err)

	return account.Nonce
}

// GenerateTransaction will generate a transaction based on input data
func GenerateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   txVersion,
		Signature: []byte(mockTxSignature),
	}
}

// GetBLSKeyStatus will return the bls key status
func GetBLSKeyStatus(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte) string {
	scQuery := &process.SCQuery{
		ScAddress:  vm.StakingSCAddress,
		FuncName:   "getBLSKeyStatus",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, OkReturnCode, result.ReturnCode)

	return string(result.ReturnData[0])
}

// GetAllNodeStates will return the status of all the nodes that belong to the provided address
func GetAllNodeStates(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte) map[string]string {
	scQuery := &process.SCQuery{
		ScAddress:  address,
		FuncName:   "getAllNodeStates",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, OkReturnCode, result.ReturnCode)

	m := make(map[string]string)
	status := ""
	for _, resultData := range result.ReturnData {
		if len(resultData) != 96 {
			// not a BLS key
			status = string(resultData)
			continue
		}

		m[hex.EncodeToString(resultData)] = status
	}

	return m
}

// CheckValidatorStatus will compare the status of the provided bls key with the provided expected status
func CheckValidatorStatus(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, blsKey string, expectedStatus string) {
	err := cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)

	validatorsStatistics, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	require.Equal(t, expectedStatus, validatorsStatistics[blsKey].ValidatorStatus)
}
