#!/usr/bin/env bash

# From config.sh
export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

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

updateGenesisMidas() {
    python3 $SCRIPT_PATH/pyScripts/update_genesis_midas.py $TESTNETDIR
}

updateProxyMidas() {
    python3 $SCRIPT_PATH/pyScripts/update_proxy.py $TESTNETDIR
}

###############################################################
# From sovereignBridge deployMainChainContractsAndSetupObserver
# Rest of the steps are done manually:
# deployEsdtSafeContract - contract already deployed on Devnet
# deployFeeMarketContract - contract already deployed on Devnet
# setFeeMarketAddress - transaction executed on Devnet
# disableFeeMarketContract - transaction executed on Devnet
# unpauseEsdtSafeContract - transaction executed on Devnet
# setGenesisContract - genesisSmartContracts.json updated to deploy Esdt Safe and Fee Market on Sovereign
# updateSovereignConfig - sovereignConfig.toml updated with correct contract addresses for events

echo 'Preparing Observer...'
prepareObserver

##############################################
# Partial from sovereignBridge sovereignDeploy
echo 'Updating NotifierNotarizationRound...'
updateNotifierNotarizationRound

echo 'Running config.sh...'
./config.sh

copySovereignNodeConfig

echo 'Updating genesis.json file for Midas'
updateGenesisMidas

echo 'Updating proxy config'
updateProxyMidas
