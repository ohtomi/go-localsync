#!/bin/bash

MAIN_PACKAGE=lsync
REL_TO_ROOT=..

TEST_ENVIRONMENT="LSYNC_TRACE=1 LSYNC_LONG_RUN_TEST=$3"

GOX_ALL_OS="darwin linux windows"
GOX_ALL_ARCH="386 amd64"
GOX_MAIN_OS="darwin"
GOX_MAIN_ARCH="amd64"


function usage() {
  echo "
Usage: $0 [fmt|stringer|compile|test|package|release]
"
}


case "$1" in
  "fmt")
    gofmt -w .
    ;;
  "stringer")
    cd command
    stringer -type ExitCode -output meta_exitcode_string.go meta.go
    ;;
  "compile")
    $0 stringer

    cd ${MAIN_PACKAGE}
    gox \
      -ldflags "-X main.GitCommit=$(git describe --always)" \
      -os="${GOX_MAIN_OS}" \
      -arch="${GOX_MAIN_ARCH}" \
      -output "${REL_TO_ROOT}/pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"
    ;;
  "test")
    echo
    echo testing ...
    env "${TEST_ENVIRONMENT}" go test ./... $2
    ;;
  "package")
    $0 stringer

    cd ${MAIN_PACKAGE}
    rm -fr "${REL_TO_ROOT}/pkg"
    gox \
      -ldflags "-X main.GitCommit=$(git describe --always)" \
      -os="${GOX_ALL_OS}" \
      -arch="${GOX_ALL_ARCH}" \
      -output "${REL_TO_ROOT}/pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

    repo=$(grep "const Name " version.go | sed -E 's/.*"(.+)"$/\1/')
    version=$(grep "const Version " version.go | sed -E 's/.*"(.+)"$/\1/')
    cd ${REL_TO_ROOT}

    rm -fr "./dist/${version}"
    mkdir -p "./dist/${version}"
    for platform in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
      platform_name=$(basename ${platform})
      archive_name=${repo}_${version}_${platform_name}
      pushd ${platform}
      zip "../../dist/${version}/${archive_name}.zip" ./*
      popd
    done

    pushd "./dist/${version}"
    shasum -a 256 * > "./${version}_SHASUMS"
    popd
    ;;
  "release")
    version=$(grep "const Version " ${MAIN_PACKAGE}/version.go | sed -E 's/.*"(.+)"$/\1/')
    ghr "${version}" "./dist/${version}"
    ;;
  *)
    usage
    exit 1
    ;;
esac
