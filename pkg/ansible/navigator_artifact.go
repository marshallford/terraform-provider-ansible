package ansible

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PlaybookArtifact struct {
	Status string   `json:"status"`
	Stdout []string `json:"stdout"`
}

func getPlaybookArtifact(dir string) (*PlaybookArtifact, error) {
	path := filepath.Join(dir, playbookArtifactFilename)

	contents, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook artifact, %w", err)
	}

	var artifact PlaybookArtifact
	if err = json.Unmarshal(contents, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse playbook artifact, %w", err)
	}

	return &artifact, nil
}

func GetStatusFromPlaybookArtifact(dir string) (string, error) {
	artifact, err := getPlaybookArtifact(dir)
	if err != nil {
		return "", err
	}

	return artifact.Status, nil
}

func GetStdoutFromPlaybookArtifact(dir string) (string, error) {
	artifact, err := getPlaybookArtifact(dir)
	if err != nil {
		return "", err
	}

	return strings.Join(artifact.Stdout, "\n"), nil
}
