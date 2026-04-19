// Package github provides a client for interacting with the GitHub API,
// including fetching authenticated user information and contribution data.
package github

import (
	"fmt"
	"time"

	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/types"
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

// FetchOrgContributions retrieves contribution data filtered by organization for a given username and year.
//
// GitHub's GraphQL API limits contributionsByRepository queries to 100 repositories per request.
// To work around this, we query each quarter separately and merge the results. This allows us to
// capture contributions across up to 400 repos per year (100 per quarter), which covers virtually
// all real-world usage patterns. The quarterly approach also reduces the chance of hitting the
// per-repo contribution limit (100 nodes) since contributions are spread across shorter time ranges.
func (c *Client) FetchOrgContributions(username string, org string, year int) (*types.OrgContributionsResponse, error) {
	if username == "" {
		return nil, errors.New(errors.ValidationError, "username cannot be empty", nil)
	}
	if org == "" {
		return nil, errors.New(errors.ValidationError, "org cannot be empty", nil)
	}
	if year < 2008 {
		return nil, errors.New(errors.ValidationError, "year cannot be before GitHub's launch (2008)", nil)
	}

	// Query each quarter separately to work around the 100 repo limit per query.
	// Each query can return up to 100 unique repos, so quarterly queries give us up to 400 repos/year.
	quarters := []struct {
		start string
		end   string
	}{
		{fmt.Sprintf("%d-01-01T00:00:00Z", year), fmt.Sprintf("%d-03-31T23:59:59Z", year)},
		{fmt.Sprintf("%d-04-01T00:00:00Z", year), fmt.Sprintf("%d-06-30T23:59:59Z", year)},
		{fmt.Sprintf("%d-07-01T00:00:00Z", year), fmt.Sprintf("%d-09-30T23:59:59Z", year)},
		{fmt.Sprintf("%d-10-01T00:00:00Z", year), fmt.Sprintf("%d-12-31T23:59:59Z", year)},
	}

	query := `
    query ContributionsByRepo($username: String!, $from: DateTime!, $to: DateTime!) {
        user(login: $username) {
            login
            contributionsCollection(from: $from, to: $to) {
                commitContributionsByRepository(maxRepositories: 100) {
                    repository {
                        name
                        owner { login }
                    }
                    contributions(first: 100) {
                        totalCount
                        nodes { occurredAt }
                    }
                }
                issueContributionsByRepository(maxRepositories: 100) {
                    repository { owner { login } }
                    contributions(first: 100) { nodes { occurredAt } }
                }
                pullRequestContributionsByRepository(maxRepositories: 100) {
                    repository { owner { login } }
                    contributions(first: 100) { nodes { occurredAt } }
                }
                pullRequestReviewContributionsByRepository(maxRepositories: 100) {
                    repository { owner { login } }
                    contributions(first: 100) { nodes { occurredAt } }
                }
            }
        }
    }`

	var merged types.OrgContributionsResponse

	for _, q := range quarters {
		variables := map[string]interface{}{
			"username": username,
			"from":     q.start,
			"to":       q.end,
		}

		var response types.OrgContributionsResponse
		err := c.api.Do(query, variables, &response)
		if err != nil {
			return nil, errors.New(errors.NetworkError, "failed to fetch org contributions", err)
		}

		if response.User.Login == "" {
			return nil, errors.New(errors.ValidationError, "received empty username from GitHub API", nil)
		}

		// Merge quarterly results into the combined response
		merged.User.Login = response.User.Login
		merged.User.ContributionsCollection.CommitContributionsByRepository = append(
			merged.User.ContributionsCollection.CommitContributionsByRepository,
			response.User.ContributionsCollection.CommitContributionsByRepository...)
		merged.User.ContributionsCollection.IssueContributionsByRepository = append(
			merged.User.ContributionsCollection.IssueContributionsByRepository,
			response.User.ContributionsCollection.IssueContributionsByRepository...)
		merged.User.ContributionsCollection.PullRequestContributionsByRepository = append(
			merged.User.ContributionsCollection.PullRequestContributionsByRepository,
			response.User.ContributionsCollection.PullRequestContributionsByRepository...)
		merged.User.ContributionsCollection.PullRequestReviewContributionsByRepository = append(
			merged.User.ContributionsCollection.PullRequestReviewContributionsByRepository,
			response.User.ContributionsCollection.PullRequestReviewContributionsByRepository...)
	}

	return &merged, nil
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

	// Validate that the API returned a real creation date
	if response.User.CreatedAt.IsZero() {
		return 0, errors.New(errors.ValidationError, "invalid join date received from GitHub API", nil)
	}

	// Parse the join date
	joinYear := response.User.CreatedAt.Year()

	return joinYear, nil
}
