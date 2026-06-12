package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

const (
	bufferCap      = 100
	flushCountN    = 10
	flushIntervalT = 60 * time.Second
	flushTimeout   = 2 * time.Second
)

func main() {
	os.Exit(run(os.Args[1:], func(s string) { fmt.Print(s) }))
}

func run(args []string, stdout func(string)) int {
	if len(args) == 0 {
		stdout("usage: skills-telemetry <ingest|flush|status|version>\n")
		return 2
	}
	switch args[0] {
	case "version":
		stdout(version + "\n")
		return 0
	case "ingest":
		agent, endpoint := parseFlags(args[1:])
		s, err := DefaultSpool()
		if err != nil {
			fmt.Fprintln(os.Stderr, "spool:", err)
			return 0 // never fail the hook
		}
		raw, _ := io.ReadAll(os.Stdin)
		return ingest(s, agent, endpoint, raw, gitRemote)
	case "flush":
		_, endpoint := parseFlags(args[1:])
		s, err := DefaultSpool()
		if err != nil {
			fmt.Fprintln(os.Stderr, "spool:", err)
			return 0
		}
		if _, err := Flush(s, endpoint, resolveToken(), flushTimeout); err != nil {
			fmt.Fprintln(os.Stderr, "flush:", err)
		}
		return 0
	case "status":
		s, err := DefaultSpool()
		if err != nil {
			fmt.Fprintln(os.Stderr, "spool:", err)
			return 0
		}
		names, err := s.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, "list:", err)
		}
		last := "never"
		if fi, err := os.Stat(filepath.Join(s.Dir, markerName)); err == nil {
			last = fi.ModTime().UTC().Format(time.RFC3339)
		}
		stdout(fmt.Sprintf("spool: %s\nbuffered: %d\nlast_flush_attempt: %s\n", s.Dir, len(names), last))
		return 0
	default:
		stdout("unknown command: " + args[0] + "\n")
		return 2
	}
}

// parseFlags reads --agent= and --endpoint= without a flag framework (minimal).
func parseFlags(args []string) (agent, endpoint string) {
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--agent="):
			agent = strings.TrimPrefix(a, "--agent=")
		case strings.HasPrefix(a, "--endpoint="):
			endpoint = strings.TrimPrefix(a, "--endpoint=")
		}
	}
	return agent, endpoint
}

// ingest is the per-event path: parse, enqueue, rotate, opportunistic flush.
// It returns 0 even on error — a hook must never fail the agent turn.
func ingest(s *Spool, agent, endpoint string, stdin []byte, remote remoteResolver) int {
	events, err := Dispatch(agent, stdin, remote)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dispatch:", err)
		return 0
	}
	for _, ev := range events {
		if err := s.Enqueue(ev); err != nil {
			fmt.Fprintln(os.Stderr, "enqueue:", err)
		}
	}
	if _, err := s.Rotate(bufferCap); err != nil {
		fmt.Fprintln(os.Stderr, "rotate:", err)
	}
	if shouldFlush(s, flushCountN, flushIntervalT) {
		touchMarker(s)
		if _, err := Flush(s, endpoint, resolveToken(), flushTimeout); err != nil {
			fmt.Fprintln(os.Stderr, "flush:", err)
		}
	}
	return 0
}

// shouldFlush is true when there is something to send AND either enough has
// piled up or enough time has passed since the last attempt.
func shouldFlush(s *Spool, countN int, intervalT time.Duration) bool {
	names, err := s.List()
	if err != nil || len(names) == 0 {
		return false
	}
	if len(names) >= countN {
		return true
	}
	fi, err := os.Stat(filepath.Join(s.Dir, markerName))
	if err != nil {
		return true // no prior attempt recorded
	}
	return time.Since(fi.ModTime()) >= intervalT
}

// touchMarker records the time of a flush attempt (success or failure) so the
// throttle bounds retry frequency against a dead collector.
func touchMarker(s *Spool) {
	p := filepath.Join(s.Dir, markerName)
	now := time.Now()
	if err := os.WriteFile(p, []byte(now.UTC().Format(time.RFC3339)), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "marker:", err)
		return
	}
	_ = os.Chtimes(p, now, now)
}

// gitRemote best-effort resolves origin URL for a working dir; "" on failure.
func gitRemote(cwd string) string {
	cmd := exec.Command("git", "-C", cwd, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
