package process

import (
	"io"
	"os"
	"strings"
	"sync"
)

func NewManager(commands <-chan string, limit uint, output io.Writer) *Manager {
	return &Manager{commands, []*process{}, limit, output}
}

type Manager struct {
	commands  <-chan string
	processes []*process
	limit     uint
	output    io.Writer
}

func (m *Manager) Start() {
	/* Acknowledge commands */
	var id int
	for command := range m.commands {
		if strings.HasPrefix(command, "#") {
			continue
		}
		id++
		proc := newProcess(id, command, m.output)
		proc.ack()
		m.processes = append(m.processes, proc)
	}

	/* Create buffered channel to dispatch commands */
	capacity := len(m.processes)
	if m.limit > 0 {
		capacity = int(m.limit)
	}
	ch := make(chan *process, capacity)

	/* Dispatch commands asynchronously */
	go m.dispatch(ch)

	/* Start asynchronous consumers to limit concurrent command execution */
	wg := &sync.WaitGroup{}
	wg.Add(len(m.processes))
	for i := 0; i < capacity; i++ {
		go m.consume(ch, wg)
	}

	/* Wait until all commands have been run */
	wg.Wait()
}

func (m *Manager) Signal(s os.Signal) {
	for _, p := range m.processes {
		if p.cmd.Process != nil {
			p.cmd.Process.Signal(s)
		}
	}
}

func (m *Manager) dispatch(ch chan<- *process) {
	for _, proc := range m.processes {
		ch <- proc
	}
	close(ch)
}

func (m *Manager) consume(ch <-chan *process, wg *sync.WaitGroup) {
	for proc := range ch {
		proc.Run()
		wg.Done()
	}
}
