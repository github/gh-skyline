package stl

import (
	"fmt"
	"sync"

	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/logger"
	"github.com/github/gh-skyline/internal/stl/geometry"
	"github.com/github/gh-skyline/internal/types"
)

// GenerateSTL creates a 3D model from GitHub contribution data and writes it to an STL file.
// It's a convenience wrapper around GenerateSTLRange for single year processing.
func GenerateSTL(contributions [][]types.ContributionDay, outputPath, username string, year int, baseType string) error {
	// Wrap single year data in the format expected by GenerateSTLRange
	contributionsRange := [][][]types.ContributionDay{contributions}
	return GenerateSTLRange(contributionsRange, outputPath, username, year, year, baseType)
}

// GenerateSTLRange creates a 3D model from multiple years of GitHub contribution data.
// It handles the complete process from data validation through geometry generation to file output.
// Parameters:
//   - contributions: 3D slice of contribution data ([year][week][day])
//   - outputPath: destination path for the STL file
//   - username: GitHub username for the contribution data
//   - startYear: first year in the range
//   - endYear: last year in the range
//   - baseType: type of the model base (flat or slanted)
func GenerateSTLRange(contributions [][][]types.ContributionDay, outputPath, username string, startYear, endYear int, baseType string) error {
	log := logger.GetLogger()
	if err := log.Debug("Starting STL generation for user %s, years %d-%d", username, startYear, endYear); err != nil {
		return errors.Wrap(err, "failed to log debug message")
	}

	if err := validateInput(contributions[0], outputPath, username); err != nil {
		return errors.Wrap(err, "input validation failed")
	}

	dimensions, err := calculateDimensions(len(contributions))
	if err != nil {
		return errors.Wrap(err, "failed to calculate dimensions")
	}

	// Find global max contribution across all years
	maxContribution := findMaxContributionsAcrossYears(contributions)

	modelTriangles, err := generateModelGeometry(contributions, dimensions, maxContribution, username, startYear, endYear, baseType)
	if err != nil {
		return errors.Wrap(err, "failed to generate geometry")
	}

	if err := log.Info("Model generation complete: %d total triangles", len(modelTriangles)); err != nil {
		return errors.Wrap(err, "failed to log info message")
	}
	if err := log.Debug("Writing STL file to: %s", outputPath); err != nil {
		return errors.Wrap(err, "failed to log debug message")
	}

	if err := WriteSTLBinary(outputPath, modelTriangles); err != nil {
		return errors.Wrap(err, "failed to write STL file")
	}

	if err := log.Info("STL file written successfully to: %s", outputPath); err != nil {
		return errors.Wrap(err, "failed to log info message")
	}
	return nil
}

// modelDimensions represents the core measurements of the 3D model.
// All measurements are in millimeters.
type modelDimensions struct {
	innerWidth float64 // Width of the contribution grid
	innerDepth float64 // Depth of the contribution grid
	imagePath  string  // Path to the logo image
}

func validateInput(contributions [][]types.ContributionDay, outputPath, username string) error {
	if len(contributions) == 0 {
		return errors.New(errors.ValidationError, "contributions data cannot be empty", nil)
	}
	if len(contributions) > geometry.GridSize {
		return errors.New(errors.ValidationError, "contributions data exceeds maximum grid size", nil)
	}
	if outputPath == "" {
		return errors.New(errors.ValidationError, "output path cannot be empty", nil)
	}
	if username == "" {
		return errors.New(errors.ValidationError, "username cannot be empty", nil)
	}
	return nil
}

func calculateDimensions(yearCount int) (modelDimensions, error) {
	if yearCount <= 0 {
		return modelDimensions{}, errors.New(errors.ValidationError, "year count must be positive", nil)
	}

	var width, depth float64
	width, depth = geometry.CalculateMultiYearDimensions(yearCount)

	dims := modelDimensions{
		innerWidth: width,
		innerDepth: depth,
		imagePath:  "assets/invertocat.png",
	}

	if dims.innerWidth <= 0 || dims.innerDepth <= 0 {
		return modelDimensions{}, errors.New(errors.ValidationError, "invalid model dimensions", nil)
	}

	return dims, nil
}

