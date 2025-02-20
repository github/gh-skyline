// Package utils are utility functions for the GitHub Skyline project
package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Constants for GitHub launch year and default output file format
const (
	githubLaunchYear = 2008
	outputFileFormat = "%s-%s-github-skyline.stl"
)

// ParseYearRange parses whether a year is a single year or a range of years.
func ParseYearRange(yearRange string) (startYear, endYear int, err error) {
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

// FormatYearRange returns a formatted string representation of the year range
func FormatYearRange(startYear, endYear int) string {
	if startYear == endYear {
		return fmt.Sprintf("%d", startYear)
	}
	// Use YYYY-YY format for multi-year ranges
	return fmt.Sprintf("%04d-%02d", startYear, endYear%100)
}

// GenerateOutputFilename creates a filename for the STL output based on the user and year range.
func GenerateOutputFilename(username string, startYear, endYear int, customPath string, ytdEnd string) string {
	if customPath != "" {
		return customPath
	}

	now := time.Now()
	isYTD := endYear == now.Year() && startYear == now.AddDate(0, -12, 0).Year()

	var filename string
	switch {
	case isYTD:
		// For YTD mode, use a date range format
		endDate := now
		if ytdEnd != "" {
			if parsed, err := time.Parse("2006-01-02", ytdEnd); err == nil {
				endDate = parsed
			}
		}
		filename = fmt.Sprintf("%s-contributions-ytd-%s.stl", username, endDate.Format("2006-01-02"))
	case startYear == endYear:
		filename = fmt.Sprintf("%s-contributions-%d.stl", username, startYear)
	default:
		filename = fmt.Sprintf("%s-contributions-%d-%d.stl", username, startYear, endYear)
	}

	return filename
}
