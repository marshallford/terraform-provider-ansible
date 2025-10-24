package ansible

import (
	"fmt"
)

const (
	inventoriesDir = "inventories"
)

type Inventory struct {
	Name     string
	Contents string
}

func ResolvedInventoryPaths(runDir *RunDir, invs []Inventory) map[string]string {
	paths := make(map[string]string, len(invs))

	for _, i := range invs {
		paths[i.Name] = runDir.ResolvedJoin(inventoriesDir, i.Name)
	}

	return paths
}

func CreateInventories(runDir *RunDir, inventories []Inventory) error {
	for _, inventory := range inventories {
		err := writeFile(runDir.HostJoin(inventoriesDir, inventory.Name), inventory.Contents)
		if err != nil {
			return fmt.Errorf("failed to create ansible inventory file for run, %w", err)
		}
	}

	return nil
}
