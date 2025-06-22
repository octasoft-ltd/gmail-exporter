# Multi-Account Setup Quick Reference

This document provides a quick reference for setting up Gmail Exporter with multiple accounts.

## Bug Fix: Import Duplication Issue

**Issue:** The import command was sending emails as new messages instead of importing them into the mailbox,
causing recipients to receive duplicate emails.

**Fix:** Updated the importer to use `Users.Messages.Import()` instead of `Users.Messages.Send()`,
which properly imports emails without sending them.

## Multi-Account Authentication

### Quick Setup

1. **Source Account (for export):**

   ```bash
   ./gmail-exporter auth setup --credentials source-creds.json --token source-token.json
   ./gmail-exporter auth login --token source-token.json
   ```

2. **Destination Account (for import):**

   ```bash
   ./gmail-exporter auth setup --credentials dest-creds.json --token dest-token.json
   ./gmail-exporter auth login --token dest-token.json
   ```

### Usage

**Export from source:**

```bash
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir exports/
```

**Import to destination:**

```bash
./gmail-exporter import \
  --input-dir exports/ \
  --import-credentials dest-creds.json \
  --import-token dest-token.json
```

### New Flags

- `--import-credentials`: Credentials file for destination account
- `--import-token`: Token file for destination account

If not specified, defaults to main credentials (same account import).

## Testing Workflow

Always test with small numbers first:

```bash
# Test export
./gmail-exporter export --output-dir test/ --limit 1

# Test import
./gmail-exporter import --input-dir test/ --import-credentials dest-creds.json --limit 1

# Verify email appears correctly in destination account
```

## Key Changes Made

1. **Fixed import duplication bug** - now uses proper Gmail API Import
2. **Added multi-account support** - separate credentials for import
3. **Updated documentation** - comprehensive multi-account examples
4. **Enhanced CLI help** - clear authentication instructions
5. **Added validation** - proper error handling for missing credentials

## Files Updated

- `internal/cli/import.go` - Added import-credentials and import-token flags
- `internal/cli/cli_test.go` - Updated tests for new flags
- `README.md` - Comprehensive multi-account documentation
- `USAGE.md` - Detailed examples and troubleshooting
- `MULTI_ACCOUNT_SETUP.md` - This quick reference guide

All tests passing ✅
