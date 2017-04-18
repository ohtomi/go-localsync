package command

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"time"
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

func TestWatchCommand__no_recursive(t *testing.T) {

	go func() {
		outStream, errStream, inStream := new(bytes.Buffer), new(bytes.Buffer), strings.NewReader("")
		meta := NewTestMeta(outStream, errStream, inStream)
		command := &WatchCommand{
			Meta: *meta,
		}

		args := []string{"--src", "../testdata/src", "--dest", "../testdata/dest"}
		exitStatus := command.Run(args)

		if DebugMode {
			t.Log(outStream.String())
			t.Log(errStream.String())
		}

		if ExitCode(exitStatus) != ExitCodeOK {
			t.Fatalf("ExitStatus is %s, but want %s", ExitCode(exitStatus), ExitCodeOK)
		}
	}()

	creator := func(p string) {
		fd, err := os.Create(p)
		if err != nil {
			t.Fatalf("err must be nil, but %+v", err)
		}
		defer fd.Close()
	}

	checker := func(p string) {
		time.Sleep(10 * time.Millisecond)

		_, err := os.Stat(p)
		if err != nil {
			t.Fatalf("missing %q", p)
		}
	}

	creator("../testdata/src/foo")
	checker("../testdata/dest/foo")
}
