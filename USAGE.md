# Gmail Exporter Usage Guide

This guide provides step-by-step instructions for setting up and using the Gmail Exporter tool.

## Prerequisites

1. **Go 1.21 or later** installed on your system
2. **Gmail account** with API access
3. **Google Cloud Console** access to create OAuth credentials

## Step 1: Enable Gmail API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Gmail API"
   - Click "Enable"
4. Create OAuth 2.0 credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth 2.0 Client IDs"
   - Choose "Desktop application"
   - Download the credentials JSON file

## Step 2: Install Gmail Exporter

### Option A: Build from Source
```bash
git clone https://github.com/your-username/gmail-exporter
cd gmail-exporter
go build -o gmail-exporter ./cmd/gmail-exporter
```

### Option B: Using Go Install
```bash
go install github.com/your-username/gmail-exporter/cmd/gmail-exporter@latest
```

## Step 3: Initial Setup

### Set up OAuth credentials
```bash
./gmail-exporter auth setup --credentials-file path/to/your/credentials.json
```

### Authenticate with Gmail
```bash
./gmail-exporter auth login
```

This will:
1. Open a browser window for OAuth authentication
2. Ask you to grant permissions to the application
3. Save the authentication token for future use

## Step 4: Basic Usage Examples

### Export all emails to a specific recipient
```bash
./gmail-exporter export --to "user@example.com" --output-dir ./exports
```

### Export emails with multiple filters
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --from "sender@example.com" \
  --subject "Invoice" \
  --has-attachment \
  --output-dir ./exports
```

### Export emails with size and date filters
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --size-greater-than "5MB" \
  --date-within "30d" \
  --output-dir ./exports
```

### Export and organize by labels
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --organize-by-labels \
  --output-dir ./exports
```

## Step 5: Advanced Usage

### Parallel Processing
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --parallel-workers 5 \
  --output-dir ./exports
```

### Different Export Formats
```bash
# Export as JSON
./gmail-exporter export --to "user@example.com" --format json --output-dir ./exports

# Export as EML (default)
./gmail-exporter export --to "user@example.com" --format eml --output-dir ./exports
```

### Resume Failed Exports
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --resume \
  --state-file ./exports/.state.json \
  --output-dir ./exports
```

## Step 6: Configuration File

Create a configuration file at `~/.gmail-exporter.yaml`:

```yaml
# Gmail API Configuration
credentials_file: "~/.gmail-exporter/credentials.json"
token_file: "~/.gmail-exporter/token.json"

# Default Export Settings
output_dir: "./exports"
organize_by_labels: false
parallel_workers: 3

# Default Filters
filters:
  exclude_chats: true
  search_scope: "all_mail"

# Metrics Configuration
metrics:
  enabled: true
  format: "json"
  output_file: "metrics.json"

# Logging
log_level: "info"
log_file: ""
```

## Step 7: Monitoring and Metrics

### View Operation Status
```bash
./gmail-exporter status --state-file ./exports/.state.json
```

### Check Authentication Status
```bash
./gmail-exporter auth status
```

### View Metrics
After an export operation, check the metrics file:
```bash
cat ./exports/metrics.json
```

## Common Use Cases

### 1. Backup Emails to Specific Recipients
```bash
./gmail-exporter export \
  --to "important-client@example.com" \
  --output-dir ./backups/important-client \
  --format eml \
  --include-attachments
```

### 2. Export Large Attachments
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --has-attachment \
  --size-greater-than "10MB" \
  --output-dir ./large-attachments
```

### 3. Export Recent Communications
```bash
./gmail-exporter export \
  --to "user@example.com" \
  --date-within "7d" \
  --output-dir ./recent-emails
```

### 4. Export by Label
```bash
./gmail-exporter export \
  --labels "important,work" \
  --organize-by-labels \
  --output-dir ./labeled-emails
```

### 5. Complete Workflow (Export + Forward + Archive)
```bash
./gmail-exporter workflow \
  --to "user@example.com" \
  --destination "backup@example.com" \
  --cleanup-action archive \
  --output-dir ./exports
```

## Troubleshooting

### Authentication Issues
```bash
# Refresh expired token
./gmail-exporter auth refresh

# Check authentication status
./gmail-exporter auth status

# Re-authenticate if needed
./gmail-exporter auth login
```

### Rate Limiting
If you encounter rate limiting:
```bash
# Reduce parallel workers
./gmail-exporter export --to "user@example.com" --parallel-workers 1

# Add delays between requests (future feature)
./gmail-exporter export --to "user@example.com" --delay-between-requests 1s
```

### Large Exports
For large exports:
```bash
# Split by date ranges
./gmail-exporter export --to "user@example.com" --date-after "2024-01-01" --date-before "2024-06-30"
./gmail-exporter export --to "user@example.com" --date-after "2024-07-01" --date-before "2024-12-31"

# Enable compression
./gmail-exporter export --to "user@example.com" --compress-exports
```

