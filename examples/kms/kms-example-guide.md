# Working with SOPS-Diff and AWS KMS Encryption: A Practical Guide

This guide demonstrates how to create encrypted secrets using SOPS with AWS KMS keys and how to compare them using the `sops-diff` utility.

## 1. Set Up AWS KMS

First, ensure you have AWS CLI installed and configured:

```bash
# Verify AWS CLI is installed
aws --version

# Configure AWS CLI with your credentials if not already done
aws configure
```

Next, create a KMS key for encrypting your secrets:

```bash
# Create a new KMS key
aws kms create-key --description "Key for SOPS encryption"

# The output will include a KeyId, which you'll need for the next steps
# Example: "KeyId": "arn:aws:kms:us-east-1:123456789012:key/abcd1234-ab12-cd34-ef56-abcdef123456"
```

Optionally, create an alias for easier reference:

```bash
aws kms create-alias \
    --alias-name alias/sops-key \
    --target-key-id arn:aws:kms:us-east-1:123456789012:key/abcd1234-ab12-cd34-ef56-abcdef123456
```

## 2. Configure SOPS

Create a `.sops.yaml` configuration file in your project root:

```bash
cat > .sops.yaml <<EOF
creation_rules:
  - kms: 'arn:aws:kms:us-east-1:123456789012:key/abcd1234-ab12-cd34-ef56-abcdef123456'
EOF
```

You can also specify the key by its alias:

```bash
cat > .sops.yaml <<EOF
creation_rules:
  - kms: 'alias/sops-key'
EOF
```

## 3. Create and Encrypt Files

### Create the original secrets file

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

### Create a second file with modified secrets

```bash
# Decrypt to a temporary file
sops -d secrets.yaml > secrets_temp.yaml

# Modify a value
sed -i 's/super_secret_password/new_password_123/' secrets_temp.yaml
# For macOS use: sed -i '' 's/super_secret_password/new_password_123/' secrets_temp.yaml

# Encrypt the modified file
sops -e secrets_temp.yaml > secrets_new.yaml

# Clean up the temporary unencrypted file
rm secrets_temp.yaml
```

## 4. Using SOPS-Diff to Compare Files

### Full Comparison Mode

The default mode shows both keys and their values:

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

### Summary Mode

Use the `--summary` flag to show only which keys changed without revealing values:

```bash
sops-diff --summary secrets.yaml secrets_new.yaml
```

Output:
```
Summary of key changes:
! = modified key, + = added key, - = removed key
--------------------------------------
! database.password
```

### Git Integration

If you're using Git, you can compare with previous versions:

```bash
# Add and commit the original file
git add secrets.yaml
git commit -m "Add encrypted secrets"

# After making changes to the file
sops-diff --git HEAD:secrets.yaml secrets.yaml
```

## 5. Setting Up Git Integration

For seamless integration with Git, add these configurations:

```bash
# Add Git attributes
cat > .gitattributes <<EOF
*.yaml diff=sopsdiffer
*.yml diff=sopsdiffer
*.json diff=sopsdiffer
EOF

# Configure Git to use sops-diff

# Option 1: Full diff (shows all values)
git config diff.sopsdiffer.command "sops-diff --git"

# Option 2: Summary mode (only shows which keys changed without revealing values)
# This is more secure for public code reviews or when working in environments
# where you want to avoid exposing sensitive data
git config diff.sopsdiffer.command "sops-diff --git --summary"
```

Now you can use regular `git diff` commands and Git will automatically use `sops-diff` for encrypted files.

## 6. AWS KMS-Specific Features

### Using Multiple AWS Regions

You can specify KMS keys from different regions:

```yaml
creation_rules:
  - path_regex: us-secrets\.yaml$
    kms: 'arn:aws:kms:us-east-1:123456789012:key/abcd1234-ab12-cd34-ef56-abcdef123456'
  - path_regex: eu-secrets\.yaml$
    kms: 'arn:aws:kms:eu-west-1:123456789012:key/efgh5678-ef56-gh78-ij90-efghij456789'
```

### Using KMS with AWS Profiles

If you have multiple AWS profiles, you can specify which one to use:

```bash
# Set the AWS profile
export AWS_PROFILE=production

# Then run sops-diff
sops-diff secrets.yaml secrets_new.yaml
```

### Cross-Account KMS Keys

For cross-account scenarios, ensure your IAM role has permissions to use the KMS key:

```bash
# Example policy allowing another account to use the key
aws kms create-grant \
    --key-id arn:aws:kms:us-east-1:123456789012:key/abcd1234-ab12-cd34-ef56-abcdef123456 \
    --grantee-principal arn:aws:iam::987654321098:role/YourCrossAccountRole \
    --operations Decrypt Encrypt ReEncryptFrom ReEncryptTo
```

## 7. CI/CD Integration for AWS KMS

For GitHub Actions, add these steps to your workflow:

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v1
  with:
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1

- name: Generate SOPS diff
  run: |
    sops-diff --summary ${{ github.event.pull_request.base.sha }}:secrets.enc.yaml secrets.enc.yaml > diff.txt
```

For GitLab CI:

```yaml
sops-diff-job:
  image: aws-cli:latest
  script:
    - aws configure set aws_access_key_id $AWS_ACCESS_KEY_ID
    - aws configure set aws_secret_access_key $AWS_SECRET_ACCESS_KEY
    - aws configure set region us-east-1
    - sops-diff --summary origin/main:secrets.enc.yaml secrets.enc.yaml > diff.txt
```

## 8. Security Considerations for AWS KMS

- Use IAM roles with the principle of least privilege
- Consider using key policies to restrict which IAM principals can use the KMS key
- Regularly rotate AWS KMS keys
- For highly sensitive environments, consider using AWS CloudTrail to audit KMS key usage
- Use separate KMS keys for different environments (dev, staging, production)

## 9. Troubleshooting AWS KMS Issues

### Permission Errors

If you see errors like `KMS access denied`, check:

```bash
# Verify you have access to the KMS key
aws kms describe-key --key-id arn:aws:kms:us-east-1:123456789012:key/abcd1234-ab12-cd34-ef56-abcdef123456

# Check which IAM identity you're using
aws sts get-caller-identity
```

### Region Issues

If your KMS key is in a different region:

```bash
# Explicitly set the region
export AWS_REGION=us-east-1
sops-diff secrets.yaml secrets_new.yaml
```

By using `sops-diff` with AWS KMS, you can safely review changes to encrypted files without exposing sensitive information during the code review process, while leveraging AWS's robust key management infrastructure.
