#!/usr/bin/env bash
### every exit != 0 fails the script
set -ex

ARCH=$(arch | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g')
echo "Install Audio Requirements"

zypper install -ny curl git
zypper install -yn ffmpeg pulseaudio-utils


mkdir -p /var/run/pulse

cd $STARTUPDIR
mkdir jsmpeg
wget -qO- https://kasmweb-build-artifacts.s3.amazonaws.com/kasm_websocket_relay/f7efb82dc59a02d1b99e2e2b3c6d127dc548ba72/kasm_websocket_relay_${ARCH}_develop.f7efb8.tar.gz | tar xz --strip 1 -C $STARTUPDIR/jsmpeg
chmod +x $STARTUPDIR/jsmpeg/kasm_audio_out-linux
