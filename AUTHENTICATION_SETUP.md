# Authentication Setup Guide

This guide walks you through setting up Gmail API authentication for Gmail Exporter, including handling Google's "unverified app" warnings.

## 📋 Prerequisites

- Gmail account
- Access to [Google Cloud Console](https://console.cloud.google.com/)
- Gmail Exporter installed on your system

## 🚀 Step-by-Step Setup

### Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" → "New Project"
3. Enter project name (e.g., "Gmail Exporter Personal")
4. Click "Create"

### Step 2: Enable Gmail API

1. In your project, go to "APIs & Services" → "Library"
2. Search for "Gmail API"
3. Click on "Gmail API" → "Enable"

### Step 3: Configure OAuth Consent Screen

This step is **crucial** to avoid authentication issues:

1. Go to "APIs & Services" → "OAuth consent screen"
2. Choose "External" (unless you have a Google Workspace account)
3. Fill in required fields:
   - **App name**: "Gmail Exporter" (or your preferred name)
   - **User support email**: Your email address
   - **Developer contact information**: Your email address
4. Click "Save and Continue"
5. **Scopes**: Click "Save and Continue" (we'll add scopes automatically)
6. **Test users** (IMPORTANT):
   - Click "Add Users"
   - Add your Gmail address(es) that you want to use with the tool
   - This prevents the "unverified app" blocking
   - You can add up to 100 test users
7. Click "Save and Continue"
8. Review and click "Back to Dashboard"

### Step 4: Create OAuth Credentials

1. Go to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "OAuth 2.0 Client IDs"
3. Choose "Desktop application"
4. Name it "Gmail Exporter Desktop"
5. Click "Create"
6. **Download the JSON file** - this is your `credentials.json`

### Step 5: Set Up Gmail Exporter

1. **Place credentials file:**
   ```bash
   # Copy your downloaded credentials file
   cp ~/Downloads/client_secret_*.json ./gmail-credentials.json
   ```

2. **Set up authentication:**
   ```bash
   ./gmail-exporter auth setup --credentials-file gmail-credentials.json
   ```

3. **Authenticate:**
   ```bash
   ./gmail-exporter auth login
   ```

## 🔐 Authentication Flow

When you run `auth login`, here's what happens:

### Automatic Flow (Recommended)
1. **Security information displayed** - Review what permissions are being requested
2. **Browser opens automatically** - OAuth consent screen loads
3. **Grant permissions** - Click through Google's OAuth flow
4. **Automatic completion** - Browser shows success page, terminal completes
5. **Ready to use** - Authentication token saved locally

### Manual Flow (Fallback)
If automatic flow fails:
1. **Copy URL** - Manually open the provided URL in your browser
2. **Complete OAuth** - Grant permissions in browser
3. **Copy auth code** - From the browser URL after redirect
4. **Paste in terminal** - Enter the code when prompted

## ⚠️ Handling "Unverified App" Warning

You'll see this warning because Gmail Exporter isn't published to Google's app store:

### What You'll See
```
Google hasn't verified this app
This app hasn't been verified by Google yet. Only proceed if you know and trust the developer.
```

### How to Proceed Safely

1. **Verify you added yourself as a test user** (Step 3.6 above)
2. **Review the source code** (see SECURITY.md)
3. **Click "Advanced"**
4. **Click "Go to Gmail Exporter (unsafe)"**
5. **Grant the requested permissions**

### Why This is Safe
- You're running the code locally on your machine
- You can review all source code
- No data is transmitted to external servers
- You control the OAuth application in your Google Cloud project

## 🔧 Troubleshooting

### "Access blocked: This app's request is invalid"
**Cause:** OAuth consent screen not properly configured
**Solution:**
1. Ensure you completed Step 3 (OAuth consent screen)
2. Add yourself as a test user
3. Make sure Gmail API is enabled

### "The redirect URI in the request does not match"
**Cause:** OAuth client configuration issue
**Solution:**
1. Recreate OAuth credentials (Step 4)
2. Choose "Desktop application" type
3. Don't manually set redirect URIs

### "invalid_grant" error
**Cause:** Clock skew or expired authorization code
**Solution:**
1. Check your system clock is correct
2. Try authentication again quickly after getting the URL
3. Use manual flow if automatic flow times out

### Browser doesn't open automatically
**Cause:** No default browser or permission issues
**Solution:**
1. Copy the URL from terminal
2. Manually open in your browser
3. Continue with manual flow

### "Token expired" after some time
**Cause:** Refresh token expired (normal after long periods)
**Solution:**
```bash
./gmail-exporter auth refresh
# If that fails:
./gmail-exporter auth login
```

## 🔄 Multi-Account Setup

For multiple Gmail accounts:

### Account 1 (Source)
```bash
./gmail-exporter auth setup --credentials source-creds.json --token source-token.json
./gmail-exporter auth login --token source-token.json
```

### Account 2 (Destination)
```bash
./gmail-exporter auth setup --credentials dest-creds.json --token dest-token.json
./gmail-exporter auth login --token dest-token.json
```

**Note:** Each account needs to be added as a test user in the OAuth consent screen.

## 📱 Mobile/Remote Access

If you're running Gmail Exporter on a remote server:

1. **Use manual flow** - automatic browser opening won't work
2. **Copy URL to local browser** - Open the OAuth URL on your local machine
3. **Complete authentication** - Grant permissions on your local browser
4. **Copy auth code** - From the redirect URL
5. **Paste on server** - Enter the code in the remote terminal

## 🔒 Security Best Practices

1. **Keep credentials secure:**
   ```bash
   chmod 600 gmail-credentials.json
   chmod 600 ~/.gmail-exporter/token.json
   ```

2. **Don't commit credentials to version control:**
   ```bash
   echo "*.json" >> .gitignore
   echo "gmail-credentials.json" >> .gitignore
   ```

3. **Regularly review access:**
   - Check [Google Account permissions](https://myaccount.google.com/permissions)
   - Remove access when no longer needed

4. **Use separate projects for different purposes:**
   - Personal use: One project
   - Work use: Separate project
   - Testing: Another separate project

## 📋 Quick Verification

After setup, verify everything works:

```bash
# Check authentication status
./gmail-exporter auth status

# Test with a small export
./gmail-exporter export --output-dir test-export --limit 1

# Verify the export worked
ls -la test-export/
```

## 🆘 Getting Help

If you're still having issues:

1. **Check the logs** - Run with `--log-level debug`
2. **Review SECURITY.md** - Understand what the tool does
3. **Try with a test Gmail account** - Before using your main account
4. **Create an issue** - Include error messages and steps you've tried

---

**Remember:** The initial setup is the most complex part. Once authenticated, Gmail Exporter will automatically refresh tokens and handle authentication seamlessly.
