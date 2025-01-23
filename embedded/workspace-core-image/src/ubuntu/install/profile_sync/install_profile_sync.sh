#!/usr/bin/env bash

set -eo pipefail

ARCH=$(arch)
BRANCH="develop"
COMMIT_ID="42424ea385a0d10fa7bb5749e207ee70b2a44ae2"

COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)
BINARY_NAME="opensuse_15_${BRANCH}_${COMMIT_ID_SHORT}_${ARCH}-kasm-profile-sync"
BUILD_URL="https://kasmweb-build-artifacts.s3.amazonaws.com/profile-sync/${COMMIT_ID}/${BINARY_NAME}"

cd /usr/bin/
wget "$BUILD_URL"
chmod +x "$BINARY_NAME"
ln -s "$BINARY_NAME" kasm-profile-sync

