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
	FetchContributionsForDateRange(username string, from, to time.Time) (*types.ContributionsResponse, error)
}

// GenerateSkyline creates a 3D model with ASCII art preview of GitHub contributions for the specified year range, or "full lifetime" of the user
func GenerateSkyline(startYear, endYear int, targetUser string, full bool, output string, artOnly bool, ytdEnd string) error {
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

	// Handle YTD mode (last 12 months)
	now := time.Now()
	isYTD := endYear == now.Year() && startYear == now.AddDate(0, -12, 0).Year()
	var allContributions [][][]types.ContributionDay

	if isYTD {
		// For YTD mode, fetch a single continuous period from 12 months ago until now
		endDate := now
		if ytdEnd != "" {
			parsedEnd, err := time.Parse("2006-01-02", ytdEnd)
			if err != nil {
				return fmt.Errorf("invalid ytd-end date format, expected YYYY-MM-DD: %v", err)
			}
			if parsedEnd.After(now) {
				return fmt.Errorf("ytd-end date cannot be in the future")
			}
			endDate = parsedEnd
		}
		startDate := endDate.AddDate(0, -12, 0)

		response, err := client.FetchContributionsForDateRange(targetUser, startDate, endDate)
		if err != nil {
			return err
		}
		contributions := convertResponseToGrid(response)
		allContributions = append(allContributions, contributions)

		// Generate ASCII art for the YTD period
		// Use both years for YTD display
		yearDisplay := fmt.Sprintf("%d-%d", startDate.Year(), endDate.Year())
		asciiArt, err := ascii.GenerateASCII(contributions, targetUser, startYear, true, !artOnly)
		if err != nil {
			if warnErr := log.Warning("Failed to generate ASCII preview: %v", err); warnErr != nil {
				return warnErr
			}
		} else {
			// Replace the year display in ASCII art
			lines := strings.Split(asciiArt, "\n")
			for i, line := range lines {
				if strings.Contains(line, fmt.Sprintf("%d", startYear)) {
					lines[i] = strings.ReplaceAll(line, fmt.Sprintf("%d", startYear), yearDisplay)
					break
				}
			}
			fmt.Println(strings.Join(lines, "\n"))
		}
	} else {
		// Handle regular year-based contributions
		for year := startYear; year <= endYear; year++ {
			response, err := client.FetchContributions(targetUser, year)
			if err != nil {
				return err
			}
			contributions := convertResponseToGrid(response)
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
	}

	if !artOnly {
		// Generate filename with custom end date for YTD mode
		outputPath := utils.GenerateOutputFilename(targetUser, startYear, endYear, output, ytdEnd)

		// Generate the STL file
		if len(allContributions) == 1 {
			if isYTD {
				// For YTD mode, pass both years to the STL generator
				return stl.GenerateSTLRange(allContributions, outputPath, targetUser, startYear, endYear)
			}
			return stl.GenerateSTL(allContributions[0], outputPath, targetUser, startYear)
		}
		return stl.GenerateSTLRange(allContributions, outputPath, targetUser, startYear, endYear)
	}

	return nil
}

// convertResponseToGrid converts a GitHub API response to a 2D grid of contributions
func convertResponseToGrid(response *types.ContributionsResponse) [][]types.ContributionDay {
	weeks := response.User.ContributionsCollection.ContributionCalendar.Weeks
	contributionGrid := make([][]types.ContributionDay, len(weeks))
	for i, week := range weeks {
		contributionGrid[i] = week.ContributionDays
	}
	return contributionGrid
}
