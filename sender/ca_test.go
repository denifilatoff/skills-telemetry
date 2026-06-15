package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// selfSignedPEM returns a valid self-signed certificate in PEM form for tests.
func selfSignedPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-ca"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestCopyCAFileWritesCanonicalCert(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	src := filepath.Join(t.TempDir(), "source.crt")
	pemBytes := selfSignedPEM(t)
	if err := os.WriteFile(src, pemBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyCAFile(dir, src); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "ca.crt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(pemBytes) {
		t.Fatal("copied cert bytes differ from source")
	}
}

func TestCopyCAFileRejectsMissingSource(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	if err := copyCAFile(dir, filepath.Join(t.TempDir(), "nope.crt")); err == nil {
		t.Fatal("want error for missing source")
	}
}

func TestCopyCAFileRejectsNonPEM(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	src := filepath.Join(t.TempDir(), "bad.crt")
	if err := os.WriteFile(src, []byte("not a certificate"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyCAFile(dir, src); err == nil {
		t.Fatal("want error for non-PEM input")
	}
}

func TestCATLSConfigNilWhenNoCertFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	cfg, err := caTLSConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Fatal("want nil config when no ca.crt (fall back to system trust)")
	}
}

func TestCATLSConfigBuildsPoolWhenCertPresent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), pkgName)
	src := filepath.Join(t.TempDir(), "source.crt")
	if err := os.WriteFile(src, selfSignedPEM(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := copyCAFile(dir, src); err != nil {
		t.Fatal(err)
	}
	cfg, err := caTLSConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.RootCAs == nil {
		t.Fatal("want a TLS config with a populated root pool")
	}
}
