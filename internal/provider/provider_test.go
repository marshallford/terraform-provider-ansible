package provider_test

import (
	"fmt"
	"os"
	"os/exec"
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

func testAccLookPath(t *testing.T, file string) string {
	t.Helper()

	path, err := exec.LookPath(file)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath(programPath); err != nil {
		t.Fatalf("%s program not installed via Makefile", filepath.Base(programPath))
	}

	testAccLookPath(t, "docker")
}

func testAccAbsProgramPath(t *testing.T) string {
	t.Helper()

	absProgramPath, err := filepath.Abs(programPath)
	if err != nil {
		t.Fatal(err)
	}

	return absProgramPath
}

func testAccPrependProgramToPath(t *testing.T) {
	t.Helper()

	t.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Dir(testAccAbsProgramPath(t)), os.Getenv("PATH")))
}

func testAccFixture(t *testing.T, name string, format ...any) string {
	t.Helper()

	fixture, err := os.ReadFile(fmt.Sprintf("test-fixtures/%s.tf", name))
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(string(fixture), format...)
}
