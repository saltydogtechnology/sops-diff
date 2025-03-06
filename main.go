package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/getsops/sops/v3/decrypt"
	"github.com/mattn/go-isatty"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// Version of the sops-diff utility
	Version = "0.1.0"
)

var (
	// Command line flags
	summaryMode      bool
	outputFormat     string
	colorOutput      bool
	diffTool         string
	gitSupport       bool
	errorOnDecrypted bool
)

type DiffOptions struct {
	SummaryMode      bool
	OutputFormat     string
	ColorOutput      bool
	DiffTool         string
	GitSupport       bool
	ErrorOnDecrypted bool
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "sops-diff [flags] FILE1 FILE2",
		Short: "Compare two SOPS-encrypted files",
		Long: `sops-diff - A utility to compare two SOPS-encrypted files
		
This tool decrypts two SOPS-encrypted files and shows the differences between them.
It supports YAML, JSON, and ENV file formats, with options for summary mode (keys only)
or full mode (keys and values). The output is formatted for easy review.

Examples:
  sops-diff secret1.enc.yaml secret2.enc.yaml
  sops-diff --summary secret1.enc.yaml secret2.enc.yaml
  sops-diff HEAD:secrets.enc.yaml secrets.enc.yaml
  sops-diff --format=json secret1.enc.json secret2.enc.json
  sops-diff --format=env config1.env config2.env
`,
		Version: Version,
		// NOTE: Changed from ExactArgs(2) to handle Git diff arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			options := DiffOptions{
				SummaryMode:      summaryMode,
				OutputFormat:     outputFormat,
				ColorOutput:      colorOutput,
				DiffTool:         diffTool,
				GitSupport:       gitSupport,
				ErrorOnDecrypted: errorOnDecrypted,
			}

			// Handle Git diff invocation with special argument pattern
			if gitSupport && len(args) >= 7 {
				// Git passes: path old-file old-hex old-mode new-file new-hex new-mode
				// We need old-file (args[1]) and the actual file path (args[0] or args[4])

				// Use old-file (temporary blob file) for first arg
				oldFile := args[1]

				// For the second file, use the path from args[0]
				// (This handles the case when comparing working copy with staged/committed)
				newFile := args[0]

				// If new-hex (args[5]) is not all zeros, we're comparing different revisions
				if args[5] != "0000000000000000000000000000000000000000" {
					// When comparing different revisions, use args[4] for new file
					newFile = args[4]
				}

				fmt.Fprintf(os.Stderr, "Git diff mode: comparing %s with %s\n", oldFile, newFile)
				return runDiff(oldFile, newFile, options)
			}

			// Regular (non-Git) invocation requires exactly 2 args
			if len(args) != 2 {
				return fmt.Errorf("accepts 2 arg(s), received %d", len(args))
			}

			return runDiff(args[0], args[1], options)
		},
	}

	// Define flags
	rootCmd.Flags().BoolVarP(&summaryMode, "summary", "s", false, "Display only keys that have changed, without sensitive values")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "auto", "Output format: auto, yaml, json, env")
	rootCmd.Flags().BoolVarP(&colorOutput, "color", "c", true, "Use colored output when supported")
	rootCmd.Flags().StringVarP(&diffTool, "diff-tool", "d", "", "Use an external diff tool (e.g. 'vimdiff')")
	rootCmd.Flags().BoolVarP(&gitSupport, "git", "g", false, "Enable Git revision comparison support")
	rootCmd.Flags().BoolVar(&errorOnDecrypted, "error-on-decrypted", true, "Return error if any file is found to be decrypted")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Compare two sets of data and show only changed keys
