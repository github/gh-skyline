package stl

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/logger"
	"github.com/github/gh-skyline/internal/stl/geometry"
	"github.com/github/gh-skyline/internal/types"
)

// GenerateSTL creates a 3D model from GitHub contribution data and writes it to an STL file.
// It's a convenience wrapper around GenerateSTLRange for single year processing.
func GenerateSTL(contributions [][]types.ContributionDay, outputPath, username string, year int) error {
	// Wrap single year data in the format expected by GenerateSTLRange
	contributionsRange := [][][]types.ContributionDay{contributions}
	return GenerateSTLRange(contributionsRange, outputPath, username, year, year)
}

// GenerateSTLRange creates a 3D model from multiple years of GitHub contribution data.
// It handles the complete process from data validation through geometry generation to file output.
// Parameters:
//   - contributions: 3D slice of contribution data ([year][week][day])
//   - outputPath: destination path for the STL file
//   - username: GitHub username for the contribution data
//   - startYear: first year in the range
//   - endYear: last year in the range
func GenerateSTLRange(contributions [][][]types.ContributionDay, outputPath, username string, startYear, endYear int) error {
	log := logger.GetLogger()
	if err := log.Debug("Starting STL generation for user %s, years %d-%d", username, startYear, endYear); err != nil {
		return errors.Wrap(err, "failed to log debug message")
	}

	dimensions, maxContribution, err := prepareModel(contributions, outputPath, username)
	if err != nil {
		return err
	}

	modelTriangles, err := generateModelGeometry(contributions, dimensions, maxContribution, username, startYear, endYear)
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

// GenerateMultiPartSTL writes the model as three STL files that share one coordinate system,
// so a slicer can load them as parts of one object and give each part its own filament:
//   - {name}-base.stl: the base
//   - {name}-text.stl: the logo, username and year(s) on the front of the base
//   - {name}-skyline.stl: the contribution columns
//
// {name} is outputPath without its extension. It returns the paths it wrote, in that order.
func GenerateMultiPartSTL(contributions [][][]types.ContributionDay, outputPath, username string, startYear, endYear int) ([]string, error) {
	dimensions, maxContribution, err := prepareModel(contributions, outputPath, username)
	if err != nil {
		return nil, err
	}

	parts, err := generateModelParts(contributions, dimensions, maxContribution, username, startYear, endYear)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate geometry")
	}

	stem := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	files := []struct {
		path      string
		triangles []types.Triangle
	}{
		{stem + "-base.stl", parts.base},
		{stem + "-text.stl", append(append([]types.Triangle{}, parts.text...), parts.logo...)},
		{stem + "-skyline.stl", parts.columns},
	}

	paths := make([]string, 0, len(files))
	for _, f := range files {
		if err := WriteSTLBinary(f.path, f.triangles); err != nil {
			return nil, errors.Wrap(err, "failed to write STL file")
		}
		if err := logger.GetLogger().Info("STL part written successfully to: %s", f.path); err != nil {
			return nil, errors.Wrap(err, "failed to log info message")
		}
		paths = append(paths, f.path)
	}

	return paths, nil
}