### Debug Mode
```bash
./gmail-exporter export --to "user@example.com" --log-level debug --verbose
```

## Filter Reference

| Filter | Description | Example |
|--------|-------------|---------|
| `--to` | Recipient email address | `--to "user@example.com"` |
| `--from` | Sender email address | `--from "sender@example.com"` |
| `--subject` | Subject contains text | `--subject "Invoice"` |
| `--includes-words` | Email body contains words | `--includes-words "urgent important"` |
| `--excludes-words` | Email body excludes words | `--excludes-words "spam promotional"` |
| `--size-greater-than` | Email size threshold | `--size-greater-than "5MB"` |
| `--size-less-than` | Email size threshold | `--size-less-than "10MB"` |
| `--date-within` | Date range | `--date-within "30d"` |
| `--date-after` | After specific date | `--date-after "2024-01-01"` |
| `--date-before` | Before specific date | `--date-before "2024-12-31"` |
| `--has-attachment` | Has attachments | `--has-attachment` |
| `--no-attachment` | No attachments | `--no-attachment` |
| `--exclude-chats` | Exclude chat messages | `--exclude-chats` |
| `--labels` | Specific labels | `--labels "important,work"` |

## Security Notes

- OAuth tokens are stored securely in `~/.gmail-exporter/`
- Credentials are never logged or included in metrics
- All API communications use HTTPS
- Local exports can be encrypted (future feature)

## Performance Tips

1. Use parallel workers for large exports: `--parallel-workers 5`
2. Filter by date ranges to reduce API calls
3. Use specific filters to reduce the number of emails processed
4. Monitor rate limits and adjust accordingly
5. Use resume functionality for interrupted operations

## Multi-Account Authentication

Gmail Exporter supports exporting from one Gmail account and importing into another. This is useful for:
- Migrating between personal and work accounts
- Consolidating multiple Gmail accounts
- Backing up emails to a different account
- Testing import functionality safely

### Setting Up Multiple Accounts

#### 1. Create Separate Credentials for Each Account

For each Gmail account, you need separate OAuth credentials:

**Source Account (Export):**
```bash
# Set up credentials for source account
./gmail-exporter auth setup \
  --credentials source-gmail-credentials.json \
  --token source-gmail-token.json

# Authenticate source account
./gmail-exporter auth login --token source-gmail-token.json
```

**Destination Account (Import):**
```bash
# Set up credentials for destination account
./gmail-exporter auth setup \
  --credentials dest-gmail-credentials.json \
  --token dest-gmail-token.json

# Authenticate destination account
./gmail-exporter auth login --token dest-gmail-token.json
```

#### 2. Verify Authentication

Check that both accounts are properly authenticated:

```bash
# Check source account
./gmail-exporter auth status --token source-gmail-token.json

# Check destination account
./gmail-exporter auth status --token dest-gmail-token.json
```

### Complete Migration Example

Here's a complete example of migrating emails from one account to another:

```bash
# Step 1: Export from source account
./gmail-exporter export \
  --credentials-file source-gmail-credentials.json \
  --token-file source-gmail-token.json \
  --output-dir migration-2024/ \
  --from "important-client@company.com" \
  --date-after "2024-01-01" \
  --organize-by-labels \
  --format eml

# Step 2: Test import with a small number first
./gmail-exporter import \
  --input-dir migration-2024/ \
  --import-credentials dest-gmail-credentials.json \
  --import-token dest-gmail-token.json \
  --limit 5 \
  --preserve-dates

# Step 3: If test successful, import all
./gmail-exporter import \
  --input-dir migration-2024/ \
  --import-credentials dest-gmail-credentials.json \
  --import-token dest-gmail-token.json \
  --preserve-dates \
  --parallel-workers 3

# Step 4: Optional - Clean up source account
./gmail-exporter cleanup \
  --credentials-file source-gmail-credentials.json \
  --token-file source-gmail-token.json \
  --action archive \
  --filter-file migration-2024/processed_emails.json \
  --dry-run  # Remove this flag when ready to execute
```

### Account-Specific Configuration

You can create separate config files for each account:

**source-config.yaml:**
```yaml
credentials_file: "source-gmail-credentials.json"
token_file: "source-gmail-token.json"
output_dir: "./exports"
parallel_workers: 5
```

**dest-config.yaml:**
```yaml
credentials_file: "dest-gmail-credentials.json"
token_file: "dest-gmail-token.json"
parallel_workers: 3
preserve_dates: true
```

Then use them with:
```bash
# Export with source config
./gmail-exporter export --config source-config.yaml --from "sender@example.com"

# Import with destination config
./gmail-exporter import --config dest-config.yaml --input-dir exports/
```

## Advanced Filtering Examples

### Date-Based Migration

Migrate emails by year:
```bash
# Export 2023 emails
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir migration-2023/ \
  --date-after "2023-01-01" \
  --date-before "2024-01-01"

# Export 2024 emails
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir migration-2024/ \
  --date-after "2024-01-01"
```

### Label-Based Migration

