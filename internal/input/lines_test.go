package input

import (
	"bytes"
	"testing"
)

func TestLinesEmpty(t *testing.T) {
	var buf bytes.Buffer
	ch := make(chan string)
	Lines(&buf, ch)
	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0", len(lines))
	}
}

func TestLinesNoTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte("hello"))
	ch := make(chan string)
	Lines(&buf, ch)
	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0] != "hello" {
		t.Errorf("got %q, want %q", lines[0], "hello")
	}
}

func TestLinesWhitespaceOnly(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte("   \n     \n"))
	ch := make(chan string)
	Lines(&buf, ch)
	var lines []string
	for line := range ch {
		lines = append(lines, line)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0", len(lines))
	}
}

func TestFetchLines(t *testing.T) {
	var stdin bytes.Buffer
	stdin.Write([]byte("line1\n\nline2 \n"))
	ch := make(chan string)
	Lines(&stdin, ch)

	lines := []string{}
	for line := range ch {
		lines = append(lines, line)
	}
	if len(lines) != 2 {
		t.Errorf("got %d, wanted %d", len(lines), 2)
	}
	if lines[0] != "line1" {
		t.Errorf("got [%s], wanted [%s]", lines[0], "line1")
	}
	if lines[1] != "line2" {
		t.Errorf("got [%s], wanted [%s]", lines[1], "line2")
	}
}
