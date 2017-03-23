package command

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type WatchAgent struct {
	src       string
	dest      string
	recursive bool
	verbose   bool
	watcher   *fsnotify.Watcher
	channels  map[string]chan interface{}

	meta Meta
}

func NewWatchAgent(src, dest string, recursive, verbose bool, meta Meta) (*WatchAgent, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	srcdir, err := filepath.EvalSymlinks(src)
	if err != nil {
		return nil, err
	}

	destdir, err := filepath.EvalSymlinks(dest)
	if err != nil {
		return nil, err
	}

	return &WatchAgent{srcdir, destdir, recursive, verbose, watcher, map[string]chan interface{}{}, meta}, nil
}

func (w *WatchAgent) Close() error {
	return w.watcher.Close()
}

func (w *WatchAgent) Start() error {
	// TODO traverse dest dir to remove difference between src and dest.

	walker := func(rawpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if w.toAbs(rawpath) == w.toAbs(w.src) {
				if ch, err := w.watch(rawpath); err != nil {
					return err
				} else {
					w.channels[rawpath] = ch
				}
				return nil
			}

			if !w.recursive {
				return filepath.SkipDir
			}

			if err := w.createDir(rawpath); err != nil {
				return err
			}
			if ch, err := w.watch(rawpath); err != nil {
				return err
			} else {
				w.channels[rawpath] = ch
			}
			return nil

		} else {
			return w.copyFile(rawpath)
		}
	}

	return filepath.Walk(w.src, walker)
}

func (w *WatchAgent) Stop() error {
	// TODO
	return nil
}

//

func (w *WatchAgent) watch(root string) (chan interface{}, error) {
	if err := w.watcher.Add(root); err != nil {
		return nil, err
	}

	ch := make(chan interface{})
	go func() {
		for {
			select {
			case event := <-w.watcher.Events:
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					w.handleCreateEvent(event)
				case event.Op&fsnotify.Write == fsnotify.Write:
					w.handleWriteEvent(event)
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					w.handleRemoveEvent(event)
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					w.handleRenameEvent(event)
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					w.handleChmodEvent(event)
				}
			case err := <-w.watcher.Errors:
				w.handleErrorEvent(err)
			case <-ch:
				return
			}
		}
	}()

	return ch, nil
}

func (w *WatchAgent) unwatch(root string) error {
	done, ok := w.channels[root]
	if ok {
		done <- true
		delete(w.channels, root)
	}

	return nil
}

func (w *WatchAgent) handleCreateEvent(event fsnotify.Event) {
	w.meta.Ui.Output(fmt.Sprintf("create %s", event.Name))
	info, err := os.Stat(event.Name)
	if err != nil {
		w.meta.Ui.Error(fmt.Sprintf("error %s", err))
		return
	}
	if info.IsDir() {
		if err := w.createDir(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
		if ch, err := w.watch(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		} else {
			w.channels[event.Name] = ch
		}
		// TODO copy files under the new directory
	} else {
		if err := w.copyFile(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
	}
}

func (w *WatchAgent) handleWriteEvent(event fsnotify.Event) {
	w.meta.Ui.Output(fmt.Sprintf("write %s", event.Name))
	if err := w.copyFile(event.Name); err != nil {
		w.meta.Ui.Error(fmt.Sprintf("error %s", err))
		return
	}
}

func (w *WatchAgent) handleRemoveEvent(event fsnotify.Event) {
	w.meta.Ui.Output(fmt.Sprintf("remove %s", event.Name))
	destpath := path.Join(w.dest, w.toRel(w.src, event.Name))
	info, err := os.Stat(destpath)
	if err != nil {
		return
	}
	if info.IsDir() {
		if err := w.deleteDir(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
		if err := w.unwatch(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
	} else {
		if err := w.deleteFile(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
	}
}

func (w *WatchAgent) handleRenameEvent(event fsnotify.Event) {
	w.meta.Ui.Output(fmt.Sprintf("rename %s", event.Name))
	destpath := path.Join(w.dest, w.toRel(w.src, event.Name))
	info, err := os.Stat(destpath)
	if err != nil {
		return
	}
	if info.IsDir() {
		if err := w.deleteDir(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
		if err := w.unwatch(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
	} else {
		if err := w.deleteFile(event.Name); err != nil {
			w.meta.Ui.Error(fmt.Sprintf("error %s", err))
			return
		}
	}
}

func (w *WatchAgent) handleChmodEvent(event fsnotify.Event) {}

func (w *WatchAgent) handleErrorEvent(err error) {
	if err != nil {
		w.meta.Ui.Error(fmt.Sprintf("error %s", err))
	}
}

//

func (w *WatchAgent) createDir(srcdir string) error {
	srcinfo, err := os.Stat(srcdir)
	if err != nil {
		return err
	}
	destdir := path.Join(w.dest, w.toRel(w.src, srcdir))
	return os.MkdirAll(destdir, srcinfo.Mode())
}

func (w *WatchAgent) copyFile(srcfile string) error {
	srcfd, err := os.Open(srcfile)
	if err != nil {
		return err
	}
	defer srcfd.Close()

	destfile := path.Join(w.dest, w.toRel(w.src, srcfile))
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

func (w *WatchAgent) deleteDir(srcdir string) error {
	destdir := path.Join(w.dest, w.toRel(w.src, srcdir))
	return os.RemoveAll(destdir)
}

func (w *WatchAgent) deleteFile(srcfile string) error {
	destfile := path.Join(w.dest, w.toRel(w.src, srcfile))
	return os.Remove(destfile)
}

//

func (w *WatchAgent) toAbs(rawpath string) string {
	abs, err := filepath.Abs(rawpath)
	if err != nil {
		panic(fmt.Sprintf("unexpected error in toAbs. detail: %q", err))
	}
	return abs
}

func (w *WatchAgent) toRel(basepath, targetpath string) string {
	rel, err := filepath.Rel(w.toAbs(basepath), w.toAbs(targetpath))
	if err != nil {
		panic(fmt.Sprintf("unexpected error in toRel. detail: %q", err))
	}
	return rel
}
