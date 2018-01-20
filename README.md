# go-localsync

This tool synchronizes local directories.

## Description

This is a tool to synchronize two local directories.

- copy files from `src` directory to `dest` directory.
- create or delete directories under `dest` directory same as `src` directory.

## Usage

### Watch file system events of the specified directory

```bash
$ lsync -watch -h
usage: lsync [--version] [--help] <command> [<args>]

Available commands are:
    version    Print lsync version and quit
    watch      Watch file system events of the specified directory
```

### Environment Variables

- `LSYNC_DEBUG`: whether or not print stack trace at error.
- `LSYNC_LONG_RUN_TEST`: execute long-run test.

## Install

To install, use `go get`:

```bash
$ go get -u github.com/ohtomi/go-localsync/lsync
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

## License

MIT

## Author

[Kenichi Ohtomi](https://github.com/ohtomi)
