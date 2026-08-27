package geometry

import (
	"fmt"
	"image/png"
	"os"

	"github.com/fogleman/gg"
	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/types"
)

const (
	baseWidthVoxelResolution = 2000 // Number of voxels across the skyline face
	voxelDepth               = 1.0  // Distance to come out of face

	logoScale      = 0.4  // Percent
	logoTopOffset  = 0.15 // Percent
	logoLeftOffset = 0.03 // Percent

	textEdgeOffsetPercent = 0.03
)

// textRenderConfig controls how text is sized and positioned horizontally.
// edgeOffsetPercent is measured from the edge implied by justification: left
// text is offset from the left edge and right text from the right edge.
type textRenderConfig struct {
	fontSize          float64
	justification     string
	edgeOffsetPercent float64
}

var (
	usernameTextConfig = textRenderConfig{
		fontSize:          120.0,
		justification:     "left",
		edgeOffsetPercent: textEdgeOffsetPercent,
	}
	yearTextConfig = textRenderConfig{
		fontSize:          100.0,
		justification:     "right",
		edgeOffsetPercent: textEdgeOffsetPercent,
	}
)

// Create3DText generates 3D text geometry for the username and year.
func Create3DText(username string, year string, baseWidth float64, baseHeight float64) ([]types.Triangle, error) {
	if username == "" {
		username = "anonymous"
	}

	usernameTriangles, err := renderText(
		username,
		usernameTextConfig,
		baseWidth,
		baseHeight,
	)
	if err != nil {
		return nil, err
	}

	yearTriangles, err := renderText(
		year,
		yearTextConfig,
		baseWidth,
		baseHeight,
	)
	if err != nil {
		return nil, err
	}

	return append(usernameTriangles, yearTriangles...), nil
}

// renderText places text on the face of a skyline, offset from its nearest edge and vertically-aligned.
// The function takes the text to be displayed and its rendering configuration.
// It returns an array of types.Triangle.
//
// Parameters:
//
//	text (string): The text to be displayed on the skyline's front face.
//	config (textRenderConfig): Font size, justification, and distance from the corresponding edge.
//
// Returns:
//
//	([]types.Triangle, error): A slice of triangles representing text.
func renderText(text string, config textRenderConfig, baseWidth float64, baseHeight float64) ([]types.Triangle, error) {
	// Create a rendering context for the face of the skyline
	faceWidthRes := baseWidthVoxelResolution
	faceHeightRes := int(float64(faceWidthRes) * baseHeight / baseWidth)

	// Create image representing the skyline face
	dc := gg.NewContext(faceWidthRes, faceHeightRes)
	dc.SetRGB(0, 0, 0)
	dc.Clear()
	dc.SetRGB(1, 1, 1)

	// Load font into context
	fontPath, cleanup, err := writeTempFont(PrimaryFont)
	if err != nil {
		// Try fallback font
		fontPath, cleanup, err = writeTempFont(FallbackFont)
		if err != nil {
			return nil, errors.New(errors.IOError, "failed to load any fonts", err)
		}
	}
	if err := dc.LoadFontFace(fontPath, config.fontSize); err != nil {
		return nil, errors.New(errors.IOError, "failed to load font", err)
	}

	// Draw text on image at desired location
	var triangles []types.Triangle

	horizontalPosition, justificationPercent := config.horizontalPlacement(faceWidthRes)

	dc.DrawStringAnchored(
		text,
		horizontalPosition,         // Offset from the corresponding edge
		float64(faceHeightRes)*0.5, // Offset from top
		justificationPercent,       // Justification (0.0=left, 0.5=center, 1.0=right)
		0.5,                        // Vertically aligned
	)

	// Convert context image pixels into voxels
	for x := 0; x < faceWidthRes; x++ {
		for y := 0; y < faceHeightRes; y++ {
			if isPixelActive(dc, x, y) {
				voxel, err := createVoxelOnFace(
					float64(x),
					float64(y),
					voxelDepth,
					baseWidth,
					baseHeight,
				)
				if err != nil {
					return nil, errors.New(errors.STLError, "failed to create cube", err)
				}

				triangles = append(triangles, voxel...)
			}
		}
	}

	defer cleanup()

	return triangles, nil
}