func compareData(data1, data2 interface{}) (string, error) {
	flat1 := make(map[string]interface{})
	flat2 := make(map[string]interface{})

	flatten(data1, "", flat1)
	flatten(data2, "", flat2)

	var changed []string

	// Find keys that exist in data1 but not in data2 or have different values
	for k, v1 := range flat1 {
		if v2, exists := flat2[k]; !exists {
			changed = append(changed, fmt.Sprintf("- %s", k))
		} else if fmt.Sprintf("%v", v1) != fmt.Sprintf("%v", v2) {
			changed = append(changed, fmt.Sprintf("! %s", k))
		}
	}

	// Find keys that exist in data2 but not in data1
	for k := range flat2 {
		if _, exists := flat1[k]; !exists {
			changed = append(changed, fmt.Sprintf("+ %s", k))
		}
	}

	sort.Strings(changed)

	var buffer strings.Builder
	for _, line := range changed {
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	return buffer.String(), nil
}

// Compare two env files and show only changed keys
func compareEnvData(data1, data2 map[string]string) (string, error) {
	var changed []string

	// Find keys that exist in data1 but not in data2 or have different values
	for k, v1 := range data1 {
		if v2, exists := data2[k]; !exists {
			changed = append(changed, fmt.Sprintf("- %s", k))
		} else if v1 != v2 {
			changed = append(changed, fmt.Sprintf("! %s", k))
		}
	}

	// Find keys that exist in data2 but not in data1
	for k := range data2 {
		if _, exists := data1[k]; !exists {
			changed = append(changed, fmt.Sprintf("+ %s", k))
		}
	}

	sort.Strings(changed)

	var buffer strings.Builder
	for _, line := range changed {
		buffer.WriteString(line)
		buffer.WriteString("\n")
	}

	return buffer.String(), nil
}

// runDiff is the main function that handles the diff operation
func runDiff(file1Path, file2Path string, options DiffOptions) error {
	// Keep all the existing code for reading and decrypting files
	var file1Content, file2Content []byte
	var err error

	// Handle Git references if enabled
	if options.GitSupport && (strings.Contains(file1Path, ":") || strings.Contains(file2Path, ":")) {
		file1Content, err = readGitFile(file1Path)
		if err != nil {
			return fmt.Errorf("error reading Git file %s: %w", file1Path, err)
		}

		file2Content, err = readGitFile(file2Path)
		if err != nil {
			return fmt.Errorf("error reading Git file %s: %w", file2Path, err)
		}
	} else {
		// Regular file reading
		file1Content, err = ioutil.ReadFile(file1Path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", file1Path, err)
		}

		file2Content, err = ioutil.ReadFile(file2Path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", file2Path, err)
		}
	}

	// Determine file format
	format1 := detectFormat(file1Path, options.OutputFormat)
	format2 := detectFormat(file2Path, options.OutputFormat)

	// Use the explicitly specified format or the detected one
	format := options.OutputFormat
	if format == "auto" {
		// If any of the files is .env, use env format
		if format1 == "env" || format2 == "env" {
			format = "env"
		} else if format1 != format2 {
			return fmt.Errorf("files appear to be different formats: %s and %s", format1, format2)
		} else {
			format = format1
		}
	}

	// Decrypt files
	decryptFormat := format
	if format == "env" {
		decryptFormat = "dotenv"
	}

	// Try to decrypt both files
	var decrypted1, decrypted2 []byte
	var decryptErr1, decryptErr2 error

	decrypted1, decryptErr1 = decrypt.Data(file1Content, decryptFormat)
	decrypted2, decryptErr2 = decrypt.Data(file2Content, decryptFormat)

	// Handle cases where files are already decrypted (has no SOPS metadata)
	var file1Decrypted, file2Decrypted bool

	if decryptErr1 != nil && strings.Contains(decryptErr1.Error(), "sops metadata not found") {
		decrypted1 = file1Content
		decryptErr1 = nil
		file1Decrypted = true

		// Print warning for potentially unencrypted sensitive content
		fmt.Fprintf(os.Stderr, "\033[33mWARNING: File '%s' appears to be decrypted (no SOPS metadata found)!\033[0m\n", file1Path)
		fmt.Fprintf(os.Stderr, "\033[33m         Make sure you don't commit decrypted sensitive files.\033[0m\n")

		// If configured to error on decrypted files, return an error
		if options.ErrorOnDecrypted {
			return fmt.Errorf("file '%s' is decrypted, aborting as --error-on-decrypted is enabled", file1Path)
		}
	}

	if decryptErr2 != nil && strings.Contains(decryptErr2.Error(), "sops metadata not found") {
		// Print warning for potentially unencrypted sensitive content
		fmt.Fprintf(os.Stderr, "\033[33mWARNING: File '%s' appears to be decrypted (no SOPS metadata found)!\033[0m\n", file2Path)
		fmt.Fprintf(os.Stderr, "\033[33m         Make sure you don't commit decrypted sensitive files.\033[0m\n")

		// If configured to error on decrypted files, return an error
		if options.ErrorOnDecrypted {
			return fmt.Errorf("file '%s' is decrypted, aborting as --error-on-decrypted is enabled", file2Path)
		}

		decrypted2 = file2Content
		decryptErr2 = nil
		file2Decrypted = true
	}

	// If both files were already decrypted, show a message
	if file1Decrypted && file2Decrypted && !options.SummaryMode {
		fmt.Println("\033[33mBoth files appear to be already decrypted. Comparing as plain text.\033[0m")
	} else if (file1Decrypted || file2Decrypted) && !options.SummaryMode {
		// If one file is encrypted and one is decrypted, warn about potential false positives
		fmt.Fprintf(os.Stderr, "\033[33mNote: Comparing encrypted and decrypted files may show structural differences\033[0m\n")
		fmt.Fprintf(os.Stderr, "\033[33min addition to actual content changes.\033[0m\n")
	}

	// If decryption fails with dotenv format, try other formats for .env files
	if format == "env" && (decryptErr1 != nil || decryptErr2 != nil) {
		// Try with yaml format first
		if decryptErr1 != nil {
			decrypted1, decryptErr1 = decrypt.Data(file1Content, "yaml")
		}
		if decryptErr2 != nil {
			decrypted2, decryptErr2 = decrypt.Data(file2Content, "yaml")
		}

		// If still failing, try json format
		if decryptErr1 != nil {
			decrypted1, decryptErr1 = decrypt.Data(file1Content, "json")
		}
		if decryptErr2 != nil {
			decrypted2, decryptErr2 = decrypt.Data(file2Content, "json")
		}
	}

	// Return the first error encountered if decryption still failed
	if decryptErr1 != nil {
		return fmt.Errorf("error decrypting %s: %w", file1Path, decryptErr1)
	}

	if decryptErr2 != nil {
		return fmt.Errorf("error decrypting %s: %w", file2Path, decryptErr2)
	}

	// For env files, we need to handle differently since they might have been encrypted using different formats
	if format == "env" {
		// Parse .env files directly as text
		data1Map, err := parseEnv(decrypted1)
		if err != nil {
			return fmt.Errorf("error parsing ENV from %s: %w", file1Path, err)
		}

		data2Map, err := parseEnv(decrypted2)
		if err != nil {
			return fmt.Errorf("error parsing ENV from %s: %w", file2Path, err)
		}

		// If using an external diff tool
		if options.DiffTool != "" {
			return diffWithExternalTool(data1Map, data2Map, format, options)
		}

		// Generate formatted output for comparison
		if options.SummaryMode {
			// Direct comparison of data for summary mode using the specialized env comparison
			summaryOutput, err := compareEnvData(data1Map, data2Map)
			if err != nil {
				return fmt.Errorf("error generating summary comparison: %w", err)
			}

			// If there are no changes, inform the user
			if summaryOutput == "" {
				fmt.Println("No changes detected in keys")
			} else {
				fmt.Println("Summary of key changes:")
				fmt.Println("! = modified key, + = added key, - = removed key")
				fmt.Println("--------------------------------------")
				fmt.Print(summaryOutput)
			}
			return nil
		} else {
			// Full mode - show keys and values
			output1, err := formatFull(data1Map, format)
			if err != nil {
				return fmt.Errorf("error formatting data for %s: %w", file1Path, err)
			}

			output2, err := formatFull(data2Map, format)
			if err != nil {
				return fmt.Errorf("error formatting data for %s: %w", file2Path, err)
			}

			// Generate and display the diff
			diff := generateDiff(file1Path, file2Path, output1, output2, options)
			fmt.Print(diff)
		}
		return nil
	}

	// For non-env formats, continue with the normal process
	var data1, data2 interface{}
	switch format {
	case "yaml":
		err = yaml.Unmarshal(decrypted1, &data1)
		if err != nil {
			return fmt.Errorf("error parsing YAML from %s: %w", file1Path, err)
		}

		err = yaml.Unmarshal(decrypted2, &data2)
		if err != nil {
			return fmt.Errorf("error parsing YAML from %s: %w", file2Path, err)
		}
	case "json":
		err = json.Unmarshal(decrypted1, &data1)
		if err != nil {
			return fmt.Errorf("error parsing JSON from %s: %w", file1Path, err)
		}

		err = json.Unmarshal(decrypted2, &data2)
		if err != nil {
			return fmt.Errorf("error parsing JSON from %s: %w", file2Path, err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// If using an external diff tool
	if options.DiffTool != "" {
		return diffWithExternalTool(data1, data2, format, options)
	}

	// Generate formatted output for comparison
	if options.SummaryMode {
		// Direct comparison of data for summary mode
		summaryOutput, err := compareData(data1, data2)
		if err != nil {
			return fmt.Errorf("error generating summary comparison: %w", err)
		}

		// If there are no changes, inform the user
		if summaryOutput == "" {
			fmt.Println("No changes detected in keys")
		} else {
			fmt.Println("Summary of key changes:")
			fmt.Println("! = modified key, + = added key, - = removed key")
			fmt.Println("--------------------------------------")
			fmt.Print(summaryOutput)
		}
		return nil
	} else {
		// Full mode - show keys and values
		var output1, output2 string

		output1, err = formatFull(data1, format)
		if err != nil {
			return fmt.Errorf("error formatting data for %s: %w", file1Path, err)
		}

		output2, err = formatFull(data2, format)
		if err != nil {
			return fmt.Errorf("error formatting data for %s: %w", file2Path, err)
		}

		// Generate and display the diff
		diff := generateDiff(file1Path, file2Path, output1, output2, options)
		fmt.Print(diff)
	}

	return nil
}

// detectFormat detects the file format based on extension or specified format
func detectFormat(filePath, specifiedFormat string) string {
	if specifiedFormat != "auto" {
		return specifiedFormat
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".env":
		return "env"
	default:
		// Default to YAML if can't detect
		return "yaml"
	}
}

// parseEnv parses an environment file into a map
func parseEnv(data []byte) (map[string]string, error) {
	result := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, comments, and lines obviously not in .env format
		if line == "" ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "{") ||
			strings.HasPrefix(line, "[") ||
			strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "sops:") ||
			strings.Contains(line, ": |") {
			continue
		}

		// Find the first equals sign
		idx := strings.Index(line, "=")
		if idx <= 0 {
			// Skip lines without = or if = is the first character
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Handle quoted values
		if len(value) > 1 && (value[0] == '"' && value[len(value)-1] == '"' ||
			value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}

		result[key] = value
	}

	return result, nil
}

// formatSummary formats data showing only the keys (for summary mode)
func formatSummary(data interface{}, format string) (string, error) {
	// Flatten the data structure to get all keys
	flatMap := make(map[string]interface{})
	flatten(data, "", flatMap)

	var keys []string
	for k := range flatMap {
		keys = append(keys, k)
	}

	// Sort keys for consistent output
	sort.Strings(keys)

	var buffer strings.Builder
	for _, k := range keys {
		buffer.WriteString(k)
		buffer.WriteString("\n")
	}

	return buffer.String(), nil
}

// formatFull formats data showing keys and values (for full mode)
func formatFull(data interface{}, format string) (string, error) {
	var output []byte
	var err error

	switch format {
	case "yaml":
		output, err = yaml.Marshal(data)
	case "json":
		output, err = json.MarshalIndent(data, "", "  ")
	case "env":
		// For ENV format, convert to a string representation
		if m, ok := data.(map[string]string); ok {
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var buffer strings.Builder
			for _, k := range keys {
				buffer.WriteString(k)
				buffer.WriteString("=")
				buffer.WriteString(m[k])
				buffer.WriteString("\n")
			}
			return buffer.String(), nil
		} else {
			return "", fmt.Errorf("expected map[string]string for ENV format, got %T", data)
		}
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return "", err
	}

	return string(output), nil
}

// generateDiff creates a diff output between two strings
func generateDiff(file1, file2, text1, text2 string, options DiffOptions) string {
	fromFile := "a/" + filepath.Base(file1)
	toFile := "b/" + filepath.Base(file2)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(text1),
		B:        difflib.SplitLines(text2),
		FromFile: fromFile,
		ToFile:   toFile,
		Context:  3,
		Eol:      "\n",
	}

	result, _ := difflib.GetUnifiedDiffString(diff)

	// Apply colors if enabled and output is to a terminal
	if options.ColorOutput && isatty.IsTerminal(os.Stdout.Fd()) {
		result = colorDiff(result)
	}

	return result
}

// colorDiff applies ANSI color codes to make diff output more readable
func colorDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var colored []string

	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			// Green for additions
			colored = append(colored, "\033[32m"+line+"\033[0m")
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			// Red for deletions
			colored = append(colored, "\033[31m"+line+"\033[0m")
		} else if strings.HasPrefix(line, "@@") {
			// Cyan for line information
			colored = append(colored, "\033[36m"+line+"\033[0m")
		} else {
			colored = append(colored, line)
		}
	}

	return strings.Join(colored, "\n")
}

