# Working with SOPS-Diff and GPG Encryption: A Practical Guide

This guide demonstrates how to create encrypted secrets using SOPS with GPG keys and how to compare them using the `sops-diff` utility.

## 1. Generate a GPG Key

First, let's create a GPG key that we'll use for encryption:

```bash
# Generate a new GPG key (non-interactive mode)
gpg --batch --gen-key <<EOF
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Test User
Name-Email: test@example.com
Expire-Date: 0
%no-protection
EOF

# List the keys to get the fingerprint
gpg --list-keys test@example.com
```

Note the fingerprint from the output, which looks something like: `1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S0`

## 2. Configure SOPS

Create a `.sops.yaml` configuration file in your project root:

```bash
cat > .sops.yaml <<EOF
creation_rules:
  - pgp: 'YOUR_FINGERPRINT_HERE'  # Replace with your actual fingerprint
EOF
```

## 3. Create and Encrypt Files

### Create and encrypt YAML files

```bash
# Create a YAML file with secrets
cat > secrets.yaml <<EOF
database:
  username: admin
  password: super_secret_password
api:
  key: 1234567890abcdef
  endpoint: https://api.example.com
EOF

# Encrypt the file with SOPS
sops -e -i secrets.yaml
```

### Create and encrypt JSON files

```bash
# Create a JSON file with secrets
cat > secrets.json <<EOF
{
  "database": {
    "username": "admin",
    "password": "super_secret_password"
  },
  "api": {
    "key": "1234567890abcdef",
    "endpoint": "https://api.example.com"
  }
}
EOF

# Encrypt the file with SOPS
sops -e -i secrets.json
```

### Create and encrypt .env files

```bash
# Create a .env file with secrets
cat > .env <<EOF
DB_USERNAME=admin
DB_PASSWORD=super_secret_password
API_KEY=1234567890abcdef
API_ENDPOINT=https://api.example.com
EOF

# Encrypt the file with SOPS
sops -e -i .env
```

### Create a modified version for comparison

```bash
# Example with YAML
sops -d secrets.yaml > secrets_temp.yaml
# sed -i 's/super_secret_password/new_password_123/' secrets_temp.yaml  # Linux
sed -i '' 's/super_secret_password/new_password_123/' secrets_temp.yaml  # macOS
sops -e secrets_temp.yaml > secrets_new.yaml
rm secrets_temp.yaml

# Example with JSON
sops -d secrets.json > secrets_temp.json
# sed -i 's/super_secret_password/new_password_123/' secrets_temp.json  # Linux
sed -i '' 's/super_secret_password/new_password_123/' secrets_temp.json  # macOS
sops -e secrets_temp.json > secrets_new.json
rm secrets_temp.json

# Example with .env
sops -d .env > env_temp
# sed -i 's/DB_PASSWORD=.*/DB_PASSWORD=new_password_123/' env_temp  # Linux
sed -i '' 's/DB_PASSWORD=.*/DB_PASSWORD=new_password_123/' env_temp  # macOS
sops -e env_temp > env_new
rm env_temp
```

## 4. Using SOPS-Diff to Compare Files

### Full Comparison Mode

The default mode shows both keys and their values:

#### YAML Example

```bash
sops-diff secrets.yaml secrets_new.yaml
```

Output:
```diff
--- a/secrets.yaml
+++ b/secrets_new.yaml
@@ -2,6 +2,6 @@
     endpoint: https://api.example.com
     key: 1234567890abcdef
 database:
-    password: super_secret_password
+    password: new_password_123
     username: admin
```

#### JSON Example

```bash
sops-diff secrets.json secrets_new.json
```

Output:
```diff
--- a/secrets.json
+++ b/secrets_new.json
@@ -1,7 +1,7 @@
 {
   "database": {
     "username": "admin",
-    "password": "super_secret_password"
+    "password": "new_password_123"
   },
   "api": {
     "key": "1234567890abcdef",
```

#### .env Example

```bash
sops-diff .env env_new
```

Output:
```diff
--- a/.env
+++ b/env_new
@@ -1,4 +1,4 @@
 API_ENDPOINT=https://api.example.com
 API_KEY=1234567890abcdef
-DB_PASSWORD=super_secret_password
+DB_PASSWORD=new_password_123
 DB_USERNAME=admin
```

### Summary Mode

Use the `--summary` flag to show only which keys changed without revealing values:

```bash
# YAML example
sops-diff --summary secrets.yaml secrets_new.yaml

# JSON example
sops-diff --summary secrets.json secrets_new.json

# .env example
sops-diff --summary .env env_new
```

Example output for YAML/JSON:
```
Summary of key changes:
! = modified key, + = added key, - = removed key
--------------------------------------
! database.password
```

Example output for .env:
```
Summary of key changes:
! = modified key, + = added key, - = removed key
--------------------------------------
! DB_PASSWORD
```

### Explicit Format Specification

You can specify the format explicitly:

```bash
# YAML format (explicit)
sops-diff --format=yaml secrets.yaml secrets_new.yaml

# JSON format (explicit)
sops-diff --format=json secrets.json secrets_new.json

# ENV format (explicit)
sops-diff --format=env .env env_new
```

### Git Integration

If you're using Git, you can compare with previous versions:

```bash
# Add and commit the original file
git add secrets.yaml
git commit -m "Add encrypted secrets"

# After making changes to the file, compare with the committed version
sops-diff --git HEAD:secrets.yaml secrets.yaml
```

## 5. Setting Up Git Integration

For seamless integration with Git, add these configurations:

```bash
# Add Git attributes for different file types
cat > .gitattributes <<EOF
*.yaml diff=sopsdiffer
*.yml diff=sopsdiffer
*.json diff=sopsdiffer
*.env diff=sopsdiffer
EOF

# Configure Git to use sops-diff

# Option 1: Full diff (shows all values)
git config diff.sopsdiffer.command "sops-diff --git"

# Option 2: Summary mode (only shows which keys changed without revealing values)
# This is more secure for public code reviews or when working in environments
# where you want to avoid exposing sensitive data
git config diff.sopsdiffer.command "sops-diff --git --summary"

# Option 3: Allow viewing diffs of decrypted files (disable the default safety feature)
git config diff.sopsdiffer.command "sops-diff --git --error-on-decrypted=false"
```

Now you can use regular `git diff` commands and Git will automatically use `sops-diff` for encrypted files.

## 6. Additional Usage Examples

### Use External Diff Tools

```bash
# Use vimdiff with YAML files
sops-diff --diff-tool=vimdiff secrets.yaml secrets_new.yaml

# Use meld with JSON files
sops-diff --diff-tool=meld secrets.json secrets_new.json

# Use Visual Studio Code with .env files
sops-diff --diff-tool="code --diff" .env env_new
```

### Combining Options

You can combine various options:

```bash
# Summary mode with explicit format
sops-diff --summary --format=json secrets.json secrets_new.json

# External diff tool with explicit format
sops-diff --diff-tool=vimdiff --format=env .env env_new

# Disable color output
sops-diff --color=false secrets.yaml secrets_new.yaml
```

## 7. Security Considerations

- The `sops-diff` utility never writes decrypted content to disk, only to memory
- For public code reviews, always use the `--summary` mode
- The tool respects your SOPS configuration and key management setup
- Always treat secret files with appropriate caution, even when encrypted
- When using external diff tools, be aware that they might cache or store temporary files

By using `sops-diff`, you can safely review changes to encrypted files without exposing sensitive information during the code review process.
