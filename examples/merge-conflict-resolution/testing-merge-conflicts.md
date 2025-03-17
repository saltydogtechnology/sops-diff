# Testing SOPS-Diff Merge Conflict Resolution

This document provides test scenarios for verifying the merge conflict resolution functionality of the SOPS-Diff tool with encrypted files in various formats (YAML, JSON, and ENV).

## Prerequisites

- SOPS installed and configured with access to encryption keys
- Git repository initialized
- SOPS-Diff built with merge conflict resolution feature

## Basic Merge Conflict Test

This sequence creates a simple merge conflict in a SOPS-encrypted file:

### YAML Example

```bash
# 1. Create a base branch for testing
git checkout -b base

# 2. Create a test encrypted file with multiple values
cat > test_secrets.yaml << EOF
SECRET_KEY: my_secret_value
API_URL: https://api.example.com
DEBUG_MODE: false
DATABASE:
  USER: db_user
  HOST: db.example.com
  PORT: 5432
EOF
sops -e -i test_secrets.yaml

# 3. Add and commit the initial version
git add test_secrets.yaml
git commit -m "Add initial encrypted secrets"

# 4. Create first test branch
git checkout -b test/merge-conflicts-1

# 5. Modify the encrypted file in this branch
sops -d test_secrets.yaml > temp.yaml
cat > temp.yaml << EOF
SECRET_KEY: modified_in_branch1
API_URL: https://api.example.com
DEBUG_MODE: true
EXTRA_1: extra_in_branch1
DATABASE:
  USER: db_user
  HOST: db.example.com
  PORT: 5432
EOF
sops -e temp.yaml > test_secrets.yaml
rm temp.yaml
git add test_secrets.yaml
git commit -m "Change secret in branch1"

# 6. Return to base branch
git checkout base

# 7. Create second test branch
git checkout -b test/merge-conflicts-2

# 8. Modify the same file differently in this branch
sops -d test_secrets.yaml > temp.yaml
cat > temp.yaml << EOF
SECRET_KEY: modified_in_branch2
API_URL: https://api-v2.example.com
DEBUG_MODE: false
DATABASE:
  USER: db_user_new
  HOST: db.example.com
  PORT: 5432
EOF
sops -e temp.yaml > test_secrets.yaml
rm temp.yaml
git add test_secrets.yaml
git commit -m "Change secret in branch2"

# 9. Merge the first branch into base
git checkout base
git merge test/merge-conflicts-1

# 10. Try to merge the second branch - this should cause a conflict
git merge test/merge-conflicts-2
```

After running these commands, Git will report a merge conflict in the respective file. You can now test the tool with the `git-conflicts` command:

```bash
# View conflict in console with colored output (default)
sops-diff git-conflicts test_secrets.yaml  # for YAML

# View conflicts in Git diff format rather than with conflict markers
sops-diff git-conflicts test_secrets.yaml --view-as-diff

# Or save to file with --output flag
sops-diff git-conflicts test_secrets.yaml --output diff.yaml
```

### JSON Example

```bash
# 1. Create a base branch for testing (if needed)
git checkout -b base-json

# 2. Create a test encrypted JSON file with multiple values
cat > test_secrets.json << EOF
{
  "SECRET_KEY": "my_secret_value",
  "API_URL": "https://api.example.com",
  "DEBUG_MODE": false,
  "DATABASE": {
    "USER": "db_user",
    "HOST": "db.example.com",
    "PORT": 5432
  }
}
EOF
sops -e -i test_secrets.json

# 3. Add and commit the initial version
git add test_secrets.json
git commit -m "Add initial encrypted JSON secrets"

# 4. Create first test branch
git checkout -b test/json-conflicts-1

# 5. Modify the encrypted file in this branch
sops -d test_secrets.json > temp.json
cat > temp.json << EOF
{
  "SECRET_KEY": "modified_in_branch1",
  "API_URL": "https://api.example.com",
  "DEBUG_MODE": true,
  "DATABASE": {
    "USER": "db_user",
    "HOST": "db.example.com",
    "PORT": 5432
  }
}
EOF
sops -e temp.json > test_secrets.json
rm temp.json
git add test_secrets.json
git commit -m "Change JSON secret in branch1"

# 6. Return to base branch
git checkout base-json

# 7. Create second test branch
git checkout -b test/json-conflicts-2

# 8. Modify the same file differently in this branch
sops -d test_secrets.json > temp.json
cat > temp.json << EOF
{
  "SECRET_KEY": "modified_in_branch2",
  "API_URL": "https://api-v2.example.com",
  "DEBUG_MODE": false,
  "DATABASE": {
    "USER": "db_user_new",
    "HOST": "db.example.com",
    "PORT": 5432
  }
}
EOF
sops -e temp.json > test_secrets.json
rm temp.json
git add test_secrets.json
git commit -m "Change JSON secret in branch2"

# 9. Merge the first branch into base
git checkout base-json
git merge test/json-conflicts-1

# 10. Try to merge the second branch - this should cause a conflict
git merge test/json-conflicts-2
```

After running these commands, Git will report a merge conflict in the respective file. You can now test the tool with the `git-conflicts` command:

```bash
# View conflict in console with colored output (default)
sops-diff git-conflicts test_secrets.json  # for JSON

# View conflicts in Git diff format
sops-diff git-conflicts test_secrets.json --view-as-diff

# Or save to file with --output flag
sops-diff git-conflicts test_secrets.yaml --output diff.yaml
```

### ENV Example