func findMaxContributions(contributions [][]types.ContributionDay) int {
	maxContrib := 0
	for _, week := range contributions {
		for _, day := range week {
			if day.ContributionCount > maxContrib {
				maxContrib = day.ContributionCount
			}
		}
	}
	return maxContrib
}

// findMaxContributionsAcrossYears finds the maximum contribution count across all years
func findMaxContributionsAcrossYears(contributionsPerYear [][][]types.ContributionDay) int {
	maxContrib := 0
	for _, yearContributions := range contributionsPerYear {
		yearMax := findMaxContributions(yearContributions)
		if yearMax > maxContrib {
			maxContrib = yearMax
		}
	}
	return maxContrib
}

// geometryResult holds the output of geometry generation operations.
// It includes both the generated triangles and any errors that occurred.
type geometryResult struct {
	triangles []types.Triangle
	err       error
}

// generateModelGeometry orchestrates the concurrent generation of all model components.
// It manages four parallel processes for generating the base, columns, text, and logo.
func generateModelGeometry(contributionsPerYear [][][]types.ContributionDay, dims modelDimensions, maxContrib int, username string, startYear, endYear int, baseType string) ([]types.Triangle, error) {
	if len(contributionsPerYear) == 0 {
		return nil, errors.New(errors.ValidationError, "contributions data cannot be empty", nil)
	}

	// Create channels for each geometry component
	channels := map[string]chan geometryResult{
		"base":    make(chan geometryResult),
		"columns": make(chan geometryResult),
		"text":    make(chan geometryResult),
		"image":   make(chan geometryResult),
	}

	var wg sync.WaitGroup
	wg.Add(len(channels))

	// Launch goroutines for each component
	go generateBase(dims, baseType, channels["base"], &wg)
	go generateColumnsForYearRange(contributionsPerYear, maxContrib, channels["columns"], &wg)
	go generateText(username, startYear, endYear, dims, baseType, channels["text"], &wg)
	go generateLogo(dims, baseType, channels["image"], &wg)

	// Collect results from all channels
	modelTriangles := make([]types.Triangle, 0, estimateTriangleCount(contributionsPerYear[0])*len(contributionsPerYear))
	for componentName := range channels {
		result := <-channels[componentName]
		if result.err != nil {
			return nil, errors.Wrap(result.err, fmt.Sprintf("failed to generate %s geometry", componentName))
		}
		modelTriangles = append(modelTriangles, result.triangles...)
	}

	// Clean up
	wg.Wait()
	for _, ch := range channels {
		close(ch)
	}

	return modelTriangles, nil
}

func generateBase(dims modelDimensions, baseType string, ch chan<- geometryResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	var baseTriangles []types.Triangle
	var err error
	
	if baseType == "slanted" {
		baseTriangles, err = geometry.CreateSlantedBase(dims.innerWidth, dims.innerDepth)
	} else {
		baseTriangles, err = geometry.CreateCuboidBase(dims.innerWidth, dims.innerDepth)
	}

	if err != nil {
		if logErr := logger.GetLogger().Warning("Failed to generate base geometry: %v. Continuing without base.", err); logErr != nil {
			ch <- geometryResult{triangles: []types.Triangle{}, err: logErr}
			return
		}
		ch <- geometryResult{triangles: []types.Triangle{}}
		return
	}

	ch <- geometryResult{triangles: baseTriangles}
}

