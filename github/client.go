// Package github provides a client for interacting with the GitHub API,
// including fetching authenticated user information and contribution data.
package github

import (
	"fmt"
	"time"

	"github.com/github/gh-skyline/errors"
	"github.com/github/gh-skyline/types"
)

// APIClient interface defines the methods we need from the client
type APIClient interface {
	Do(query string, variables map[string]interface{}, response interface{}) error
}

// Client holds the API client
type Client struct {
	api APIClient
}

// NewClient creates a new GitHub client
func NewClient(apiClient APIClient) *Client {
	return &Client{api: apiClient}
}

// GetAuthenticatedUser fetches the authenticated user's login name from GitHub.
func (c *Client) GetAuthenticatedUser() (string, error) {
	// GraphQL query to fetch the authenticated user's login.
	query := `
    query {
        viewer {
            login
        }
    }`

	var response struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}

	// Execute the GraphQL query.
	err := c.api.Do(query, nil, &response)
	if err != nil {
		return "", errors.New(errors.NetworkError, "failed to fetch authenticated user", err)
	}

	if response.Viewer.Login == "" {
		return "", errors.New(errors.ValidationError, "received empty username from GitHub API", nil)
	}

	return response.Viewer.Login, nil
}

// FetchContributions retrieves the contribution data for a given username and year from GitHub.
func (c *Client) FetchContributions(username string, year int) (*types.ContributionsResponse, error) {
	if username == "" {
		return nil, errors.New(errors.ValidationError, "username cannot be empty", nil)
	}

	if year < 2008 {
		return nil, errors.New(errors.ValidationError, "year cannot be before GitHub's launch (2008)", nil)
	}

	startDate := fmt.Sprintf("%d-01-01T00:00:00Z", year)
	endDate := fmt.Sprintf("%d-12-31T23:59:59Z", year)

	// GraphQL query to fetch the user's contributions within the specified date range.
	query := `
    query ContributionGraph($username: String!, $from: DateTime!, $to: DateTime!) {
        user(login: $username) {
            login
            contributionsCollection(from: $from, to: $to) {
                contributionCalendar {
                    totalContributions
                    weeks {
                        contributionDays {
                            contributionCount
                            date
                        }
                    }
                }
            }
        }
    }`

	variables := map[string]interface{}{
		"username": username,
		"from":     startDate,
		"to":       endDate,
	}

	var response types.ContributionsResponse

	// Execute the GraphQL query.
	err := c.api.Do(query, variables, &response)
	if err != nil {
		return nil, errors.New(errors.NetworkError, "failed to fetch contributions", err)
	}

	if response.User.Login == "" {
		return nil, errors.New(errors.ValidationError, "received empty username from GitHub API", nil)
	}

	return &response, nil
}

// GetUserJoinYear fetches the year a user joined GitHub using the GitHub API.
func (c *Client) GetUserJoinYear(username string) (int, error) {
	if username == "" {
		return 0, errors.New(errors.ValidationError, "username cannot be empty", nil)
	}

	// GraphQL query to fetch the user's account creation date.
	query := `
    query UserJoinDate($username: String!) {
        user(login: $username) {
            createdAt
        }
    }`

	variables := map[string]interface{}{
		"username": username,
	}

	var response struct {
		User struct {
			CreatedAt time.Time `json:"createdAt"`
		} `json:"user"`
	}

	// Execute the GraphQL query.
	err := c.api.Do(query, variables, &response)
	if err != nil {
		return 0, errors.New(errors.NetworkError, "failed to fetch user's join date", err)
	}

	// Parse the join date
	joinYear := response.User.CreatedAt.Year()
	if joinYear == 0 {
		return 0, errors.New(errors.ValidationError, "invalid join date received from GitHub API", nil)
	}

	return joinYear, nil
}
