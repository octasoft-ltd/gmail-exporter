# Gmail Exporter

A powerful command-line tool for exporting, importing, and managing Gmail emails with advanced filtering capabilities.

## Features

- **Export emails** from Gmail with advanced filtering (supports all Gmail search operators)
- **Import emails** into Gmail accounts (supports cross-account transfers)
- **Cleanup emails** from source account after export
- **Multiple formats**: EML, JSON, mbox
- **Parallel processing** for high performance
- **Progress tracking** with real-time indicators
- **Resumable operations** with state management
- **Comprehensive metrics** collection (JSON and Prometheus formats)
- **OAuth 2.0 authentication** with Google Gmail API
- **Cross-account support** for migrating between Gmail accounts

## Installation

### Prerequisites

- Go 1.19 or later
- Gmail API credentials (see Authentication section)

### Build from Source

```bash
git clone https://github.com/octasoft-ltd/gmail-exporter.git
cd gmail-exporter
go build -o gmail-exporter ./cmd/gmail-exporter
```

## Authentication

Gmail Exporter supports two authentication scenarios:

### Single Account (Export and Import to Same Account)

1. **Set up Gmail API credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one
   - Enable Gmail API
   - Create OAuth 2.0 credentials (Desktop application)
   - **Important:** Add yourself as a test user to avoid "unverified app" issues
   - Download the credentials JSON file

2. **Authenticate:**
   ```bash
   ./gmail-exporter auth setup --credentials gmail-credentials.json
   ./gmail-exporter auth login
   ```

📚 **Detailed Setup Guide:** See [AUTHENTICATION_SETUP.md](AUTHENTICATION_SETUP.md) for step-by-step instructions including handling Google's "unverified app" warnings.

### Multi-Account (Export from One Account, Import to Another)

For migrating emails between different Gmail accounts:

1. **Set up credentials for source account (export):**
   ```bash
   ./gmail-exporter auth setup --credentials source-credentials.json --token source-token.json
   ./gmail-exporter auth login --token source-token.json
   ```

2. **Set up credentials for destination account (import):**
   ```bash
   ./gmail-exporter auth setup --credentials dest-credentials.json --token dest-token.json
   ./gmail-exporter auth login --token dest-token.json
   ```

3. **Export from source account:**
   ```bash
   ./gmail-exporter export \
     --credentials-file source-credentials.json \
     --token-file source-token.json \
     --output-dir exports/ \
     --from "important-sender@example.com"
   ```

4. **Import to destination account:**
   ```bash
   ./gmail-exporter import \
     --input-dir exports/ \
     --import-credentials dest-credentials.json \
     --import-token dest-token.json
   ```

## Usage Examples

### Basic Export

```bash
# Export all emails to EML format
./gmail-exporter export --output-dir exports/

# Export with filters
./gmail-exporter export \
  --output-dir exports/ \
  --from "sender@example.com" \
  --subject "Important" \
  --date-after "2024-01-01" \
  --has-attachment
```

### Advanced Filtering

```bash
# Complex filter example
./gmail-exporter export \
  --output-dir exports/ \
  --from "notifications@github.com" \
  --includes-words "pull request merged" \
  --date-within "30d" \
  --size-greater-than "1MB" \
  --labels "work,important" \
  --exclude-chats
```

### Cross-Account Migration

```bash
# 1. Export from source account
./gmail-exporter export \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --output-dir migration/ \
  --search-scope "all_mail"

# 2. Import to destination account
./gmail-exporter import \
  --input-dir migration/ \
  --import-credentials dest-creds.json \
  --import-token dest-token.json

# 3. Optional: Clean up source account
./gmail-exporter cleanup \
  --credentials-file source-creds.json \
  --token-file source-token.json \
  --action archive \
  --filter-file migration/processed_emails.json
```

### Testing with Limits

```bash
# Test with a small number of messages first
./gmail-exporter export --output-dir test/ --limit 5
./gmail-exporter import --input-dir test/ --limit 5
```

## Configuration

### Command-line Flags

