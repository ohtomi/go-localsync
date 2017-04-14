package command

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func SetTestEnv(key, newValue string) func() {

	prevValue := os.Getenv(key)
	os.Setenv(key, newValue)
	reverter := func() {
		os.Setenv(key, prevValue)
	}
	return reverter
}

func NewTestMeta(outStream, errStream io.Writer, inStream io.Reader) *Meta {

	return &Meta{
		Ui: &cli.BasicUi{
			Writer:      outStream,
			ErrorWriter: errStream,
			Reader:      inStream,
		}}
}

func TestWatchCommand__dummy(t *testing.T) {

	outStream, errStream, inStream := new(bytes.Buffer), new(bytes.Buffer), strings.NewReader("")
	meta := NewTestMeta(outStream, errStream, inStream)
	command := &WatchCommand{
		Meta: *meta,
	}

	args := []string{"--src", "/path/to/src", "--dest", "/path/to/dest"}
	exitStatus := command.Run(args)

	if DebugMode {
		t.Log(outStream.String())
		t.Log(errStream.String())
	}

	if ExitCode(exitStatus) != ExitCodeOK {
		t.Fatalf("ExitStatus is %s, but want %s", ExitCode(exitStatus), ExitCodeOK)
	}
}
