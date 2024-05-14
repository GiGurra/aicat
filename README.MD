# aicat

`aicat` is a very basic application for building LLM context out of text files on disk. It is a command-line tool that
concatenates and optionally transforms the contents of files based on specified patterns. It allows you to filter
files using patterns and file type, and apply transformations using shell commands.

## features

 - Match files using patterns. 
 - Filter files by type (e.g., regular files). 
 - Optionally apply patterns to match specific file names. 
 - Concatenate file contents. 
 - Optionally transform file contents using shell commands. (untested/planned)
 - Handle binary files with an optional flag.

## installation

To install aicat, you need to have Go installed. Then, you can build and install the application using:

`go install github.com/gigurra/aicat`

## usage

`aicat [root dir] [flags]`

```
>aicat --help

A CLI tool to concatenate and optionally transform file contents based on specified patterns.

Usage:
  aicat [root] [flags]

Flags:
  -t, --file-type string   Type of files to search for (f for regular files) (env: FILE_TYPE) (default "f")
  -b, --binary             Print binary files (env: BINARY) (default false)
  -p, --patterns strings   Pattern to match file names (env: PATTERNS)
      --transform string   Optional shell command to transform file contents (env: TRANSFORM)
  -h, --help               help for aicat
```

## examples

1. Concatenate all .go and .mod files:

    aicat -p "*.go,*.mod"

2. Concatenate and transform file contents:

    aicat -p "*.txt" --transform "sed 's/foo/bar/g'"

3. Include binary (just identified as non utf8 contents) files:

    aicat . -b

4. Cat the relevant go source files in this repo
```
aicat . -p "*.go,*.mod"
--- FILE: go.mod ---
module github.com/gigurra/aicat

go 1.22.2

require (
	github.com/GiGurra/boa v0.0.32
	...



--- FILE: main.go ---
package main

import (
	"fmt"
	...
)

var params struct {
	Root      boa.Optional[string]   `descr:"Root directory to start from" pos:"true"`
	FileType  boa.Required[string]   `short:"t" descr:"Type of files to search for (f for regular files)" default:"f"`
	Binary    boa.Required[bool]     `descr:"Print binary files" default:"false"`
	Patterns  boa.Optional[[]string] `descr:"Pattern to match file names"`
	Transform boa.Optional[string]   `descr:"Optional shell command to transform file contents"`
}

func main() {
	boa.Wrap{
		Use:    "aicat",
		Short:  "Concatenate and optionally transform file contents",
		Long:   `A CLI tool to concatenate and optionally transform file contents based on specified patterns.`,
		Params: &params,
		Run: func(cmd *cobra.Command, args []string) {

...
```

## license

This project is licensed under the MIT License. See the LICENSE file for details.

## contributing

In the unexpected event that this application is used by anyone else,
contributions are welcome! Please open an issue or submit a pull request on GitHub.
