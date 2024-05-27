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

# From sovereignBridge stopAndCleanSovereign
echo 'Stoping all Blockchain components...'
./stop.sh

echo 'Stopping Sovereign Bridge Service...'
screen -S sovereignBridgeService -X kill

echo 'Stopping Observer...'
stopObserver
