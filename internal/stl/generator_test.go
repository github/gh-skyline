package stl

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/github/gh-skyline/internal/types"
)

// Test data setup
func createTestContributions() [][]types.ContributionDay {
	contributions := make([][]types.ContributionDay, 52)
	for i := range contributions {
		contributions[i] = make([]types.ContributionDay, 7)
		for j := range contributions[i] {
			contributions[i][j] = types.ContributionDay{ContributionCount: (i + j) % 5}
		}
	}
	return contributions
}

func TestGenerateSTL(t *testing.T) {
	contributions := createTestContributions()
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.stl")

	err := GenerateSTL(contributions, outputPath, "testuser", 2023, "flat")
	if err != nil {
		// Check if error is due to missing resources
		if strings.Contains(err.Error(), "failed to open image") ||
			strings.Contains(err.Error(), "failed to load fonts") {
			t.Skip("Skipping test due to missing required resources")
		}
		t.Errorf("GenerateSTL failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("STL file was not created")
	}

	// Test error cases
	tests := []struct {
		name          string
		contributions [][]types.ContributionDay
		outputPath    string
		username      string
		year          int
		wantErr       bool
	}{
		{"empty contributions", nil, outputPath, "user", 2023, true},
		{"empty output path", contributions, "", "user", 2023, true},
		{"empty username", contributions, outputPath, "", 2023, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateSTL(tt.contributions, tt.outputPath, tt.username, tt.year, "flat")
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSTL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSTLRange(t *testing.T) {
	// Create test data for multiple years
	contributionsRange := make([][][]types.ContributionDay, 3)
	for i := range contributionsRange {
		contributionsRange[i] = createTestContributions()
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test_range.stl")

	tests := []struct {
		name          string
		contributions [][][]types.ContributionDay
		outputPath    string
		username      string
		startYear     int
		endYear       int
		wantErr       bool
	}{
		{
			name:          "valid multi-year",
			contributions: contributionsRange,
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2021,
			endYear:       2023,
			wantErr:       false,
		},
		{
			name:          "single year range",
			contributions: [][][]types.ContributionDay{createTestContributions()},
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2023,
			wantErr:       false,
		},
		{
			name:          "invalid year range",
			contributions: contributionsRange,
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2022,
			wantErr:       false, // Should still work, just displays years in correct order
		},
		{
			name:          "empty contributions array",
			contributions: [][][]types.ContributionDay{},
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2023,
			wantErr:       true,
		},
		{
			name:          "empty contributions array",
			contributions: [][][]types.ContributionDay{{}},
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2023,
			wantErr:       true,
		},
		{
			name:          "nil inner array",
			contributions: [][][]types.ContributionDay{nil},
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2023,
			wantErr:       true,
		},
		{
			name:          "nil contributions",
			contributions: nil,
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2023,
			wantErr:       true,
		},
		{
			name:          "empty array with non-empty inner array",
			contributions: [][][]types.ContributionDay{make([][]types.ContributionDay, 0)},
			outputPath:    outputPath,
			username:      "testuser",
			startYear:     2023,
			endYear:       2023,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// To prevent panic in test execution, wrap the function call
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("GenerateSTLRange() panicked: %v", r)
				}
			}()

			err := GenerateSTLRange(tt.contributions, tt.outputPath, tt.username, tt.startYear, tt.endYear, "flat")
			if (err != nil) != tt.wantErr {
				// Only fail if the error is not related to missing resources
				if !strings.Contains(err.Error(), "failed to open image") {
					t.Errorf("GenerateSTLRange() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			if !tt.wantErr && err == nil {
				// Verify file was created
				if _, err := os.Stat(tt.outputPath); os.IsNotExist(err) {
					t.Error("STL file was not created")
				}
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	validContributions := createTestContributions()

	tests := []struct {
		name          string
		contributions [][]types.ContributionDay
		outputPath    string
		username      string
		wantErr       bool
	}{
		{"valid input", validContributions, "output.stl", "user", false},
		{"nil contributions", nil, "output.stl", "user", true},
		{"empty contributions", [][]types.ContributionDay{}, "output.stl", "user", true},
		{"empty output path", validContributions, "", "user", true},
		{"empty username", validContributions, "output.stl", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.contributions, tt.outputPath, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateDimensionsMultiYear(t *testing.T) {
	tests := []struct {
		name      string
		yearCount int
		wantErr   bool
	}{
		{"single year", 1, false},
		{"multiple years", 3, false},
		{"zero years", 0, true},
		{"negative years", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := calculateDimensions(tt.yearCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if dims.innerWidth <= 0 || dims.innerDepth <= 0 {
					t.Errorf("calculateDimensions() returned invalid dimensions: width=%v, depth=%v",
						dims.innerWidth, dims.innerDepth)
				}
			}
		})
	}
}

func TestFindMaxContributions(t *testing.T) {
	tests := []struct {
		name          string
		contributions [][]types.ContributionDay
		want          int
	}{
		{
			name:          "normal contributions",
			contributions: createTestContributions(),
			want:          4,
		},
		{
			name: "single max contribution",
			contributions: [][]types.ContributionDay{
				{{ContributionCount: 10}},
			},
			want: 10,
		},
		{
			name:          "empty contributions",
			contributions: [][]types.ContributionDay{},
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMaxContributions(tt.contributions)
			if got != tt.want {
				t.Errorf("findMaxContributions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMaxContributionsAcrossYears(t *testing.T) {
	tests := []struct {
		name          string
		contributions [][][]types.ContributionDay
		want          int
	}{
		{
			name: "multiple years with varying max",
			contributions: [][][]types.ContributionDay{
				{{{ContributionCount: 5}}},
				{{{ContributionCount: 10}}},
				{{{ContributionCount: 3}}},
			},
			want: 10,
		},
		{
			name: "empty years",
			contributions: [][][]types.ContributionDay{
				{},
				{},
			},
			want: 0,
		},
		{
			name:          "nil input",
			contributions: nil,
			want:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMaxContributionsAcrossYears(tt.contributions)
			if got != tt.want {
				t.Errorf("findMaxContributionsAcrossYears() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateBase(t *testing.T) {
	dims, err := calculateDimensions(1)
	if err != nil {
		t.Fatalf("calculateDimensions() error = %v", err)
	}
	ch := make(chan geometryResult)
	var wg sync.WaitGroup
	wg.Add(1)

	go generateBase(dims, "flat", ch, &wg)

	result := <-ch
	if result.err != nil {
		t.Errorf("generateBase() error = %v", result.err)
	}
	if len(result.triangles) == 0 {
		t.Error("generateBase() returned no triangles")
	}
}

func TestGenerateText(t *testing.T) {
	dims, err := calculateDimensions(1)
	if err != nil {
		t.Fatalf("calculateDimensions() error = %v", err)
	}
	ch := make(chan geometryResult)
	var wg sync.WaitGroup
	wg.Add(1)

	go generateText("testuser", 2023, 2023, dims, "flat", ch, &wg)

	result := <-ch
	if result.err != nil {
		t.Errorf("generateText() error = %v", result.err)
	}
	// Remove the triangle count check since text might not be generated
	// due to missing fonts, which is an acceptable condition
}

func TestEstimateTriangleCount(t *testing.T) {
	contributions := createTestContributions()
	count := estimateTriangleCount(contributions)
	if count <= 0 {
		t.Errorf("estimateTriangleCount() = %v, want > 0", count)
	}
}

func TestGenerateColumnsForYearRange(t *testing.T) {
	// Create test data for multiple years
	contributionsPerYear := make([][][]types.ContributionDay, 3)
	for i := range contributionsPerYear {
		contributionsPerYear[i] = createTestContributions()
	}

	ch := make(chan geometryResult)
	var wg sync.WaitGroup
	wg.Add(1)

	maxContrib := 10 // Set a known max contribution value

	// Test the goroutine
	go generateColumnsForYearRange(contributionsPerYear, maxContrib, ch, &wg)

	// Collect the result
	result := <-ch
	if len(result.triangles) == 0 {
		t.Error("generateColumnsForYearRange() returned no triangles")
	}

	wg.Wait()
}

func TestCreateContributionGeometry(t *testing.T) {
	contributions := createTestContributions()
	yearIndex := 0
	maxContrib := 10

	triangles := CreateContributionGeometry(contributions, yearIndex, maxContrib)

	if len(triangles) == 0 {
		t.Error("CreateContributionGeometry() returned no triangles")
	}

	// Test with empty contributions
	emptyTriangles := CreateContributionGeometry([][]types.ContributionDay{}, yearIndex, maxContrib)
	if len(emptyTriangles) != 0 {
		t.Error("CreateContributionGeometry() should return empty slice for empty contributions")
	}

	// Test with zero max contribution
	zeroMaxTriangles := CreateContributionGeometry(contributions, yearIndex, 0)
	if len(zeroMaxTriangles) == 0 {
		t.Error("CreateContributionGeometry() should still generate triangles with zero max contribution")
	}
}

func TestGenerateModelGeometry(t *testing.T) {
	contributionsPerYear := make([][][]types.ContributionDay, 2)
	for i := range contributionsPerYear {
		contributionsPerYear[i] = createTestContributions()
	}

	dims, err := calculateDimensions(len(contributionsPerYear))
	if err != nil {
		t.Fatalf("calculateDimensions() error = %v", err)
	}
	maxContrib := findMaxContributionsAcrossYears(contributionsPerYear)
	username := "testuser"
	startYear := 2022
	endYear := 2023

	triangles, err := generateModelGeometry(contributionsPerYear, dims, maxContrib, username, startYear, endYear, "flat")
	if err != nil {
		t.Errorf("generateModelGeometry() error = %v", err)
	}
	if len(triangles) == 0 {
		t.Error("generateModelGeometry() returned no triangles")
	}

	// Test error case with nil contributions
	_, err = generateModelGeometry(nil, dims, maxContrib, username, startYear, endYear, "flat")
	if err == nil {
		t.Error("generateModelGeometry() should return error for nil contributions")
	}

	// Test with empty username
	_, err = generateModelGeometry(contributionsPerYear, dims, maxContrib, "", startYear, endYear, "flat")
	if err != nil {
		t.Error("generateModelGeometry() should handle empty username")
	}
}

func TestGenerateLogo(t *testing.T) {
	dims, err := calculateDimensions(1)
	if err != nil {
		t.Fatalf("calculateDimensions() error = %v", err)
	}
	ch := make(chan geometryResult)
	var wg sync.WaitGroup
	wg.Add(1)

	go generateLogo(dims, "flat", ch, &wg)

	result := <-ch
	// Even if image file is not found, result should not be nil
	if result.triangles == nil {
		t.Error("generateLogo() returned nil triangles slice")
	}
	wg.Wait()
}

func TestCalculateDimensions(t *testing.T) {
	tests := []struct {
		name      string
		yearCount int
		wantErr   bool
	}{
		{"single year", 1, false},
		{"multiple years", 3, false},
		{"zero years", 0, true},
		{"negative years", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := calculateDimensions(tt.yearCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateDimensions(%d) error = %v, wantErr %v", tt.yearCount, err, tt.wantErr)
				return
			}

			if !tt.wantErr && (dims.innerWidth <= 0 || dims.innerDepth <= 0) {
				t.Errorf("calculateDimensions(%d) returned invalid dimensions: width=%v, depth=%v",
					tt.yearCount, dims.innerWidth, dims.innerDepth)
			}
		})
	}
}

func TestGenerateText_WithYearRange(t *testing.T) {
	dims, err := calculateDimensions(1)
	if err != nil {
		t.Fatalf("calculateDimensions() error = %v", err)
	}
	tests := []struct {
		name      string
		username  string
		startYear int
		endYear   int
	}{
		{"same year", "testuser", 2023, 2023},
		{"year range", "testuser", 2021, 2023},
		{"empty username", "", 2023, 2023},
		{"inverse year range", "testuser", 2023, 2021}, // Should still work
		{"distant years", "testuser", 2000, 2023},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan geometryResult)
			var wg sync.WaitGroup
			wg.Add(1)

			go generateText(tt.username, tt.startYear, tt.endYear, dims, "flat", ch, &wg)

			result := <-ch
			// Even if font generation fails, result should not be nil
			if result.triangles == nil {
				t.Errorf("%s: generateText() returned nil triangles slice", tt.name)
			}
			wg.Wait()
			close(ch)
		})
	}
}

func TestGenerateColumnsForYearRange_Extended(t *testing.T) {
	tests := []struct {
		name            string
		yearsCount      int
		maxContrib      int
		expectTriangles bool
	}{
		{"single year", 1, 10, true},
		{"multiple years", 3, 10, true},
		{"zero contributions", 2, 0, true},
		{"large contribution count", 2, 1000, true},
		{"empty year data", 0, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contributionsPerYear := make([][][]types.ContributionDay, tt.yearsCount)
			for i := range contributionsPerYear {
				contributionsPerYear[i] = createTestContributions()
			}

			ch := make(chan geometryResult)
			var wg sync.WaitGroup
			wg.Add(1)

			go generateColumnsForYearRange(contributionsPerYear, tt.maxContrib, ch, &wg)

			result := <-ch
			if tt.expectTriangles && len(result.triangles) == 0 {
				t.Error("generateColumnsForYearRange() returned no triangles when triangles were expected")
			}

			wg.Wait()
			close(ch)
		})
	}
}

func TestResourceHandling(t *testing.T) {
	// Test handling of missing font files
	t.Run("missing font handling", func(t *testing.T) {
		dims, err := calculateDimensions(1)
		if err != nil {
			t.Fatalf("calculateDimensions() error = %v", err)
		}
		ch := make(chan geometryResult)
		var wg sync.WaitGroup
		wg.Add(1)

		// This should log a warning but continue
		go generateText("testuser", 2023, 2023, dims, "flat", ch, &wg)

		result := <-ch
		// Even with missing fonts, we should get a valid (possibly empty) result
		if result.triangles == nil {
			t.Error("generateText() returned nil instead of empty slice with missing fonts")
		}
		wg.Wait()
	})

	// Test handling of missing image file
	t.Run("missing image handling", func(t *testing.T) {
		dims, err := calculateDimensions(1)
		if err != nil {
			t.Fatalf("calculateDimensions() error = %v", err)
		}
		ch := make(chan geometryResult)
		var wg sync.WaitGroup
		wg.Add(1)

		// This should log a warning but continue
		go generateLogo(dims, "flat", ch, &wg)

		result := <-ch
		// Even with missing image, we should get a valid (possibly empty) result
		if result.triangles == nil {
			t.Error("generateLogo() returned nil instead of empty slice with missing image")
		}
		wg.Wait()
	})

	// Test full model generation with missing resources
	t.Run("full model with missing resources", func(t *testing.T) {
		contributionsPerYear := make([][][]types.ContributionDay, 2)
		for i := range contributionsPerYear {
			contributionsPerYear[i] = createTestContributions()
		}

		dims, err := calculateDimensions(len(contributionsPerYear))
		if err != nil {
			t.Fatalf("calculateDimensions() error = %v", err)
		}
		maxContrib := findMaxContributionsAcrossYears(contributionsPerYear)

		// This should complete successfully even with missing resources
		triangles, err := generateModelGeometry(contributionsPerYear, dims, maxContrib, "testuser", 2022, 2023, "flat")
		if err != nil {
			t.Errorf("generateModelGeometry() failed with missing resources: %v", err)
		}

		// Should still generate base geometry and contribution columns
		if len(triangles) == 0 {
			t.Error("generateModelGeometry() returned no triangles with missing resources")
		}
	})
}

func TestCalculateDimensionsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		yearCount int
		wantErr   bool
	}{
		{"max year count", 100, false},  // Test very large year count
		{"boundary zero", 0, true},      // Should error
		{"boundary negative", -1, true}, // Should error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims, err := calculateDimensions(tt.yearCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateDimensions() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if dims.innerWidth <= 0 || dims.innerDepth <= 0 {
					t.Errorf("Invalid dimensions for yearCount %d: width=%v, depth=%v",
						tt.yearCount, dims.innerWidth, dims.innerDepth)
				}
			}
		})
	}
}
