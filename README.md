# SOPS-Diff

A command-line utility to simplify the code review process for SOPS-encrypted secrets by automatically decrypting and displaying differences between files.

## Overview

SOPS-Diff addresses a common challenge in secure code review: reviewing changes to encrypted files. By automating the decryption and comparison of SOPS-encrypted files, this tool makes it easier to identify and verify changes to sensitive configuration without exposing those secrets to unnecessary risk.

## Features

- **Automated Comparison**: Decrypt two SOPS-encrypted files and show only their differences
- **Multiple Format Support**: Parse and compare YAML, JSON, and ENV file formats
- **Output Modes**:
  - Full mode: Show both keys and values that have changed
  - Summary mode: Display only the changed keys without sensitive values
- **Git Integration**:
  - Compare between Git revisions (e.g., `sops-diff --git HEAD:secrets.enc.yaml secrets.enc.yaml`)
  - Git attribute support for automatic invocation during `git diff`
- **CI/CD Integration**: Generate diff reports in formats suitable for PR comments
- **Security-Focused**:
  - No decrypted content written to disk
  - Minimized exposure of secrets
- **Security-Focused**:
  - No decrypted content written to disk
  - Minimized exposure of secrets
  - Automatic detection and warning for decrypted files

## Installation

### From Binary Releases

```bash
# For Linux (amd64)
curl -L https://github.com/saltydogtechnology/sops-diff/releases/latest/download/sops-diff-linux-amd64 -o /usr/local/bin/sops-diff
chmod +x /usr/local/bin/sops-diff

# For macOS (amd64)
curl -L https://github.com/saltydogtechnology/sops-diff/releases/latest/download/sops-diff-darwin-amd64 -o /usr/local/bin/sops-diff
chmod +x /usr/local/bin/sops-diff

# For Windows (amd64)
# Download from https://github.com/saltydogtechnology/sops-diff/releases/latest/download/sops-diff-windows-amd64.exe
```

### From Source

```bash
go install github.com/saltydogtechnology/sops-diff@latest
```

## Usage

### Basic Usage

```bash
# Compare two encrypted files
sops-diff secret1.enc.yaml secret2.enc.yaml

# Show only keys that have changed (without values)
sops-diff --summary secret1.enc.yaml secret2.enc.yaml

# Compare different formats
sops-diff --format=json config1.enc.json config2.enc.json
```

### Git Integration

```bash
# Compare between different Git revisions
sops-diff --git HEAD:secrets.enc.yaml secrets.enc.yaml

# Compare between branches
sops-diff main:secrets.enc.yaml feature/new-secret:secrets.enc.yaml
```

### Using External Diff Tools

```bash
# Use your preferred diff tool
sops-diff --diff-tool=vimdiff secret1.enc.yaml secret2.enc.yaml

# Use with a graphical diff tool
sops-diff --diff-tool=meld secret1.enc.yaml secret2.enc.yaml
```

### Supported Formats

- YAML (`.yaml`, `.yml`)
- JSON (`.json`)
- Environment files (`.env`)

## Setting Up Git Integration

### 1. Add a Git Attributes Configuration

Add the following to your repository's `.gitattributes` file:

```
*.enc.json diff=sopsdiffer
*.enc.yaml diff=sopsdiffer
*.enc.yml diff=sopsdiffer
*.enc.env diff=sopsdiffer
```

### 2. Configure Git to Use SOPS-Diff

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

After this setup, `git diff` will automatically use SOPS-Diff for files matching the patterns.

## CI/CD Integration

### GitHub Actions

A sample GitHub Actions workflow is provided in `examples/github/workflows/sops-diff.yaml` that:

1. Detects changes to encrypted files in pull requests
2. Generates diffs using SOPS-Diff
3. Posts a summary of changes as a PR comment

See the [GitHub Workflow](./examples/github/workflows/sops-diff.yaml) file for details.

### GitLab CI

A sample GitLab CI configuration is provided in `examples/gitlab/sops-diff.yaml` that:

1. Detects changes to encrypted files in merge requests
2. Generates diffs using SOPS-Diff
3. Saves the summary of changes as an artifact

See the [GitLab CI Configuration](./examples/gitlab/sops-diff.yaml) file for details.

## Security Considerations

- SOPS-Diff does not write decrypted content to disk
- In CI/CD environments, use `--summary` mode to avoid exposing sensitive values
- The tool works with your existing SOPS encryption/decryption configuration (AWS KMS, GCP KMS, age, PGP)
- Ensure proper access controls for CI/CD environments that need to decrypt files

## Requirements

- Go 1.17 or higher
- SOPS v3.7.x or higher
- For Git integration: Git 2.20.0 or higher

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.