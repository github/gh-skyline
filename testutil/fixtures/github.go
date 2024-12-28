// Package fixtures provides test utilities and mock data generators
// for testing the gh-skyline application.
package fixtures

import (
	"time"

	"github.com/github/gh-skyline/types"
)

// GenerateContributionsResponse creates a mock contributions response
func GenerateContributionsResponse(username string, year int) *types.ContributionsResponse {
	response := &types.ContributionsResponse{}
	response.User.Login = username
	response.User.ContributionsCollection.ContributionCalendar.TotalContributions = 100

	// Create sample weeks with contribution days
	weeks := make([]struct {
		ContributionDays []types.ContributionDay `json:"contributionDays"`
	}, 52)

	for i := range weeks {
		days := make([]types.ContributionDay, 7)
		for j := range days {
			days[j] = types.ContributionDay{
				ContributionCount: (i + j) % 10,
				Date:              time.Date(year, 1, 1+i*7+j, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
			}
		}
		weeks[i].ContributionDays = days
	}

	response.User.ContributionsCollection.ContributionCalendar.Weeks = weeks
	return response
}

// CreateMockContributionDay creates a mock contribution day
func CreateMockContributionDay(date time.Time, count int) types.ContributionDay {
	return types.ContributionDay{
		ContributionCount: count,
		Date:              date.Format("2006-01-02"),
	}
}
