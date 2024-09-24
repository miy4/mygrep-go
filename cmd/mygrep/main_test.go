package main

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name string
		args []string
		in   string
		out  string
		err  string
		want int
	}{
		{
			name: "no args",
			args: []string{},
			in:   "",
			out:  "",
			err:  "Usage: mygrep PATTERN [FILE]\n",
			want: EXIT_ERROR,
		},
		{
			name: "valid pattern",
			args: []string{"a"},
			in:   "a\nb\nc\n",
			out:  "a\n",
			err:  "",
			want: EXIT_OK,
		},
	}

	for _, tt := range tests {
		outBuffer := &strings.Builder{}
		errBuffer := &strings.Builder{}
		cli := &cli{
			in:  strings.NewReader(tt.in),
			out: outBuffer,
			err: errBuffer,
		}
		exit := cli.run(tt.args)
		if exit != tt.want {
			t.Errorf("%s: exit = %d; want %d", tt.name, exit, tt.want)
		} else if outBuffer.String() != tt.out {
			t.Errorf("%s: out = %q; want %q", tt.name, outBuffer.String(), tt.out)
		} else if errBuffer.String() != tt.err {
			t.Errorf("%s: err = %q; want %q", tt.name, errBuffer.String(), tt.err)
		}
	}
}
