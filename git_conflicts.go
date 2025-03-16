package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// colorizeConflictOutput adds ANSI color codes to conflict markers and content
// for better readability in terminal output
func colorizeConflictOutput(content string) string {
	lines := strings.Split(content, "\n")
	var colored []string

	inOurs := false
	inTheirs := false

	for _, line := range lines {
		// Colorize conflict markers
		if strings.HasPrefix(line, "<<<<<<< ") {
			// Cyan color for start marker
			colored = append(colored, "\033[36m"+line+"\033[0m")
			inOurs = true
			continue
		}

		if line == "=======" {
			// Cyan color for separator marker
			colored = append(colored, "\033[36m"+line+"\033[0m")
			inOurs = false
			inTheirs = true
			continue
		}

		if strings.HasPrefix(line, ">>>>>>> ") {
			// Cyan color for end marker
			colored = append(colored, "\033[36m"+line+"\033[0m")
			inTheirs = false
			continue
		}

		// Colorize content
		if inOurs {
			// Red color for "our" changes
			colored = append(colored, "\033[31m"+line+"\033[0m")
		} else if inTheirs {
			// Green color for "their" changes
			colored = append(colored, "\033[32m"+line+"\033[0m")
		} else {
			// Normal text without color
			colored = append(colored, line)
		}
	}

	return strings.Join(colored, "\n")
}

// getCurrentBranchName returns the name of the current branch
func getCurrentBranchName() string {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "your branch"
	}
	return strings.TrimSpace(string(output))
}

// getMergingBranchName returns the name of the branch being merged
func getMergingBranchName() string {
	// Check if MERGE_HEAD exists (we're in the middle of a merge)
	_, err := os.Stat(".git/MERGE_HEAD")
	if os.IsNotExist(err) {
		return "incoming changes"
	}

	// Get the branch name from the MERGE_HEAD
	cmd := exec.Command("git", "name-rev", "--name-only", "MERGE_HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "incoming changes"
	}

	branchName := strings.TrimSpace(string(output))
	return "incoming changes from " + branchName
}