Migrate specific labels:
```bash
# Export work emails
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir work-emails/ \
  --labels "work,projects,meetings" \
  --organize-by-labels

# Export personal emails
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir personal-emails/ \
  --labels "personal,family,friends"
```

### Size-Based Migration

Handle large emails separately:
```bash
# Export small emails first (faster)
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir small-emails/ \
  --size-less-than "10MB"

# Export large emails with fewer workers
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir large-emails/ \
  --size-greater-than "10MB" \
  --parallel-workers 1
```

## Testing and Validation

### Safe Testing Approach

Always test with small numbers first:

```bash
# 1. Test export with 1 message
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir test-export/ \
  --limit 1

# 2. Test import with that 1 message
./gmail-exporter import \
  --input-dir test-export/ \
  --import-credentials dest-creds.json \
  --import-token dest-token.json \
  --limit 1

# 3. Verify the email appeared correctly in destination account
# 4. Test cleanup with dry-run
./gmail-exporter cleanup \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --action archive \
  --filter-file test-export/processed_emails.json \
  --dry-run \
  --limit 1

# 5. If all looks good, increase limits gradually
```

### Validation Checklist

Before running large migrations:

- [ ] Both accounts authenticated successfully
- [ ] Test export/import with 1-5 messages
- [ ] Verify emails appear correctly in destination
- [ ] Check that attachments are preserved
- [ ] Verify dates are preserved (if using --preserve-dates)
- [ ] Test cleanup with --dry-run first
- [ ] Have backups of important data

## Troubleshooting Multi-Account Issues

### Authentication Problems

**Problem:** "Invalid credentials" error
```bash
# Solution: Verify credentials file is for correct account
./gmail-exporter auth status --token source-token.json
./gmail-exporter auth status --token dest-token.json
```

**Problem:** "Token expired" error
```bash
# Solution: Refresh tokens for both accounts
./gmail-exporter auth refresh --token source-token.json
./gmail-exporter auth refresh --token dest-token.json
```

### Import Issues

**Problem:** Emails not appearing in destination account
- Verify you're using correct destination credentials
- Check that import completed without errors
- Look in "All Mail" folder, not just Inbox

**Problem:** Duplicate emails during import
- This was a bug in earlier versions - ensure you're using latest version
- The tool now uses Gmail API Import instead of Send

### Performance Issues

**Problem:** Import/export too slow
```bash
# Increase parallel workers
--parallel-workers 5

# Or decrease if hitting rate limits
--parallel-workers 1
```

**Problem:** Rate limiting errors
```bash
# Reduce parallel workers and add delays
--parallel-workers 1
```

## Best Practices

### Security
- Use separate credentials files for each account
- Store credentials securely (not in version control)
- Use specific scopes (readonly for export, full for import)
- Regularly rotate OAuth tokens

### Performance
- Start with small test batches
- Use appropriate parallel worker counts
- Monitor rate limits
- Split large migrations by date/size

### Organization
- Use descriptive output directory names
- Include dates in directory names
- Keep export and import logs
- Document your migration process

### Backup
- Always backup important data first
- Test restore procedures
- Keep multiple copies of exports
- Verify data integrity after migration

## Example Scripts

### Complete Migration Script

```bash
#!/bin/bash
set -e

SOURCE_CREDS="source-gmail-credentials.json"
SOURCE_TOKEN="source-gmail-token.json"
DEST_CREDS="dest-gmail-credentials.json"
DEST_TOKEN="dest-gmail-token.json"
MIGRATION_DIR="migration-$(date +%Y%m%d)"

echo "Starting Gmail migration to $MIGRATION_DIR"

# Step 1: Export
echo "Exporting emails..."
./gmail-exporter export \
  --credentials-file "$SOURCE_CREDS" \
  --token-file "$SOURCE_TOKEN" \
  --output-dir "$MIGRATION_DIR" \
  --date-after "2024-01-01" \
  --organize-by-labels

# Step 2: Test import
echo "Testing import with 5 messages..."
./gmail-exporter import \
  --input-dir "$MIGRATION_DIR" \
  --import-credentials "$DEST_CREDS" \
  --import-token "$DEST_TOKEN" \
  --limit 5

read -p "Test import successful? Continue with full import? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  # Step 3: Full import
  echo "Running full import..."
  ./gmail-exporter import \
    --input-dir "$MIGRATION_DIR" \
    --import-credentials "$DEST_CREDS" \
    --import-token "$DEST_TOKEN"

  echo "Migration completed successfully!"
else
  echo "Migration cancelled."
fi
```

### Incremental Backup Script

```bash
#!/bin/bash
# Daily incremental backup script

BACKUP_DIR="backups/$(date +%Y%m%d)"
YESTERDAY=$(date -d "yesterday" +%Y-%m-%d)

./gmail-exporter export \
  --output-dir "$BACKUP_DIR" \
  --date-after "$YESTERDAY" \
  --format eml \
  --include-attachments

echo "Daily backup completed: $BACKUP_DIR"
```
