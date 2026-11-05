package build

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RemoteCertBundle holds PEM material for a BYO BuildKit host.
type RemoteCertBundle struct {
	CA         []byte
	ServerCert []byte
	ServerKey  []byte
	ClientCert []byte
	ClientKey  []byte
}

// GenerateRemoteCertBundle creates a small mTLS PKI for one remote builder.
func GenerateRemoteCertBundle(host string) (RemoteCertBundle, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "dockyard-builder"
	}
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return RemoteCertBundle{}, err
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Dockyard Build CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return RemoteCertBundle{}, err
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return RemoteCertBundle{}, err
	}
	serverTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(825 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{host},
		IPAddresses:  []net.IP{net.ParseIP(host)},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caKey)
	if err != nil {
		return RemoteCertBundle{}, err
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return RemoteCertBundle{}, err
	}
	clientTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "dockyard-build-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(825 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTmpl, caTmpl, &clientKey.PublicKey, caKey)
	if err != nil {
		return RemoteCertBundle{}, err
	}

	return RemoteCertBundle{
		CA:         caPEM,
		ServerCert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER}),
		ServerKey:  marshalECKey(serverKey),
		ClientCert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER}),
		ClientKey:  marshalECKey(clientKey),
	}, nil
}

func marshalECKey(key *ecdsa.PrivateKey) []byte {
	b, _ := x509.MarshalECPrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
}

// RemoteBuilderRunLine is the docker run snippet for the remote host.
func RemoteBuilderRunLine(hostPort string) string {
	return fmt.Sprintf(
		"docker run -d --name dockyard-builder --privileged --restart unless-stopped "+
			"-p %s -v /etc/dockyard-builder:/certs:ro -v dockyard-builder-data:/var/lib/buildkit "+
			"moby/buildkit:buildx-stable-1 --addr tcp://0.0.0.0:%s "+
			"--tlscacert /certs/ca.pem --tlscert /certs/cert.pem --tlskey /certs/key.pem",
		hostPort, strings.TrimPrefix(hostPort, ":"),
	)
}

// EnsureRemoteBuildx registers a buildx remote driver (best effort).
func EnsureRemoteBuildx(ctx context.Context, builderName, endpoint string, certDir string) error {
	builderName = strings.TrimSpace(builderName)
	if builderName == "" {
		return fmt.Errorf("builder name required")
	}
	check := exec.CommandContext(ctx, "docker", "buildx", "inspect", builderName)
	if check.Run() == nil {
		return nil
	}
	cacert := filepath.Join(certDir, "ca.pem")
	cert := filepath.Join(certDir, "cert.pem")
	key := filepath.Join(certDir, "key.pem")
	create := exec.CommandContext(ctx, "docker", "buildx", "create",
		"--name", builderName,
		"--driver", "remote",
		"--driver-opt", "cacert="+cacert,
		"--driver-opt", "cert="+cert,
		"--driver-opt", "key="+key,
		endpoint,
	)
	out, err := create.CombinedOutput()
	if err != nil {
		return fmt.Errorf("buildx create: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// WriteCertDir writes client TLS material for buildx (0600 on keys).
func WriteCertDir(dir string, bundle RemoteCertBundle) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	files := map[string][]byte{
		"ca.pem":   bundle.CA,
		"cert.pem": bundle.ClientCert,
		"key.pem":  bundle.ClientKey,
	}
	for name, data := range files {
		mode := os.FileMode(0o644)
		if strings.HasSuffix(name, ".pem") && name == "key.pem" {
			mode = 0o600
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, mode); err != nil {
			return err
		}
	}
	return nil
}

// TestRemoteBuilder runs buildx inspect on a registered builder name.
func TestRemoteBuilder(ctx context.Context, builderName string) (platforms string, err error) {
	cmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", builderName, "--bootstrap")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
