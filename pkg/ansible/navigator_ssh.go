package ansible

import (
	"fmt"
	"path/filepath"
)

const navigatorPrivateKeysDir = "/tmp/private-keys"

type PrivateKey struct {
	Name string
	Data string
}

func CreatePrivateKeys(dir string, keys []PrivateKey, settings *NavigatorSettings) error {
	for _, key := range keys {
		err := writeFile(filepath.Join(dir, key.Name), key.Data)
		if err != nil {
			return fmt.Errorf("failed to create private key file for run, %w", err)
		}
	}

	// TODO better option?
	if settings.VolumeMounts == nil {
		settings.VolumeMounts = map[string]string{}
	}

	settings.VolumeMounts[dir] = navigatorPrivateKeysDir

	return nil
}
