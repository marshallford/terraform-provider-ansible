package ansible

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	jq "github.com/itchyny/gojq"
)

func programExistsOnPath(program string) error {
	if _, err := exec.LookPath(program); err != nil {
		return err
	}

	return nil
}

func writeFile(path string, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o600) //nolint:gomnd,mnd
}

func jqJSON(data []byte, filter string) ([]string, error) {
	var blob any
	if err := json.Unmarshal(data, &blob); err != nil {
		return nil, err
	}

	query, err := jq.Parse(filter)
	if err != nil {
		return nil, err
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

			return nil, err
		}

		result, err := jq.Marshal(value)
		if err != nil {
			return nil, err
		}

		results = append(results, string(result))
	}

	return results, nil
}
