package ansible

import (
	"fmt"
	"os"
)

type ArtifactQuery struct {
	JQFilter string
	Results  []string
}

func QueryPlaybookArtifact(runDir *RunDir, queries map[string]ArtifactQuery) error {
	path := runDir.HostJoin(playbookArtifactFilename)

	contents, err := os.ReadFile(path) // #nosec G304
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
