# mygrep-go

This repository contains a CLI application that implements a part of the `grep` command. It has been heavily influenced by the grep track on codecrafters.io.

## Features

- CLI interface for searching patterns in files/stdin
- Tiny implementation of support for regular expressions
  - Start/end of string anchor: `^`, `$`
  - Quantifier: `+`, `*`, `?`
  - Wildcard: `.`
  - Meta characters: `\d`, `\w`
  - Positive/negative character group: `[abc]`, `[^abc]`
  - Alternation: `(abc|def)`

## Getting Started

To get started with mygrep-go, follow these steps:

1. Clone the repository: `git clone https://github.com/miy4/mygrep-go.git`
1. Install the necessary dependencies: `go get -d ./...`
1. Build the application: `go build`
1. Run the application: `./mygrep pattern file`

## Acknowledgement

I would like to acknowledge the following resources that have been instrumental in the development of this project:

- The [grep track](https://app.codecrafters.io/courses/grep) on codecrafters.io: This track provided valuable guidance and inspiration for implementing the `grep` functionality in this CLI application.
- The [article](https://rhaeguard.github.io/posts/regex/) by [rhaeguard](https://github.com/rhaeguard/): It was a great reference for understanding regular expressions and implementing support for them in this project.
