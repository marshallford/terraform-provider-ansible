package provider_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/marshallford/terraform-provider-ansible/internal/provider"
)

const programPath = "../../.venv/bin/ansible-navigator" // TODO improve

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){ //nolint:gochecknoglobals
	"ansible": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(programPath); errors.Is(err, os.ErrNotExist) {
		t.Fatal("ansible-navigator program not installed via Makefile")
	}
}

func testAccPrependProgramToPath(t *testing.T) {
	t.Helper()

	absPath, err := filepath.Abs(programPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Dir(absPath), os.Getenv("PATH")))
}

func testAccNavigatorRunResourceConfigUsePath(t *testing.T, basename string, workingDirectory string) string {
	t.Helper()

	resource, err := os.ReadFile(fmt.Sprintf("test-fixtures/%s.tf", basename))
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(string(resource), workingDirectory)
}

func testAccNavigatorRunResourceConfig(t *testing.T, basename string, workingDirectory string) string {
	t.Helper()

	resource, err := os.ReadFile(fmt.Sprintf("test-fixtures/%s.tf", basename))
	if err != nil {
		t.Fatal(err)
	}

	absProgramPath, err := filepath.Abs(programPath)
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(string(resource), absProgramPath, workingDirectory)
}
