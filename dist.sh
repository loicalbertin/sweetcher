#!/usr/bin/env bash
set -e

export GO111MODULE=on

if [[ "${CI}" == "true" ]] && [[ -z "${TRAVIS_TAG}" ]] ; then
    echo "Skipping dist build as we are not building a tag"
    exit 0
fi

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${scriptDir}
export CGO_ENABLED=0

XC_ARCH=${XC_ARCH:-"386 amd64 arm arm64"}
XC_OS=${XC_OS:-"solaris darwin freebsd linux windows"}

echo "==> Building..."
"$(which gox)" \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -osarch="!darwin/arm !darwin/arm64" \
    -output "build/{{.OS}}_{{.Arch}}/{{.Dir}}" \
    -tags="${GOTAGS}" \
    .

for PLATFORM in $(find ./build -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename ${PLATFORM})
    echo "--> ${OSARCH}"

    pushd $PLATFORM >/dev/null 2>&1
    zip ../sweetcher_${OSARCH}.zip ./*
    tar czvf ../sweetcher_${OSARCH}.tgz ./*
    rm ./*
    popd >/dev/null 2>&1
done
