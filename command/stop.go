package command

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type StopCommand struct {
	Meta
}

func (c *StopCommand) Run(args []string) int {

	var (
		pid string
	)

	flags := flag.NewFlagSet("stop", flag.ContinueOnError)
	flags.Usage = func() {
		c.Ui.Error(c.Help())
	}

	flags.StringVar(&pid, "pid", DefaultPid, "")
	flags.StringVar(&pid, "p", DefaultPid, "")

	if err := flags.Parse(args); err != nil {
		return int(ExitCodeParseFlagsError)
	}

	if len(pid) == 0 {
		pid = DefaultPid
	}

	// process

	pidFileInfo, err := os.Stat(pid)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to get info of pid file. cause: %s", err))
		return int(ExitCodeError)
	}

	if pidFileInfo == nil {
		return int(ExitCodeOK)
	}

	err = os.Remove(pid)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to remove pid file. cause: %s", err))
		return int(ExitCodeError)
	}

	return int(ExitCodeOK)
}

func (c *StopCommand) Synopsis() string {
	return "Stop agent(s) to synchronize two local directories"
}

func (c *StopCommand) Help() string {
	helpText := `usage: lsync stop [options...]

Options:
  --pid, -p   Path to process id file for the agent.
`
	return strings.TrimSpace(helpText)
}
