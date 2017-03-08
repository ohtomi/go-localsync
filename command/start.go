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

func (c *StartCommand) Run(args []string) int {

	var (
		src  string
		dest string

		pid string
	)

	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	flags.Usage = func() {
		c.Ui.Error(c.Help())
	}

	flags.StringVar(&pid, "pid", DefaultPid, "")
	flags.StringVar(&pid, "p", DefaultPid, "")

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

	c.Ui.Output(fmt.Sprintf("%s, %s, %s", src, dest, pid))

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
					c.Ui.Output(fmt.Sprintf("%s", event.Name))
				case event.Op&fsnotify.Write == fsnotify.Write:
					c.Ui.Output(fmt.Sprintf("%s", event.Name))
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					c.Ui.Output(fmt.Sprintf("%s", event.Name))
					if event.Name == pidFileInfo.Name() {
						done <- true
					}
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					c.Ui.Output(fmt.Sprintf("%s", event.Name))
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					c.Ui.Output(fmt.Sprintf("%s", event.Name))
				}
			case err := <-watcher.Errors:
				c.Ui.Output(fmt.Sprintf("error occurred. %s", err))
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

	<-done

	return int(ExitCodeOK)
}

func (c *StartCommand) Synopsis() string {
	return "Start agent to synchronize two local directories"
}

func (c *StartCommand) Help() string {
	helpText := `usage: lsync start [options...] SRC DEST

Options:
	--pid, -p   Path to process id file for the agent.
`
	return strings.TrimSpace(helpText)
}
