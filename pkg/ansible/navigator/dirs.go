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

// runDirs records the run directory as seen by each process that consumes a
// path from it: this program (and the container engine it may invoke), the
// ansible-navigator process, and the ansible-playbook process.
type runDirs struct {
	host      runDir
	navigator runDir
	playbook  runDir
}

func newRunDirs(mode Mode, hostDir string) runDirs {
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
