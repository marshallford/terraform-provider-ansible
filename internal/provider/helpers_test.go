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
	gossh "golang.org/x/crypto/ssh"
)

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

// avoid this issue: https://github.com/hashicorp/terraform-plugin-testing/issues/277
func testAccAbs(t *testing.T, programPath string) string { //nolint:unparam
	t.Helper()

	absPath, err := filepath.Abs(programPath)
	if err != nil {
		t.Fatal(err)
	}

	return absPath
}

func testAccPrependProgramsToPath(t *testing.T) {
	t.Helper()

	t.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Dir(testAccAbs(t, navigatorProgramPath)), os.Getenv("PATH")))
}

func testAccResource(t *testing.T, name string, format ...any) string {
	t.Helper()

	baseRunDirectory := t.TempDir()

	providerData, err := os.ReadFile(filepath.Join("testdata", "provider.tf"))
	if err != nil {
		t.Fatal(err)
	}

	resourceData, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("%s.tf", name)))
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(string(providerData), baseRunDirectory) + fmt.Sprintf(string(resourceData), format...)
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
