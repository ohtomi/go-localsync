package command

import (
	"os"

	"github.com/mitchellh/cli"
)

type ExitCode int

const (
	ExitCodeOK ExitCode = iota
	ExitCodeError
	ExitCodeParseFlagsError
	ExitCodeBadArgs
)

const (
	EnvDebug       = "LSYNC_DEBUG"
	EnvLongRunTest = "LSYNC_LONG_RUN_TEST"
)

var (
	DebugMode       = os.Getenv(EnvDebug) != ""
	LongRunTestMode = os.Getenv(EnvLongRunTest) != ""
)

const (
	DefaultPid = "./lsync.pid"
)

// Meta contain the meta-option that nearly all subcommand inherited.
type Meta struct {
	Ui cli.Ui
}
