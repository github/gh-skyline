package ascii

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/github/gh-skyline/types"
)

// ErrInvalidGrid is returned when the contribution grid is invalid
var ErrInvalidGrid = errors.New("invalid contribution grid")

// GenerateASCII creates a 2D ASCII art representation of the contribution data.
// It returns the generated ASCII art as a string and an error if the operation fails.
// When includeHeader is true, the output includes the header template.
func GenerateASCII(contributionGrid [][]types.ContributionDay, username string, year int, includeHeader bool, includeUserInfo bool) (string, error) {
	if len(contributionGrid) == 0 {
		return "", ErrInvalidGrid
	}

	var buffer bytes.Buffer

	// Only include header if requested
	if includeHeader {
		for _, line := range strings.Split(HeaderTemplate, "\n") {
			buffer.WriteString(line + "\n")
		}
		buffer.WriteString("\n")
	}

	// Find max contribution count for normalization
	maxContributions := 0
	for _, week := range contributionGrid {
		for _, day := range week {
			if day.ContributionCount > maxContributions {
				maxContributions = day.ContributionCount
			}
		}
	}

	// Initialize the ASCII grid (7 rows x 53 columns)
	asciiGrid := make([][]rune, 7)
	for i := range asciiGrid {
		asciiGrid[i] = make([]rune, len(contributionGrid))
	}

	// Get current time for future date comparison
	now := time.Now()

	// Process each week
	for weekIdx, week := range contributionGrid {
		// Update to receive nonZeroCount
		sortedDays, nonZeroCount := sortContributionDays(week, now)

		// Fill the column for this week
		for dayIdx, day := range sortedDays {
			if day.ContributionCount == -1 {
				asciiGrid[dayIdx][weekIdx] = FutureBlock
			} else {
				normalized := float64(day.ContributionCount) / float64(maxContributions)
				asciiGrid[dayIdx][weekIdx] = getBlock(normalized, dayIdx, nonZeroCount)
			}
		}
	}

	// Write the contribution grid
	for i := len(asciiGrid) - 1; i >= 0; i-- {
		for _, ch := range asciiGrid[i] {
			buffer.WriteRune(ch)
		}
		buffer.WriteRune('\n')
	}

	if includeUserInfo {
		// Add centered user info below
		buffer.WriteString("\n")
		buffer.WriteString(centerText(username))
		buffer.WriteString(centerText(fmt.Sprintf("%d", year)))
	}

	return buffer.String(), nil
}

// sortContributionDays sorts the contribution days within a week.
// It places non-zero contributions first, followed by zero contributions, and future dates last.
func sortContributionDays(week []types.ContributionDay, now time.Time) ([]types.ContributionDay, int) {
	sortedDays := make([]types.ContributionDay, 7)
	nonZeroContributions := []types.ContributionDay{}
	zeroContributions := []types.ContributionDay{}
	futureDates := []types.ContributionDay{}

	// Separate contributions
	for _, day := range week {
		switch {
		case day.IsAfter(now):
			futureDates = append(futureDates, types.ContributionDay{
				ContributionCount: -1,
				Date:              day.Date,
			})
		case day.ContributionCount > 0:
			nonZeroContributions = append(nonZeroContributions, day)
		default:
			zeroContributions = append(zeroContributions, day)
		}
	}

	// Build sortedDays from bottom to top
	idx := 0
	for _, day := range nonZeroContributions {
		sortedDays[idx] = day
		idx++
	}
	for _, day := range zeroContributions {
		sortedDays[idx] = day
		idx++
	}
	for _, day := range futureDates {
		sortedDays[idx] = day
		idx++
	}

	return sortedDays, len(nonZeroContributions)
}

// getBlockType determines the contribution level category based on the normalized value
func getBlockType(normalized float64) int {
	switch {
	case normalized < LowThreshold:
		return 0 // Low
	case normalized < MediumThreshold:
		return 1 // Medium
	default:
		return 2 // High
	}
}

// blockSets defines the block characters for different positions and intensity levels
var blockSets = map[string][3]rune{
	"foundation": {FoundationLow, FoundationMed, FoundationHigh},
	"middle":     {MiddleLow, MiddleMed, MiddleHigh},
	"top":        {TopLow, TopMed, TopHigh},
}

// getBlock determines the appropriate block character based on position and contribution level
func getBlock(normalized float64, dayIdx, nonZeroIdx int) rune {
	if normalized == 0 {
		return EmptyBlock
	}

	blockType := getBlockType(normalized)

	// Single block column uses foundation style
	if nonZeroIdx == 1 {
		return blockSets["foundation"][blockType]
	}

	switch {
	case dayIdx == nonZeroIdx-1: // Top block
		return blockSets["top"][blockType]
	case dayIdx == 0: // Bottom block
		return blockSets["foundation"][blockType]
	default: // Middle blocks
		return blockSets["middle"][blockType]
	}
}
