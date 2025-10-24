package ansible

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PlaybookArtifact struct {
	Status string   `json:"status"`
	Stdout []string `json:"stdout"`
}

func getPlaybookArtifact(runDir *RunDir) (*PlaybookArtifact, error) {
	path := runDir.HostJoin(playbookArtifactFilename)

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

func GetStatusFromPlaybookArtifact(runDir *RunDir) (string, error) {
	artifact, err := getPlaybookArtifact(runDir)
	if err != nil {
		return "", err
	}

	return artifact.Status, nil
}

func GetStdoutFromPlaybookArtifact(runDir *RunDir) (string, error) {
	artifact, err := getPlaybookArtifact(runDir)
	if err != nil {
		return "", err
	}

	return strings.Join(artifact.Stdout, "\n"), nil
}
