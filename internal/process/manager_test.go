package process

import (
	"bytes"
	"testing"
)

func makeCommands(cmds []string) <-chan string {
	ch := make(chan string, len(cmds))
	for _, cmd := range cmds {
		ch <- cmd
	}
	close(ch)
	return ch
}

func countEvents(t *testing.T, output, eventType string) int {
	t.Helper()
	count := 0
	for _, e := range parseEvents(t, output) {
		if e.Event == eventType {
			count++
		}
	}
	return count
}

func TestManagerSkipsComments(t *testing.T) {
	var buf bytes.Buffer
	ch := makeCommands([]string{"# this is a comment", "echo hello"})
	m := NewManager(ch, 0, &buf)
	m.Start()

	if n := countEvents(t, buf.String(), "ack"); n != 1 {
		t.Errorf("ack count: got %d, want 1", n)
	}
}

func TestManagerRunsAllCommands(t *testing.T) {
	var buf bytes.Buffer
	ch := makeCommands([]string{"echo a", "echo b", "echo c"})
	m := NewManager(ch, 0, &buf)
	m.Start()

	if n := countEvents(t, buf.String(), "exit"); n != 3 {
		t.Errorf("exit count: got %d, want 3", n)
	}
}

func TestManagerConcurrencyLimit(t *testing.T) {
	var buf bytes.Buffer
	ch := makeCommands([]string{"echo a", "echo b", "echo c", "echo d"})
	m := NewManager(ch, 2, &buf)
	m.Start()

	if n := countEvents(t, buf.String(), "exit"); n != 4 {
		t.Errorf("exit count: got %d, want 4", n)
	}
}
