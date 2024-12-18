// Package types provides data structures and functions for handling
// GitHub contribution data and 3D geometry for STL file generation.
package types

import (
	"errors"
	"math"
	"time"
)

// ContributionDay represents a single day of GitHub contributions.
type ContributionDay struct {
	ContributionCount int    `json:"contributionCount"`
	Date              string `json:"date"`
}

// IsAfter checks if the contribution day is after the given time
func (c ContributionDay) IsAfter(t time.Time) bool {
	date, err := time.Parse("2006-01-02", c.Date)
	if err != nil {
		return false
	}
	return date.After(t)
}

// Validate checks if the ContributionDay has valid data.
// Returns an error if the date is not in the correct format or if
// the contribution count is negative.
func (c ContributionDay) Validate() error {
	if _, err := time.Parse("2006-01-02", c.Date); err != nil {
		return errors.New("invalid date format, expected YYYY-MM-DD")
	}
	if c.ContributionCount < 0 {
		return errors.New("contribution count cannot be negative")
	}
	return nil
}

// ContributionsResponse represents the contribution data returned by the GitHub API.
type ContributionsResponse struct {
	User struct {
		Login                   string `json:"login"`
		ContributionsCollection struct {
			ContributionCalendar struct {
				TotalContributions int `json:"totalContributions"`
				Weeks              []struct {
					ContributionDays []ContributionDay `json:"contributionDays"`
				} `json:"weeks"`
			} `json:"contributionCalendar"`
		} `json:"contributionsCollection"`
	} `json:"user"`
}

// Point3D represents a point in 3D space using float64 for accuracy in calculations.
// Each coordinate (X, Y, Z) represents a position in 3D space.
type Point3D struct {
	X, Y, Z float64
}

// IsValid checks if the point coordinates are valid (not NaN or Inf)
func (p Point3D) IsValid() bool {
	return !math.IsNaN(p.X) && !math.IsInf(p.X, 0) &&
		!math.IsNaN(p.Y) && !math.IsInf(p.Y, 0) &&
		!math.IsNaN(p.Z) && !math.IsInf(p.Z, 0)
}

// Point3DFloat32 represents a point in 3D space using float32 for STL output.
// This type is specifically used for STL file format compatibility.
type Point3DFloat32 struct {
	X, Y, Z float32
}

// ToFloat32 converts a Point3D to Point3DFloat32.
// The conversion from float64 to float32 is necessary for STL file format compatibility,
// as the STL binary format specifically requires 32-bit floating-point numbers.
// While calculations are done in float64 for better precision, the final output
// must conform to the STL specification.
func (p Point3D) ToFloat32() Point3DFloat32 {
	return Point3DFloat32{
		X: float32(p.X),
		Y: float32(p.Y),
		Z: float32(p.Z),
	}
}

// Triangle represents a triangle in 3D space using float64 coordinates.
// It consists of a normal vector and three vertices defining the triangle.
type Triangle struct {
	Normal     Point3D
	V1, V2, V3 Point3D
}

// Validate checks if the triangle is valid by verifying all points
// are valid and the normal vector is properly normalized.
func (t Triangle) Validate() error {
	if !t.Normal.IsValid() || !t.V1.IsValid() || !t.V2.IsValid() || !t.V3.IsValid() {
		return errors.New("triangle contains invalid coordinates")
	}

	// Check if normal vector is normalized (length ≈ 1)
	normalLength := math.Sqrt(t.Normal.X*t.Normal.X + t.Normal.Y*t.Normal.Y + t.Normal.Z*t.Normal.Z)
	if math.Abs(normalLength-1.0) > 1e-6 {
		return errors.New("normal vector is not normalized")
	}

	return nil
}

// TriangleFloat32 represents a triangle with float32 coordinates for STL output.
// This type is specifically used for STL file format compatibility.
type TriangleFloat32 struct {
	Normal     Point3DFloat32
	V1, V2, V3 Point3DFloat32
}

// ToFloat32 converts a Triangle to TriangleFloat32.
// This conversion is required for STL file format compliance, which mandates
// the use of 32-bit floating-point numbers. While internal calculations use
// float64 for improved accuracy, the final STL output must use float32 values
// to maintain compatibility with CAD and 3D printing software.
func (t Triangle) ToFloat32() TriangleFloat32 {
	return TriangleFloat32{
		Normal: t.Normal.ToFloat32(),
		V1:     t.V1.ToFloat32(),
		V2:     t.V2.ToFloat32(),
		V3:     t.V3.ToFloat32(),
	}
}
