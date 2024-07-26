package ansible

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"k8s.io/client-go/util/jsonpath"
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

func jsonPathParse(expression string) (*jsonpath.JSONPath, error) {
	jsonPath := jsonpath.New(expression)
	jsonPath.AllowMissingKeys(true)

	err := jsonPath.Parse(fmt.Sprintf("{%s}", expression))
	if err != nil {
		return nil, err
	}

	return jsonPath, nil
}

func jsonPath(data []byte, expression string) (string, error) {
	var blob interface{}
	if err := json.Unmarshal(data, &blob); err != nil {
		return "", err
	}

	jsonPath, err := jsonPathParse(expression)
	if err != nil {
		return "", err
	}

	output := new(bytes.Buffer)
	if err := jsonPath.Execute(output, blob); err != nil {
		return "", err
	}

	return output.String(), nil
}
