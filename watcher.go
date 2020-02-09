package main

import (
	"path"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/syncthing/notify"
)

type watcher struct {
	ignore *ignore.GitIgnore
	notify chan notify.EventInfo
}

func newWatcher(pwd string, patterns []string) (*watcher, error) {
	i, err := ignore.CompileIgnoreLines(patterns...)
	if err != nil {
		return nil, err
	}

	w := &watcher{
		notify: make(chan notify.EventInfo, 1),
		ignore: i,
	}

	err = notify.WatchWithFilter(path.Join(pwd, "..."), w.notify, i.MatchesPath, notify.All)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *watcher) watch() chan notify.EventInfo {
	return w.notify
}
