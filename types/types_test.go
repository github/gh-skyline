package types

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

// TestContributionDaySerialization verifies that ContributionDay structs can be correctly
// deserialized from JSON responses.
func TestContributionDaySerialization(t *testing.T) {
	testCases := []struct {
		name     string
		jsonData string
		expected ContributionDay
	}{
		{
			name:     "should parse regular contribution day",
			jsonData: `{"contributionCount": 5, "date": "2024-03-21"}`,
			expected: ContributionDay{ContributionCount: 5, Date: "2024-03-21"},
		},
		{
			name:     "should parse day with no contributions",
			jsonData: `{"contributionCount": 0, "date": "2024-03-22"}`,
			expected: ContributionDay{ContributionCount: 0, Date: "2024-03-22"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var actual ContributionDay
			if err := json.Unmarshal([]byte(tc.jsonData), &actual); err != nil {
				t.Errorf("failed to unmarshal ContributionDay: %v", err)
			}
			if actual != tc.expected {
				t.Errorf("got %+v, want %+v", actual, tc.expected)
			}
		})
	}
}

// TestContributionsResponseParsing ensures the complete GitHub API response structure
// is properly parsed with nested fields.
func TestContributionsResponseParsing(t *testing.T) {
	sampleResponse := `{
		"user": {
			"login": "testuser",
			"contributionsCollection": {
				"contributionCalendar": {
					"totalContributions": 100,
					"weeks": [
						{
							"contributionDays": [
								{
									"contributionCount": 5,
									"date": "2024-03-21"
								}
							]
						}
					]
				}
			}
		}
	}`

	var parsedResponse ContributionsResponse
	if err := json.Unmarshal([]byte(sampleResponse), &parsedResponse); err != nil {
		t.Fatalf("failed to unmarshal ContributionsResponse: %v", err)
	}

	// Verify key fields
	expectedUsername := "testuser"
	expectedTotalContributions := 100

	if parsedResponse.User.Login != expectedUsername {
		t.Errorf("username mismatch: got %q, want %q", parsedResponse.User.Login, expectedUsername)
	}
	if parsedResponse.User.ContributionsCollection.ContributionCalendar.TotalContributions != expectedTotalContributions {
		t.Errorf("total contributions mismatch: got %d, want %d",
			parsedResponse.User.ContributionsCollection.ContributionCalendar.TotalContributions,
			expectedTotalContributions)
	}
}

// TestPoint3D validates the basic structure and comparison of 3D points used
// for the contribution graph visualization.
func TestPoint3D(t *testing.T) {
	testCases := []struct {
		name     string
		point    Point3D
		expected Point3D
	}{
		{
			name:     "should handle origin point",
			point:    Point3D{0, 0, 0},
			expected: Point3D{0, 0, 0},
		},
		{
			name:     "should handle arbitrary coordinates",
			point:    Point3D{1.5, -2.5, 3.0},
			expected: Point3D{1.5, -2.5, 3.0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.point != tc.expected {
				t.Errorf("Point3D = %+v, want %+v", tc.point, tc.expected)
			}
		})
	}
}

// TestTriangle ensures triangles are correctly created with proper normal and vertex assignments
// for 3D mesh generation.
func TestTriangle(t *testing.T) {
	// Define triangle components
	expectedNormal := Point3D{0, 0, 1}
	vertex1 := Point3D{0, 0, 0}
	vertex2 := Point3D{1, 0, 0}
	vertex3 := Point3D{0, 1, 0}

	triangle := Triangle{
		Normal: expectedNormal,
		V1:     vertex1,
		V2:     vertex2,
		V3:     vertex3,
	}

	// Verify triangle properties
	if triangle.Normal != expectedNormal {
		t.Errorf("Triangle normal = %+v, want %+v", triangle.Normal, expectedNormal)
	}
	if triangle.V1 != vertex1 || triangle.V2 != vertex2 || triangle.V3 != vertex3 {
		t.Errorf("Triangle vertices do not match expected values")
	}
}

