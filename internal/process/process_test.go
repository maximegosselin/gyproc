package process

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func parseEvent(t *testing.T, line string) event {
	t.Helper()
	var e event
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		t.Fatalf("invalid JSON %q: %v", line, err)
	}
	return e
}

func parseEvents(t *testing.T, output string) []event {
	t.Helper()
	var events []event
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" {
			events = append(events, parseEvent(t, line))
		}
	}
	return events
}

func TestOutputAckLine(t *testing.T) {
	var buf bytes.Buffer
	p := newProcess(3, "echo hello", &buf)
	p.ack()

	e := parseEvent(t, strings.TrimSpace(buf.String()))
	if e.Seq != 3 {
		t.Errorf("seq: got %d, want 3", e.Seq)
	}
	if e.Event != "ack" {
		t.Errorf("event: got %s, want ack", e.Event)
	}
	if e.Command == nil || *e.Command != "echo hello" {
		t.Errorf("command: got %v, want echo hello", e.Command)
	}
}

func TestRunSuccess(t *testing.T) {
	var buf bytes.Buffer
	p := newProcess(1, "echo hello", &buf)
	p.Run()

	events := parseEvents(t, buf.String())
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d:\n%s", len(events), buf.String())
	}
	if events[0].Event != "run" {
		t.Errorf("event[0]: got %s, want run", events[0].Event)
	}
	if events[1].Event != "out" {
		t.Errorf("event[1]: got %s, want out", events[1].Event)
	}
	if events[2].Event != "exit" {
		t.Errorf("event[2]: got %s, want exit", events[2].Event)
	}
	if events[2].ExitCode == nil || *events[2].ExitCode != 0 {
		t.Errorf("exit code: got %v, want 0", events[2].ExitCode)
	}
}

func TestRunNonZeroExitCode(t *testing.T) {
	var buf bytes.Buffer
	p := newProcess(1, "false", &buf)
	p.Run()

	events := parseEvents(t, buf.String())
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d:\n%s", len(events), buf.String())
	}
	if events[0].Event != "run" {
		t.Errorf("event[0]: got %s, want run", events[0].Event)
	}
	if events[1].Event != "exit" {
		t.Errorf("event[1]: got %s, want exit", events[1].Event)
	}
	if events[1].ExitCode == nil || *events[1].ExitCode != 1 {
		t.Errorf("exit code: got %v, want 1", events[1].ExitCode)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	p := newProcess(1, "nonexistentcommand", &buf)
	p.Run()

	events := parseEvents(t, buf.String())
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d:\n%s", len(events), buf.String())
	}
	if events[0].Event != "run" {
		t.Errorf("event[0]: got %s, want run", events[0].Event)
	}
	if events[1].Event != "fail" {
		t.Errorf("event[1]: got %s, want fail", events[1].Event)
	}
	if events[1].Message == nil || !strings.Contains(*events[1].Message, "nonexistentcommand") {
		t.Errorf("fail message: got %v, want message containing command name", events[1].Message)
	}
}
