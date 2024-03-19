package ansible

import (
	"fmt"
	"os"
	"path/filepath"
)

type ArtifactQuery struct {
	JSONPath string
	Result   string
}

func QueryPlaybookArtifact(dir string, queries map[string]ArtifactQuery) error {
	path := filepath.Join(dir, playbookArtifactFilename)

	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	for name, query := range queries {
		result, err := jsonPath(contents, query.JSONPath)
		if err != nil {
			return fmt.Errorf("failed to query playbook artifact with JSONPath, %w", err)
		}

		query.Result = result
		queries[name] = query
	}

	return nil
}
