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
}

// GenerateSkyline creates a 3D model with ASCII art preview of GitHub contributions for the specified year range, or "full lifetime" of the user.
// The 'label' parameter customizes the display name in the output; if left empty, it defaults to the targetUser.
func GenerateSkyline(startYear, endYear int, targetUser string, label string, full bool, output string, artOnly bool) error {
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
	if label == "" {
		label = targetUser
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
		contributions, err := fetchContributionData(client, targetUser, year)
		if err != nil {
			return err
		}
		allContributions = append(allContributions, contributions)

		// Generate ASCII art for each year
		asciiArt, err := ascii.GenerateASCII(contributions, label, year, (year == startYear) && !artOnly, !artOnly)
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
			return stl.GenerateSTL(allContributions[0], outputPath, label, startYear)
		}
		return stl.GenerateSTLRange(allContributions, outputPath, label, startYear, endYear)
	}

	return nil
}

// fetchContributionData retrieves and formats the contribution data for the specified year.
func fetchContributionData(client *github.Client, username string, year int) ([][]types.ContributionDay, error) {
	response, err := client.FetchContributions(username, year)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contributions: %w", err)
	}

	// Convert weeks data to 2D array for STL generation
	weeks := response.User.ContributionsCollection.ContributionCalendar.Weeks
	contributionGrid := make([][]types.ContributionDay, len(weeks))
	for i, week := range weeks {
		contributionGrid[i] = week.ContributionDays
	}

	return contributionGrid, nil
}
