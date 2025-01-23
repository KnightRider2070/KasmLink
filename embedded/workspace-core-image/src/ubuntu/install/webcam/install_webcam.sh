#!/usr/bin/env bash
### every exit != 0 fails the script
set -e

STARTUPDIR="/dockerstartup"
DISTRO=${DISTRO:-"default_distro_value"}

echo "DISTRO is set to: ${DISTRO}"

# Normalize DISTRO names
if [ "${DISTRO}" == "oracle7" ]; then
  DISTRO="centos"
elif [ "${DISTRO}" == "oracle8" ]; then
  DISTRO="oracle"
fi

COMMIT_ID="23df7e27fe5c1536bd08da6bd58d65d1d7facb1b"
BRANCH="develop"
COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)

ARCH=$(arch | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g')

mkdir -p $STARTUPDIR/webcam

echo "Downloading and extracting webcam server tarball..."
if wget -qO- https://kasmweb-build-artifacts.s3.amazonaws.com/kasm_webcam_server/${COMMIT_ID}/kasm_webcam_server_${ARCH}_${BRANCH}.${COMMIT_ID_SHORT}.tar.gz \
| tar -xvz -C $STARTUPDIR/webcam/; then
    echo "Download and extraction successful."
else
    echo "Error: Failed to download or extract kasm_webcam_server tarball."
    exit 1
fi

echo "${BRANCH}:${COMMIT_ID}" > $STARTUPDIR/webcam/kasm_webcam_server.version
echo "Webcam setup complete."
