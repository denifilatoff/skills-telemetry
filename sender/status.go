package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// statusReport is the read-only diagnosis the provisioning skill reads to decide
// what, if anything, is missing. It never sends anything (see selftest for the
// live check).
type statusReport struct {
	Version     string
	ConfigDir   string
	Endpoint    string
	Provisioned bool
	CAFound     bool
	Buffered    int
	LastFlush   string
}

// gatherStatus inspects the spool and config dir against an already-resolved
// endpoint. A machine is "provisioned" once it has an endpoint to send to.
func gatherStatus(s *Spool, configDir, endpoint string) statusReport {
	r := statusReport{
		Version:     version,
		ConfigDir:   configDir,
		Endpoint:    endpoint,
		Provisioned: endpoint != "",
		LastFlush:   "never",
	}
	if configDir != "" {
		if _, err := os.Stat(filepath.Join(configDir, caFileName)); err == nil {
			r.CAFound = true
		}
	}
	if names, err := s.List(); err == nil {
		r.Buffered = len(names)
	}
	if fi, err := os.Stat(filepath.Join(s.Dir, flushStampName)); err == nil {
		r.LastFlush = fi.ModTime().UTC().Format(time.RFC3339)
	}
	return r
}

// formatStatus renders the report for a human and, when the machine is not yet
// provisioned, says so plainly so the next step is obvious.
func formatStatus(r statusReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "version: %s\n", r.Version)
	fmt.Fprintf(&b, "config_dir: %s\n", r.ConfigDir)
	endpoint := r.Endpoint
	if endpoint == "" {
		endpoint = "(unset)"
	}
	fmt.Fprintf(&b, "endpoint: %s\n", endpoint)
	fmt.Fprintf(&b, "ca: %s\n", caState(r.CAFound))
	fmt.Fprintf(&b, "buffered: %d\n", r.Buffered)
	fmt.Fprintf(&b, "last_flush_attempt: %s\n", r.LastFlush)
	if r.Provisioned {
		fmt.Fprint(&b, "state: provisioned\n")
	} else {
		fmt.Fprint(&b, "state: not provisioned — run `skills-telemetry provision` to set the endpoint\n")
	}
	return b.String()
}

func caState(found bool) string {
	if found {
		return "ca.crt found"
	}
	return "none (system trust store)"
}