// generateText creates 3D text geometry for the model
func generateText(username string, startYear int, endYear int, dims modelDimensions, baseType string, ch chan<- geometryResult, wg *sync.WaitGroup) {
	defer wg.Done()
	embossedYear := fmt.Sprintf("%d", endYear)

	// If start year and end year are the same, only show one year
	if startYear != endYear {
		// Make the year 'YYYY-YY'
		embossedYear = fmt.Sprintf("%04d-%02d", startYear, endYear%100)
	}

	textTriangles, err := geometry.Create3DText(username, embossedYear, dims.innerWidth, geometry.BaseHeight, baseType)
	if err != nil {
		if logErr := logger.GetLogger().Warning("Failed to generate text geometry: %v. Continuing without text.", err); logErr != nil {
			ch <- geometryResult{triangles: []types.Triangle{}, err: logErr}
			return
		}
		ch <- geometryResult{triangles: []types.Triangle{}}
		return
	}
	ch <- geometryResult{triangles: textTriangles}
}

// generateLogo handles the generation of the GitHub logo geometry
func generateLogo(dims modelDimensions, baseType string, ch chan<- geometryResult, wg *sync.WaitGroup) {
	defer wg.Done()
	logoTriangles, err := geometry.GenerateImageGeometry(dims.innerWidth, geometry.BaseHeight, baseType)
	if err != nil {
		// Log warning and continue without logo instead of failing
		if logErr := logger.GetLogger().Warning("Failed to generate logo geometry: %v. Continuing without logo.", err); logErr != nil {
			ch <- geometryResult{triangles: []types.Triangle{}, err: logErr}
			return
		}
		ch <- geometryResult{triangles: []types.Triangle{}}
		return
	}
	ch <- geometryResult{triangles: logoTriangles}
}

func estimateTriangleCount(contributions [][]types.ContributionDay) int {
	totalContributions := 0
	for _, week := range contributions {
		for _, day := range week {
			if day.ContributionCount > 0 {
				totalContributions++
			}
		}
	}

	baseTrianglesCount := 12
	columnsTrianglesCount := totalContributions * 12
	textTrianglesEstimate := 1000
	return baseTrianglesCount + columnsTrianglesCount + textTrianglesEstimate
}

// generateColumnsForYearRange generates contribution columns for multiple years
func generateColumnsForYearRange(contributionsPerYear [][][]types.ContributionDay, maxContrib int, ch chan<- geometryResult, wg *sync.WaitGroup) {
	defer wg.Done()
	var yearTriangles []types.Triangle

	// Process years in reverse order so most recent year is at the front
	for i := len(contributionsPerYear) - 1; i >= 0; i-- {
		yearOffset := len(contributionsPerYear) - 1 - i
		triangles, err := geometry.CreateContributionGeometry(contributionsPerYear[i], yearOffset, maxContrib)
		if err != nil {
			if logErr := logger.GetLogger().Warning("Failed to generate column geometry for year %d: %v. Skipping year.", i, err); logErr != nil {
				return
			}
			continue
		}
		yearTriangles = append(yearTriangles, triangles...)
	}

	ch <- geometryResult{triangles: yearTriangles}
}

// CreateContributionGeometry generates geometry for a single year's worth of contributions
func CreateContributionGeometry(contributions [][]types.ContributionDay, yearIndex int, maxContrib int) []types.Triangle {
	var triangles []types.Triangle

	// Calculate the Y offset for this year's grid
	// Each subsequent year is placed further back (larger Y value)
	baseYOffset := float64(yearIndex) * (geometry.YearOffset + geometry.YearSpacing)

	// Generate contribution columns
	for weekIdx, week := range contributions {
		for dayIdx, day := range week {
			if day.ContributionCount > 0 {
				height := geometry.NormalizeContribution(day.ContributionCount, maxContrib)
				x := float64(weekIdx) * geometry.CellSize
				y := baseYOffset + float64(dayIdx)*geometry.CellSize

				columnTriangles, err := geometry.CreateColumn(x, y, height, geometry.CellSize)
				if err != nil {
					if logErr := logger.GetLogger().Warning("Failed to generate column geometry: %v. Skipping column.", err); logErr != nil {
						return nil
					}
					continue
				}
				triangles = append(triangles, columnTriangles...)
			}
		}
	}

	return triangles
}
