package command

import (
	"flag"
	"fmt"
	"strings"
)

type WatchCommand struct {
	Meta
}

func (c *WatchCommand) Run(args []string) int {

	var (
		src       string
		dest      string
		recursive bool
		verbose   bool
	)

	flags := flag.NewFlagSet("start", flag.ContinueOnError)
	flags.Usage = func() {
		c.Ui.Error(c.Help())
	}

	flags.StringVar(&src, "src", "", "")
	flags.StringVar(&src, "s", "", "")
	flags.StringVar(&dest, "dest", "", "")
	flags.StringVar(&dest, "d", "", "")
	flags.BoolVar(&recursive, "recursive", false, "")
	flags.BoolVar(&recursive, "r", false, "")
	flags.BoolVar(&verbose, "verbose", false, "")

	if err := flags.Parse(args); err != nil {
		return int(ExitCodeParseFlagsError)
	}

	if len(src) == 0 {
		c.Ui.Error("missing SRC.")
		return int(ExitCodeBadArgs)
	}

	if len(dest) == 0 {
		c.Ui.Error("missing DEST.")
		return int(ExitCodeBadArgs)
	}

	// process

	c.Ui.Output(fmt.Sprintf(`starting a watch agent...
        src: %s
       dest: %s
  recursive: %v
    verbose: %v

press Ctrl+C to stop the watch agent.
`, src, dest, recursive, verbose))

	return int(ExitCodeOK)
}

func (c *WatchCommand) Synopsis() string {
	return "Watch file system events of the specified directory"
}

func (c *WatchCommand) Help() string {
	helpText := `usage: lsync watch [options...]

Options:
  --src, -s        Path to SRC directory.
  --dest, -d       Path to DEST directory.
  --recursive, -r  Watch recursively under SRC.
  --verbose        Report file system event verbosely.
`
	return strings.TrimSpace(helpText)
}
