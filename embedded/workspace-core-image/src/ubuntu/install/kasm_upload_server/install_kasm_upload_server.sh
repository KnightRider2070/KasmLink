set -ex

COMMIT_ID="b234364fdef2dc96ac106ac5e4350c86c1f45021"
BRANCH="develop"
COMMIT_ID_SHORT=$(echo "${COMMIT_ID}" | cut -c1-6)
ARCH=$(arch | sed 's/aarch64/arm64/g' | sed 's/x86_64/amd64/g')

mkdir -p $STARTUPDIR/upload_server
if ! wget --quiet https://kasmweb-build-artifacts.s3.amazonaws.com/kasm_upload_service/${COMMIT_ID}/kasm_upload_service_${ARCH}_${BRANCH}.${COMMIT_ID_SHORT}.tar.gz -O /tmp/kasm_upload_server.tar.gz; then
    echo "Error: Failed to download kasm_upload_service tarball."
    exit 1
fi

if ! tar -xvf /tmp/kasm_upload_server.tar.gz -C $STARTUPDIR/upload_server; then
    echo "Error: Failed to extract tarball."
    exit 1
fi

rm /tmp/kasm_upload_server.tar.gz
echo "${BRANCH}:${COMMIT_ID}" > $STARTUPDIR/upload_server/kasm_upload_service.version
