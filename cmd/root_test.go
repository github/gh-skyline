package cmd

import (
	"fmt"
	"testing"
	"time"

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
	expectedFlags := []string{
		"year",
		"user",
		"full",
		"debug",
		"web",
		"art-only",
		"output",
		"year-to-date",
		"until",
	}
	for _, flag := range expectedFlags {
		if flags.Lookup(flag) == nil {
			t.Errorf("expected flag %s to be initialized", flag)
		}
	}
}

func TestMutuallyExclusiveFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "year and year-to-date are mutually exclusive",
			args:    []string{"--year", "2024", "--year-to-date"},
			wantErr: true,
		},
		{
			name:    "year and full are mutually exclusive",
			args:    []string{"--year", "2024", "--full"},
			wantErr: true,
		},
		{
			name:    "full and year-to-date are mutually exclusive",
			args:    []string{"--full", "--year-to-date"},
			wantErr: true,
		},
		{
			name:    "until requires year-to-date",
			args:    []string{"--until", "2024-03-21"},
			wantErr: true,
		},
		{
			name:    "valid year-to-date with until",
			args:    []string{"--year-to-date", "--until", "2024-03-21"},
			wantErr: false,
		},
		{
			name:    "valid single year",
			args:    []string{"--year", "2024"},
			wantErr: false,
		},
		{
			name:    "valid year-to-date",
			args:    []string{"--year-to-date"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYearToDatePeriodCalculation(t *testing.T) {
	tests := []struct {
		name          string
		untilDate     string
		wantStartYear int
		wantEndYear   int
		wantErr       bool
	}{
		{
			name:          "default to current date",
			untilDate:     "",
			wantStartYear: time.Now().AddDate(0, -12, 0).Year(),
			wantEndYear:   time.Now().Year(),
			wantErr:       false,
		},
		{
			name:          "specific date in current year",
			untilDate:     time.Now().Format("2006-01-02"),
			wantStartYear: time.Now().AddDate(0, -12, 0).Year(),
			wantEndYear:   time.Now().Year(),
			wantErr:       false,
		},
		{
			name:          "invalid date format",
			untilDate:     "2024/03/21",
			wantStartYear: 0,
			wantEndYear:   0,
			wantErr:       true,
		},
		{
			name:          "future date",
			untilDate:     time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			wantStartYear: 0,
			wantEndYear:   0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			yearToDate = true
			until = tt.untilDate
			yearRange = ""

			err := handleSkylineCommand(rootCmd, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleSkylineCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check if the yearRange was set correctly
				expectedYearRange := fmt.Sprintf("%d-%d", tt.wantStartYear, tt.wantEndYear)
				if yearRange != expectedYearRange {
					t.Errorf("yearRange = %v, want %v", yearRange, expectedYearRange)
				}
			}
		})
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
