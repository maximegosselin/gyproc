package main

import (
	"flag"
	"fmt"
	"github.com/maximegosselin/gyproc/internal/input"
	"github.com/maximegosselin/gyproc/internal/process"
	"os"
	"os/signal"
	"syscall"
)

var (
	file  string
	limit uint
)

func main() {
	flag.StringVar(&file, "file", "", "file containing commands")
	flag.UintVar(&limit, "limit", 0, "limit of concurrent commands")
	flag.Parse()

	/* Use file as input if provided with --file argument */
	in := os.Stdin
	if file != "" {
		r, err := os.Open(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		in = r
	}

	/* Send input lines on the commands channel */
	commands := make(chan string)
	input.Lines(in, commands)

	/* Create process manager */
	pm := process.NewManager(commands, limit, os.Stdout)

	/* Forward signals to processes */
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		pm.Signal(<-signals)
	}()

	/* Start process manager */
	pm.Start()
}
