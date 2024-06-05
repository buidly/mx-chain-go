#!/usr/bin/env bash

# From config.sh
export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export SOVEREIGN_BRIDGE_PATH="$MULTIVERSXTESTNETSCRIPTSDIR/sovereignBridge"
export SCRIPT_PATH=$SOVEREIGN_BRIDGE_PATH

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"
export USE_ELASTICSEARCH=0

source "$MULTIVERSXTESTNETSCRIPTSDIR/include/config.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/build.sh"

updateGenesisMidas() {
    python3 $SCRIPT_PATH/pyScripts/update_genesis_midas.py $TESTNETDIR
}

echo 'Running config.sh...'
./config.sh

echo 'Updating genesis.json file for Midas'
updateGenesisMidas