// diffWithExternalTool uses an external tool for diffing
func diffWithExternalTool(data1, data2 interface{}, format string, options DiffOptions) error {
	// Create temporary files for the decrypted content
	tmpFile1, err := ioutil.TempFile("", "sops-diff-*")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	tmpPath1 := tmpFile1.Name()
	defer os.Remove(tmpPath1)

	tmpFile2, err := ioutil.TempFile("", "sops-diff-*")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	tmpPath2 := tmpFile2.Name()
	defer os.Remove(tmpPath2)

	// Format and write the content
	if options.SummaryMode {
		// For summary mode with external diff tool, we'll output to a single file
		var summaryOutput string
		var err error

		// Use appropriate comparison function based on data type
		if _, ok := data1.(map[string]string); ok && format == "env" {
			// For env files
			summaryOutput, err = compareEnvData(data1.(map[string]string), data2.(map[string]string))
		} else {
			// For other formats
			summaryOutput, err = compareData(data1, data2)
		}
		if err != nil {
			return fmt.Errorf("error generating summary comparison: %w", err)
		}

		if summaryOutput == "" {
			summaryOutput = "No changes detected in keys\n"
		} else {
			summaryOutput = "Summary of key changes:\n! = modified key, + = added key, - = removed key\n--------------------------------------\n" + summaryOutput
		}

		if _, err := tmpFile1.WriteString(summaryOutput); err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}
		if err := tmpFile1.Close(); err != nil {
			return fmt.Errorf("error closing temporary file: %w", err)
		}

		// For viewing a single file result
		cmd := exec.Command(options.DiffTool, tmpPath1)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	} else {
		// Full mode with external diff tool
		formattedData1, err := formatFull(data1, format)
		if err != nil {
			return fmt.Errorf("error formatting first file for external diff tool: %w", err)
		}
		formattedData2, err := formatFull(data2, format)
		if err != nil {
			return fmt.Errorf("error formatting second file for external diff tool: %w", err)
		}

		if _, err := tmpFile1.WriteString(formattedData1); err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}
		if err := tmpFile1.Close(); err != nil {
			return fmt.Errorf("error closing temporary file: %w", err)
		}

		if _, err := tmpFile2.WriteString(formattedData2); err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}
		if err := tmpFile2.Close(); err != nil {
			return fmt.Errorf("error closing temporary file: %w", err)
		}

		// Run the external diff tool
		cmd := exec.Command(options.DiffTool, tmpPath1, tmpPath2)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}
}

// readGitFile reads content from a Git revision (e.g., HEAD:path/to/file)
func readGitFile(gitPath string) ([]byte, error) {
	parts := strings.SplitN(gitPath, ":", 2)
	if len(parts) != 2 {
		// Not a Git path, treat as a regular file
		return ioutil.ReadFile(gitPath)
	}

	revision := parts[0]
	path := parts[1]

	// Use git show to get the content
	cmd := exec.Command("git", "show", revision+":"+path)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git show command failed: %w", err)
	}

	return output.Bytes(), nil
}

// flatten recursively flattens a nested data structure into a map with dot notation keys
func flatten(data interface{}, prefix string, result map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flatten(val, newKey, result)
		}
	case map[interface{}]interface{}:
		for k, val := range v {
			strKey, ok := k.(string)
			if !ok {
				strKey = fmt.Sprintf("%v", k)
			}

			newKey := strKey
			if prefix != "" {
				newKey = prefix + "." + strKey
			}
			flatten(val, newKey, result)
		}
	case []interface{}:
		for i, val := range v {
			newKey := fmt.Sprintf("%s[%d]", prefix, i)
			flatten(val, newKey, result)
		}
	default:
		result[prefix] = v
	}
}
