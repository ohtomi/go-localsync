package command

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type StartCommand struct {
	Meta
}

func (c *StartCommand) ProcessCreate(store *FileInfoStore, filename string, pid string, verbose bool) bool {

	if verbose && filename != pid {
		c.Ui.Output(fmt.Sprintf("created %s", filename))
	}

	if err := store.Add(filename); err != nil {
		c.Ui.Error(fmt.Sprintf("error occurred. %s", err))
		return true
	}

	// TODO copy it to DEST

	return false
}

func (c *StartCommand) ProcessWrite(store *FileInfoStore, filename string, pid string, verbose bool) bool {

	if verbose && filename != pid {
		c.Ui.Output(fmt.Sprintf("modified %s", filename))
	}

	// TODO copy it to DEST

	return false
}

func (c *StartCommand) ProcessRemove(store *FileInfoStore, filename string, pid string, verbose bool) bool {

	if verbose && filename != pid {
		c.Ui.Output(fmt.Sprintf("removed %s", filename))
	}

	if err := store.Remove(filename); err != nil {
		c.Ui.Error(fmt.Sprintf("error occurred. %s", err))
		return true
	}

	// TODO remove it from DEST

	return filename == pid
}

func (c *StartCommand) ProcessRename(store *FileInfoStore, filename string, pid string, verbose bool) bool {

	if verbose && filename != pid {
		c.Ui.Output(fmt.Sprintf("renamed %s", filename))
	}

	if err := store.Remove(filename); err != nil {
		c.Ui.Error(fmt.Sprintf("error occurred. %s", err))
		return true
	}

	// TODO remove it from DEST

	return false
}

func (c *StartCommand) ProcessChmod(store *FileInfoStore, filename string, pid string, verbose bool) bool {
	return false
}

func (c *StartCommand) ProcessError(err error) bool {

	c.Ui.Error(fmt.Sprintf("error occurred. %s", err))
	return true
}

func (c *StartCommand) Run(args []string) int {

	var (
		src  string
		dest string

		pid       string
		recursive bool
		verbose   bool
	)

	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	flags.Usage = func() {
		c.Ui.Error(c.Help())
	}

	flags.StringVar(&pid, "pid", DefaultPid, "")
	flags.StringVar(&pid, "p", DefaultPid, "")
	flags.BoolVar(&recursive, "recursive", false, "")
	flags.BoolVar(&recursive, "r", false, "")
	flags.BoolVar(&verbose, "verbose", false, "")

	if err := flags.Parse(args); err != nil {
		return int(ExitCodeParseFlagsError)
	}

	parsedArgs := flags.Args()
	if len(parsedArgs) != 2 {
		c.Ui.Error("you must set SRC and DEST.")
		return int(ExitCodeBadArgs)
	}
	src, dest = parsedArgs[0], parsedArgs[1]

	if len(src) == 0 {
		c.Ui.Error("missing SRC.")
		return int(ExitCodeSrcNotFound)
	}
	if len(dest) == 0 {
		c.Ui.Error("missing DEST.")
		return int(ExitCodeDestNotFound)
	}

	if len(pid) == 0 {
		pid = DefaultPid
	}

	// process

	if same, err := FilePath(src).IsSameFilePath(FilePath(dest)); same || err != nil {
		c.Ui.Error("SRC is DEST.")
		return int(ExitCodeBadArgs)
	}

	c.Ui.Output(fmt.Sprintf("%s, %s, %s", src, dest, pid))

	store := &FileInfoStore{src, []*FileInfo{}}
	store.Load(recursive)

	pidFile, err := os.Create(pid)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to create pid file. cause: %s", err))
		return int(ExitCodeError)
	}
	pidFileInfo, err := pidFile.Stat()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to get stat of pid file. cause: %s", err))
		return int(ExitCodeError)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to start lsync agent. cause: %s", err))
		return int(ExitCodeError)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					done <- c.ProcessCreate(store, event.Name, pidFileInfo.Name(), verbose)
				case event.Op&fsnotify.Write == fsnotify.Write:
					done <- c.ProcessWrite(store, event.Name, pidFileInfo.Name(), verbose)
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					done <- c.ProcessRemove(store, event.Name, pidFileInfo.Name(), verbose)
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					done <- c.ProcessRename(store, event.Name, pidFileInfo.Name(), verbose)
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					done <- c.ProcessChmod(store, event.Name, pidFileInfo.Name(), verbose)
				}
			case err := <-watcher.Errors:
				done <- c.ProcessError(err)
			}
		}
	}()

	err = watcher.Add(src)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to start watching SRC. cause: %s", err))
		return int(ExitCodeError)
	}

	err = watcher.Add(pid)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to start watching pid file. cause: %s", err))
		return int(ExitCodeError)
	}

	for {
		result := <-done
		if result {
			return int(ExitCodeOK)
		}
	}
}

func (c *StartCommand) Synopsis() string {
	return "Start agent to synchronize two local directories"
}

func (c *StartCommand) Help() string {
	helpText := `usage: lsync start [options...] SRC DEST

Options:
  --pid, -p        Path to process id file for the agent.
  --recursive, -r  Watch recursively under SRC.
  --verbose        Report watch event.
`
	return strings.TrimSpace(helpText)
}
