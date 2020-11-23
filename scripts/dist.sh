#!/usr/bin/env bash
set -e

export GO111MODULE=on

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${scriptDir}/..
export CGO_ENABLED=0

XC_ARCH=${XC_ARCH:-"386 amd64 arm arm64"}
XC_OS=${XC_OS:-"solaris darwin freebsd linux windows"}

echo "==> Building..."
"$(which gox)" \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -osarch="!darwin/arm !darwin/arm64 !darwin/386" \
    -output "build/{{.OS}}_{{.Arch}}/{{.Dir}}" \
    -tags="${GOTAGS}" \
    .

for PLATFORM in $(find ./build -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename ${PLATFORM})
    echo "--> ${OSARCH}"

    pushd $PLATFORM >/dev/null 2>&1
    if [[ ${OSARCH} = windows* ]] ; then
        zip ../sweetcher_${OSARCH}.zip ./*
    else
        tar czvf ../sweetcher_${OSARCH}.tgz ./*
    fi
    rm ./*
    popd >/dev/null 2>&1
done
