package main

import (
	"crypto/tls"
	"errors"
	"time"
)

// selftestSkill is the marker skill name carried by a probe event. The
// collector and dashboards filter on it so a probe never counts as real
// skill usage.
const selftestSkill = "__selftest__"

// selftestResult reports what the live probe proved.
type selftestResult struct {
	Delivered bool // the collector accepted the probe and it left the spool
	Sent      int  // events sent in the flush that carried the probe
}

// runSelftest sends one real, marked probe event and confirms the pipeline
// works end to end up to ingest: the collector accepted it (HTTP 200) and the
// probe left the spool. This is the guarantee available without read access to
// the store. An empty endpoint is a configuration error, not a delivery
// failure — the machine is not provisioned.
func runSelftest(s *Spool, endpoint, token string, tlsConfig *tls.Config, timeout time.Duration) (selftestResult, error) {
	if endpoint == "" {
		return selftestResult{}, errors.New("no endpoint: machine is not provisioned")
	}
	probe := SkillEvent{
		Agent:     "selftest",
		SessionID: newUUID(),
		Skill:     selftestSkill,
		TS:        time.Now().UTC(),
	}
	if err := s.Enqueue(probe); err != nil {
		return selftestResult{}, err
	}
	sent, err := Flush(s, endpoint, token, tlsConfig, timeout)
	if err != nil {
		return selftestResult{Sent: sent}, err
	}
	return selftestResult{Delivered: probesRemaining(s) == 0, Sent: sent}, nil
}

// probesRemaining counts probe events still buffered — used to confirm the
// probe actually left the spool after a flush.
func probesRemaining(s *Spool) int {
	names, err := s.List()
	if err != nil {
		return 0
	}
	n := 0
	for _, name := range names {
		if ev, err := s.Read(name); err == nil && ev.Skill == selftestSkill {
			n++
		}
	}
	return n
}
