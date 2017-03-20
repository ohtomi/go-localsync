package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
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

	interrupted := c.chanToTrapCtrlC()

	if err := c.startWatcher(src, dest, recursive, verbose); err != nil {
		c.Ui.Error(fmt.Sprintf("failed to start watcher. cause: %q", err))
		return int(ExitCodeError)
	}

	<-interrupted

	return int(ExitCodeOK)
}

func (c *WatchCommand) chanToTrapCtrlC() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	return ch
}

func (c *WatchCommand) startWatcher(src, dest string, recursive, verbose bool) error {

	toAbs := func(rawpath string) string {
		resolved, err := filepath.EvalSymlinks(rawpath)
		if err != nil {
			// TODO
		}
		abs, err := filepath.Abs(resolved)
		if err != nil {
			// TODO
		}
		return abs
	}

	toRel := func(basepath, targetpath string) string {
		rel, err := filepath.Rel(toAbs(basepath), toAbs(targetpath))
		if err != nil {
			// TODO
		}
		return rel
	}

	createDir := func(dir string) error {
		// TODO add watcher
		// TODO copy files if exists (for rename event)
		return os.MkdirAll(dir, os.ModeDir)
	}

	// deleteDir := func(dir string) error {
	// 	// TODO delete watcher
	// 	return os.RemoveAll(dir)
	// }

	copyFile := func(srcfile string) error {
		srcfd, err := os.Open(srcfile)
		if err != nil {
			return err
		}
		defer srcfd.Close()

		destfile := path.Join(dest, toRel(src, srcfile))
		destfd, err := os.Create(destfile)
		if err != nil {
			return err
		}
		defer destfd.Close()

		_, err = io.Copy(destfd, srcfd)
		if err != nil {
			return err
		}

		return nil
	}

	// deleteFile := func(file string) error {
	// 	return os.Remove(file)
	// }

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	watch := func(root string) error {
		err := watcher.Add(root)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					switch {
					case event.Op&fsnotify.Create == fsnotify.Create:
						c.Ui.Output("create " + event.Name)
					case event.Op&fsnotify.Write == fsnotify.Write:
						c.Ui.Output("write " + event.Name)
					case event.Op&fsnotify.Remove == fsnotify.Remove:
						c.Ui.Output("remove " + event.Name)
					case event.Op&fsnotify.Rename == fsnotify.Rename:
						// TODO
					case event.Op&fsnotify.Chmod == fsnotify.Chmod:
						// TODO
					}
					// case err := <-watcher.Errors:
					// TODO
				}
			}
		}()

		return nil
	}

	walk := func(rawpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if toAbs(rawpath) == toAbs(src) {
				return watch(rawpath)
			}
			if !recursive {
				return filepath.SkipDir
			}
			if err := createDir(rawpath); err != nil {
				return err
			}
			if err := watch(rawpath); err != nil {
				return err
			}
			return nil
		} else {
			return copyFile(rawpath)
		}
	}

	return filepath.Walk(src, walk)
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
