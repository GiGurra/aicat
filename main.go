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
	Pattern      boa.Optional[string] `descr:"Glob pattern to match files" pos:"true"`
	FileType     boa.Required[string] `short:"t" descr:"Type of files to search for (f for regular files)" default:"f"`
	Binary       boa.Required[bool]   `descr:"Print binary files" default:"false"`
	NamePattern  boa.Optional[string] `descr:"Pattern to match file names"`
	TransformCmd boa.Optional[string] `descr:"Optional shell command to transform file contents"`
}

func globPatternToFileList(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func main() {
	boa.Wrap{
		Use:    "aicat",
		Short:  "Concatenate and optionally transform file contents",
		Long:   `A CLI tool to concatenate and optionally transform file contents based on specified patterns.`,
		Params: &params,
		Run: func(cmd *cobra.Command, args []string) {

			globPattern := func() string {
				if params.Pattern.HasValue() {
					return *params.Pattern.Value()
				}
				return "*"
			}()

			filesByGlobalPattern, err := globPatternToFileList(globPattern)
			if err != nil {
				panic(fmt.Errorf("error globbing pattern '%s': %w", params.Pattern.Value(), err))
			}

			// iterate all files and filter based on file type and pattern
			var files []string
			for _, file := range filesByGlobalPattern {
				fileInfo, err := os.Stat(file)
				if err != nil {
					slog.Error(fmt.Sprintf("Error stating file: %s", file), err)
					continue
				}

				switch params.FileType.Value() {
				case "f":
					if !fileInfo.Mode().IsRegular() || fileInfo.IsDir() {
						continue
					}
				default:
					panic(fmt.Errorf("unknown file type: %s", params.FileType.Value()))
				}

				if params.NamePattern.HasValue() {
					if !matchPattern(filepath.Base(file), *params.NamePattern.Value()) {
						continue
					}
				}

				files = append(files, file)
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
				if params.TransformCmd.Value() != nil {
					transformedContent, err := runTransformCommand(*params.TransformCmd.Value(), file, string(content))
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
