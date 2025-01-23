#!/usr/bin/env bash
set -e

echo "Install KasmVNC server"
cd /tmp
BUILD_ARCH=$(uname -p)
COMMIT_ID="6c368aa746bf16bab692535597e1d031affc7c77"
BRANCH="release" # just use 'release' for a release branch
KASMVNC_VER="1.3.2"
COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)

# Naming scheme is now different between an official release and feature branch
KASM_VER_NAME_PART="${KASMVNC_VER}_${BRANCH}_${COMMIT_ID_SHORT}"
if [[ "${BRANCH}" == "release" ]] ; then
  KASM_VER_NAME_PART="${KASMVNC_VER}"
fi

if [[ "$(arch)" =~ ^x86_64$ ]] ; then
        BUILD_URL="https://kasmweb-build-artifacts.s3.amazonaws.com/kasmvnc/${COMMIT_ID}/kasmvncserver_opensuse_15_${KASM_VER_NAME_PART}_x86_64.rpm"
else
        BUILD_URL="https://kasmweb-build-artifacts.s3.amazonaws.com/kasmvnc/${COMMIT_ID}/kasmvncserver_opensuse_15_${KASM_VER_NAME_PART}_aarch64.rpm"
fi


mkdir -p /etc/pki/tls/private
wget "${BUILD_URL}" -O kasmvncserver.rpm
zypper install -y \
    libdrm_amdgpu1 \
    libdrm_radeon1
if [ "${BUILD_ARCH}" == "x86_64" ]; then
    zypper install -y libdrm_intel1
fi
zypper install -y --allow-unsigned-rpm ./kasmvncserver.rpm
rm kasmvncserver.rpm

mkdir -p $KASM_VNC_PATH/www/Downloads
chown -R 0:0 $KASM_VNC_PATH
chmod -R og-w $KASM_VNC_PATH
ln -sf /home/kasm-user/Downloads $KASM_VNC_PATH/www/Downloads/Downloads
chown -R 1000:0 $KASM_VNC_PATH/www/Downloads
