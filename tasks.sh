#!/bin/bash

MAIN_PACKAGE=lsync

function usage() {
  echo "
Usage: $0 [fmt|stringer|compile|test|package|release]
"
}

if [ $# -ne 1 ]; then
  usage
  exit 1
fi

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

    cd $MAIN_PACKAGE
    gox \
      -ldflags "-X main.GitCommit=$(git describe --always)" \
      -os="darwin" \
      -arch="amd64" \
      -output "../pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"
    ;;
  "test")
    echo
    echo testing ...
    env go test ./... -v
    ;;
  "package")
    $0 stringer

    cd $MAIN_PACKAGE
    rm -fr ../pkg
    gox \
      -ldflags "-X main.GitCommit=$(git describe --always)" \
      -os="darwin linux windows" \
      -arch="386 amd64" \
      -output "../pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"

    repo=$(grep "const Name " version.go | sed -E 's/.*"(.+)"$/\1/')
    version=$(grep "const Version " version.go | sed -E 's/.*"(.+)"$/\1/')
    cd ..

    rm -fr ./dist/${version}
    mkdir -p ./dist/${version}
    for platform in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
      platform_name=$(basename ${platform})
      archive_name=${repo}_${version}_${platform_name}
      pushd ${platform}
      zip ../../dist/${version}/${archive_name}.zip ./*
      popd
    done

    pushd ./dist/${version}
    shasum -a 256 * > ./${version}_SHASUMS
    popd
    ;;
  "release")
    version=$(grep "const Version " ${MAIN_PACKAGE}/version.go | sed -E 's/.*"(.+)"$/\1/')
    ghr ${version} ./dist/${version}
    ;;
  *)
    usage
    exit 1
    ;;
esac
