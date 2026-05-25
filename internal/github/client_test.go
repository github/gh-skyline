package github

import (
	"testing"

	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/testutil/mocks"
	"github.com/github/gh-skyline/internal/types"
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
		mockError     error
		expectedYear  int
		expectedError bool
	}{
		{
			name:          "successful response",
			username:      "testuser",
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
		{
			name:     "zero join date",
			username: "testuser",
			// expectedYear is 0, so mock.Do leaves CreatedAt as zero time.Time, triggering IsZero guard
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

func TestFetchOrgContributions(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		org           string
		year          int
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful response",
			username:      "testuser",
			org:           "testorg",
			year:          2024,
			expectedError: false,
		},
		{
			name:          "empty username",
			username:      "",
			org:           "testorg",
			year:          2024,
			expectedError: true,
		},
		{
			name:          "empty org",
			username:      "testuser",
			org:           "",
			year:          2024,
			expectedError: true,
		},
		{
			name:          "invalid year",
			username:      "testuser",
			org:           "testorg",
			year:          2007,
			expectedError: true,
		},
		{
			name:          "network error",
			username:      "testuser",
			org:           "testorg",
			year:          2024,
			mockError:     errors.New(errors.NetworkError, "network error", nil),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&mocks.MockGitHubClient{
				Username: tt.username,
				Err:      tt.mockError,
			})

			resp, err := client.FetchOrgContributions(tt.username, tt.org, tt.year)
			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if !tt.expectedError {
				if resp == nil {
					t.Error("expected response but got nil")
				} else if resp.User.Login != tt.username {
					t.Errorf("expected user %s, got %s", tt.username, resp.User.Login)
				}
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
