#!/usr/bin/env bash

# From config.sh
export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export TESTNETMODE=$1

export SOVEREIGN_BRIDGE_PATH="$MULTIVERSXTESTNETSCRIPTSDIR/sovereignBridge"
export SCRIPT_PATH=$SOVEREIGN_BRIDGE_PATH

# From sovereignBridge/script.sh
source $SOVEREIGN_BRIDGE_PATH/config/configs.cfg
source $SOVEREIGN_BRIDGE_PATH/config/helper.cfg
source $SOVEREIGN_BRIDGE_PATH/config/esdt-safe.snippets.sh
source $SOVEREIGN_BRIDGE_PATH/config/fee-market.snippets.sh
source $SOVEREIGN_BRIDGE_PATH/config/multisig-verifier.snippets.sh
source $SOVEREIGN_BRIDGE_PATH/config/token.snippets.sh
source $SOVEREIGN_BRIDGE_PATH/config/common.snippets.sh
source $SOVEREIGN_BRIDGE_PATH/config/py.snippets.sh
source $SOVEREIGN_BRIDGE_PATH/observer/deployObserver.sh

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/config.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/build.sh"

WALLET_ADDRESS=$(echo "$(head -n 1 $(eval echo ${WALLET}))" | sed -n 's/.* for \([^-]*\)-----.*/\1/p')

echo 'Wallet address:'
echo $WALLET_ADDRESS

##############################################
# Partial from sovereignBridge sovereignDeploy
echo 'Deploying Multisig Verifier contract on Mainchain'
deployMultisigVerifierContract
setEsdtSafeAddress
setMultisigAddress

echo 'Starting Bridge Service' $WALLET $PROXY $ESDT_SAFE_ADDRESS $MULTISIG_VERIFIER_ADDRESS
updateAndStartBridgeService

setSovereignBridgeAddress

echo 'Starting Midas Sovereign Chain...'
./sovereignStart.sh $TESTNETMODE

echo 'Deploying observer for Mainchain...'
deployObserver

echo 'Sending transactions to Chain Esdt and Fee Market contracts on Sovereign'
setFeeMarketAddressSovereign
disableFeeMarketContractSovereign
unpauseEsdtSafeContractSovereign
