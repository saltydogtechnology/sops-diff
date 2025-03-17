# Installing SOPS-Diff

This document provides detailed installation instructions for SOPS-Diff on various platforms.

## Prerequisites

Before installing SOPS-Diff, ensure you have:

1. **SOPS** installed (v3.7.x or higher)
   - [SOPS Installation Guide](https://github.com/getsops/sops#installing-from-source)

2. **Required Access**
   - For AWS KMS: Proper IAM permissions for the KMS keys you're using
   - For GCP KMS, PGP, or Age: Appropriate credentials configured

## Installation Methods

### Method 1: Pre-built Binaries

#### Linux

```bash
export SOPS_DIFF_VERSION=v0.2.0
```

```bash
# For amd64 architecture
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/$SOPS_DIFF_VERSION/sops-diff-$SOPS_DIFF_VERSION-linux-amd64.tar.gz | tar xz
chmod +x sops-diff-linux-amd64
sudo mv sops-diff-linux-amd64 /usr/local/bin/sops-diff

# For arm64 architecture
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/$SOPS_DIFF_VERSION/sops-diff-$SOPS_DIFF_VERSION-linux-arm64.tar.gz | tar xz
chmod +x sops-diff-linux-arm64
sudo mv sops-diff-linux-arm64 /usr/local/bin/sops-diff
```

#### macOS

```bash
export SOPS_DIFF_VERSION=v0.2.0
```

```bash
# For amd64 architecture
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/$SOPS_DIFF_VERSION/sops-diff-$SOPS_DIFF_VERSION-darwin-amd64.tar.gz | tar xz
chmod +x sops-diff-darwin-amd64
sudo mv sops-diff-darwin-amd64 /usr/local/bin/sops-diff

# For arm64 architecture (Apple Silicon)
curl -L https://github.com/saltydogtechnology/sops-diff/releases/download/$SOPS_DIFF_VERSION/sops-diff-$SOPS_DIFF_VERSION-darwin-arm64.tar.gz| tar xz
chmod +x sops-diff-darwin-arm64
sudo mv sops-diff-darwin-arm64 /usr/local/bin/sops-diff

```

#### Windows

1. Download the appropriate binary from the [releases page](https://github.com/saltydogtechnology/sops-diff/releases)
   - For 64-bit systems: `sops-diff-windows-amd64.exe`

2. Rename the file to `sops-diff.exe`

3. Add the file to a location in your PATH:
   - Create a directory like `C:\Tools` if it doesn't exist
   - Move `sops-diff.exe` to this directory
   - Add the directory to your PATH environment variable

### Method 2: Building from Source

If you prefer to build from source, follow these steps:

#### Requirements

- Go 1.23.3 or higher
- Git

#### Build Steps

```bash
# Clone the repository
git clone https://github.com/saltydogtechnology/sops-diff.git
cd sops-diff

# Build the binary
go build -o sops-diff .

# Install it
sudo mv sops-diff /usr/local/bin/
```

## Verifying Installation

After installation, verify that SOPS-Diff is correctly installed:

```bash
sops-diff --version
```

You should see output indicating the version of SOPS-Diff.

## Configuration

### Environment Variables

SOPS-Diff respects the same environment variables as SOPS, including:

- `SOPS_KMS_ARN`: AWS KMS key ARN
- `SOPS_PGP_FP`: PGP fingerprint
- `SOPS_AGE_RECIPIENTS`: Age recipients
- `SOPS_GCP_KMS_IDS`: Google Cloud KMS key IDs

See the [SOPS documentation](https://github.com/getsops/sops#encrypting-using-sops) for details on these environment variables.

### Git Configuration

To set up Git integration, follow these steps:

1. Add a Git attributes configuration:

   ```bash
   echo '*.enc.yaml diff=sopsdiffer' >> ~/.gitattributes
   echo '*.enc.json diff=sopsdiffer' >> ~/.gitattributes
   echo '*.enc.yml diff=sopsdiffer' >> ~/.gitattributes
   echo '*.enc.env diff=sopsdiffer' >> ~/.gitattributes
   git config --global core.attributesfile ~/.gitattributes
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

   ```bash
   # Option 3: Allow viewing diffs of decrypted files (disable the default safety feature)
   git config diff.sopsdiffer.command "sops-diff --git --error-on-decrypted=false"
   ```

## Troubleshooting

### Common Issues

1. **SOPS decryption failures**
   - Ensure you have proper access to the encryption keys
   - Check that the SOPS configuration file (`.sops.yaml`) is correctly set up
   - Verify that the environment has the necessary credentials

2. **Git integration not working**
   - Confirm that the Git attributes and diff driver are correctly configured
   - Check file patterns in your `.gitattributes` file

3. **Format detection issues**
   - Use the `--format` flag to explicitly specify the file format

4. **Decrypted file errors**
   - By default, sops-diff errors when it detects decrypted files
   - Use `--error-on-decrypted=false` to disable this safety feature when needed

### Getting Help

If you encounter issues not covered here:

1. Check the [GitHub Issues](https://github.com/saltydogtechnology/sops-diff/issues) for similar problems
2. Open a new issue with details about your environment and the problem

## Upgrading

To upgrade SOPS-Diff to the latest version:

```bash
# Using pre-built binaries
curl -L https://github.com/saltydogtechnology/sops-diff/releases/latest/download/sops-diff-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o /usr/local/bin/sops-diff
chmod +x /usr/local/bin/sops-diff

# Using Go
go install github.com/saltydogtechnology/sops-diff@latest

# Using Homebrew
brew upgrade sops-diff
```
