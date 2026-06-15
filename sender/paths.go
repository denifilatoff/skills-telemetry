package main

import (
	"os"
	"path/filepath"
)

const pkgName = "qubership-skills-telemetry"

// pkgConfigDir is the per-machine config directory holding durable provisioning
// state: the env file, the CA certificate, the token, and the machine id. It is
// os.UserConfigDir() across platforms (Application Support on macOS, ~/.config
// on Linux, %AppData% on Windows). Returns "" when no config dir is available.
func pkgConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		return ""
	}
	return filepath.Join(base, pkgName)
}

// pkgConfigPath joins name onto the package config dir, or "" if there is none.
func pkgConfigPath(name string) string {
	dir := pkgConfigDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, name)
}

// pkgEnv loads the provisioned env file from the package config dir.
func pkgEnv() map[string]string {
	return loadEnvFile(pkgConfigPath("env"))
}
