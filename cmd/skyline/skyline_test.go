package skyline

import (
	"testing"

	"github.com/github/gh-skyline/internal/github"
	"github.com/github/gh-skyline/internal/testutil/fixtures"
	"github.com/github/gh-skyline/internal/testutil/mocks"
)

func TestGenerateSkyline(t *testing.T) {
	// Save original initializer
	originalInit := github.InitializeGitHubClient
	defer func() {
		github.InitializeGitHubClient = originalInit
	}()

	tests := []struct {
		name       string
		startYear  int
		endYear    int
		targetUser string
		label      string
		full       bool
		mockClient *mocks.MockGitHubClient
		wantErr    bool
	}{
		{
			name:       "single year",
			startYear:  2024,
			endYear:    2024,
			targetUser: "testuser",
			label:      "",
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
			label:      "",
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
			label:      "",
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
			// Create a closure that returns our mock client
			github.InitializeGitHubClient = func() (*github.Client, error) {
				return github.NewClient(tt.mockClient), nil
			}

			err := GenerateSkyline(tt.startYear, tt.endYear, tt.targetUser, tt.label, tt.full, "", false)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSkyline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
