package main

import (
	"fmt"
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var params struct {
	RootDir      boa.Required[string] `descr:"Root directory to search for files"`
	FileType     boa.Required[string] `descr:"Type of files to search for (f for regular files)"`
	NamePattern  boa.Required[string] `descr:"Pattern to match file names"`
	TransformCmd boa.Optional[string] `descr:"Optional shell command to transform file contents"`
}

func main() {
	boa.Wrap{
		Use:    "aicat",
		Short:  "Concatenate and optionally transform file contents",
		Long:   `A CLI tool to concatenate and optionally transform file contents based on specified patterns.`,
		Params: &params,
		Run: func(cmd *cobra.Command, args []string) {
			// Find files matching the pattern
			var files []string
			err := filepath.Walk(params.RootDir.Value(), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if params.FileType.Value() == "f" && !info.IsDir() && matchPattern(info.Name(), params.NamePattern.Value()) {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				fmt.Println("Error walking the path:", err)
				return
			}

			// Concatenate file contents with headers
			for _, file := range files {
				content, err := ioutil.ReadFile(file)
				if err != nil {
					fmt.Println("Error reading file:", file, err)
					continue
				}

				// Print file header
				fmt.Printf("--- FILE: %s ---\n", file)

				// Optionally transform content
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
