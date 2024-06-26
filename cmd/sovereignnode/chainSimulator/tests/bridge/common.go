package bridge

import (
	"encoding/hex"
	"testing"

	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	chainSim "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-sdk-abi-go/abi"
	"github.com/stretchr/testify/require"
)

const (
	deposit = "deposit"
)

var serializer, _ = abi.NewSerializer(abi.ArgsNewSerializer{
	PartsSeparator: "@",
})

// ArgsBridgeSetup holds the arguments for bridge setup
type ArgsBridgeSetup struct {
	ESDTSafeAddress  []byte
	FeeMarketAddress []byte
}

// DeploySovereignBridgeSetup will deploy all bridge contracts
// This function will:
// - deploy esdt-safe contract
// - deploy fee-market contract
// - set the fee-market address inside esdt-safe contract
// - disable fee in fee-market contract
// - unpause esdt-safe contract so deposit operations can start
func DeploySovereignBridgeSetup(
	t *testing.T,
	cs chainSim.ChainSimulator,
	wallet dtos.WalletAddress,
	esdtSafeWasmPath string,
	feeMarketWasmPath string,
) *ArgsBridgeSetup {
	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress := chainSim.GetSysAccBytesAddress(t, nodeHandler)
	nonce := GetNonce(t, nodeHandler, wallet.Bech32)

	esdtSafeArgs := "@01" + // is_sovereign_chain
		"@" + // min_valid_signers
		"@" + hex.EncodeToString(wallet.Bytes) // initiator_address
	esdtSafeAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, esdtSafeArgs, esdtSafeWasmPath)

	feeMarketArgs := "@" + hex.EncodeToString(esdtSafeAddress) + // esdt_safe_address
		"@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad" + // price_aggregator_address
		"@" + hex.EncodeToString([]byte("USDC-350c4e")) + // usdc_token_id
		"@" + hex.EncodeToString([]byte("WEGLD-a28c59")) // wegld_token_id
	feeMarketAddress := chainSim.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, feeMarketArgs, feeMarketWasmPath)

	setFeeMarketAddressData := "setFeeMarketAddress" +
		"@" + hex.EncodeToString(feeMarketAddress)
	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, setFeeMarketAddressData, uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, feeMarketAddress, chainSim.ZeroValue, "disableFee", uint64(10000000))

	chainSim.SendTransaction(t, cs, wallet.Bytes, &nonce, esdtSafeAddress, chainSim.ZeroValue, "unpause", uint64(10000000))

	return &ArgsBridgeSetup{
		ESDTSafeAddress:  esdtSafeAddress,
		FeeMarketAddress: feeMarketAddress,
	}
}

// GetNonce returns account's nonce
func GetNonce(t *testing.T, nodeHandler process.NodeHandler, address string) uint64 {
	acc, _, err := nodeHandler.GetFacadeHandler().GetAccount(address, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)

	return acc.Nonce
}

// Deposit will deposit tokens in the bridge sc safe contract
func Deposit(
	t *testing.T,
	cs chainSim.ChainSimulator,
	sender []byte,
	nonce *uint64,
	contract []byte,
	tokens []chainSim.ArgsDepositToken,
	receiver []byte,
	transferData *sovereign.TransferData,
) {
	require.True(t, len(tokens) > 0)

	args := make([]any, 0)
	args = append(args, &abi.AddressValue{Value: contract})
	args = append(args, &abi.U32Value{Value: uint32(len(tokens))})
	for _, token := range tokens {
		args = append(args, &abi.StringValue{Value: token.Identifier})
		args = append(args, &abi.U64Value{Value: token.Nonce})
		args = append(args, &abi.BigUIntValue{Value: token.Amount})
	}
	args = append(args, &abi.StringValue{Value: deposit})
	args = append(args, &abi.AddressValue{Value: receiver})
	args = append(args, &abi.OptionalValue{Value: getTransferDataValue(transferData)})

	multiTransferArg, err := serializer.Serialize(args)
	require.Nil(t, err)
	depositArgs := core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + multiTransferArg

	chainSim.SendTransaction(t, cs, sender, nonce, sender, chainSim.ZeroValue, depositArgs, uint64(20000000))
}

func getTransferDataValue(transferData *sovereign.TransferData) any {
	if transferData == nil {
		return nil
	}

	arguments := make([]abi.SingleValue, len(transferData.Args))
	for i, arg := range transferData.Args {
		arguments[i] = &abi.BytesValue{Value: arg}
	}
	return &abi.MultiValue{
		Items: []any{
			&abi.U64Value{Value: transferData.GasLimit},
			&abi.BytesValue{Value: transferData.Function},
			&abi.ListValue{Items: arguments},
		},
	}
}
