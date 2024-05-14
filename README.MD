# aicat

`aicat` is a very basic application for building LLM context out of text files on disk. It is a command-line tool that
concatenates and optionally transforms the contents of files based on specified patterns. It allows you to filter
files using patterns and file type, and apply transformations using shell commands.

## features

 - Match files using patterns. 
 - Filter files by type (e.g., regular files). 
 - Optionally apply patterns to match specific file names. 
 - Concatenate file contents. 
 - Optionally transform file contents using shell commands. 
 - Handle binary files with an optional flag.

## installation

To install aicat, you need to have Go installed. Then, you can build and install the application using:

`go install github.com/gigurra/aicat`

## usage

aicat [root dir] [flags]

## examples

1 Concatenate all .go and .mod files:

    aicat -p "*.go,*.mod"

2 Concatenate and transform file contents:

    aicat "*.txt" --transform "sed 's/foo/bar/g'"

3 Include binary (just identified as non utf8 contents) files:

    aicat -b

## license

This project is licensed under the MIT License. See the LICENSE file for details.

## contributing

In the unexpected event that this application is used by anyone else,
contributions are welcome! Please open an issue or submit a pull request on GitHub.
