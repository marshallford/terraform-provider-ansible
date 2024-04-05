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
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/marshallford/terraform-provider-ansible/internal/provider"
	gossh "golang.org/x/crypto/ssh"
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

func testAccResource(t *testing.T, name string, format ...any) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("%s.tf", name)))
	if err != nil {
		t.Fatal(err)
	}

	return fmt.Sprintf(string(data), format...)
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
