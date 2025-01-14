// Package geometry provides 3D geometry generation functions for STL models.
package geometry

import (
	"github.com/github/gh-skyline/errors"
	"github.com/github/gh-skyline/types"
)

// CreateQuad creates two triangles forming a quadrilateral from four vertices.
// Returns an error if the vertices form a degenerate quad or contain invalid coordinates.
func CreateQuad(v1, v2, v3, v4 types.Point3D) ([]types.Triangle, error) {
	normal, err := calculateNormal(v1, v2, v3)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate quad normal")
	}

	return []types.Triangle{
		{Normal: normal, V1: v1, V2: v2, V3: v3},
		{Normal: normal, V1: v1, V2: v3, V3: v4},
	}, nil
}

// CreateCuboidBase generates triangles for a rectangular base.
func CreateCuboidBase(width, depth float64) ([]types.Triangle, error) {
	// The base starts at Z = -BaseHeight and extends to Z = 0
	return createBox(0, 0, -BaseHeight, width, depth, BaseHeight)
}

// CreateColumn generates triangles for a vertical column at the specified position.
// The column extends from the base height to the specified height.
func CreateColumn(x, y, height, size float64) ([]types.Triangle, error) {
	// Start at z=0 since the base's top surface is at z=0
	return createBox(x, y, 0, size, size, height)
}

// CreateCube generates triangles forming a cube at the specified position with given dimensions.
// The cube is created in a right-handed coordinate system where:
//   - X increases to the right
//   - Y increases moving away from the viewer
//   - Z increases moving upward
//
// The specified position (x,y,z) defines the front bottom left corner of the cube.
// Returns a slice of triangles that form all six faces of the cube.
func CreateCube(x, y, z, width, height, depth float64) ([]types.Triangle, error) {
	return createBox(x, y, z, width, height, depth)
}

// createBox is an internal helper function that generates triangles for a box shape.
// The box is created in a right-handed coordinate system where:
//   - X increases to the right
//   - Y increases moving away from the viewer
//   - Z increases moving upward
//
// Parameters:
//   - x, y, z: coordinates of the front bottom left corner
//   - width: size along X axis
//   - height: size along Y axis
//   - depth: size along Z axis
//
// All faces are oriented with normals pointing outward from the box.
func createBox(x, y, z, width, height, depth float64) ([]types.Triangle, error) {
	// Validate dimensions
	if width < 0 || height < 0 || depth < 0 {
		return nil, errors.New(errors.ValidationError, "negative dimensions not allowed", nil)
	}

	// Pre-allocate with exact capacity needed
	const facesCount = 6
	const trianglesPerFace = 2
	triangles := make([]types.Triangle, 0, facesCount*trianglesPerFace)

	vertices := make([]types.Point3D, 8) // Pre-allocate vertices array
	quads := [6][4]int{
		{0, 3, 2, 1}, // front (viewed from front)
		{5, 6, 7, 4}, // back (viewed from back)
		{4, 7, 3, 0}, // left (viewed from left)
		{1, 2, 6, 5}, // right (viewed from right)
		{3, 7, 6, 2}, // top (viewed from top)
		{4, 0, 1, 5}, // bottom (viewed from bottom)
	}

	// Fill vertices array
	vertices[0] = types.Point3D{X: x, Y: y, Z: z}
	vertices[1] = types.Point3D{X: x + width, Y: y, Z: z}
	vertices[2] = types.Point3D{X: x + width, Y: y + height, Z: z}
	vertices[3] = types.Point3D{X: x, Y: y + height, Z: z}
	vertices[4] = types.Point3D{X: x, Y: y, Z: z + depth}
	vertices[5] = types.Point3D{X: x + width, Y: y, Z: z + depth}
	vertices[6] = types.Point3D{X: x + width, Y: y + height, Z: z + depth}
	vertices[7] = types.Point3D{X: x, Y: y + height, Z: z + depth}

	// Generate triangles
	for _, quad := range quads {
		quadTriangles, err := CreateQuad(
			vertices[quad[0]],
			vertices[quad[1]],
			vertices[quad[2]],
			vertices[quad[3]],
		)

		if err != nil {
			return nil, errors.New(errors.STLError, "failed to create quad", err)
		}

		triangles = append(triangles, quadTriangles...)
	}

	return triangles, nil
}