// horizontalPlacement returns the horizontal draw position and anchor used by
// gg.DrawStringAnchored. Right-justified text measures edgeOffsetPercent from
// the right edge so left- and right-aligned labels can share one edge distance.
func (config textRenderConfig) horizontalPlacement(faceWidth int) (position float64, anchor float64) {
	switch config.justification {
	case "center":
		return float64(faceWidth) * 0.5, 0.5
	case "right":
		return float64(faceWidth) * (1 - config.edgeOffsetPercent), 1.0
	default:
		return float64(faceWidth) * config.edgeOffsetPercent, 0.0
	}
}

// createVoxelOnFace creates a voxel on the face of a skyline by generating a cube at the specified coordinates.
// The function takes in the x, y coordinates and height.
// It returns a slice of types.Triangle representing the cube and an error if the cube creation fails.
//
// Parameters:
//
//	x (float64): The x-coordinate on the skyline face (left to right).
//	y (float64): The y-coordinate on the skyline face (top to bottom).
//	height (float64): Distance coming out of the face.
//
// Returns:
//
//	([]types.Triangle, error): A slice of triangles representing the cube and an error if any.
func createVoxelOnFace(x float64, y float64, height float64, baseWidth float64, baseHeight float64) ([]types.Triangle, error) {
	// Mapping resolution
	xResolution := float64(baseWidthVoxelResolution)
	yResolution := xResolution * baseHeight / baseWidth

	// Pixel size
	voxelSize := 1.0

	// Scale coordinate to face resolution
	x = (x / xResolution) * baseWidth
	y = (y / yResolution) * baseHeight
	voxelSizeX := (voxelSize / xResolution) * baseWidth
	voxelSizeY := (voxelSize / yResolution) * baseHeight

	cube, err := CreateCube(
		// Location (from top left corner of skyline face)
		x,             // x - Left to right
		-height,       // y - Negative comes out of face. Positive goes into face.
		-voxelSizeY-y, // z - Bottom to top

		// Size
		voxelSizeX, // x length - left to right from specified point
		height,     // thickness - distance coming out of face
		voxelSizeY, // y length - bottom to top from specified point
	)

	return cube, err
}

// GenerateImageGeometry creates 3D geometry from the embedded logo image.
func GenerateImageGeometry(baseWidth float64, baseHeight float64) ([]types.Triangle, error) {
	// Get temporary image file
	imgPath, cleanup, err := getEmbeddedImage()
	if err != nil {
		return nil, err
	}

	defer cleanup()

	return renderImage(
		imgPath,
		logoScale,
		voxelDepth,
		logoLeftOffset,
		logoTopOffset,
		baseWidth,
		baseHeight,
	)
}

// renderImage generates 3D geometry for the given image configuration.
func renderImage(filePath string, scale float64, height float64, leftOffsetPercent float64, topOffsetPercent float64, baseWidth float64, baseHeight float64) ([]types.Triangle, error) {

	// Get voxel resolution of base face
	faceWidthRes := baseWidthVoxelResolution
	faceHeightRes := int(float64(faceWidthRes) * baseHeight / baseWidth)

	// Load image from file
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New(errors.IOError, "failed to open image", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			closeErr := errors.New(errors.IOError, "failed to close reader", err)
			// Log the error or handle it appropriately
			fmt.Println(closeErr)
		}
	}()
	img, err := png.Decode(reader)
	if err != nil {
		return nil, errors.New(errors.IOError, "failed to decode PNG", err)
	}

	// Get image size
	bounds := img.Bounds()
	logoWidth := bounds.Max.X
	logoHeight := bounds.Max.Y

	// Transfer image pixels onto face of skyline as voxels
	var triangles []types.Triangle
	for x := 0; x < logoWidth; x++ {
		for y := logoHeight - 1; y >= 0; y-- {
			// Get pixel color and alpha
			r, _, _, a := img.At(x, y).RGBA()

			// If pixel is active (white) and not fully transparent, create a voxel
			if a > 32768 && r > 32768 {

				voxel, err := createVoxelOnFace(
					(leftOffsetPercent*float64(faceWidthRes))+float64(x)*scale,
					(topOffsetPercent*float64(faceHeightRes))+float64(y)*scale,
					height,
					baseWidth,
					baseHeight,
				)

				if err != nil {
					return nil, errors.New(errors.STLError, "failed to create cube", err)
				}

				triangles = append(triangles, voxel...)
			}
		}
	}

	return triangles, nil
}

// isPixelActive checks if a pixel is active (white) in the given context.
func isPixelActive(dc *gg.Context, x, y int) bool {
	r, _, _, _ := dc.Image().At(x, y).RGBA()
	return r > 32768
}
