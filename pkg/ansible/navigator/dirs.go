package navigator

import (
	"path"
	"path/filepath"
)

type runDir struct {
	root      string
	container bool
}

func (d runDir) join(parts ...string) string {
	parts = append([]string{d.root}, parts...)

	if d.container {
		return path.Join(parts...)
	}

	return filepath.Join(parts...)
}

// runDirs records the run directory as seen by each process that reads a path
// from it. The three can differ, because navigator and the playbook may each
// run inside a container.
type runDirs struct {
	host      runDir
	navigator runDir
	playbook  runDir
}

func newRunDirs(hostDir string, mode Mode) runDirs {
	host := runDir{root: filepath.Clean(hostDir)}

	playbook := host
	if mode.UsesEE() {
		playbook = runDir{root: containerRunDir, container: true}
	}

	return runDirs{
		host:      host,
		navigator: host,
		playbook:  playbook,
	}
}
