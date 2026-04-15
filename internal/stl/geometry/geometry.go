package geometry

import (
	"math"

	"github.com/github/gh-skyline/internal/types"
)

// Model dimension constants define the basic measurements for the 3D model.
const (
	BaseHeight    float64 = 10.0     // Height of the base in model units
	MaxHeight     float64 = 25.0     // Maximum height for contribution columns
	CellSize      float64 = 2.5      // Size of each contribution cell
	GridSize      int     = 53       // Number of weeks in a year
	BaseThickness float64 = 10.0     // Total thickness of the base
	MinHeight     float64 = CellSize // Minimum height for any contribution column
	BaseSlant     float64 = 3.64     // Slant offset for the base (BaseHeight * tan(20°))
)

// Text rendering constants control the appearance and positioning of text.
const (
	TextPadding  float64 = CellSize * 2   // Increased padding
	TextWidthPct float32 = 0.6            // Reduced to ensure text fits
	TextDepth    float64 = 2.0 * CellSize // More prominent depth
)

// Font file paths for text rendering.
const (
	PrimaryFont  = "monasans-medium.ttf"
	FallbackFont = "monasans-regular.ttf"
)

// Additional constants for year range styling
const (
	YearSpacing float64 = 0.0 // Remove gap between years
	YearOffset  float64 = 7.0 * CellSize
)

// ModelDimensions defines the inner dimensions of the model.
type ModelDimensions struct {
	InnerWidth float64
	InnerDepth float64
}

// NormalizeContribution converts a contribution count to a normalized height value.
// Returns 0 for no contributions, or a value between MinHeight and MaxHeight for active contributions.
func NormalizeContribution(count, maxCount int) float64 {
	if count == 0 {
		return 0 // No contribution means no column
	}
	if maxCount <= 0 {
		return MinHeight // Avoid division by zero, return minimum height
	}

	// Calculate the available height range for columns
	heightRange := MaxHeight - MinHeight

	// Use square root to create more visual variation in height
	// This creates a more pronounced difference between low and high contribution counts
	normalizedValue := math.Sqrt(float64(count)) / math.Sqrt(float64(maxCount))

	// Scale to fit between MinHeight and MaxHeight
	return MinHeight + (normalizedValue * heightRange)
}

// CreateContributionGeometry generates geometry for a single year's contributions
func CreateContributionGeometry(contributions [][]types.ContributionDay, yearIndex int, maxContrib int) ([]types.Triangle, error) {
	var triangles []types.Triangle

	// Base Y offset includes padding and positions each year accordingly
	baseYOffset := 2*CellSize + float64(yearIndex)*7*CellSize

	for weekIdx, week := range contributions {
		for dayIdx, day := range week {
			if day.ContributionCount > 0 {
				height := NormalizeContribution(day.ContributionCount, maxContrib)
				x := 2*CellSize + float64(weekIdx)*CellSize
				y := baseYOffset + float64(dayIdx)*CellSize

				columnTriangles, err := CreateColumn(x, y, height, CellSize)
				if err != nil {
					return nil, err
				}
				triangles = append(triangles, columnTriangles...)
			}
		}
	}

	return triangles, nil
}

// CalculateMultiYearDimensions calculates dimensions for multiple years
func CalculateMultiYearDimensions(yearCount int) (width, depth float64) {
	// Total width: grid size + padding on both sides
	width = float64(GridSize)*CellSize + 4*CellSize
	// Total depth: (7 days * number of years) + padding on both sides
	depth = float64(7*yearCount)*CellSize + 4*CellSize
	return width, depth
}