#### Export Command
- `--output-dir, -o`: Output directory for exported emails
- `--format`: Export format (eml, json, mbox) [default: eml]
- `--organize-by-labels`: Organize emails by labels in folder structure
- `--parallel-workers`: Number of parallel workers [default: 3]
- `--include-attachments`: Include email attachments [default: true]
- `--limit, -l`: Limit number of messages to process (useful for testing)

#### Import Command
- `--input-dir, -i`: Input directory containing exported emails
- `--import-credentials`: Gmail API credentials for destination account
- `--import-token`: OAuth token file for destination account
- `--parallel-workers`: Number of parallel workers [default: 3]
- `--preserve-dates`: Preserve original email dates [default: true]
- `--limit, -l`: Limit number of messages to process (useful for testing)

#### Cleanup Command
- `--action`: Action to perform (archive, delete) [default: archive]
- `--filter-file`: JSON file containing processed email IDs
- `--dry-run`: Show what would be done without making changes
- `--limit, -l`: Limit number of messages to process

### Filter Options

All Gmail search operators are supported:

- `--to`: Recipient email address
- `--from`: Sender email address
- `--subject`: Subject contains text
- `--includes-words`: Email body contains words
- `--excludes-words`: Email body excludes words
- `--size-greater-than`: Email size greater than (e.g., 5MB)
- `--size-less-than`: Email size less than (e.g., 10MB)
- `--date-within`: Date within period (e.g., 30d, 1w, 6m)
- `--date-after`: After specific date (YYYY-MM-DD)
- `--date-before`: Before specific date (YYYY-MM-DD)
- `--has-attachment`: Has attachments
- `--no-attachment`: No attachments
- `--exclude-chats`: Exclude chat messages [default: true]
- `--labels`: Specific labels (comma-separated)
- `--search-scope`: Search scope (all_mail, inbox, sent, drafts, spam, trash)

## Output Formats

### EML Format (Default)
Standard email format that preserves all email data including headers, body, and attachments.

### JSON Format
Structured format containing the complete Gmail API message object.

### Mbox Format
Unix mailbox format for compatibility with email clients.

## Security and Privacy

- **OAuth 2.0**: Secure authentication without storing passwords
- **Local storage**: All credentials and tokens stored locally
- **API permissions**: Requests only necessary Gmail API scopes
- **No data transmission**: All processing happens locally

🔒 **Security Documentation:** See [SECURITY.md](SECURITY.md) for detailed information about:
- What permissions are requested and why
- How to verify the code is safe
- Handling Google's "unverified app" warnings
- Security best practices

## Troubleshooting

### Authentication Issues

```bash
# Check authentication status
./gmail-exporter auth status

# Refresh expired tokens
./gmail-exporter auth refresh

# Re-authenticate if needed
./gmail-exporter auth login
```

### Common Issues

1. **"Invalid credentials"**: Ensure credentials file is valid JSON from Google Cloud Console
2. **"Token expired"**: Run `./gmail-exporter auth refresh`
3. **"Permission denied"**: Ensure Gmail API is enabled in Google Cloud Console
4. **"Rate limit exceeded"**: Reduce parallel workers with `--parallel-workers 1`

### Multi-Account Issues

1. **Wrong account**: Verify you're using correct credentials/token files
2. **Cross-contamination**: Use separate credential files for each account
3. **Token confusion**: Always specify both credentials and token files explicitly

## Performance

- **Parallel processing**: Configurable worker pools for optimal performance
- **Resumable operations**: Continue interrupted exports/imports
- **Progress tracking**: Real-time progress indicators
- **Memory efficient**: Streams large emails without loading entirely into memory

## Metrics

The tool collects comprehensive metrics:

```json
{
  "operation": "export",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T10:35:30Z",
  "duration": "5m30s",
  "total_matched": 1500,
  "total_processed": 1500,
  "total_failed": 0,
  "total_size": 2147483648,
  "performance": {
    "emails_per_second": 4.5,
    "bytes_per_second": 6553600
  }
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- **Issues**: Report bugs and feature requests on GitHub
- **Documentation**: See USAGE.md for detailed examples
- **Security**: Report security issues privately (see [SECURITY.md](SECURITY.md))
- **Authentication**: See [AUTHENTICATION_SETUP.md](AUTHENTICATION_SETUP.md) for detailed setup instructions 