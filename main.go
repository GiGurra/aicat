package main

import (
	"encoding/json"
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

type Params struct {
	Root      boa.Optional[string]   `descr:"Root directory to start from" pos:"true"`
	FileType  boa.Required[string]   `short:"t" descr:"Type of files to search for (f for regular files)" default:"f"`
	Binary    boa.Required[bool]     `descr:"Print binary files" default:"false"`
	Patterns  boa.Optional[[]string] `descr:"Pattern to match file names"`
	Transform boa.Optional[string]   `descr:"Optional shell command to transform file contents"`
}

var rootParams Params
var storeParams Params

func main() {
	boa.Wrap{
		Use:    "aicat",
		Short:  "Concatenate and optionally transform file contents",
		Long:   `A CLI tool to concatenate and optionally transform file contents based on specified patterns.`,
		Params: &rootParams,
		SubCommands: []*cobra.Command{
			boa.Wrap{
				Use:   "template",
				Short: "Manage templates",
				SubCommands: []*cobra.Command{
					boa.Wrap{
						Use:    "store [name]",
						Short:  "Store a new template",
						Params: &storeParams,
						Run:    storeTemplate,
					}.ToCmd(),
					{
						Use:   "delete [name]",
						Short: "Delete an existing template",
						Run:   deleteTemplate,
					},
					{
						Use:   "list",
						Short: "List all templates",
						Run:   listTemplates,
					},
				},
			}.ToCmd(),
		},
		Run: func(cmd *cobra.Command, args []string) {
			rootDir := func() string {
				if rootParams.Root.HasValue() {
					return *rootParams.Root.Value()
				}
				return "."
			}()

			params := SelectedParams{
				FileType: rootParams.FileType.Value(),
				Binary:   rootParams.Binary.Value(),
				Patterns: func() []string {
					if rootParams.Patterns.HasValue() {
						return *rootParams.Patterns.Value()
					}
					return nil
				}(),
				Transform: func() string {
					if rootParams.Transform.HasValue() {
						return *rootParams.Transform.Value()
					}
					return ""
				}(),
			}

			// walk the file tree and collect all files
			var files []string
			err := filepath.Walk(rootDir, func(file string, info os.FileInfo, err error) error {
				fileInfo, err := os.Stat(file)
				if err != nil {
					panic(fmt.Sprintf("error stating file: %s: %v", file, err))
				}

				switch rootParams.FileType.Value() {
				case "f":
					if !fileInfo.Mode().IsRegular() || fileInfo.IsDir() {
						return nil
					}
				default:

					// Find template
					templateDir := getTemplateDir()
					templatePath := filepath.Join(templateDir, rootParams.FileType.Value()+".json")
					templateData, err := os.ReadFile(templatePath)
					if err != nil {
						slog.Error(fmt.Sprintf("error reading template file: %s\n", err))
						os.Exit(1)
					}

					storedParams := &StoredParams{}
					err = json.Unmarshal(templateData, storedParams)
					if err != nil {
						panic(fmt.Errorf("error unmarshalling template file: %s", err))
					}

					if storedParams.FileType != nil && !rootParams.FileType.HasValue() {
						params.FileType = *storedParams.FileType
					}

					if storedParams.Patterns != nil && !rootParams.Patterns.HasValue() {
						params.Patterns = *storedParams.Patterns
					}

					if storedParams.Transform != nil && !rootParams.Transform.HasValue() {
						params.Transform = *storedParams.Transform
					}

					if storedParams.Binary != nil && !rootParams.Binary.HasValue() {
						params.Binary = *storedParams.Binary
					}
				}

				if params.Patterns != nil {
					foundMatch := false
					for _, pattern := range params.Patterns {
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
					panic(fmt.Errorf("error reading file: %s", err))
				}

				// Print file header
				fmt.Printf("--- FILE: %s ---\n", file)

				// check if the file is valid utf8
				if !utf8.ValidString(string(content)) && !params.Binary {
					slog.Warn(fmt.Sprintf(" - Contents of '%s' not valid utf8, assumed binary, skipping\n", file))
					continue
				}

				// Optionally transform content
				if params.Transform != "" {
					transformedContent, err := runTransformCommand(params.Transform, file, string(content))
					if err != nil {
						panic(fmt.Errorf("error running transformation command: %s", err))
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

func getTemplateDir() string {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("error getting user home directory: %s", err))
	}

	res := filepath.Join(homeDir, ".aicat", "templates")

	err = os.MkdirAll(res, 0755)

	if err != nil {
		panic(fmt.Errorf("error creating template directory: %s", err))
	}

	return res
}

type StoredParams struct {
	FileType  *string   `json:"fileType,omitempty"`
	Binary    *bool     `json:"binary,omitempty"`
	Patterns  *[]string `json:"patterns,omitempty"`
	Transform *string   `json:"transform,omitempty"`
}

type SelectedParams struct {
	FileType  string   `json:"fileType"`
	Binary    bool     `json:"binary"`
	Patterns  []string `json:"patterns"`
	Transform string   `json:"transform"`
}

func toPtr[T any](t T) *T {
	return &t
}

// storeTemplate stores a new template
func storeTemplate(_ *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Template name is required")
		return
	}
	name := args[0]
	templateDir := getTemplateDir()
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		fmt.Println("Error creating template directory:", err)
		return
	}

	templatePath := filepath.Join(templateDir, name+".json")
	template := StoredParams{
		FileType: toPtr(storeParams.FileType.Value()),
		Binary:   toPtr(storeParams.Binary.Value()),
		Patterns: func() *[]string {
			if storeParams.Patterns.HasValue() {
				return storeParams.Patterns.Value()
			}
			return nil
		}(),
		Transform: func() *string {
			if storeParams.Transform.HasValue() {
				return storeParams.Transform.Value()
			}
			return nil
		}(),
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling template:", err)
		return
	}

	err = os.WriteFile(templatePath, data, 0644)
	if err != nil {
		fmt.Println("Error writing template file:", err)
		return
	}

	fmt.Println("Template stored successfully")
}

// deleteTemplate deletes an existing template
func deleteTemplate(_ *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Template name is required")
		return
	}
	name := args[0]
	templatePath := filepath.Join(getTemplateDir(), name+".json")
	err := os.Remove(templatePath)
	if err != nil {
		fmt.Println("Error deleting template:", err)
		return
	}

	fmt.Println("Template deleted successfully")
}

// listTemplates lists all stored templates
func listTemplates(_ *cobra.Command, _ []string) {
	templateDir := getTemplateDir()
	files, err := os.ReadDir(templateDir)
	if err != nil {
		fmt.Println("Error reading template directory:", err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			templateName := strings.TrimSuffix(file.Name(), ".json")
			templateData, err := os.ReadFile(filepath.Join(templateDir, file.Name()))
			if err != nil {
				panic(fmt.Errorf("error reading template file: %s", err))
			}

			// format in compact json form
			var data map[string]interface{}
			err = json.Unmarshal(templateData, &data)
			if err != nil {
				panic(fmt.Errorf("error unmarshalling template file: %s", err))
			}

			templateData, err = json.Marshal(data)
			if err != nil {
				panic(fmt.Errorf("error marshalling template file: %s", err))
			}

			paddedName := fmt.Sprintf("%-20s", templateName)

			fmt.Printf("%s %s\n", paddedName, string(templateData))
		}
	}
}