// prepareModel validates the input and returns the model dimensions and the highest
// daily contribution count, which every column height is scaled against.
func prepareModel(contributions [][][]types.ContributionDay, outputPath, username string) (modelDimensions, int, error) {
	if len(contributions) == 0 {
		return modelDimensions{}, 0, errors.New(errors.ValidationError, "contributions data cannot be empty", nil)
	}

	if err := validateInput(contributions[0], outputPath, username); err != nil {
		return modelDimensions{}, 0, errors.Wrap(err, "input validation failed")
	}

	// Apply the same size bounds to every remaining year.
	// outputPath and username are shared across all years and have already been validated above.
	for i := 1; i < len(contributions); i++ {
		if len(contributions[i]) == 0 {
			return modelDimensions{}, 0, errors.New(errors.ValidationError, fmt.Sprintf("contributions data for year index %d cannot be empty", i), nil)
		}
		if len(contributions[i]) > geometry.GridSize {
			return modelDimensions{}, 0, errors.New(errors.ValidationError, fmt.Sprintf("contributions data for year index %d exceeds maximum grid size", i), nil)
		}
	}

	dimensions, err := calculateDimensions(len(contributions))
	if err != nil {
		return modelDimensions{}, 0, errors.Wrap(err, "failed to calculate dimensions")
	}

	return dimensions, findMaxContributionsAcrossYears(contributions), nil
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

// modelParts holds the triangles of each model component.
type modelParts struct {
	base    []types.Triangle
	columns []types.Triangle
	text    []types.Triangle
	logo    []types.Triangle
}

// generateModelGeometry returns the triangles of all model components in one slice,
// appended base → columns → text → image for a reproducible STL output across runs.
func generateModelGeometry(contributionsPerYear [][][]types.ContributionDay, dims modelDimensions, maxContrib int, username string, startYear, endYear int) ([]types.Triangle, error) {
	parts, err := generateModelParts(contributionsPerYear, dims, maxContrib, username, startYear, endYear)
	if err != nil {
		return nil, err
	}

	modelTriangles := make([]types.Triangle, 0, len(parts.base)+len(parts.columns)+len(parts.text)+len(parts.logo))
	modelTriangles = append(modelTriangles, parts.base...)
	modelTriangles = append(modelTriangles, parts.columns...)
	modelTriangles = append(modelTriangles, parts.text...)
	modelTriangles = append(modelTriangles, parts.logo...)

	return modelTriangles, nil
}

// generateModelParts orchestrates the concurrent generation of all model components.
// It manages four parallel processes for generating the base, columns, text, and logo.
// Channels are buffered so every goroutine can send and exit even if an error causes
// an early return, preventing goroutine leaks.
func generateModelParts(contributionsPerYear [][][]types.ContributionDay, dims modelDimensions, maxContrib int, username string, startYear, endYear int) (modelParts, error) {
	if len(contributionsPerYear) == 0 {
		return modelParts{}, errors.New(errors.ValidationError, "contributions data cannot be empty", nil)
	}

	baseCh := make(chan geometryResult, 1)
	columnsCh := make(chan geometryResult, 1)
	textCh := make(chan geometryResult, 1)
	logoCh := make(chan geometryResult, 1)

	// Launch goroutines for each component
	go generateBase(dims, baseCh)
	go generateColumnsForYearRange(contributionsPerYear, maxContrib, columnsCh)
	go generateText(username, startYear, endYear, dims, textCh)
	go generateLogo(dims, logoCh)

	// Collect results in a fixed order so errors are reported consistently.
	var parts modelParts
	components := []struct {
		name string
		ch   chan geometryResult
		dst  *[]types.Triangle
	}{
		{"base", baseCh, &parts.base},
		{"columns", columnsCh, &parts.columns},
		{"text", textCh, &parts.text},
		{"image", logoCh, &parts.logo},
	}
	for _, component := range components {
		result := <-component.ch
		if result.err != nil {
			return modelParts{}, errors.Wrap(result.err, fmt.Sprintf("failed to generate %s geometry", component.name))
		}
		*component.dst = result.triangles
	}

	return parts, nil
}

func generateBase(dims modelDimensions, ch chan<- geometryResult) {
	baseTriangles, err := geometry.CreateCuboidBase(dims.innerWidth, dims.innerDepth)

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
func generateText(username string, startYear int, endYear int, dims modelDimensions, ch chan<- geometryResult) {
	embossedYear := fmt.Sprintf("%d", endYear)

	// If start year and end year are the same, only show one year
	if startYear != endYear {
		// Make the year 'YYYY-YY'
		embossedYear = fmt.Sprintf("%04d-%02d", startYear, endYear%100)
	}

	textTriangles, err := geometry.Create3DText(username, embossedYear, dims.innerWidth, geometry.BaseHeight)
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
func generateLogo(dims modelDimensions, ch chan<- geometryResult) {
	logoTriangles, err := geometry.GenerateImageGeometry(dims.innerWidth, geometry.BaseHeight)
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
func generateColumnsForYearRange(contributionsPerYear [][][]types.ContributionDay, maxContrib int, ch chan<- geometryResult) {
	var yearTriangles []types.Triangle

	// Process years in reverse order so most recent year is at the front
	for i := len(contributionsPerYear) - 1; i >= 0; i-- {
		yearOffset := len(contributionsPerYear) - 1 - i
		triangles, err := geometry.CreateContributionGeometry(contributionsPerYear[i], yearOffset, maxContrib)
		if err != nil {
			if logErr := logger.GetLogger().Warning("Failed to generate column geometry for year %d: %v. Skipping year.", i, err); logErr != nil {
				// logErr is secondary; report the original geometry error to the caller.
				ch <- geometryResult{triangles: []types.Triangle{}, err: err}
				return
			}
			continue
		}
		yearTriangles = append(yearTriangles, triangles...)
	}

	ch <- geometryResult{triangles: yearTriangles}
}