```bash
# 1. Create a base branch for testing
git checkout -b base-env

# 2. Create a test encrypted ENV file with multiple values
cat > test_secrets.env << EOF
SECRET_KEY=my_secret_value
API_URL=https://api.example.com
DEBUG_MODE=false
DB_USER=db_user
DB_HOST=db.example.com
DB_PORT=5432
EOF
sops -e -i test_secrets.env

# 3. Add and commit the initial version
git add test_secrets.env
git commit -m "Add initial encrypted ENV secrets"

# 4. Create first test branch
git checkout -b test/env-conflicts-1

# 5. Modify the encrypted file in this branch
sops -d test_secrets.env > temp.env
cat > temp.env << EOF
SECRET_KEY=modified_in_branch1
API_URL=https://api.example.com
DEBUG_MODE=true
DB_USER=db_user
DB_HOST=db.example.com
DB_PORT=5432
EOF
sops -e temp.env > test_secrets.env
rm temp.env
git add test_secrets.env
git commit -m "Change ENV secret in branch1"

# 6. Return to base branch
git checkout base-env

# 7. Create second test branch
git checkout -b test/env-conflicts-2

# 8. Modify the same file differently in this branch
sops -d test_secrets.env > temp.env
cat > temp.env << EOF
SECRET_KEY=modified_in_branch2
API_URL=https://api-v2.example.com
DEBUG_MODE=false
DB_USER=db_user_new
DB_HOST=db.example.com
DB_PORT=5432
EOF
sops -e temp.env > test_secrets.env
rm temp.env
git add test_secrets.env
git commit -m "Change ENV secret in branch2"

# 9. Merge the first branch into base
git checkout base-env
git merge test/env-conflicts-1

# 10. Try to merge the second branch - this should cause a conflict
git merge test/env-conflicts-2
```

After running these commands, Git will report a merge conflict in the respective file. You can now test the tool with the `git-conflicts` command:

```bash
# View conflict in console with colored output (default)
sops-diff git-conflicts test_secrets.env   # for ENV

# Or save to file with --output flag
sops-diff git-conflicts test_secrets.yaml --output diff.yaml
```

The output will show branch names where available, making it easier to identify changes:

```
<<<<<<< HEAD (main branch)
SECRET_KEY: modified_in_branch1
API_URL: https://api.example.com
DEBUG_MODE: true
DATABASE:
    USER: db_user
    HOST: db.example.com
    PORT: 5432
=======
SECRET_KEY: modified_in_branch2
API_URL: https://api-v2.example.com
DEBUG_MODE: false
DATABASE:
    USER: db_user_new
    HOST: db.example.com
    PORT: 5432
>>>>>>> OTHER (incoming changes from test/merge-conflicts-2)
```

After resolving the conflict, complete the merge:

```bash
git add test_secrets.yaml  # or .json or .env
git merge --continue
```

## Testing Git Integration

To test the automatic Git integration for conflict resolution:

```bash
# Set up Git to use SOPS-Diff for merge conflicts
sops-diff setup-git-merge-tool

# Add the required configuration to .gitattributes
echo "*.yaml merge=sops" >> .gitattributes
echo "*.json merge=sops" >> .gitattributes
echo "*.env merge=sops" >> .gitattributes
git add .gitattributes
git commit -m "Configure Git to use SOPS-Diff for merge resolution"
```

Now when you encounter a merge conflict, Git should automatically use SOPS-Diff to resolve it:

```bash
# Create a conflict as in previous examples
# ...

# When Git reports a conflict, use Git's mergetool
git mergetool --tool=sops
```

Git should invoke SOPS-Diff to help resolve the conflict.

## Expected Outcomes

- **Conflict Resolution**: The tool should decrypt both versions of the conflicted file and present them in a decrypted form for easier resolution.
- **Diff View Option**: When using `--view-as-diff`, the tool should use Git's merge-file to provide a more Git-like conflict view.
- **Colored Output**: When using the standard output (not saving to file), the conflict markers and content will be colorized for better readability.
- **Git Integration**: Git should seamlessly invoke SOPS-Diff when a conflict is detected in encrypted files.
- **Format Support**: The tool should work correctly for all supported formats (YAML, JSON, ENV).

## Cleanup

After testing, you can clean up the test branches:

```bash
# Ensure you're on a different branch first
git checkout main

# Delete the test branches
git branch -D base base-json base-env
git branch -D test/merge-conflicts-1 test/merge-conflicts-2
git branch -D test/json-conflicts-1 test/json-conflicts-2
git branch -D test/env-conflicts-1 test/env-conflicts-2
```

## Troubleshooting

If you encounter issues:

1. Check that SOPS can decrypt the files with your credentials
2. Verify Git attributes are properly configured
3. Ensure SOPS-Diff has the necessary permissions to execute
4. Check for error messages in the console output
5. For Git integration issues, check that the Git configuration was properly set with `git config --global -l | grep sops`

## Common Issues and Solutions

1. **Decryption Fails**: Ensure you have access to the encryption keys used to encrypt the files
   ```bash
   # Test manual decryption
   sops -d test_secrets.yaml
   ```

2. **Git Integration Not Working**: Verify the Git attributes and config 
   ```bash
   # Check Git configuration
   git config --global -l | grep sops
   
   # Verify .gitattributes content
   cat .gitattributes
   ```

3. **Conflict Resolution Incomplete**: If the conflict isn't fully resolved, check the error messages and ensure both sides of the conflict could be properly decrypted

4. **Flag Issues**: Remember that each subcommand has its own flags. For example:
   ```bash
   # This is correct:
   sops-diff git-conflicts file.yaml --output out.yaml
   
   # This won't work (flag not recognized by subcommand):
   sops-diff --output out.yaml git-conflicts file.yaml
   ```
