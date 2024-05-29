package basic

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/process"
	sovereignChainSimulator "github.com/multiversx/mx-chain-go/sovereignnode/chainSimulator"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../node/config/"
	sovereignConfigPath        = "../../../config/"
	adderWasmPath              = "../testdata/adder.wasm"
)

var oneEgld = big.NewInt(1000000000000000000)
var initialMinting = big.NewInt(0).Mul(oneEgld, big.NewInt(100))

func TestSmartContract_Adder(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs, err := sovereignChainSimulator.NewSovereignChainSimulator(sovereignChainSimulator.ArgsSovereignChainSimulator{
		SovereignConfigPath: sovereignConfigPath,
		ArgsChainSimulator: &chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck: false,
			TempDir:                t.TempDir(),
			PathToInitialConfig:    defaultPathToInitialConfig,
			NumOfShards:            1,
			GenesisTimestamp:       time.Now().Unix(),
			RoundDurationInMillis:  uint64(6000),
			RoundsPerEpoch:         core.OptionalUint64{},
			ApiInterface:           api.NewNoApiInterface(),
			MinNodesPerShard:       2,
			ConsensusGroupSize:     2,
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(200)
	require.Nil(t, err)

	nodeHandler := cs.GetNodeHandler(core.SovereignChainShardId)
	systemScAddress, _ := nodeHandler.GetCoreComponents().AddressPubKeyConverter().Decode("erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu")

	wallet, err := cs.GenerateAndMintWalletAddress(core.SovereignChainShardId, initialMinting)
	require.Nil(t, err)
	nonce := uint64(0)

	initSum := big.NewInt(123)
	initArgs := "@" + hex.EncodeToString(initSum.Bytes())
	deployedContractAddress := chainSimulatorIntegrationTests.DeployContract(t, cs, wallet.Bytes, &nonce, systemScAddress, initArgs, adderWasmPath)

	res, _, err := nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum := big.NewInt(0).SetBytes(res.ReturnData[0])
	require.Equal(t, initSum, sum)

	addedSum := big.NewInt(10)
	addTxData := "add@" + hex.EncodeToString(addedSum.Bytes())
	chainSimulatorIntegrationTests.SendTransaction(t, cs, wallet.Bytes, &nonce, deployedContractAddress, big.NewInt(0), addTxData, uint64(10000000))

	res, _, err = nodeHandler.GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress:  deployedContractAddress,
		FuncName:   "getSum",
		CallerAddr: nil,
		BlockNonce: core.OptionalUint64{},
	})
	require.Nil(t, err)
	sum = big.NewInt(0).SetBytes(res.ReturnData[0])
	require.Equal(t, initSum.Add(initSum, addedSum), sum)
}
