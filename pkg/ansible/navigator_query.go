package ansible

import (
	"fmt"
	"os"
	"path/filepath"

	jq "github.com/itchyny/gojq"
)

type ArtifactQuery struct {
	JQFilter string
	Results  []string
}

func QueryPlaybookArtifact(dir string, queries map[string]ArtifactQuery) error {
	path := filepath.Join(dir, playbookArtifactFilename)

	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	for name, query := range queries {
		results, err := jqJSON(contents, query.JQFilter)
		if err != nil {
			return fmt.Errorf("failed to query playbook artifact, %w", err)
		}

		query.Results = results
		queries[name] = query
	}

	return nil
}

func ValidateJQFilter(filter string) error {
	if _, err := jq.Parse(filter); err != nil {
		return fmt.Errorf("failed to parse JQ filter, %w", err)
	}

	return nil
}
