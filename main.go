// Package main provides the entry point for the GitHub Skyline Generator.
// It generates a 3D model of GitHub contributions in STL format.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/github/gh-skyline/ascii"
	"github.com/github/gh-skyline/errors"
	"github.com/github/gh-skyline/github"
	"github.com/github/gh-skyline/logger"
	"github.com/github/gh-skyline/stl"
	"github.com/github/gh-skyline/types"
	"github.com/spf13/cobra"
)

// Browser interface matches browser.Browser functionality
type Browser interface {
	Browse(url string) error
}

// GitHubClientInterface defines the methods for interacting with GitHub API
type GitHubClientInterface interface {
	GetAuthenticatedUser() (string, error)
	GetUserJoinYear(username string) (int, error)
	FetchContributions(username string, year int) (*types.ContributionsResponse, error)
}

// Constants for GitHub launch year and default output file format
const (
	githubLaunchYear = 2008
	outputFileFormat = "%s-%s-github-skyline.stl"
)

// Command line variables and root command configuration
var (
	yearRange string
	user      string
	full      bool
	debug     bool
	web       bool
	artOnly   bool
	output    string // new output path flag

	rootCmd = &cobra.Command{
		Use:   "skyline",
		Short: "Generate a 3D model of a user's GitHub contribution history",
		Long: `GitHub Skyline creates 3D printable STL files from GitHub contribution data.
It can generate models for specific years or year ranges for the authenticated user or an optional specified user.

While the STL file is being generated, an ASCII preview will be displayed in the terminal.

ASCII Preview Legend:
  ' ' Empty/Sky     - No contributions
  '.' Future dates  - What contributions could you make?
  '░' Low level     - Light contribution activity
  '▒' Medium level  - Moderate contribution activity
  '▓' High level    - Heavy contribution activity
  '╻┃╽' Top level   - Last block with contributions in the week (Low, Medium, High)

Layout:
Each column represents one week. Days within each week are reordered vertically
to create a "building" effect, with empty spaces (no contributions) at the top.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			log := logger.GetLogger()
			if debug {
				log.SetLevel(logger.DEBUG)
				if err := log.Debug("Debug logging enabled"); err != nil {
					return err
				}
			}

			client, err := initializeGitHubClient()
			if err != nil {
				return errors.New(errors.NetworkError, "failed to initialize GitHub client", err)
			}

			if web {
				b := browser.New("", os.Stdout, os.Stderr)
				if err := openGitHubProfile(user, client, b); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				return nil
			}

			startYear, endYear, err := parseYearRange(yearRange)
			if err != nil {
				return fmt.Errorf("invalid year range: %v", err)
			}

			return generateSkyline(startYear, endYear, user, full)
		},
	}
)

// init sets up command line flags for the skyline CLI tool
func init() {
	rootCmd.Flags().StringVarP(&yearRange, "year", "y", fmt.Sprintf("%d", time.Now().Year()), "Year or year range (e.g., 2024 or 2014-2024)")
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "GitHub username (optional, defaults to authenticated user)")
	rootCmd.Flags().BoolVarP(&full, "full", "f", false, "Generate contribution graph from join year to current year")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.Flags().BoolVarP(&web, "web", "w", false, "Open GitHub profile (authenticated or specified user).")
	rootCmd.Flags().BoolVarP(&artOnly, "art-only", "a", false, "Generate only ASCII preview")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (optional)")
}

// main initializes and executes the root command for the GitHub Skyline CLI
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// formatYearRange returns a formatted string representation of the year range
func formatYearRange(startYear, endYear int) string {
	if startYear == endYear {
		return fmt.Sprintf("%d", startYear)
	}
	return fmt.Sprintf("%02d-%02d", startYear%100, endYear%100)
}

// generateOutputFilename creates a consistent filename for the STL output
func generateOutputFilename(user string, startYear, endYear int) string {
	if output != "" {
		// Ensure the filename ends with .stl
		if !strings.HasSuffix(strings.ToLower(output), ".stl") {
			return output + ".stl"
		}
		return output
	}
	yearStr := formatYearRange(startYear, endYear)
	return fmt.Sprintf(outputFileFormat, user, yearStr)
}

// generateSkyline creates a 3D model with ASCII art preview of GitHub contributions for the specified year range, or "full lifetime" of the user
func generateSkyline(startYear, endYear int, targetUser string, full bool) error {
	log := logger.GetLogger()

	client, err := initializeGitHubClient()
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
		contributions, err := fetchContributionData(client, targetUser, year)
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
					if strings.Contains(line, string(ascii.EmptyBlock)) ||
						strings.Contains(line, string(ascii.FoundationLow)) {
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
		outputPath := generateOutputFilename(targetUser, startYear, endYear)

		// Generate the STL file
		if len(allContributions) == 1 {
			return stl.GenerateSTL(allContributions[0], outputPath, targetUser, startYear)
		}
		return stl.GenerateSTLRange(allContributions, outputPath, targetUser, startYear, endYear)
	}

	return nil
}

// Variable for client initialization - allows for testing
var initializeGitHubClient = defaultGitHubClient

// defaultGitHubClient is the default implementation of client initialization
func defaultGitHubClient() (*github.Client, error) {
	apiClient, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}
	return github.NewClient(apiClient), nil
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

// Parse year range string (e.g., "2024" or "2014-2024")
func parseYearRange(yearRange string) (startYear, endYear int, err error) {
	if strings.Contains(yearRange, "-") {
		parts := strings.Split(yearRange, "-")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid year range format")
		}
		startYear, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		endYear, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
	} else {
		year, err := strconv.Atoi(yearRange)
		if err != nil {
			return 0, 0, err
		}
		startYear, endYear = year, year
	}
	return startYear, endYear, validateYearRange(startYear, endYear)
}

// validateYearRange checks if the years are within the range
// of GitHub's launch year to the current year and if
// the start year is not greater than the end year.
func validateYearRange(startYear, endYear int) error {
	currentYear := time.Now().Year()
	if startYear < githubLaunchYear || endYear > currentYear {
		return fmt.Errorf("years must be between %d and %d", githubLaunchYear, currentYear)
	}
	if startYear > endYear {
		return fmt.Errorf("start year cannot be after end year")
	}
	return nil
}

// openGitHubProfile opens the GitHub profile page for the specified user or authenticated user
func openGitHubProfile(targetUser string, client GitHubClientInterface, b Browser) error {
	if targetUser == "" {
		username, err := client.GetAuthenticatedUser()
		if err != nil {
			return errors.New(errors.NetworkError, "failed to get authenticated user", err)
		}
		targetUser = username
	}

	profileURL := fmt.Sprintf("https://github.com/%s", targetUser)
	return b.Browse(profileURL)
}
