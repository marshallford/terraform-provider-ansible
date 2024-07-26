package provider_test

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gliderlabs/ssh"
	"github.com/hashicorp/terraform-plugin-testing/config"
	gossh "golang.org/x/crypto/ssh"
)

const (
	navigatorProgramPath = "../../.venv/bin/ansible-navigator" // TODO improve
)

func testAccDefaultConfigVariables(t *testing.T) config.Variables {
	t.Helper()

	return config.Variables{
		"base_run_directory":       config.StringVariable(t.TempDir()),
		"ansible_navigator_binary": config.StringVariable(navigatorProgramPath),
	}
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

	if _, err := exec.LookPath(navigatorProgramPath); err != nil {
		t.Fatalf("%s program not installed via Makefile", filepath.Base(navigatorProgramPath))
	}

	testAccLookPath(t, "docker")
}

func testAccAbs(t *testing.T, programPath string) string {
	t.Helper()

	absPath, err := filepath.Abs(programPath)
	if err != nil {
		t.Fatal(err)
	}

	return absPath
}

func testAccPrependProgramsToPath(t *testing.T) {
	t.Helper()

	t.Setenv("PATH", fmt.Sprintf("%s%c%s", filepath.Dir(testAccAbs(t, navigatorProgramPath)), os.PathListSeparator, os.Getenv("PATH")))
}

func testAccFile(t *testing.T, name string) string {
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

func sshKeygen(t *testing.T) (string, string) {
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

func sshServer(t *testing.T, publicKey string) int {
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
			allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
			if err != nil {
				t.Fatal(err)
			}

			return ssh.KeysEqual(key, allowed)
		}),
	)
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
