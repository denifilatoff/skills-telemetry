package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
)

// caFileName is the well-known name the sender auto-discovers under the config
// dir. Provisioning copies the deployment CA here; the flush adds it to the
// trust pool when present.
const caFileName = "ca.crt"

// copyCAFile validates that src is a PEM bundle containing at least one
// certificate, then copies it to <configDir>/ca.crt. Validating up front turns
// a wrong path into a clear provisioning error instead of a silent TLS failure
// at send time. The cert is not secret, so 0644 is fine.
func copyCAFile(configDir, src string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(b) {
		return errors.New("no PEM certificate found in " + src)
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	dst := filepath.Join(configDir, caFileName)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

// caTLSConfig auto-discovers the CA at <configDir>/ca.crt. When the file is
// absent it returns (nil, nil): the caller leaves TLS at its default, which
// trusts the system store. When present it appends the CA to a copy of the
// system pool, so trust is additive — the machine keeps trusting public and
// corporate roots and adds the private CA on top. A missing config dir or an
// unreadable system pool degrades to a pool holding just the CA.
func caTLSConfig(configDir string) (*tls.Config, error) {
	if configDir == "" {
		return nil, nil
	}
	b, err := os.ReadFile(filepath.Join(configDir, caFileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(b) {
		return nil, errors.New("ca.crt holds no PEM certificate")
	}
	return &tls.Config{RootCAs: pool}, nil
}
