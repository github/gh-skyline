package github

import (
	"testing"
	"time"

	"github.com/github/gh-skyline/errors"
	"github.com/github/gh-skyline/testutil/mocks"
	"github.com/github/gh-skyline/types"
)

func TestGetAuthenticatedUser(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		mockError     error
		expectedUser  string
		expectedError bool
	}{
		{
			name:          "successful response",
			mockResponse:  "testuser",
			expectedUser:  "testuser",
			expectedError: false,
		},
		{
			name:          "empty username",
			mockResponse:  "",
			expectedError: true,
		},
		{
			name:          "network error",
			mockError:     errors.New(errors.NetworkError, "network error", nil),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mocks.MockGitHubClient{
				Username: tt.mockResponse,
				Err:      tt.mockError,
			})

			user, err := client.GetAuthenticatedUser()
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if user != tt.expectedUser {
				t.Errorf("expected user %q, got %q", tt.expectedUser, user)
			}
		})
	}
}

func TestGetUserJoinYear(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		mockResponse  time.Time
		mockError     error
		expectedYear  int
		expectedError bool
	}{
		{
			name:          "successful response",
			username:      "testuser",
			mockResponse:  time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedYear:  2015,
			expectedError: false,
		},
		{
			name:          "empty username",
			username:      "",
			expectedError: true,
		},
		{
			name:          "network error",
			username:      "testuser",
			mockError:     errors.New(errors.NetworkError, "network error", nil),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mocks.MockGitHubClient{
				JoinYear: tt.expectedYear,
				Err:      tt.mockError,
			})

			year, err := client.GetUserJoinYear(tt.username)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if !tt.expectedError && year != tt.expectedYear {
				t.Errorf("expected year %d, got %d", tt.expectedYear, year)
			}
		})
	}
}

func TestFetchContributions(t *testing.T) {
	mockContributions := &types.ContributionsResponse{
		User: struct {
			Login                   string `json:"login"`
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []types.ContributionDay `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		}{
			Login: "chrisreddington",
			ContributionsCollection: struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []types.ContributionDay `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			}{
				ContributionCalendar: struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []types.ContributionDay `json:"contributionDays"`
					} `json:"weeks"`
				}{
					TotalContributions: 100,
					Weeks: []struct {
						ContributionDays []types.ContributionDay `json:"contributionDays"`
					}{
						{
							ContributionDays: []types.ContributionDay{
								{
									ContributionCount: 5,
									Date:              "2023-01-01",
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name          string
		username      string
		year          int
		mockResponse  *types.ContributionsResponse
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful response",
			username:      "testuser",
			year:          2023,
			mockResponse:  mockContributions,
			expectedError: false,
		},
		{
			name:          "empty username",
			username:      "",
			year:          2023,
			expectedError: true,
		},
		{
			name:          "invalid year",
			username:      "testuser",
			year:          2007,
			expectedError: true,
		},
		{
			name:          "network error",
			username:      "testuser",
			year:          2023,
			mockError:     errors.New(errors.NetworkError, "network error", nil),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mocks.MockGitHubClient{
				Username: tt.username,
				MockData: tt.mockResponse,
				Err:      tt.mockError,
			})

			resp, err := client.FetchContributions(tt.username, tt.year)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if !tt.expectedError {
				if resp == nil {
					t.Error("expected response but got nil")
				} else if resp.User.Login != "testuser" {
					t.Errorf("expected user testuser, got %s", resp.User.Login)
				}
			}
		})
	}
}
