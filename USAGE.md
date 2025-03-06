# SOPS-Diff Usage Guide

This guide provides comprehensive instructions for using the SOPS-Diff utility for comparing encrypted files.

## Command-line Options

```
sops-diff [flags] FILE1 FILE2

Flags:
  -s, --summary        Display only keys that have changed, without sensitive values
  -f, --format string  Output format: auto, yaml, json, env (default "auto")
  -c, --color          Use colored output when supported (default true)
  -d, --diff-tool      Use an external diff tool (e.g. 'vimdiff')
  -g, --git            Enable Git revision comparison support
  -h, --help           Help for sops-diff
  -v, --version        Version for sops-diff
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

#### Setting Up Git Integration

1. Configure Git attributes:

   Add to your `.gitattributes` file:
   ```
   *.enc.json diff=sopsdiffer
   *.enc.yaml diff=sopsdiffer
   *.enc.yml diff=sopsdiffer
   *.enc.env diff=sopsdiffer
   ```

2. Configure Git to use SOPS-Diff:

   ```bash
   # Option 1: Full diff (shows all values)
   git config diff.sopsdiffer.command "sops-diff --git"
   ```

   ```bash
   # Option 2: Summary mode (only shows which keys changed without revealing values)
   # This is more secure for public code reviews or when working in environments
   # where you want to avoid exposing sensitive data
   git config diff.sopsdiffer.command "sops-diff --git --summary"
   ```

3. Now `git diff` will automatically use SOPS-Diff for encrypted files.

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

### Case 2: Reviewing PR Changes

When reviewing changes in a pull request:

```bash
# Use summary mode for public PR reviews
sops-diff --summary origin/main:secrets.enc.yaml feature/update-secrets:secrets.enc.yaml
```

Output will only show key changes:
```diff
--- a/secrets.enc.yaml
+++ b/secrets.enc.yaml
@@ -1,4 +1,4 @@
 API_KEY
 DATABASE_URL
-OLD_SECRET
+NEW_SECRET
```

### Case 3: Comparing Different Environments

Compare secrets between environments:

```bash
sops-diff staging/secrets.enc.yaml production/secrets.enc.yaml
```

This helps identify configuration differences between environments.

## CI/CD Integration Examples

### GitHub Actions

For a GitHub Actions workflow that comments on PRs with encrypted file changes:

```yaml
# In your workflow
- name: Generate SOPS diff
  run: |
    sops-diff --summary ${{ github.event.pull_request.base.sha }}:secrets.enc.yaml secrets.enc.yaml > diff.txt
    
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

### GitLab CI

For GitLab CI integration:

```yaml
# In your .gitlab-ci.yml
sops-diff-job:
  script:
    - sops-diff --summary origin/main:secrets.enc.yaml secrets.enc.yaml > diff.txt
    - 'curl --request POST --header "PRIVATE-TOKEN: $CI_API_TOKEN" --data-urlencode "body=### SOPS-Diff Summary\n```diff\n$(cat diff.txt)\n```" "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${CI_MERGE_REQUEST_IID}/notes"'
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

1. **Use summary mode for public code reviews**
   - `--summary` flag shows only key changes, not values

2. **Use color output for better readability**
   - The `--color` flag is enabled by default for terminal output

3. **When using Git integration, ensure you have access to the necessary keys**
   - Git operations might need access to KMS or other key management systems

4. **For large files, consider using an external diff tool**
   - `--diff-tool=meld` or similar for better visualization

5. **In CI/CD environments, use environment variables for key access**
   - Set up `SOPS_KMS_ARN`, `SOPS_AGE_RECIPIENTS`, etc. in your CI environment

6. **For complex diffs, pipe output to a file for easier review**
   - `sops-diff file1.enc.yaml file2.enc.yaml > diff.txt`

7. **When comparing files with many changes, use context to focus on specific sections**
   - Consider using grep or other tools to focus on relevant parts of the diff
