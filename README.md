# go-localsync

This tool synchronizes local directories.

## Description

This is a tool to synchronize two local directories.

- copy files from `src` directory to `dest` directory.
- create or delete directories under `dest` directory same as `src` directory.

## Usage

```bash
$ lsync start /path/to/src /path/to/dest

$ lsync start --pid /path/to/pid.file /path/to/src /path/to/dest

$ lsync start --verbose /path/to/src /path/to/dest

$ lsync start --recursive /path/to/src /path/to/dest

$ lsync stop

$ lsync stop --pid /path/to/pid.file

```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/ohtomi/go-localsync
```

Or get binary from [release page](../../releases/latest).

## Contribution

1. Fork ([https://github.com/ohtomi/go-localsync/fork](https://github.com/ohtomi/go-localsync/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[Kenichi Ohtomi](https://github.com/ohtomi)
