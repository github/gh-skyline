// Package cmd is a package that contains the root command (entrypoint) for the GitHub Skyline CLI tool.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/github/gh-skyline/cmd/skyline"
	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/github"
	"github.com/github/gh-skyline/internal/logger"
	"github.com/github/gh-skyline/internal/utils"
	"github.com/spf13/cobra"
)

// Command line variables and root command configuration
var (
	yearRange  string
	user       string
	full       bool
	debug      bool
	web        bool
	artOnly    bool
	output     string
	yearToDate bool   // renamed from ytd
	until      string // renamed from ytdEnd
)

// rootCmd is the root command for the GitHub Skyline CLI tool.
var rootCmd = &cobra.Command{
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
to create a "building" effect, with empty spaces (no contributions) at the top.

Examples:
  # Generate skyline for current year
  gh skyline

  # Generate skyline for a specific year
  gh skyline --year 2023

  # Generate skyline for the last 12 months up to today
  gh skyline --year-to-date

  # Generate skyline for the last 12 months up to a specific date
  gh skyline --year-to-date --until 2024-02-29`,
	RunE: handleSkylineCommand,
}

// init initializes command line flags for the skyline CLI tool.
func init() {
	initFlags()
}

// Execute initializes and executes the root command for the GitHub Skyline CLI.
func Execute(_ context.Context) error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

// initFlags sets up command line flags for the skyline CLI tool.
func initFlags() {
	flags := rootCmd.Flags()
	flags.StringVarP(&yearRange, "year", "y", fmt.Sprintf("%d", time.Now().Year()), "Year or year range (e.g., 2024 or 2014-2024)")
	flags.StringVarP(&user, "user", "u", "", "GitHub username (optional, defaults to authenticated user)")
	flags.BoolVarP(&full, "full", "f", false, "Generate contribution graph from join year to current year")
	flags.BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	flags.BoolVarP(&web, "web", "w", false, "Open GitHub profile (authenticated or specified user).")
	flags.BoolVarP(&artOnly, "art-only", "a", false, "Generate only ASCII preview")
	flags.StringVarP(&output, "output", "o", "", "Output file path (optional)")
	flags.BoolVar(&yearToDate, "year-to-date", false, "Generate contribution graph for the last 12 months")
	flags.StringVar(&until, "until", "", "End date for year-to-date period in YYYY-MM-DD format (defaults to today)")

	// Mark mutually exclusive flags
	rootCmd.MarkFlagsMutuallyExclusive("year", "year-to-date")
	rootCmd.MarkFlagsMutuallyExclusive("year", "full")
	rootCmd.MarkFlagsMutuallyExclusive("full", "year-to-date")

	// Add validation to ensure --until is only used with --year-to-date
	rootCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed("until") && !yearToDate {
			return fmt.Errorf("--until can only be used with --year-to-date")
		}
		return nil
	}
}

// executeRootCmd is the main execution function for the root command.
func handleSkylineCommand(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger()
	if debug {
		log.SetLevel(logger.DEBUG)
		if err := log.Debug("Debug logging enabled"); err != nil {
			return err
		}
	}

	client, err := github.InitializeGitHubClient()
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

	if yearToDate {
		// Parse custom end date if provided
		now := time.Now()
		endDate := now
		if until != "" {
			parsedEnd, err := time.Parse("2006-01-02", until)
			if err != nil {
				return fmt.Errorf("invalid until date format, expected YYYY-MM-DD: %v", err)
			}
			if parsedEnd.After(now) {
				return fmt.Errorf("until date cannot be in the future")
			}
			endDate = parsedEnd
		}

		// Calculate start date as 12 months before end date
		startDate := endDate.AddDate(0, -12, 0)
		yearRange = fmt.Sprintf("%d-%d", startDate.Year(), endDate.Year())
	}

	startYear, endYear, err := utils.ParseYearRange(yearRange)
	if err != nil {
		return fmt.Errorf("invalid year range: %v", err)
	}

	return skyline.GenerateSkyline(startYear, endYear, user, full, output, artOnly, until)
}

// Browser interface matches browser.Browser functionality.
type Browser interface {
	Browse(url string) error
}

// openGitHubProfile opens the GitHub profile page for the specified user or authenticated user.
func openGitHubProfile(targetUser string, client skyline.GitHubClientInterface, b Browser) error {
	if targetUser == "" {
		username, err := client.GetAuthenticatedUser()
		if err != nil {
			return errors.New(errors.NetworkError, "failed to get authenticated user", err)
		}
		targetUser = username
	}

	hostname, _ := auth.DefaultHost()
	profileURL := fmt.Sprintf("https://%s/%s", hostname, targetUser)
	return b.Browse(profileURL)
}