// HandleGitConflicts resolves Git merge conflicts in SOPS encrypted files
func HandleGitConflicts(filePath string, options DiffOptions) error {
	// Read the file with conflicts
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Check if the file actually contains conflicts
	if !bytes.Contains(content, []byte("<<<<<<< ")) {
		return fmt.Errorf("file %s does not contain Git conflicts", filePath)
	}

	// Create the output paths
	fileExt := filepath.Ext(filePath)
	baseName := filepath.Base(filePath)
	baseNameNoExt := strings.TrimSuffix(baseName, fileExt)
	workDir := filepath.Dir(filePath)

	oursPath := filepath.Join(workDir, baseNameNoExt+".ours"+fileExt)
	theirsPath := filepath.Join(workDir, baseNameNoExt+".theirs"+fileExt)

	// Extract both versions from the conflict
	contentStr := string(content)
	oursContent := extractOursVersion(contentStr)
	theirsContent := extractTheirsVersion(contentStr)

	// Write the two versions to temporary files
	err = ioutil.WriteFile(oursPath, []byte(oursContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write 'ours' version: %w", err)
	}
	defer cleanupFile(oursPath)

	err = ioutil.WriteFile(theirsPath, []byte(theirsContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write 'theirs' version: %w", err)
	}
	defer cleanupFile(theirsPath)

	// Decrypt both versions using the sops command line and keep in memory
	oursDecrypted, err := decryptWithSopsToMemory(oursPath)
	if err != nil {
		return fmt.Errorf("failed to decrypt 'ours' version: %w", err)
	}

	theirsDecrypted, err := decryptWithSopsToMemory(theirsPath)
	if err != nil {
		return fmt.Errorf("failed to decrypt 'theirs' version: %w", err)
	}

	// Get branch names
	currentBranch := getCurrentBranchName()
	mergingBranch := getMergingBranchName()

	// Create the merged decrypted file with conflict markers and detailed branch info
	mergedContent := fmt.Sprintf("<<<<<<< HEAD (%s branch)\n%s=======\n%s>>>>>>> OTHER (%s)\n",
		currentBranch, string(oursDecrypted), string(theirsDecrypted), mergingBranch)

	// Display helpful information
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// Check if the output should go to a file or stdout
	if options.OutputFile != "" {
		// Write to file - no coloring for file output
		err = ioutil.WriteFile(options.OutputFile, []byte(mergedContent), 0600)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Println(green("✓"), cyan("Created decrypted conflict file:"), options.OutputFile)
		fmt.Println(yellow("Instructions:"))
		fmt.Println("1. Edit the decrypted file to resolve conflicts")
		fmt.Println("2. Once resolved, encrypt it using sops:")
		fmt.Printf("   sops -e -i %s\n", options.OutputFile)
		fmt.Println("3. Replace the original file with the encrypted version:")
		fmt.Printf("   mv %s %s\n", options.OutputFile+".enc", filePath)
	} else {
		// Print to stdout
		if options.ColorOutput && isatty.IsTerminal(os.Stdout.Fd()) {
			// Apply coloring if color is enabled and output is to a terminal
			fmt.Print(colorizeConflictOutput(mergedContent))
		} else {
			// Regular output without coloring
			fmt.Print(mergedContent)
		}
	}

	fmt.Println()
	fmt.Println(yellow("Note:"), "The decrypted file contains sensitive information. Delete it when no longer needed.")

	return nil
}

// HandleGitMerge handles a Git merge operation using the sops-diff tool
// This function is called by Git when merging encrypted files
func HandleGitMerge(local, base, remote, merged string, options DiffOptions) error {
	// Decrypt all the files directly without reading their content into unused variables
	localDecrypted, err := decryptWithSopsToMemory(local)
	if err != nil {
		return fmt.Errorf("failed to decrypt local version: %w", err)
	}

	baseDecrypted, err := decryptWithSopsToMemory(base)
	if err != nil {
		return fmt.Errorf("failed to decrypt base version: %w", err)
	}

	remoteDecrypted, err := decryptWithSopsToMemory(remote)
	if err != nil {
		return fmt.Errorf("failed to decrypt remote version: %w", err)
	}

	// Create temporary files for decrypted content to use with diff tool
	tmpDir, err := ioutil.TempDir("", "sops-merge-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	localDecPath := filepath.Join(tmpDir, "LOCAL")
	baseDecPath := filepath.Join(tmpDir, "BASE")
	remoteDecPath := filepath.Join(tmpDir, "REMOTE")
	mergedDecPath := filepath.Join(tmpDir, "MERGED")

	// Write decrypted content to temporary files
	if err := ioutil.WriteFile(localDecPath, localDecrypted, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted local file: %w", err)
	}

	if err := ioutil.WriteFile(baseDecPath, baseDecrypted, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted base file: %w", err)
	}

	if err := ioutil.WriteFile(remoteDecPath, remoteDecrypted, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted remote file: %w", err)
	}

	// Initial merged content with conflict markers
	mergedContent := fmt.Sprintf("<<<<<<< LOCAL\n%s=======\n%s>>>>>>> REMOTE\n",
		string(localDecrypted), string(remoteDecrypted))

	if err := ioutil.WriteFile(mergedDecPath, []byte(mergedContent), 0600); err != nil {
		return fmt.Errorf("failed to write initial merged file: %w", err)
	}

	// Launch external diff tool if specified
	if options.DiffTool != "" {
		diffCmd := exec.Command(options.DiffTool, localDecPath, remoteDecPath, mergedDecPath)
		diffCmd.Stdin = os.Stdin
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr

		if err := diffCmd.Run(); err != nil {
			return fmt.Errorf("diff tool failed: %w", err)
		}
	} else {
		fmt.Println("No diff tool specified. Using default merge with conflict markers.")
	}

	// Read the merged result
	mergedResult, err := ioutil.ReadFile(mergedDecPath)
	if err != nil {
		return fmt.Errorf("failed to read merged result: %w", err)
	}

	// Check if there are still conflict markers
	if bytes.Contains(mergedResult, []byte("<<<<<<< ")) {
		fmt.Println("Merge not complete: conflict markers still present in the merged file.")
		return fmt.Errorf("conflicts not resolved")
	}

	// Encrypt the merged result
	cmd := exec.Command("sops", "-e", "--input-type", filepath.Ext(merged)[1:], "--output-type", filepath.Ext(merged)[1:], "/dev/stdin")
	cmd.Stdin = bytes.NewReader(mergedResult)
	encryptedOutput, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("sops encryption failed: %s", exitErr.Stderr)
		}
		return fmt.Errorf("sops encryption failed: %w", err)
	}

	// Write the encrypted result to the merged file
	if err := ioutil.WriteFile(merged, encryptedOutput, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted merged file: %w", err)
	}

	fmt.Println("Successfully merged and encrypted the result.")
	return nil
}

// setupGitMergeTool configures Git to use sops-diff for resolving conflicts in encrypted files
func SetupGitMergeTool() error {
	// Configure Git to use sops-diff as a merge tool
	cmds := []struct {
		args []string
	}{
		{[]string{"config", "--global", "merge.sops.name", "SOPS merge tool"}},
		{[]string{"config", "--global", "merge.sops.driver", "sops-diff git-merge %A %O %B %P"}},
		{[]string{"config", "--global", "merge.sops.recursive", "binary"}},
		{[]string{"config", "--global", "mergetool.sops.cmd", "sops-diff git-merge --diff-tool=$EDITOR $LOCAL $BASE $REMOTE $MERGED"}},
		{[]string{"config", "--global", "mergetool.sops.trustExitCode", "true"}},
	}

	for _, cmd := range cmds {
		if err := exec.Command("git", cmd.args...).Run(); err != nil {
			return fmt.Errorf("error executing git %s: %w", strings.Join(cmd.args, " "), err)
		}
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(green("✓"), "Successfully configured Git to use sops-diff for encrypted files")
	fmt.Println(yellow("Next steps:"))
	fmt.Println("Add the following to your .gitattributes file:")
	fmt.Println("*.enc.yaml merge=sops")
	fmt.Println("*.enc.json merge=sops")
	fmt.Println("*.enc.env merge=sops")

	return nil
}

// cleanupFile safely removes a file
func cleanupFile(path string) {
	_ = ioutil.WriteFile(path, []byte{}, 0600) // Overwrite with empty content first
	_ = os.Remove(path)
}

// decryptWithSopsToMemory decrypts a file using the sops command line and returns the content
func decryptWithSopsToMemory(inputPath string) ([]byte, error) {
	cmd := exec.Command("sops", "-d", inputPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("sops decryption failed: %s", exitErr.Stderr)
		}
		return nil, fmt.Errorf("sops decryption failed: %w", err)
	}

	return output, nil
}

// extractOursVersion extracts the "our" version from the conflict
func extractOursVersion(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string

	inConflict := false
	takeOurs := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "<<<<<<< ") {
			inConflict = true
			takeOurs = true
			continue
		}

		if inConflict && line == "=======" {
			takeOurs = false
			continue
		}

		if inConflict && strings.HasPrefix(line, ">>>>>>> ") {
			inConflict = false
			takeOurs = false
			continue
		}

		if !inConflict || takeOurs {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// extractTheirsVersion extracts the "their" version from the conflict
func extractTheirsVersion(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string

	inConflict := false
	takeTheirs := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "<<<<<<< ") {
			inConflict = true
			takeTheirs = false
			continue
		}

		if inConflict && line == "=======" {
			takeTheirs = true
			continue
		}

		if inConflict && strings.HasPrefix(line, ">>>>>>> ") {
			inConflict = false
			takeTheirs = false
			continue
		}

		if !inConflict || takeTheirs {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}
