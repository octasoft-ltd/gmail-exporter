# Security Documentation

This document explains the security aspects of Gmail Exporter, what permissions it requires, and how to verify its safety.

## 🔐 Permissions Required

Gmail Exporter requests the following Gmail API scopes:

### 1. `https://www.googleapis.com/auth/gmail.readonly`
**What it does:** Read-only access to Gmail messages and settings
**Used for:** 
- Exporting emails from your account
- Reading email metadata (subject, sender, date, etc.)
- Accessing email content and attachments

### 2. `https://www.googleapis.com/auth/gmail.modify`
**What it does:** Read, compose, send, and permanently delete Gmail messages
**Used for:**
- Importing emails into your account
- Archiving emails during cleanup operations
- Deleting emails during cleanup operations (if you choose delete action)

### 3. `https://www.googleapis.com/auth/gmail.send`
**What it does:** Send email on your behalf
**Used for:**
- Importing emails (uses Gmail API Import, not Send - this scope is required by the API)

## 🛡️ Security Measures

### Local Processing Only
- **No data transmission**: All email processing happens locally on your machine
- **No external servers**: Your emails are never sent to external servers
- **Local storage**: All exports are stored locally on your filesystem

### Secure Authentication
- **OAuth 2.0**: Uses Google's secure OAuth 2.0 flow
- **No password storage**: Never stores your Gmail password
- **Token-based**: Uses refresh tokens that can be revoked
- **Local token storage**: Tokens stored locally with restricted file permissions (0600)

### Minimal Network Access
- **Gmail API only**: Only communicates with Google's Gmail API
- **HTTPS only**: All API communications use HTTPS encryption
- **No telemetry**: No usage data or analytics sent anywhere

## 🔍 Code Verification

You can verify the safety of this application by reviewing the source code:

### Key Files to Review

1. **Authentication (`internal/auth/auth.go`)**
   - OAuth 2.0 implementation
   - Token storage and management
   - No credential logging or transmission

2. **Export (`internal/exporter/exporter.go`)**
   - Read-only Gmail API calls
   - Local file writing only
   - No network transmission of email data

3. **Import (`internal/importer/importer.go`)**
   - Uses Gmail API Import (not Send)
   - Reads local files only
   - Adds emails to your mailbox without sending

4. **Cleanup (`internal/cleaner/cleaner.go`)**
   - Archive or delete operations
   - Dry-run mode for safety
   - Only processes emails you've already exported

### What to Look For

✅ **Safe patterns:**
- `gmail.NewService()` - Creates Gmail API client
- `service.Users.Messages.List()` - Lists messages (read-only)
- `service.Users.Messages.Get()` - Gets message content (read-only)
- `service.Users.Messages.Import()` - Imports messages to mailbox
- `service.Users.Messages.Modify()` - Archives messages
- `os.WriteFile()` - Writes to local filesystem
- `json.NewEncoder()` - Encodes data locally

❌ **Red flags to watch for (NOT present in this code):**
- HTTP requests to non-Google domains
- `service.Users.Messages.Send()` with new message content
- Network transmission of email data
- Credential logging or storage in plain text
- External API calls

## 🚨 Google's Unverified App Warning

When you first authenticate, Google will show a warning about an "unverified app":

### Why This Happens
- This application is not published to Google's app store
- Google requires a verification process for published apps
- Since this is open-source software you run locally, it's not verified

### How to Proceed Safely

1. **Review the source code** (see sections above)
2. **Add yourself as a test user** in Google Cloud Console:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Navigate to "APIs & Services" > "OAuth consent screen"
   - Scroll down to "Test users"
   - Add your Gmail address as a test user
3. **Proceed through the warning** by clicking "Advanced" then "Go to Gmail Exporter (unsafe)"

### Alternative: Verify the App (Advanced)
If you want to remove the warning entirely, you can:
1. Submit the app for Google's verification process
2. This requires domain verification and security review
3. Not necessary for personal use

## 🔒 Best Practices

### For Users
1. **Review the code** before running, especially if you're handling sensitive emails
2. **Use test accounts** first to verify behavior
3. **Start with small limits** (`--limit 5`) to test functionality
4. **Keep credentials secure** - don't share credentials.json or token files
5. **Revoke access** when no longer needed via [Google Account settings](https://myaccount.google.com/permissions)

### For Developers
1. **Never log credentials** or tokens
2. **Use minimal scopes** required for functionality
3. **Implement dry-run modes** for destructive operations
4. **Provide clear documentation** about what the code does
5. **Use secure file permissions** for sensitive files

## 🔧 Revoking Access

If you want to revoke Gmail Exporter's access to your account:

1. **Via Google Account:**
   - Go to [Google Account permissions](https://myaccount.google.com/permissions)
   - Find "Gmail Exporter" in the list
   - Click "Remove access"

2. **Via Local Files:**
   - Delete the token file: `rm ~/.gmail-exporter/token.json`
   - Delete credentials: `rm ~/.gmail-exporter/credentials.json`

## 📋 Security Checklist

Before using Gmail Exporter:

- [ ] I have reviewed the source code in key files listed above
- [ ] I understand what permissions are being granted
- [ ] I have added myself as a test user in Google Cloud Console
- [ ] I am comfortable with the security implications
- [ ] I will start with small test exports (`--limit 5`)
- [ ] I know how to revoke access if needed

## 🚨 Reporting Security Issues

If you discover a security vulnerability:

1. **Do NOT** create a public GitHub issue
2. **Email privately** to: security@example.com
3. **Include details** about the vulnerability
4. **Allow time** for investigation and fix before public disclosure

## 📚 Additional Resources

- [Google OAuth 2.0 Security Best Practices](https://developers.google.com/identity/protocols/oauth2/security-best-practices)
- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [Google Cloud Security](https://cloud.google.com/security)

---

**Remember:** This application runs entirely on your local machine. Your emails never leave your computer except to communicate with Google's Gmail API for the operations you explicitly request. 