package ansible

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PlaybookStdout []string

func (s PlaybookStdout) String() string {
	return strings.Join(s, "\n")
}

type PlaybookArtifact struct {
	Status string
	Stdout PlaybookStdout
}

type PlaybookArtifactQuery struct {
	JQFilter string
	Raw      bool
	Results  []string
}

type playbookArtifactFormat struct {
	Status string   `json:"status"`
	Stdout []string `json:"stdout"`
}

func ParsePlaybookArtifact(data []byte) (*PlaybookArtifact, error) {
	var format playbookArtifactFormat
	if err := json.Unmarshal(data, &format); err != nil {
		return nil, fmt.Errorf("failed to parse playbook artifact, %w", err)
	}

	return &PlaybookArtifact{
		Status: format.Status,
		Stdout: PlaybookStdout(format.Stdout),
	}, nil
}
