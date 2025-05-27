package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Authenticator handles Gmail API authentication
type Authenticator struct {
	credentialsFile string
	tokenFile       string
	config          *oauth2.Config
}

// Status represents the authentication status
type Status struct {
	Status      string     `json:"status"`
	TokenExpiry *time.Time `json:"token_expiry,omitempty"`
	Email       string     `json:"email,omitempty"`
}

// NewAuthenticator creates a new authenticator instance
func NewAuthenticator(credentialsFile, tokenFile string) (*Authenticator, error) {
	// Read credentials file
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	// Parse credentials and create OAuth config
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.GmailModifyScope, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// Set redirect URI to localhost for better UX
	config.RedirectURL = "http://localhost:8080/callback"

	return &Authenticator{
		credentialsFile: credentialsFile,
		tokenFile:       tokenFile,
		config:          config,
	}, nil
}

// Authenticate performs the OAuth 2.0 authentication flow
func (a *Authenticator) Authenticate() error {
	// Check if we already have a valid token
	token, err := a.loadToken()
	if err == nil && token.Valid() {
		logrus.Info("Using existing valid token")
		return nil
	}

	fmt.Println("🔐 Starting Gmail API authentication...")
	fmt.Println()
	fmt.Println("📋 IMPORTANT SECURITY INFORMATION:")
	fmt.Println("   This application will request the following Gmail permissions:")
	fmt.Println("   • Read your email messages and settings")
	fmt.Println("   • Modify your email messages and settings (for import/cleanup)")
	fmt.Println("   • Send email on your behalf (for import functionality)")
	fmt.Println()
	fmt.Println("   These permissions are necessary for:")
	fmt.Println("   • Exporting emails (read-only access)")
	fmt.Println("   • Importing emails into your account")
	fmt.Println("   • Archiving/deleting emails during cleanup")
	fmt.Println()
	fmt.Println("   🔍 You can review the source code at:")
	fmt.Println("   • internal/auth/auth.go (this authentication code)")
	fmt.Println("   • internal/exporter/ (export functionality)")
	fmt.Println("   • internal/importer/ (import functionality)")
	fmt.Println("   • internal/cleaner/ (cleanup functionality)")
	fmt.Println()

	// Try automatic flow first
	if token, err := a.authenticateWithLocalServer(); err == nil {
		if err := a.saveToken(token); err != nil {
			return fmt.Errorf("unable to save token: %w", err)
		}
		fmt.Println("✅ Authentication successful!")
		return nil
	}

	// Fall back to manual flow
	fmt.Println("⚠️  Automatic authentication failed, falling back to manual flow...")
	return a.authenticateManually()
}

// authenticateWithLocalServer uses a local server to capture the auth code automatically
func (a *Authenticator) authenticateWithLocalServer() (*oauth2.Token, error) {
	// Create a channel to receive the auth code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server
	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			return
		}

		// Send success page
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Gmail Exporter - Authentication Success</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); max-width: 500px; margin: 0 auto; }
        .success { color: #28a745; font-size: 24px; margin-bottom: 20px; }
        .info { color: #666; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✅ Authentication Successful!</div>
        <div class="info">You can now close this browser window and return to the terminal.</div>
        <div class="info">Gmail Exporter is now authenticated and ready to use.</div>
    </div>
</body>
</html>`)

		codeChan <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Generate auth URL
	authURL := a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("🌐 Opening browser for authentication...")
	fmt.Printf("   If the browser doesn't open automatically, visit: %s\n", authURL)
	fmt.Println()

	// Try to open browser automatically
	if err := openBrowser(authURL); err != nil {
		logrus.WithError(err).Warn("Failed to open browser automatically")
	}

	// Wait for auth code or timeout
	var authCode string
	select {
	case authCode = <-codeChan:
		// Success
	case err := <-errChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authentication timeout after 5 minutes")
	}

	// Shutdown server
	server.Shutdown(context.Background())

	// Exchange code for token
	token, err := a.config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return token, nil
}

// authenticateManually performs manual authentication flow
func (a *Authenticator) authenticateManually() error {
	fmt.Println()
	fmt.Println("📝 Manual Authentication Required")
	fmt.Println("   Follow these steps:")
	fmt.Println()

	// Generate auth URL
	authURL := a.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("1. Open this URL in your browser:\n   %s\n\n", authURL)

	fmt.Println("2. Complete the OAuth flow in your browser")
	fmt.Println("3. After granting permissions, you'll be redirected to a page that may show an error")
	fmt.Println("4. Copy the 'code' parameter from the URL in your browser's address bar")
	fmt.Println("   Example: http://localhost:8080/callback?code=COPY_THIS_PART&scope=...")
	fmt.Println()

	var authCode string
	fmt.Print("Enter the authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		return fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := a.config.Exchange(context.TODO(), authCode)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	// Save token
	if err := a.saveToken(token); err != nil {
		return fmt.Errorf("unable to save token: %w", err)
	}

	fmt.Println("✅ Authentication successful!")
	return nil
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// RefreshToken refreshes the authentication token
func (a *Authenticator) RefreshToken() error {
	token, err := a.loadToken()
	if err != nil {
		return fmt.Errorf("unable to load token: %w", err)
	}

	// Create a token source that will refresh the token
	tokenSource := a.config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("unable to refresh token: %w", err)
	}

	// Save the refreshed token
	if err := a.saveToken(newToken); err != nil {
		return fmt.Errorf("unable to save refreshed token: %w", err)
	}

	logrus.Info("Token refreshed successfully")
	return nil
}

// GetStatus returns the current authentication status
func (a *Authenticator) GetStatus() (*Status, error) {
	token, err := a.loadToken()
	if err != nil {
		return &Status{Status: "not_authenticated"}, nil
	}

	status := &Status{
		TokenExpiry: &token.Expiry,
	}

	if token.Valid() {
		status.Status = "authenticated"

		// Try to get user email
		if email, err := a.getUserEmail(token); err == nil {
			status.Email = email
		}
	} else {
		status.Status = "token_expired"
	}

	return status, nil
}

// GetClient returns an authenticated HTTP client
func (a *Authenticator) GetClient() (*http.Client, error) {
	token, err := a.loadToken()
	if err != nil {
		return nil, fmt.Errorf("unable to load token: %w", err)
	}

	if !token.Valid() {
		// Try to refresh the token
		if err := a.RefreshToken(); err != nil {
			return nil, fmt.Errorf("token expired and refresh failed: %w", err)
		}
		// Reload the refreshed token
		token, err = a.loadToken()
		if err != nil {
			return nil, fmt.Errorf("unable to load refreshed token: %w", err)
		}
	}

	return a.config.Client(context.Background(), token), nil
}

// GetGmailService returns an authenticated Gmail service
func (a *Authenticator) GetGmailService() (*gmail.Service, error) {
	client, err := a.GetClient()
	if err != nil {
		return nil, err
	}

	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %w", err)
	}

	return service, nil
}

// loadToken loads the token from file
func (a *Authenticator) loadToken() (*oauth2.Token, error) {
	f, err := os.Open(a.tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// saveToken saves the token to file
func (a *Authenticator) saveToken(token *oauth2.Token) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(a.tokenFile), 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(a.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// getUserEmail gets the authenticated user's email address
func (a *Authenticator) getUserEmail(token *oauth2.Token) (string, error) {
	client := a.config.Client(context.Background(), token)
	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return "", err
	}

	return profile.EmailAddress, nil
}
