// Package mocks provides mock implementations of interfaces used in testing
package mocks

import (
	"fmt"
	"time"

	"github.com/github/gh-skyline/internal/testutil/fixtures"
	"github.com/github/gh-skyline/internal/types"
)

// MockGitHubClient implements both GitHubClientInterface and APIClient interfaces
type MockGitHubClient struct {
	Username string
	JoinYear int
	MockData *types.ContributionsResponse
	Response interface{} // Generic response field for testing
	Err      error       // Error to return if needed
}

// GetAuthenticatedUser implements GitHubClientInterface
func (m *MockGitHubClient) GetAuthenticatedUser() (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if m.Username == "" {
		return "", fmt.Errorf("mock username not set")
	}
	return m.Username, nil
}

// GetUserJoinYear implements GitHubClientInterface
func (m *MockGitHubClient) GetUserJoinYear(_ string) (int, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	if m.JoinYear == 0 {
		return 0, fmt.Errorf("mock join year not set")
	}
	return m.JoinYear, nil
}

// FetchContributions implements GitHubClientInterface
func (m *MockGitHubClient) FetchContributions(username string, year int) (*types.ContributionsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	// Always return generated mock data with valid contributions
	return fixtures.GenerateContributionsResponse(username, year), nil
}

// FetchContributionsForDateRange implements GitHubClientInterface
func (m *MockGitHubClient) FetchContributionsForDateRange(username string, from, to time.Time) (*types.ContributionsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if from.After(to) {
		return nil, fmt.Errorf("start date must be before end date")
	}
	if m.MockData != nil {
		return m.MockData, nil
	}
	return fixtures.GenerateContributionsResponse(username, from.Year()), nil
}

// Do implements APIClient
func (m *MockGitHubClient) Do(_ string, _ map[string]interface{}, response interface{}) error {
	if m.Err != nil {
		return m.Err
	}

	switch v := response.(type) {
	case *struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}:
		v.Viewer.Login = m.Username
	case *struct {
		User struct {
			CreatedAt time.Time `json:"createdAt"`
		} `json:"user"`
	}:
		if m.JoinYear > 0 {
			v.User.CreatedAt = time.Date(m.JoinYear, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	case *types.ContributionsResponse:
		// Always use generated mock data instead of empty response
		mockResp := fixtures.GenerateContributionsResponse(m.Username, time.Now().Year())
		*v = *mockResp
	}
	return nil
}
