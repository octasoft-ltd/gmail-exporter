package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNewAuthenticator(t *testing.T) {
	// Create temporary directory and files for testing
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock credentials file
	credentialsFile := filepath.Join(tempDir, "credentials.json")
	tokenFile := filepath.Join(tempDir, "token.json")

	// Create mock credentials content
	mockCredentials := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob", "http://localhost"},
		},
	}

	credentialsData, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}

	err = os.WriteFile(credentialsFile, credentialsData, 0644)
	if err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	// Test creating authenticator
	authenticator, err := NewAuthenticator(credentialsFile, tokenFile)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	if authenticator.credentialsFile != credentialsFile {
		t.Errorf("Expected credentials file %s, got %s", credentialsFile, authenticator.credentialsFile)
	}

	if authenticator.tokenFile != tokenFile {
		t.Errorf("Expected token file %s, got %s", tokenFile, authenticator.tokenFile)
	}

	if authenticator.config == nil {
		t.Error("Expected OAuth config to be initialized")
	}
}

func TestNewAuthenticator_InvalidCredentials(t *testing.T) {
	// Test with non-existent credentials file
	_, err := NewAuthenticator("non_existent_file.json", "token.json")
	if err == nil {
		t.Error("Expected error for non-existent credentials file")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with invalid JSON
	invalidCredentialsFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidCredentialsFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid credentials file: %v", err)
	}

	_, err = NewAuthenticator(invalidCredentialsFile, "token.json")
	if err == nil {
		t.Error("Expected error for invalid JSON credentials file")
	}
}

func TestAuthenticator_saveToken_loadToken(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock authenticator
	credentialsFile := filepath.Join(tempDir, "credentials.json")
	tokenFile := filepath.Join(tempDir, "token.json")

	// Create minimal credentials file
	mockCredentials := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob"},
		},
	}

	credentialsData, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}

	err = os.WriteFile(credentialsFile, credentialsData, 0644)
	if err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	authenticator, err := NewAuthenticator(credentialsFile, tokenFile)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Create a test token
	testToken := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Test saving token
	err = authenticator.saveToken(testToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Verify token file exists
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		t.Error("Token file was not created")
	}

	// Test loading token
	loadedToken, err := authenticator.loadToken()
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}

	if loadedToken.AccessToken != testToken.AccessToken {
		t.Errorf("Expected access token %s, got %s", testToken.AccessToken, loadedToken.AccessToken)
	}

	if loadedToken.RefreshToken != testToken.RefreshToken {
		t.Errorf("Expected refresh token %s, got %s", testToken.RefreshToken, loadedToken.RefreshToken)
	}

	if loadedToken.TokenType != testToken.TokenType {
		t.Errorf("Expected token type %s, got %s", testToken.TokenType, loadedToken.TokenType)
	}
}

func TestAuthenticator_loadToken_NonExistent(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	credentialsFile := filepath.Join(tempDir, "credentials.json")
	tokenFile := filepath.Join(tempDir, "non_existent_token.json")

	// Create minimal credentials file
	mockCredentials := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob"},
		},
	}

	credentialsData, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}

	err = os.WriteFile(credentialsFile, credentialsData, 0644)
	if err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	authenticator, err := NewAuthenticator(credentialsFile, tokenFile)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Test loading non-existent token
	_, err = authenticator.loadToken()
	if err == nil {
		t.Error("Expected error when loading non-existent token file")
	}
}

func TestAuthenticator_GetStatus(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	credentialsFile := filepath.Join(tempDir, "credentials.json")
	tokenFile := filepath.Join(tempDir, "token.json")

	// Create minimal credentials file
	mockCredentials := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob"},
		},
	}

	credentialsData, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}

	err = os.WriteFile(credentialsFile, credentialsData, 0644)
	if err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	authenticator, err := NewAuthenticator(credentialsFile, tokenFile)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Test status with no token
	status, err := authenticator.GetStatus()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Status != "not_authenticated" {
		t.Errorf("Expected status 'not_authenticated', got %s", status.Status)
	}

	// Create and save a valid token
	validToken := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	err = authenticator.saveToken(validToken)
	if err != nil {
		t.Fatalf("Failed to save valid token: %v", err)
	}

	// Test status with valid token
	status, err = authenticator.GetStatus()
	if err != nil {
		t.Fatalf("Failed to get status with valid token: %v", err)
	}

	if status.Status != "authenticated" {
		t.Errorf("Expected status 'authenticated', got %s", status.Status)
	}

	if status.TokenExpiry == nil {
		t.Error("Expected token expiry to be set")
	}

	// Create and save an expired token
	expiredToken := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-time.Hour), // Expired
	}

	err = authenticator.saveToken(expiredToken)
	if err != nil {
		t.Fatalf("Failed to save expired token: %v", err)
	}

	// Test status with expired token
	status, err = authenticator.GetStatus()
	if err != nil {
		t.Fatalf("Failed to get status with expired token: %v", err)
	}

	if status.Status != "token_expired" {
		t.Errorf("Expected status 'token_expired', got %s", status.Status)
	}
}

func TestAuthenticator_TokenDirectoryCreation(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	credentialsFile := filepath.Join(tempDir, "credentials.json")
	// Token file in a subdirectory that doesn't exist yet
	tokenFile := filepath.Join(tempDir, "subdir", "token.json")

	// Create minimal credentials file
	mockCredentials := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob"},
		},
	}

	credentialsData, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}

	err = os.WriteFile(credentialsFile, credentialsData, 0644)
	if err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	authenticator, err := NewAuthenticator(credentialsFile, tokenFile)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Create a test token
	testToken := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Test saving token (should create directory)
	err = authenticator.saveToken(testToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(tokenFile)); os.IsNotExist(err) {
		t.Error("Token directory was not created")
	}

	// Verify token file exists
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		t.Error("Token file was not created")
	}
}

func TestAuthenticator_TokenFilePermissions(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "auth_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	credentialsFile := filepath.Join(tempDir, "credentials.json")
	tokenFile := filepath.Join(tempDir, "token.json")

	// Create minimal credentials file
	mockCredentials := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":     "test_client_id",
			"client_secret": "test_client_secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     "https://oauth2.googleapis.com/token",
			"redirect_uris": []string{"urn:ietf:wg:oauth:2.0:oob"},
		},
	}

	credentialsData, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("Failed to marshal mock credentials: %v", err)
	}

	err = os.WriteFile(credentialsFile, credentialsData, 0644)
	if err != nil {
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	authenticator, err := NewAuthenticator(credentialsFile, tokenFile)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	// Create a test token
	testToken := &oauth2.Token{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Test saving token
	err = authenticator.saveToken(testToken)
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Check file permissions (should be 0600 for security)
	fileInfo, err := os.Stat(tokenFile)
	if err != nil {
		t.Fatalf("Failed to stat token file: %v", err)
	}

	expectedMode := os.FileMode(0600)
	if fileInfo.Mode().Perm() != expectedMode {
		t.Errorf("Expected file permissions %v, got %v", expectedMode, fileInfo.Mode().Perm())
	}
}
