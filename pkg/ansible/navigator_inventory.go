package ansible

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	inventoriesDir = "inventories"
	// TODO assumes EE is unix-like with a /tmp dir
	eeInventoriesDir = "/tmp/inventories"
)

type Inventory struct {
	Name     string
	Contents string
	Exclude  bool
}

func InventoryPath(dir string, name string, eeEnabled bool, excluded bool) string {
	if eeEnabled && excluded {
		return strings.Join([]string{eeInventoriesDir, name}, "/") // assume EE is unix-like
	}

	return filepath.Join(dir, inventoriesDir, name)
}

func CreateInventories(dir string, inventories []Inventory, settings *NavigatorSettings) error {
	var inventoryExcluded bool
	for _, inventory := range inventories {
		if inventory.Exclude {
			inventoryExcluded = true
		}
		err := writeFile(filepath.Join(dir, inventoriesDir, inventory.Name), inventory.Contents)
		if err != nil {
			return fmt.Errorf("failed to create ansible inventory file for run, %w", err)
		}
	}

	if !settings.EEEnabled || !inventoryExcluded {
		return nil
	}

	// TODO better option?
	if settings.VolumeMounts == nil {
		settings.VolumeMounts = map[string]string{}
	}

	settings.VolumeMounts[filepath.Join(dir, inventoriesDir)] = eeInventoriesDir

	return nil
}
