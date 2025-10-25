package ansible

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	jq "github.com/itchyny/gojq"
)

func programExistsOnPath(program string) error {
	if _, err := exec.LookPath(program); err != nil {
		return fmt.Errorf("failed to find program on path, %w", err)
	}

	return nil
}

func writeFile(path string, contents string) error {
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil { //nolint:mnd
		return fmt.Errorf("failed to write file, %w", err)
	}

	return nil
}

func jqJSON(data []byte, filter string, raw bool) ([]string, error) {
	var blob any
	if err := json.Unmarshal(data, &blob); err != nil {
		return nil, fmt.Errorf("failed to parse JSON, %w", err)
	}

	query, err := jq.Parse(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JQ filter, %w", err)
	}

	var results []string

	iter := query.Run(blob)
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

		if raw {
			if s, ok := value.(string); ok {
				results = append(results, s)
				continue
			}
		}

		result, err := jq.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert JQ result into JSON, %w", err)
		}

		results = append(results, string(result))
	}

	return results, nil
}
