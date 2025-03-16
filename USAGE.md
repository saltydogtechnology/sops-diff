# SOPS-Diff Usage Guide

This guide provides comprehensive instructions for using the SOPS-Diff utility for comparing encrypted files.

## Command-line Options

```
sops-diff [flags] FILE1 FILE2

Flags:
  -c, --color                Use colored output when supported (default true)
  -d, --diff-tool string     Use an external diff tool (e.g. 'vimdiff')
      --error-on-decrypted   Return error if any file is found to be decrypted (default true)
  -f, --format string        Output format: auto, yaml, json, env (default "auto")
  -g, --git                  Enable Git revision comparison support
  -h, --help                 help for sops-diff
  -o, --output string        Save output to file instead of printing to stdout
  -s, --summary              Display only keys that have changed, without sensitive values
  -v, --version              version for sops-diff

Commands:
  git-conflicts FILE        Resolve Git merge conflicts in SOPS-encrypted files
  setup-git-merge-tool      Configure Git to use sops-diff for merge conflict resolution
```

## Basic Usage

### Comparing Two Files

The most basic usage compares two SOPS-encrypted files:

```bash
sops-diff secret1.enc.yaml secret2.enc.yaml
```

This will:
1. Decrypt both files
2. Show the differences between them
3. Format the output as a unified diff

### Summary Mode (Keys Only)

When you want to see which keys have changed without exposing the values (useful for public PR reviews):

```bash
sops-diff --summary secret1.enc.yaml secret2.enc.yaml
```

This only shows which keys were added, removed, or modified, without showing the actual values.

### Specifying File Format

SOPS-Diff automatically detects file formats based on extensions, but you can explicitly specify the format:

```bash
sops-diff --format=json config1.enc.json config2.enc.json
sops-diff --format=yaml config1.enc.yaml config2.enc.yaml
sops-diff --format=env .env.enc .env.prod.enc
```

### Saving Output to File

By default, SOPS-Diff displays results in the terminal, but you can save the output to a file:

```bash
sops-diff file1.enc.yaml file2.enc.yaml --output diff.txt
```

## Git Merge Conflict Resolution

SOPS-Diff provides specialized functionality for handling merge conflicts in encrypted files.

### Manual Conflict Resolution

When you encounter a merge conflict in an encrypted file:

```bash
# By default, output to terminal with colored markers
sops-diff git-conflicts conflicts.enc.yaml

# Save decrypted conflict to a file for editing
sops-diff git-conflicts conflicts.enc.yaml --output conflicts.decrypted.yaml
```

The decrypted output will include conflict markers with branch names where available:

```
<<<<<<< HEAD (main branch)
SECRET_KEY: value_from_current_branch
=======
SECRET_KEY: value_from_other_branch
>>>>>>> OTHER (incoming changes from feature/branch)
```

### Setting Up Git Integration

To configure Git to automatically use SOPS-Diff for merge conflicts:

```bash
# Set up Git configuration
sops-diff setup-git-merge-tool

# Add to your .gitattributes file
*.enc.yaml merge=sops
*.enc.json merge=sops
*.enc.env merge=sops
```

After setup, you can use Git's standard mergetool command:

```bash
git mergetool --tool=sops
```

## Advanced Usage

### Git Integration

#### Comparing with Git Revisions

You can compare files between different Git revisions:

```bash
# Compare current version with HEAD
sops-diff --git HEAD:secrets.enc.yaml secrets.enc.yaml

# Compare between branches
sops-diff main:secrets.enc.yaml feature/new-config:secrets.enc.yaml

# Compare between specific commits
sops-diff abc1234:secrets.enc.yaml def5678:secrets.enc.yaml
```

### Using External Diff Tools

SOPS-Diff can delegate to external diff tools for visualization:

```bash
# Use vimdiff
sops-diff --diff-tool=vimdiff secret1.enc.yaml secret2.enc.yaml

# Use meld (graphical)
sops-diff --diff-tool=meld secret1.enc.yaml secret2.enc.yaml

# Use VSCode
sops-diff --diff-tool="code --diff" secret1.enc.yaml secret2.enc.yaml
```

## Real-World Examples

### Case 1: Adding a New Secret

When adding a new secret to a file:

```bash
# Original file (secrets.enc.yaml)
# Contains: API_KEY, DATABASE_URL

# New file (secrets.new.enc.yaml)
# Contains: API_KEY, DATABASE_URL, NEW_SECRET

sops-diff secrets.enc.yaml secrets.new.enc.yaml
```

Output will show:
```diff
--- a/secrets.enc.yaml
+++ b/secrets.new.enc.yaml
@@ -1,4 +1,5 @@
 API_KEY: "abc123..."
 DATABASE_URL: "postgres://..."
+NEW_SECRET: "xyz789..."
```

### Case 2: Resolving Merge Conflicts

When resolving merge conflicts in encrypted files:

```bash
# After a merge conflict in Git
git merge feature/update-secrets
# [CONFLICT] Merge conflict in secrets.enc.yaml

# Decrypt and view the conflict
sops-diff git-conflicts secrets.enc.yaml

# Output shows decrypted content with colored conflict markers
# Edit and resolve manually, or save to file for editing
sops-diff git-conflicts secrets.enc.yaml --output decrypted.yaml

# Edit decrypted.yaml to resolve conflicts
# Then encrypt and complete the merge
sops -e -i decrypted.yaml
mv decrypted.yaml.enc secrets.enc.yaml
git add secrets.enc.yaml
git merge --continue
```

## CI/CD Integration Examples

### GitHub Actions

For a GitHub Actions workflow that comments on PRs with encrypted file changes:

```yaml
# In your workflow
- name: Generate SOPS diff
  run: |
    sops-diff --summary ${{ github.event.pull_request.base.sha }}:secrets.enc.yaml secrets.enc.yaml --output diff.txt
    
- name: Comment on PR
  uses: actions/github-script@v6
  with:
    github-token: ${{ secrets.GITHUB_TOKEN }}
    script: |
      const fs = require('fs');
      const diff = fs.readFileSync('diff.txt', 'utf8');
      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: '### SOPS-Diff Summary\n```diff\n' + diff + '\n```'
      });
```

## Working with Different Formats

### YAML Files

```bash
sops-diff config1.enc.yaml config2.enc.yaml
```

### JSON Files

```bash
sops-diff config1.enc.json config2.enc.json
```

### Environment Files

```bash
sops-diff .env.enc .env.prod.enc
```

## Tips and Best Practices

1. **Use colored output for better readability**
   - The `--color` flag is enabled by default for terminal output
   - Colors are automatically disabled when output is redirected

2. **Use subcommand flags correctly**
   - Flags for subcommands must be placed after the subcommand
   - Example: `sops-diff git-conflicts file.yaml --output out.yaml`

3. **For merge conflicts, get branch information**
   - The `git-conflicts` command shows branch names in conflict markers when available

4. **Use output files for sensitive content**
   - `--output` flag redirects content to a file instead of displaying on screen

5. **Remember to clean up decrypted files**
   - Always delete decrypted files after use to avoid security risks

6. **When using Git integration, ensure you have access to the necessary keys**
   - Git operations might need access to KMS or other key management systems

7. **For large files, consider using an external diff tool**
   - `--diff-tool=meld` or similar for better visualization
