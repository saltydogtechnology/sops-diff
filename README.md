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
  - Resolve merge conflicts in encrypted files with `git-conflicts` command
  - Set up custom Git merge tool for encrypted files with `setup-git-merge-tool` command
- **Output Options**:
  - Color-coded output for better readability in terminal
  - Save results to file with `--output` flag
- **Security-Focused**:
  - No decrypted content written to disk by default
  - Minimized exposure of secrets
  - Automatic detection and warning for decrypted files

## Installation

For detailed installation instructions for different platforms, please refer to [INSTALLATION.md](INSTALLATION.md).

### Quick Start

```
# Linux (amd64)
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/v0.2.0/sops-diff-v0.2.0-linux-amd64.tar.gz | tar xz
sudo mv sops-diff-linux-amd64 /usr/local/bin/sops-diff
```

```
# macOS (amd64)
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/v0.2.0/sops-diff-v0.2.0-darwin-amd64.tar.gz | tar xz
sudo mv sops-diff-darwin-amd64 /usr/local/bin/sops-diff
```

```
# macOS (Apple Silicon)
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/v0.2.0/sops-diff-v0.2.0-darwin-arm64.tar.gz | tar xz
sudo mv sops-diff-darwin-arm64 /usr/local/bin/sops-diff
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

# Save output to file
sops-diff secret1.enc.yaml secret2.enc.yaml --output diff.txt
```

### Git Integration

```bash
# Compare between different Git revisions
sops-diff --git HEAD:secrets.enc.yaml secrets.enc.yaml

# Compare between branches
sops-diff --git main:secrets.enc.yaml feature/new-secret:secrets.enc.yaml
```

### Resolving Merge Conflicts

```bash
# Display decrypted conflict with syntax highlighting
sops-diff git-conflicts conflicts.enc.yaml

# View conflicts in Git diff format rather than with conflict markers
sops-diff git-conflicts conflicts.enc.yaml --view-as-diff

# Save to a file for editing
sops-diff git-conflicts conflicts.enc.yaml --output resolved.yaml
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

### 1. Configure Git for Diff and Merge Operations

```bash
# Set up Git integration
sops-diff setup-git-merge-tool
```

### 2. Add a Git Attributes Configuration

Add the following to your repository's `.gitattributes` file:

```
*.enc.json diff=sopsdiffer merge=sops
*.enc.yaml diff=sopsdiffer merge=sops
*.enc.yml diff=sopsdiffer merge=sops
*.enc.env diff=sopsdiffer merge=sops
```

### 3. Configure Git to Use SOPS-Diff for Diff Operations

   ```bash
   # Option: Full diff (shows all values)
   git config diff.sopsdiffer.command "sops-diff --git"
   ```

   ```bash
   # Alternative: Summary mode (only keys changed without values)
   git config diff.sopsdiffer.command "sops-diff --git --summary"
   ```

After this setup, `git diff` will automatically use SOPS-Diff for files matching the patterns, and merge conflicts will be handled with the `git mergetool --tool=sops` command.

## Security Considerations

- SOPS-Diff does not write decrypted content to disk by default
- In CI/CD environments, use `--summary` mode to avoid exposing sensitive values
- The tool works with your existing SOPS encryption/decryption configuration (AWS KMS, GCP KMS, age, PGP)
- Ensure proper access controls for CI/CD environments that need to decrypt files

## Requirements

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