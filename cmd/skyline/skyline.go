// Package skyline provides the entry point for the GitHub Skyline Generator.
// It generates a 3D model of GitHub contributions in STL format.
package skyline

import (
	"fmt"
	"strings"
	"time"

	"github.com/github/gh-skyline/internal/ascii"
	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/github"
	"github.com/github/gh-skyline/internal/logger"
	"github.com/github/gh-skyline/internal/stl"
	"github.com/github/gh-skyline/internal/types"
	"github.com/github/gh-skyline/internal/utils"
)

// GitHubClientInterface defines the methods for interacting with GitHub API
type GitHubClientInterface interface {
	GetAuthenticatedUser() (string, error)
	GetUserJoinYear(username string) (int, error)
	FetchContributions(username string, year int) (*types.ContributionsResponse, error)
	FetchOrgContributions(username string, org string, year int) (*types.OrgContributionsResponse, error)
}

// GenerateSkyline creates a 3D model with ASCII art preview of GitHub contributions for the specified year range, or "full lifetime" of the user
func GenerateSkyline(startYear, endYear int, targetUser string, org string, full bool, output string, artOnly bool) error {
	log := logger.GetLogger()

	client, err := github.InitializeGitHubClient()
	if err != nil {
		return errors.New(errors.NetworkError, "failed to initialize GitHub client", err)
	}

	if targetUser == "" {
		if err := log.Debug("No target user specified, using authenticated user"); err != nil {
			return err
		}
		username, err := client.GetAuthenticatedUser()
		if err != nil {
			return errors.New(errors.NetworkError, "failed to get authenticated user", err)
		}
		targetUser = username
	}

	if full {
		joinYear, err := client.GetUserJoinYear(targetUser)
		if err != nil {
			return errors.New(errors.NetworkError, "failed to get user join year", err)
		}
		startYear = joinYear
		endYear = time.Now().Year()
	}

	var allContributions [][][]types.ContributionDay
	for year := startYear; year <= endYear; year++ {
		var contributions [][]types.ContributionDay
		var err error
		if org != "" {
			contributions, err = fetchOrgContributionData(client, targetUser, org, year)
		} else {
			contributions, err = fetchContributionData(client, targetUser, year)
		}
		if err != nil {
			return err
		}
		allContributions = append(allContributions, contributions)

		// Generate ASCII art for each year
		asciiArt, err := ascii.GenerateASCII(contributions, targetUser, year, (year == startYear) && !artOnly, !artOnly)
		if err != nil {
			if warnErr := log.Warning("Failed to generate ASCII preview: %v", err); warnErr != nil {
				return warnErr
			}
		} else {
			if year == startYear {
				// For first year, show full ASCII art including header
				fmt.Println(asciiArt)
			} else {
				// For subsequent years, skip the header
				lines := strings.Split(asciiArt, "\n")
				gridStart := 0
				for i, line := range lines {
					containsEmptyBlock := strings.Contains(line, string(ascii.EmptyBlock))
					containsFoundationLow := strings.Contains(line, string(ascii.FoundationLow))
					isNotOnlyEmptyBlocks := strings.Trim(line, string(ascii.EmptyBlock)) != ""

					if (containsEmptyBlock || containsFoundationLow) && isNotOnlyEmptyBlocks {
						gridStart = i
						break
					}
				}
				// Print just the grid and user info
				fmt.Println(strings.Join(lines[gridStart:], "\n"))
			}
		}
	}

	if !artOnly {
		// Generate filename
		outputPath := utils.GenerateOutputFilename(targetUser, startYear, endYear, output)

		// Generate the STL file
		if len(allContributions) == 1 {
			return stl.GenerateSTL(allContributions[0], outputPath, targetUser, startYear)
		}
		return stl.GenerateSTLRange(allContributions, outputPath, targetUser, startYear, endYear)
	}

	return nil
}

// fetchContributionData retrieves and formats the contribution data for the specified year.
func fetchContributionData(client *github.Client, username string, year int) ([][]types.ContributionDay, error) {
	response, err := client.FetchContributions(username, year)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contributions: %w", err)
	}

	weeks := response.User.ContributionsCollection.ContributionCalendar.Weeks
	contributionGrid := make([][]types.ContributionDay, len(weeks))
	for i, week := range weeks {
		contributionGrid[i] = week.ContributionDays
	}

	return contributionGrid, nil
}

// fetchOrgContributionData retrieves contributions filtered to a specific organization.
func fetchOrgContributionData(client *github.Client, username string, org string, year int) ([][]types.ContributionDay, error) {
	response, err := client.FetchOrgContributions(username, org, year)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch org contributions: %w", err)
	}

	dailyCounts := make(map[string]int)

	for _, repo := range response.User.ContributionsCollection.CommitContributionsByRepository {
		if strings.EqualFold(repo.Repository.Owner.Login, org) {
			for _, node := range repo.Contributions.Nodes {
				date := node.OccurredAt[:10]
				dailyCounts[date]++
			}
		}
	}
	for _, repo := range response.User.ContributionsCollection.IssueContributionsByRepository {
		if strings.EqualFold(repo.Repository.Owner.Login, org) {
			for _, node := range repo.Contributions.Nodes {
				date := node.OccurredAt[:10]
				dailyCounts[date]++
			}
		}
	}
	for _, repo := range response.User.ContributionsCollection.PullRequestContributionsByRepository {
		if strings.EqualFold(repo.Repository.Owner.Login, org) {
			for _, node := range repo.Contributions.Nodes {
				date := node.OccurredAt[:10]
				dailyCounts[date]++
			}
		}
	}
	for _, repo := range response.User.ContributionsCollection.PullRequestReviewContributionsByRepository {
		if strings.EqualFold(repo.Repository.Owner.Login, org) {
			for _, node := range repo.Contributions.Nodes {
				date := node.OccurredAt[:10]
				dailyCounts[date]++
			}
		}
	}

	return buildContributionGrid(year, dailyCounts), nil
}

func buildContributionGrid(year int, dailyCounts map[string]int) [][]types.ContributionDay {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	var weeks [][]types.ContributionDay
	var currentWeek []types.ContributionDay

	for d := startDate; !d.After(endDate) || len(currentWeek) > 0; d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		count := dailyCounts[dateStr]

		if d.Year() == year || (d.Year() == year-1 && d.After(startDate.AddDate(0, 0, -1))) {
			currentWeek = append(currentWeek, types.ContributionDay{
				Date:              dateStr,
				ContributionCount: count,
			})
		}

		if d.Weekday() == time.Saturday || d.Equal(endDate) {
			if len(currentWeek) > 0 {
				weeks = append(weeks, currentWeek)
				currentWeek = nil
			}
		}

		if d.After(endDate) && d.Weekday() == time.Saturday {
			break
		}
	}

	return weeks
}
