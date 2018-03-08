#!/bin/bash

set -e

VER=$1
TOKEN=$2

if [[ -z ${VER} ]]; then
    echo "Err: please specify version to release"
    exit 1
fi

if [[ -z ${TOKEN} ]]; then
    echo "Err: please set Github token for release publishing"
    exit 1
fi


LINUX_ZIP=sonm_${VER}_linux.zip
LINUX_ZIP_PATH=release/${VER}/${LINUX_ZIP}

DARWIN_ZIP=sonm_${VER}_darwin.zip
DARWIN_ZIP_PATH=release/${VER}/${DARWIN_ZIP}

WINDOWS_ZIP=sonm_${VER}_win32.zip
WINDOWS_ZIP_PATH=release/${VER}/${WINDOWS_ZIP}

CURL="curl -s"

for_linux () {
    echo " >>> building for Linux (native)...";

    # make dir structure
    mkdir -p release/${VER}/linux/cli
    mkdir -p release/${VER}/linux/node
    mkdir -p release/${VER}/linux/hub
    mkdir -p release/${VER}/linux/worker

    # configs
    cp .ci/v0.3_config/cli.yaml release/${VER}/linux/cli/cli.yaml
    cp .ci/v0.3_config/node.yaml release/${VER}/linux/node/node.yaml
    cp .ci/v0.3_config/hub.yaml release/${VER}/linux/hub/hub.yaml
    cp .ci/v0.3_config/worker.yaml release/${VER}/linux/worker/worker.yaml

    # build
    GPU_SUPPORT=true make build/insomnia VER=${VER}

    # binaries
    cp target/sonmcli_linux_x86_64 release/${VER}/linux/cli/sonmcli
    cp target/sonmnode_linux_x86_64 release/${VER}/linux/node/sonmnode
    cp target/sonmhub_linux_x86_64 release/${VER}/linux/hub/sonmhub
    cp target/sonmworker_linux_x86_64 release/${VER}/linux/worker/sonmworker

    # archive for github
    zip -r ${LINUX_ZIP_PATH} release/${VER}/linux/

    echo " >>> release for Linux is done";
}

for_windows () {
    echo " >>> building for Windows (xgo)...";

    # make dir structure
    mkdir -p release/${VER}/windows/cli
    mkdir -p release/${VER}/windows/node

    # configs
    cp .ci/v0.3_config/cli.yaml release/${VER}/windows/cli/cli.yaml
    cp .ci/v0.3_config/node.yaml release/${VER}/windows/node/node.yaml

    # build
    xgo -tags='nocgo' -ldflags='-X main.appVersion=$(VER)' --targets=windows/386 --pkg cmd/cli -out sonmcli github.com/sonm-io/core
    xgo -tags='nocgo' -ldflags='-X main.appVersion=$(VER)' --targets=windows/386 --pkg cmd/node -out sonmnode github.com/sonm-io/core

    # binaries
    cp sonmcli-windows-4.0-386.exe release/${VER}/windows/cli/sonmcli.exe
    cp sonmnode-windows-4.0-386.exe release/${VER}/windows/node/sonmnode.exe

    # archive for github
    zip -r ${WINDOWS_ZIP_PATH} release/${VER}/windows/

    echo " >>> release for Windows is done";
}

for_macos () {
    echo " >>> building for macOS (xgo)...";

    # make dir structure
    mkdir -p release/${VER}/darwin/cli
    mkdir -p release/${VER}/darwin/node

    # configs
    cp .ci/v0.3_config/cli.yaml release/${VER}/darwin/cli/cli.yaml
    cp .ci/v0.3_config/node.yaml release/${VER}/darwin/node/node.yaml

    # build
    xgo -tags='nocgo' -ldflags='-X main.appVersion=$(VER)' --targets=darwin/amd64 --pkg cmd/cli -out sonmcli github.com/sonm-io/core
    xgo -tags='nocgo' -ldflags='-X main.appVersion=$(VER)' --targets=darwin/amd64 --pkg cmd/node -out sonmnode github.com/sonm-io/core

    # binaries
    cp sonmcli-darwin-10.6-amd64 release/${VER}/darwin/cli/sonmcli
    cp sonmnode-darwin-10.6-amd64 release/${VER}/darwin/node/sonmcli

    # archive for github
    zip -r ${DARWIN_ZIP_PATH} release/${VER}/darwin/

    echo " >>> release for macOS is done";
}

upload_release () {
    echo " >>> uploading artifacts to Github...";

    GH_AUTH="-u sshaman1101:${TOKEN}"
    GH_RELEASE_REQUEST="-d '{\"tag_name\": \"${VER}\", \"name\": \"${VER}\", \"target_commitish\": \"master\", \"draft\": true, \"body\": \"created automatically by TeamCity\"}'"
    GH_CREATE_RELEASE_URI=https://api.github.com/repos/sonm-io/core/releases

    CURL_REQ="${CURL} $GH_AUTH $GH_RELEASE_REQUEST $GH_CREATE_RELEASE_URI"
    RES=$(eval ${CURL_REQ})

    UPLOAD_URL=$(echo ${RES} | jq '.upload_url' | sed -e s/'{?name,label}'//)
    echo " **** UPLOAD URL = ${UPLOAD_URL}"


    UPLOAD_LINUX_REQ="${CURL} ${GH_AUTH} -H 'Content-Type: application/zip' --data-binary @'${LINUX_ZIP_PATH}' ${UPLOAD_URL}?name=${LINUX_ZIP}"
    UPLOAD_WINDOWS_REQ="${CURL} ${GH_AUTH} -H 'Content-Type: application/zip' --data-binary @'${WINDOWS_ZIP_PATH}' ${UPLOAD_URL}?name=${WINDOWS_ZIP}"
    UPLOAD_DARWIN_REQ="${CURL} ${GH_AUTH} -H 'Content-Type: application/zip' --data-binary @'${DARWIN_ZIP_PATH}' ${UPLOAD_URL}?name=${DARWIN_ZIP}"

    eval ${UPLOAD_LINUX_REQ}
    eval ${UPLOAD_WINDOWS_REQ}
    eval ${UPLOAD_DARWIN_REQ}

    echo " >>> release uploading is complete";
}
remove_artifacts () {
    rm -f sonmcli-darwin-10.6-amd64 sonmnode-darwin-10.6-amd64
    rm -f sonmcli-windows-4.0-386.exe sonmnode-darwin-10.6-amd64

    make clean
}


for_linux
for_macos
for_windows

upload_release
remove_artifacts
