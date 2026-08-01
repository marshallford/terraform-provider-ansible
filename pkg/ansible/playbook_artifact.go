package ansible

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	jq "github.com/itchyny/gojq"
)

type PlaybookStdout []string

func (s PlaybookStdout) String() string {
	return strings.Join(s, "\n")
}

type PlaybookArtifact struct {
	Status Status
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
		Status: ParseStatus(format.Status),
		Stdout: PlaybookStdout(format.Stdout),
	}, nil
}

func QueryPlaybookArtifact(data []byte, query PlaybookArtifactQuery) ([]string, error) {
	var blob any
	if err := json.Unmarshal(data, &blob); err != nil {
		return nil, fmt.Errorf("failed to parse JSON, %w", err)
	}

	parsed, err := jq.Parse(query.JQFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JQ filter, %w", err)
	}

	var results []string

	iter := parsed.Run(blob)
	for {
		value, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := value.(error); ok {
			var haltErr *jq.HaltError
			if errors.As(err, &haltErr) && haltErr.Value() == nil {
				break
			}

			return nil, fmt.Errorf("JQ failed, %w", err)
		}

		result, err := jqValueString(value, query.Raw)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func jqValueString(value any, raw bool) (string, error) {
	if raw {
		if s, ok := value.(string); ok {
			return s, nil
		}
	}

	result, err := jq.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to convert JQ result into JSON, %w", err)
	}

	return string(result), nil
}