// TestContributionDayIsAfter validates the IsAfter method for contribution dates
func TestContributionDayIsAfter(t *testing.T) {
	testCases := []struct {
		name     string
		day      ContributionDay
		compare  time.Time
		expected bool
	}{
		{
			name:     "date is after comparison",
			day:      ContributionDay{Date: "2024-03-21"},
			compare:  time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "date is before comparison",
			day:      ContributionDay{Date: "2024-03-21"},
			compare:  time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "invalid date format",
			day:      ContributionDay{Date: "invalid-date"},
			compare:  time.Now(),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.day.IsAfter(tc.compare)
			if result != tc.expected {
				t.Errorf("IsAfter() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestContributionDayValidate tests the validation of ContributionDay structs
func TestContributionDayValidate(t *testing.T) {
	testCases := []struct {
		name        string
		day         ContributionDay
		expectError bool
	}{
		{
			name:        "valid contribution day",
			day:         ContributionDay{ContributionCount: 5, Date: "2024-03-21"},
			expectError: false,
		},
		{
			name:        "negative contribution count",
			day:         ContributionDay{ContributionCount: -1, Date: "2024-03-21"},
			expectError: true,
		},
		{
			name:        "invalid date format",
			day:         ContributionDay{ContributionCount: 0, Date: "invalid-date"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.day.Validate()
			if (err != nil) != tc.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tc.expectError)
			}
		})
	}
}

// TestPoint3DIsValid verifies the IsValid method for Point3D
func TestPoint3DIsValid(t *testing.T) {
	testCases := []struct {
		name     string
		point    Point3D
		expected bool
	}{
		{
			name:     "valid point",
			point:    Point3D{1.0, 2.0, 3.0},
			expected: true,
		},
		{
			name:     "point with NaN",
			point:    Point3D{math.NaN(), 2.0, 3.0},
			expected: false,
		},
		{
			name:     "point with Infinity",
			point:    Point3D{1.0, math.Inf(1), 3.0},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := tc.point.IsValid(); result != tc.expected {
				t.Errorf("IsValid() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestPoint3DToFloat32 tests the conversion from Point3D to Point3DFloat32
func TestPoint3DToFloat32(t *testing.T) {
	testCases := []struct {
		name     string
		input    Point3D
		expected Point3DFloat32
	}{
		{
			name:     "simple conversion",
			input:    Point3D{1.0, 2.0, 3.0},
			expected: Point3DFloat32{1.0, 2.0, 3.0},
		},
		{
			name:     "fractional values",
			input:    Point3D{1.5, 2.5, 3.5},
			expected: Point3DFloat32{1.5, 2.5, 3.5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.ToFloat32()
			if result != tc.expected {
				t.Errorf("ToFloat32() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestTriangleValidate tests the validation of Triangle structs
func TestTriangleValidate(t *testing.T) {
	testCases := []struct {
		name        string
		triangle    Triangle
		expectError bool
	}{
		{
			name: "valid triangle",
			triangle: Triangle{
				Normal: Point3D{0, 0, 1},
				V1:     Point3D{0, 0, 0},
				V2:     Point3D{1, 0, 0},
				V3:     Point3D{0, 1, 0},
			},
			expectError: false,
		},
		{
			name: "invalid normal vector length",
			triangle: Triangle{
				Normal: Point3D{0, 0, 2}, // Not normalized
				V1:     Point3D{0, 0, 0},
				V2:     Point3D{1, 0, 0},
				V3:     Point3D{0, 1, 0},
			},
			expectError: true,
		},
		{
			name: "invalid point coordinates",
			triangle: Triangle{
				Normal: Point3D{0, 0, 1},
				V1:     Point3D{math.NaN(), 0, 0},
				V2:     Point3D{1, 0, 0},
				V3:     Point3D{0, 1, 0},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.triangle.Validate()
			if (err != nil) != tc.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tc.expectError)
			}
		})
	}
}

// TestTriangleToFloat32 tests the conversion from Triangle to TriangleFloat32
func TestTriangleToFloat32(t *testing.T) {
	input := Triangle{
		Normal: Point3D{0, 0, 1},
		V1:     Point3D{0, 0, 0},
		V2:     Point3D{1, 0, 0},
		V3:     Point3D{0, 1, 0},
	}

	expected := TriangleFloat32{
		Normal: Point3DFloat32{0, 0, 1},
		V1:     Point3DFloat32{0, 0, 0},
		V2:     Point3DFloat32{1, 0, 0},
		V3:     Point3DFloat32{0, 1, 0},
	}

	result := input.ToFloat32()
	if result != expected {
		t.Errorf("ToFloat32() = %v, want %v", result, expected)
	}
}

// TestPoint3DEdgeCases tests edge cases for Point3D
func TestPoint3DEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		point    Point3D
		expected bool
	}{
		{
			name:     "zero values",
			point:    Point3D{0, 0, 0},
			expected: true,
		},
		{
			name:     "max float64 values",
			point:    Point3D{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
			expected: true,
		},
		{
			name:     "min float64 values",
			point:    Point3D{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64},
			expected: true,
		},
		{
			name:     "mixed infinity and regular values",
			point:    Point3D{math.Inf(1), 0, 1},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.point.IsValid()
			if result != tc.expected {
				t.Errorf("IsValid() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// TestPoint3DFloat32Construction tests direct construction of Point3DFloat32
func TestPoint3DFloat32Construction(t *testing.T) {
	testCases := []struct {
		name     string
		point    Point3DFloat32
		expected Point3DFloat32
	}{
		{
			name:     "zero values",
			point:    Point3DFloat32{0, 0, 0},
			expected: Point3DFloat32{0, 0, 0},
		},
		{
			name:     "max float32 values",
			point:    Point3DFloat32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32},
			expected: Point3DFloat32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32},
		},
		{
			name:     "typical values",
			point:    Point3DFloat32{1.5, 2.5, 3.5},
			expected: Point3DFloat32{1.5, 2.5, 3.5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.point != tc.expected {
				t.Errorf("Point3DFloat32 = %v, want %v", tc.point, tc.expected)
			}
		})
	}
}
