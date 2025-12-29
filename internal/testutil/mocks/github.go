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
	Username    string
	JoinYear    int
	MockData    *types.ContributionsResponse
	MockOrgData *types.OrgContributionsResponse
	Response    interface{}
	Err         error
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
	return fixtures.GenerateContributionsResponse(username, year), nil
}

// FetchOrgContributions implements GitHubClientInterface
func (m *MockGitHubClient) FetchOrgContributions(username string, _ string, _ int) (*types.OrgContributionsResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.MockOrgData != nil {
		return m.MockOrgData, nil
	}
	return GenerateOrgContributionsResponse(username, "testorg"), nil
}

// GenerateOrgContributionsResponse creates mock org contribution data for testing
func GenerateOrgContributionsResponse(username, org string) *types.OrgContributionsResponse {
	resp := &types.OrgContributionsResponse{}
	resp.User.Login = username

	resp.User.ContributionsCollection.CommitContributionsByRepository = []struct {
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"repository"`
		Contributions struct {
			TotalCount int `json:"totalCount"`
			Nodes      []struct {
				OccurredAt string `json:"occurredAt"`
			} `json:"nodes"`
		} `json:"contributions"`
	}{
		{
			Repository: struct {
				Name  string `json:"name"`
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			}{
				Name: "test-repo",
				Owner: struct {
					Login string `json:"login"`
				}{Login: org},
			},
			Contributions: struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					OccurredAt string `json:"occurredAt"`
				} `json:"nodes"`
			}{
				TotalCount: 3,
				Nodes: []struct {
					OccurredAt string `json:"occurredAt"`
				}{
					{OccurredAt: "2024-01-15T10:00:00Z"},
					{OccurredAt: "2024-01-15T14:00:00Z"},
					{OccurredAt: "2024-02-20T09:00:00Z"},
				},
			},
		},
	}

	return resp
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
		mockResp := fixtures.GenerateContributionsResponse(m.Username, time.Now().Year())
		*v = *mockResp
	case *types.OrgContributionsResponse:
		if m.MockOrgData != nil {
			*v = *m.MockOrgData
		} else {
			*v = *GenerateOrgContributionsResponse(m.Username, "testorg")
		}
	}
	return nil
}
