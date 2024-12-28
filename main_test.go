package main

import (
	"testing"

	"fmt"

	"github.com/github/gh-skyline/github"
	"github.com/github/gh-skyline/testutil/fixtures"
	"github.com/github/gh-skyline/testutil/mocks"
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

func TestFormatYearRange(t *testing.T) {
	tests := []struct {
		name      string
		startYear int
		endYear   int
		want      string
	}{
		{
			name:      "same year",
			startYear: 2024,
			endYear:   2024,
			want:      "2024",
		},
		{
			name:      "different years",
			startYear: 2020,
			endYear:   2024,
			want:      "2020-24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatYearRange(tt.startYear, tt.endYear)
			if got != tt.want {
				t.Errorf("formatYearRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name      string
		user      string
		startYear int
		endYear   int
		want      string
	}{
		{
			name:      "single year",
			user:      "testuser",
			startYear: 2024,
			endYear:   2024,
			want:      "testuser-2024-github-skyline.stl",
		},
		{
			name:      "year range",
			user:      "testuser",
			startYear: 2020,
			endYear:   2024,
			want:      "testuser-2020-24-github-skyline.stl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateOutputFilename(tt.user, tt.startYear, tt.endYear)
			if got != tt.want {
				t.Errorf("generateOutputFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseYearRange(t *testing.T) {
	tests := []struct {
		name          string
		yearRange     string
		wantStart     int
		wantEnd       int
		wantErr       bool
		wantErrString string
	}{
		{
			name:      "single year",
			yearRange: "2024",
			wantStart: 2024,
			wantEnd:   2024,
			wantErr:   false,
		},
		{
			name:      "year range",
			yearRange: "2020-2024",
			wantStart: 2020,
			wantEnd:   2024,
			wantErr:   false,
		},
		{
			name:          "invalid format",
			yearRange:     "2020-2024-2025",
			wantErr:       true,
			wantErrString: "invalid year range format",
		},
		{
			name:      "invalid number",
			yearRange: "abc-2024",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseYearRange(tt.yearRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseYearRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantErrString != "" && err.Error() != tt.wantErrString {
				t.Errorf("parseYearRange() error = %v, wantErrString %v", err, tt.wantErrString)
				return
			}
			if !tt.wantErr {
				if start != tt.wantStart {
					t.Errorf("parseYearRange() start = %v, want %v", start, tt.wantStart)
				}
				if end != tt.wantEnd {
					t.Errorf("parseYearRange() end = %v, want %v", end, tt.wantEnd)
				}
			}
		})
	}
}

func TestValidateYearRange(t *testing.T) {
	tests := []struct {
		name      string
		startYear int
		endYear   int
		wantErr   bool
	}{
		{
			name:      "valid range",
			startYear: 2020,
			endYear:   2024,
			wantErr:   false,
		},
		{
			name:      "invalid start year",
			startYear: 2007,
			endYear:   2024,
			wantErr:   true,
		},
		{
			name:      "start after end",
			startYear: 2024,
			endYear:   2020,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateYearRange(tt.startYear, tt.endYear)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateYearRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSkyline(t *testing.T) {
	// Save original client creation function
	originalInitFn := initializeGitHubClient
	defer func() {
		initializeGitHubClient = originalInitFn
	}()

	tests := []struct {
		name       string
		startYear  int
		endYear    int
		targetUser string
		full       bool
		mockClient *mocks.MockGitHubClient
		wantErr    bool
	}{
		{
			name:       "single year",
			startYear:  2024,
			endYear:    2024,
			targetUser: "testuser",
			full:       false,
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2020,
				MockData: fixtures.GenerateContributionsResponse("testuser", 2024),
			},
			wantErr: false,
		},
		{
			name:       "year range",
			startYear:  2020,
			endYear:    2024,
			targetUser: "testuser",
			full:       false,
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2020,
				MockData: fixtures.GenerateContributionsResponse("testuser", 2024),
			},
			wantErr: false,
		},
		{
			name:       "full range",
			startYear:  2008,
			endYear:    2024,
			targetUser: "testuser",
			full:       true,
			mockClient: &mocks.MockGitHubClient{
				Username: "testuser",
				JoinYear: 2008,
				MockData: fixtures.GenerateContributionsResponse("testuser", 2024),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override the client initialization for testing
			initializeGitHubClient = func() (*github.Client, error) {
				return github.NewClient(tt.mockClient), nil
			}

			err := generateSkyline(tt.startYear, tt.endYear, tt.targetUser, tt.full)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateSkyline() error = %v, wantErr %v", err, tt.wantErr)
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
