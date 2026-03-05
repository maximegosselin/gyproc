package process

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

func newProcess(seq int, command string, output io.Writer) *process {
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	proc := &process{seq, command, cmd, output}
	cmd.Stdout = &processOutputWriter{proc}
	cmd.Stderr = &processOutputWriter{proc}
	return proc
}

type process struct {
	seq     int
	command string
	cmd     *exec.Cmd
	output  io.Writer
}

func (p *process) emit(e *event) {
	b, _ := json.Marshal(e)
	fmt.Fprintln(p.output, string(b))
}

func (p *process) ack() {
	p.emit(&event{Seq: p.seq, Event: "ack", Command: &p.command, Timestamp: time.Now().UTC()})
}

func (p *process) Run() error {
	p.emit(&event{Seq: p.seq, Event: "run", Timestamp: time.Now().UTC()})
	err := p.cmd.Run()
	_, isExitError := err.(*exec.ExitError)
	if err == nil || isExitError {
		exitCode := p.cmd.ProcessState.ExitCode()
		p.emit(&event{Seq: p.seq, Event: "exit", Pid: &p.cmd.Process.Pid, ExitCode: &exitCode, Timestamp: time.Now().UTC()})
	} else {
		msg := fmt.Errorf("could not start process: %v", err).Error()
		p.emit(&event{Seq: p.seq, Event: "fail", Message: &msg, Timestamp: time.Now().UTC()})
	}
	return err
}

type processOutputWriter struct {
	proc *process
}

func (w *processOutputWriter) Write(p []byte) (n int, err error) {
	output := string(p)
	pid := w.proc.cmd.Process.Pid
	w.proc.emit(&event{Seq: w.proc.seq, Event: "out", Pid: &pid, Message: &output, Timestamp: time.Now().UTC()})
	return len(p), nil
}

type event struct {
	Seq       int       `json:"seq"`
	Event     string    `json:"event"`
	Pid       *int      `json:"pid,omitempty"`
	Command   *string   `json:"command,omitempty"`
	Message   *string   `json:"message,omitempty"`
	ExitCode  *int      `json:"code,omitempty"`
	Timestamp time.Time `json:"time"`
}
