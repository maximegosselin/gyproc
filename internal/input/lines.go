package input

import (
	"bufio"
	"io"
	"strings"
)

// Lines reads lines from a io.Reader and sends them on a channel.
func Lines(r io.Reader, lines chan<- string) {
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				lines <- line
			}
		}
		close(lines)
	}()
}
