#!/usr/bin/env bash
set -e

echo "Installing Cursors"

cd $INST_SCRIPTS/cursors

for cursor in cursor-aero.tar.gz cursor-bridge.tar.gz cursor-capitaine-r4.tar.gz; do
    if [[ -f "$cursor" ]]; then
        tar -xzf "$cursor" -C /usr/share/icons/
    else
        echo "Error: $cursor not found"
        exit 1
    fi
done
