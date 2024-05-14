package main

import (
	"fmt"
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
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

			rootDir := func() string {
				if params.Root.HasValue() {
					return *params.Root.Value()
				}
				return "."
			}()

			// walk the file tree and collect all files
			var files []string
			err := filepath.Walk(rootDir, func(file string, info os.FileInfo, err error) error {

				slog.Info(fmt.Sprintf("Processing file: %s", file))

				fileInfo, err := os.Stat(file)
				if err != nil {
					slog.Error(fmt.Sprintf("Error stating file: %s", file), err)
					return nil
				}

				switch params.FileType.Value() {
				case "f":
					if !fileInfo.Mode().IsRegular() || fileInfo.IsDir() {
						return nil
					}
				default:
					panic(fmt.Errorf("unknown file type: %s", params.FileType.Value()))
				}

				if params.Patterns.HasValue() {
					foundMatch := false
					for _, pattern := range *params.Patterns.Value() {
						if matchPattern(filepath.Base(file), pattern) {
							foundMatch = true
							break
						}
					}
					if !foundMatch {
						return nil
					}
				}

				files = append(files, file)

				return nil
			})

			if err != nil {
				panic(fmt.Errorf("error walking the path: %s", err))
			}

			// Concatenate file contents with headers
			for _, file := range files {
				content, err := os.ReadFile(file)
				if err != nil {
					fmt.Println("Error reading file:", file, err)
					continue
				}

				//// Print file header
				fmt.Printf("--- FILE: %s ---\n", file)

				// check if the file is valid utf8
				if !utf8.ValidString(string(content)) && !params.Binary.Value() {
					fmt.Printf(" - Contents not valid utf8, assumed binary, skipping\n")
					continue
				}

				//Optionally transform content
				if params.Transform.Value() != nil {
					transformedContent, err := runTransformCommand(*params.Transform.Value(), file, string(content))
					if err != nil {
						fmt.Println("Error transforming file:", file, err)
						continue
					}
					fmt.Println(transformedContent)
				} else {
					fmt.Println(string(content))
				}

				fmt.Println()
			}
		},
	}.ToApp()
}

// matchPattern checks if a file name matches the given pattern
func matchPattern(name, pattern string) bool {
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

// runTransformCommand runs the transformation command on the file content
func runTransformCommand(cmd, filePath, content string) (string, error) {
	cmd = strings.ReplaceAll(cmd, "_path_", filePath)
	cmd = strings.ReplaceAll(cmd, "_contents_", content)

	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
