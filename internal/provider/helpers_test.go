package provider_test

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	gossh "golang.org/x/crypto/ssh"
)

const (
	// TODO improve
	navigatorProgramPath = "../../.venv/bin/ansible-navigator"
	playbookProgramPath  = "../../.venv/bin/ansible-playbook"
)

var ErrTestCheckFunc = errors.New("test check func")

func testDefaultConfigVariables(t *testing.T) config.Variables {
	t.Helper()

	return config.Variables{
		"base_run_directory":       config.StringVariable(t.TempDir()),
		"ansible_navigator_binary": config.StringVariable(navigatorProgramPath),
	}
}

func testConfigVariables(t *testing.T, overrides config.Variables) config.Variables {
	t.Helper()

	variables := testDefaultConfigVariables(t)
	maps.Copy(variables, overrides)

	return variables
}

func testLookPath(t *testing.T, file string) string {
	t.Helper()

	path, err := exec.LookPath(file)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func testPreCheck(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath(navigatorProgramPath); err != nil {
		t.Fatalf("%s program not installed via Makefile", filepath.Base(navigatorProgramPath))
	}

	testLookPath(t, "docker")
}

func testAbsPath(t *testing.T, programPath string) string {
	t.Helper()

	absPath, err := filepath.Abs(programPath)
	if err != nil {
		t.Fatal(err)
	}

	return absPath
}

func testPrependNavigatorToPath(t *testing.T) {
	t.Helper()

	t.Setenv("PATH", fmt.Sprintf("%s%c%s", filepath.Dir(testAbsPath(t, navigatorProgramPath)), os.PathListSeparator, os.Getenv("PATH")))
}

func testPrependPlaybookToPath(t *testing.T) {
	t.Helper()

	t.Setenv("PATH", fmt.Sprintf("%s%c%s", filepath.Dir(testAbsPath(t, playbookProgramPath)), os.PathListSeparator, os.Getenv("PATH")))
}

func testTerraformFile(t *testing.T, name string) string {
	t.Helper()

	providerData, err := os.ReadFile(filepath.Join("testdata", "provider.tf"))
	if err != nil {
		t.Fatal(err)
	}

	fileData, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("%s.tf", name)))
	if err != nil {
		t.Fatal(err)
	}

	return string(fileData) + string(providerData)
}

func testSSHKeygen(t *testing.T) (string, string) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := gossh.MarshalPrivateKey(crypto.PrivateKey(priv), "")
	if err != nil {
		t.Fatal(err)
	}

	publicKey, err := gossh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf("ssh-ed25519 %s", base64.StdEncoding.EncodeToString(publicKey.Marshal())), string(pem.EncodeToMemory(privateKey))
}

func testSSHServer(t *testing.T, clientPublicKey string, serverPrivateKey string) int {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	sshServer := ssh.Server{
		Handler: func(s ssh.Session) {
			_, err = s.Write([]byte("hello world!"))
			if err != nil {
				t.Fatal(err)
			}
		},
	}

	err = sshServer.SetOption(
		ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(clientPublicKey))
			if err != nil {
				t.Fatal(err)
			}

			return ssh.KeysEqual(key, allowed)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = sshServer.SetOption(ssh.HostKeyPEM([]byte(serverPrivateKey)))
	if err != nil {
		t.Fatal(err)
	}

	// TODO wait until ready?
	go sshServer.Serve(listener) //nolint:errcheck

	t.Cleanup(func() {
		if err := sshServer.Close(); err != nil {
			t.Fatal(err)
		}
	})

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal()
	}

	return addr.Port
}

// https://github.com/hashicorp/terraform-provider-random/blob/main/internal/provider/resource_integer_test.go
func testExtractResourceAttr(resourceName string, attributeName string, attributeValue *string) resource.TestCheckFunc { //nolint:unparam
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%w, resource name %s not found in state", ErrTestCheckFunc, resourceName)
		}

		attrValue, ok := resourceState.Primary.Attributes[attributeName]

		if !ok {
			return fmt.Errorf("%w, attribute %s not found in resource %s state", ErrTestCheckFunc, attributeName, resourceName)
		}

		*attributeValue = attrValue

		return nil
	}
}

// https://github.com/hashicorp/terraform-provider-random/blob/main/internal/provider/resource_integer_test.go
func testCheckAttributeValuesDiffer(i *string, j *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if testStringValue(i) == testStringValue(j) {
			return fmt.Errorf("%w, attribute values are the same, got %s", ErrTestCheckFunc, testStringValue(i))
		}

		return nil
	}
}

// https://github.com/hashicorp/terraform-provider-random/blob/main/internal/provider/resource_integer_test.go
func testCheckAttributeValuesEqual(i *string, j *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if testStringValue(i) != testStringValue(j) {
			return fmt.Errorf("%w, attribute values are different, got %s and %s", ErrTestCheckFunc, testStringValue(i), testStringValue(j))
		}

		return nil
	}
}

// https://github.com/hashicorp/terraform-provider-random/blob/main/internal/provider/resource_integer_test.go
func testStringValue(sPtr *string) string {
	if sPtr == nil {
		return ""
	}

	return *sPtr
}
