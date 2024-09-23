package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	re "github.com/miy4/mygrep-go"
)

const (
	EXIT_OK        = 0
	EXIT_NOT_MATCH = 1
	EXIT_ERROR     = 2
)

// cli represents the command line interface.
type cli struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

// run executes the command.
func (c *cli) run(args []string) int {
	pattern := args[0]
	containsMatch := false
	reader := bufio.NewReader(c.in)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintf(c.err, "Failed to read input: %v\n", err)
			return EXIT_ERROR
		}

		if re.Match(line, pattern) {
			containsMatch = true
			fmt.Fprintln(c.out, line)
		}

		if err == io.EOF {
			break
		}
	}

	if !containsMatch {
		return EXIT_NOT_MATCH
	}
	return EXIT_OK
}

// main is the entry point of the command.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s PATTERN [FILE]\n", os.Args[0])
		os.Exit(EXIT_ERROR)
	}

	cli := &cli{in: os.Stdin, out: os.Stdout, err: os.Stderr}
	if len(os.Args) > 2 {
		var err error
		cli.in, err = os.Open(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Failed to open file: %v\n", os.Args[2], err)
			os.Exit(EXIT_ERROR)
		}
	}

	os.Exit(cli.run(os.Args[1:]))
}
