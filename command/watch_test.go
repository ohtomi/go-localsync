package command

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"bufio"
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

func AssertFileExists(p string, t *testing.T) {

	_, err := os.Stat(p)
	if err != nil {
		t.Fatalf("not found %q", p)
	}
}

func AssertFileNotExists(p string, t *testing.T) {

	_, err := os.Stat(p)
	if err == nil {
		t.Fatalf("found %q", p)
	}
}

func AssertFileContent(p, expected string, t *testing.T) {

	fd, err := os.Open(p)
	if err != nil {
		t.Fatalf("failed to open %q", p)
	}
	defer fd.Close()

	actual := ""

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		actual += scanner.Text()
	}
	if scanner.Err() != nil {
		t.Fatalf("failed to scan %q", p)
	}

	if actual != expected {
		t.Fatalf("content is %q, but want %q", actual, expected)
	}
}

func CreateFileUnderTestDataDir(p string, t *testing.T) {
	fd, err := os.Create(p)
	if err != nil {
		t.Fatalf("err must be nil, but %+v", err)
	}
	defer fd.Close()

	time.Sleep(10 * time.Millisecond)
}

func WriteFileUnderTestDataDir(p, content string, t *testing.T) {
	fd, err := os.OpenFile(p, os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("err must be nil, but %+v", err)
	}
	defer fd.Close()

	fd.WriteString(content)
	err = fd.Sync()
	if err != nil {
		t.Fatalf("err must be nil, but %+v", err)
	}

	time.Sleep(10 * time.Millisecond)
}

func DeleteFileUnderTestDataDir(p string, t *testing.T) {
	err := os.Remove(p)
	if err != nil {
		t.Fatalf("err must be nil, but %+v", err)
	}

	time.Sleep(10 * time.Millisecond)
}

func RenameFileUnderTestDataDir(p, q string, t *testing.T) {
	err := os.Rename(p, q)
	if err != nil {
		t.Fatalf("err must be nil, but %+v", err)
	}

	time.Sleep(10 * time.Millisecond)
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

		command.chanToTrapCtrlC()
		if ExitCode(exitStatus) != ExitCodeOK {
			t.Fatalf("ExitStatus is %s, but want %s", ExitCode(exitStatus), ExitCodeOK)
		}
	}()

	CreateFileUnderTestDataDir("../testdata/src/foo", t)
	AssertFileExists("../testdata/dest/foo", t)

	WriteFileUnderTestDataDir("../testdata/src/foo", "dummy", t)
	AssertFileContent("../testdata/dest/foo", "dummy", t)

	RenameFileUnderTestDataDir("../testdata/src/foo", "../testdata/src/bar", t)
	AssertFileExists("../testdata/dest/bar", t)

	DeleteFileUnderTestDataDir("../testdata/src/bar", t)
	AssertFileNotExists("../testdata/dest/bar", t)
}
