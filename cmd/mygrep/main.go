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
	if len(args) < 1 {
		fmt.Fprintln(c.err, "Usage: mygrep PATTERN [FILE]")
		return EXIT_ERROR
	}

	pattern := args[0]
	containsMatch := false
	scanner := bufio.NewScanner(c.in)
	for {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(c.err, "Failed to read input: %v\n", err)
				return EXIT_ERROR
			}
			break
		}

		line := scanner.Text()
		ok, err := re.Match(line, pattern)
		if err != nil {
			fmt.Fprintf(c.err, "Failed to match: %v\n", err)
			return EXIT_ERROR
		} else if ok {
			containsMatch = true
			fmt.Fprintln(c.out, line)
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
