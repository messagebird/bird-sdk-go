package bird

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/messagebird/bird-sdk-go/option"
)

// testPKI is a throwaway CA with a server and a client leaf, written to disk
// because the option takes file paths.
type testPKI struct {
	pool       *x509.CertPool
	server     tls.Certificate
	clientCert string
	clientKey  string
}

func newTestPKI(t *testing.T) testPKI {
	t.Helper()
	dir := t.TempDir()
	caKey := newKey(t)
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	ca, _ := x509.ParseCertificate(caDER)
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	leaf := func(serial int64, cn string, usage x509.ExtKeyUsage) ([]byte, *ecdsa.PrivateKey) {
		key := newKey(t)
		tmpl := &x509.Certificate{
			SerialNumber: big.NewInt(serial),
			Subject:      pkix.Name{CommonName: cn},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(time.Hour),
			ExtKeyUsage:  []x509.ExtKeyUsage{usage},
			IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}
		return der, key
	}

	serverDER, serverKey := leaf(2, "server", x509.ExtKeyUsageServerAuth)
	clientDER, clientKey := leaf(3, "sdk client", x509.ExtKeyUsageClientAuth)
	p := testPKI{
		pool:       pool,
		server:     tls.Certificate{Certificate: [][]byte{serverDER}, PrivateKey: serverKey},
		clientCert: filepath.Join(dir, "client.pem"),
		clientKey:  filepath.Join(dir, "client.key"),
	}
	writePEM(t, p.clientCert, "CERTIFICATE", clientDER)
	keyDER, err := x509.MarshalPKCS8PrivateKey(clientKey)
	if err != nil {
		t.Fatal(err)
	}
	writePEM(t, p.clientKey, "PRIVATE KEY", keyDER)
	return p
}

func newKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func writePEM(t *testing.T, path, typ string, der []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestClientCertificatePresented(t *testing.T) {
	pki := newTestPKI(t)
	var gotCN string
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCN = r.TLS.PeerCertificates[0].Subject.CommonName
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{pki.server},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pki.pool,
	}
	srv.StartTLS()
	defer srv.Close()

	c, err := NewClient(option.WithAPIKey("bk_eu1_x"), option.WithBaseURL(srv.URL), option.WithClientCertificate(pki.clientCert, pki.clientKey))
	if err != nil {
		t.Fatal(err)
	}
	if c.cfg.BaseURL != srv.URL {
		t.Errorf("explicit base URL = %q, want it unchanged: %q", c.cfg.BaseURL, srv.URL)
	}
	// The SDK verifies servers against the system roots; trust the test CA.
	c.cfg.HTTPClient.Transport.(*http.Transport).TLSClientConfig.RootCAs = pki.pool
	if err := c.Get(context.Background(), "/v1/workspace", nil); err != nil {
		t.Fatal(err)
	}
	if gotCN != "sdk client" {
		t.Errorf("server saw client CN %q, want %q", gotCN, "sdk client")
	}
}

type wrappedTransport struct{ http.RoundTripper }

func TestClientCertificateWithReplacedDefaultTransport(t *testing.T) {
	pki := newTestPKI(t)
	saved := http.DefaultTransport
	http.DefaultTransport = wrappedTransport{saved}
	t.Cleanup(func() { http.DefaultTransport = saved })

	c, err := NewClient(option.WithAPIKey("bk_eu1_x"), option.WithClientCertificate(pki.clientCert, pki.clientKey))
	if err != nil {
		t.Fatal(err)
	}
	transport := c.cfg.HTTPClient.Transport.(*http.Transport)
	if len(transport.TLSClientConfig.Certificates) != 1 || transport.Proxy == nil {
		t.Errorf("transport = %+v, want the certificate and the environment proxy", transport)
	}
}

func TestClientCertificateSelectsMTLSHost(t *testing.T) {
	pki := newTestPKI(t)
	for _, tc := range []struct {
		name string
		opts []option.RequestOption
		want string
	}{
		{"key prefix", []option.RequestOption{option.WithAPIKey("bk_us1_x")}, "https://mtls.us1.platform.bird.com"},
		{"explicit region", []option.RequestOption{option.WithAPIKey("bk_us1_x"), option.WithRegion("eu1")}, "https://mtls.eu1.platform.bird.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient(append(tc.opts, option.WithClientCertificate(pki.clientCert, pki.clientKey))...)
			if err != nil {
				t.Fatal(err)
			}
			if c.cfg.BaseURL != tc.want {
				t.Errorf("BaseURL = %q, want %q", c.cfg.BaseURL, tc.want)
			}
		})
	}
}

func TestClientCertificateRefusals(t *testing.T) {
	pki := newTestPKI(t)
	dir := t.TempDir()
	otherKey := filepath.Join(dir, "other.key")
	otherDER, err := x509.MarshalPKCS8PrivateKey(newKey(t))
	if err != nil {
		t.Fatal(err)
	}
	writePEM(t, otherKey, "PRIVATE KEY", otherDER)
	encryptedKey := filepath.Join(dir, "encrypted.key")
	writePEM(t, encryptedKey, "ENCRYPTED PRIVATE KEY", otherDER)
	missing := filepath.Join(dir, "missing.pem")

	for _, tc := range []struct {
		name string
		opts []option.RequestOption
		want string
	}{
		{"key without certificate", []option.RequestOption{option.WithClientCertificate("", pki.clientKey)}, "certificate"},
		{"certificate without key", []option.RequestOption{option.WithClientCertificate(pki.clientCert, "")}, "key"},
		{"missing file", []option.RequestOption{option.WithClientCertificate(missing, pki.clientKey)}, missing},
		{"mismatched key", []option.RequestOption{option.WithClientCertificate(pki.clientCert, otherKey)}, otherKey},
		{"encrypted key", []option.RequestOption{option.WithClientCertificate(pki.clientCert, encryptedKey)}, encryptedKey},
		{"custom HTTP client", []option.RequestOption{option.WithHTTPClient(&http.Client{}), option.WithClientCertificate(pki.clientCert, pki.clientKey)}, "WithHTTPClient"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(append([]option.RequestOption{option.WithAPIKey("bk_eu1_x")}, tc.opts...)...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want one naming %q", err, tc.want)
			}
		})
	}

	c, err := NewClient(option.WithAPIKey("bk_eu1_x"))
	if err != nil {
		t.Fatal(err)
	}
	err = c.Get(context.Background(), "/v1/workspace", nil, option.WithClientCertificate(pki.clientCert, pki.clientKey))
	if err == nil || !strings.Contains(err.Error(), "client option") {
		t.Fatalf("per-call err = %v, want the client-option refusal", err)
	}
}
