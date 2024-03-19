package ansible

import (
	"fmt"
	"path/filepath"
)

const navigatorSSHDir = "/tmp/ssh"

type SSHPrivateKey struct {
	Name string
	Data string
}

func CreateSSHPrivateKeys(dir string, keys []SSHPrivateKey, settings *NavigatorSettings, opts *RunOptions) error {
	for _, key := range keys {
		srcPath := filepath.Join(dir, key.Name)
		destPath := fmt.Sprintf("%s/%s", navigatorSSHDir, key.Name)

		err := writeFile(srcPath, key.Data)
		if err != nil {
			return fmt.Errorf("failed to create SSH private key file for run, %w", err)
		}

		// TODO better option?
		if settings.VolumeMounts == nil {
			settings.VolumeMounts = map[string]string{}
		}

		settings.VolumeMounts[srcPath] = destPath
		opts.PrivateKey = append(opts.PrivateKey, destPath)
	}

	return nil
}
