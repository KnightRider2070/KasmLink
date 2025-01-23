#!/usr/bin/env bash
### every exit != 0 fails the script
set -e

COMMIT_ID="ab327e308e000ddff2e5020a4a66e1fe4935d380"
BRANCH="develop"
COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)

ARCH=$(arch | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g')

STARTUPDIR="/dockerstartup"
mkdir -p $STARTUPDIR/gamepad

echo "Downloading and extracting gamepad server tarball..."
if wget -qO- https://kasmweb-build-artifacts.s3.amazonaws.com/kasm_gamepad_server/${COMMIT_ID}/kasm_gamepad_server_${ARCH}_${BRANCH}.${COMMIT_ID_SHORT}.tar.gz \
| tar -xvz -C $STARTUPDIR/gamepad/; then
    echo "Download and extraction successful."
else
    echo "Error: Failed to download or extract kasm_gamepad_server tarball."
    exit 1
fi

echo "Setting up gamepad resources..."
SCRIPT_PATH="$( cd "$(dirname "$0")" ; pwd -P )"
SCRIPT_PATH="$(realpath "$SCRIPT_PATH")"

mkdir -p /usr/share/extra/icons/
if [[ -f "${SCRIPT_PATH}/gamepad.svg" ]]; then
    cp "${SCRIPT_PATH}/gamepad.svg" /usr/share/extra/icons/gamepad.svg
else
    echo "Error: gamepad.svg not found in ${SCRIPT_PATH}."
    exit 1
fi

echo "${BRANCH}:${COMMIT_ID}" > $STARTUPDIR/gamepad/kasm_gamepad_server.version
echo "Gamepad setup complete."
