package cmd

import (
	"fmt"
	"testing"

	"github.com/github/gh-skyline/internal/testutil/mocks"
)

// MockBrowser implements the Browser interface
type MockBrowser struct {
	LastURL string
	Err     error
}

// Browse implements the Browser interface
func (m *MockBrowser) Browse(url string) error {
	m.LastURL = url
	return m.Err
}

func TestRootCmd(t *testing.T) {
	cmd := rootCmd
	if cmd.Use != "skyline" {
		t.Errorf("expected command use to be 'skyline', got %s", cmd.Use)
	}
	if cmd.Short != "Generate a 3D model of a user's GitHub contribution history" {
		t.Errorf("expected command short description to be 'Generate a 3D model of a user's GitHub contribution history', got %s", cmd.Short)
	}
	if cmd.Long == "" {
		t.Error("expected command long description to be non-empty")
	}
}

func TestInit(t *testing.T) {
	flags := rootCmd.Flags()
	expectedFlags := []string{"year", "user", "full", "debug", "web", "art-only", "output", "multi-part"}
	for _, flag := range expectedFlags {
		if flags.Lookup(flag) == nil {
			t.Errorf("expected flag %s to be initialized", flag)
		}
	}
}

// TestOpenGitHubProfile tests the openGitHubProfile function
func TestOpenGitHubProfile(t *testing.T) {
	tests := []struct {
		name       string
		targetUser string
		mockClient *mocks.MockGitHubClient
		wantURL    string
		wantErr    bool
	}{
		{
			name:       "specific user",
			targetUser: "testuser",
			mockClient: &mocks.MockGitHubClient{},
			wantURL:    "https://github.com/testuser",
			wantErr:    false,
		},
		{
			name:       "authenticated user",
			targetUser: "",
			mockClient: &mocks.MockGitHubClient{
				Username: "authuser",
			},
			wantURL: "https://github.com/authuser",
			wantErr: false,
		},
		{
			name:       "client error",
			targetUser: "",
			mockClient: &mocks.MockGitHubClient{
				Err: fmt.Errorf("mock error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBrowser := &MockBrowser{}
			if tt.wantErr {
				mockBrowser.Err = fmt.Errorf("mock error")
			}
			err := openGitHubProfile(tt.targetUser, tt.mockClient, mockBrowser)

			if (err != nil) != tt.wantErr {
				t.Errorf("openGitHubProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && mockBrowser.LastURL != tt.wantURL {
				t.Errorf("openGitHubProfile() URL = %v, want %v", mockBrowser.LastURL, tt.wantURL)
			}
		})
	}
}
