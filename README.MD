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
- Store and re-use templates for parameters, for example, to store a set of patterns.

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
  aicat [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  template    Manage templates

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
	boa.Cmd{
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

## Partially documented features

### Templates

You can store your parameters in reusable templates, for example:

`aicat template store go -p "*.go,*.mod"`

Then you can use the template like this:

`aicat . -t go`

## Code quality

Absolute garbage. This is a quick and dirty implementation to scratch an itch. It is not tested, and the code is not
very clean. It is not recommended to use this code as an example of how to write good Go code. It is a prototype, and it
is not intended to be used in production... Is what you could say. But it works, and it is useful. So, there's that.

It's mostly bootstrapped/written by itself, together with github copilot, and some adjustments by me.

Most of this readme is AI generated, and I have not read it. I'm sure it's great.

## Other notes

Here are some fancy unicode characters selected by google that you might enjoy:

```
(✿ ♡‿♡)ᕦ⊙෴⊙ᕤ( ͝° ͜ʖ͡°)ᕤᕙ( ͡° ͜ʖ ͡°)ᕗ└|∵|┐♪乁( ⁰͡ Ĺ̯ ⁰͡ ) ㄏ乁( . ര ʖ̯ ര . )
ㄏ♡(ӦｖӦ｡)(ﾉ◕ヮ◕)ﾉ*.✧╮(. ❛ ᴗ ❛.)╭(ノ｀Д´)ノ彡┻━┻(┛❍ᴥ❍)┛彡┻━┻(-_- )ﾉ⌒┫ ┻ ┣ ┳
ಠ︵ಠಠ ೧ ಠಠಿ_ಠ(´;︵;)(༎ຶ ෴ ༎ຶ)༎ຶ‿༎ຶ(＠_＠;)(^_^メ)(╭☞•́⍛•̀)╭☞´◔‿ゝ◔)━☞⊙.☉⁄(⁄ ⁄•⁄-⁄•⁄ ⁄)⁄
```