package geometry

import (
	"fmt"
	"image/png"
	"os"

	"github.com/fogleman/gg"
	"github.com/github/gh-skyline/errors"
	"github.com/github/gh-skyline/types"
)

// Common configuration for rendered elements
type renderConfig struct {
	startX     float64
	startY     float64
	startZ     float64
	voxelScale float64
	depth      float64
}

// TextConfig holds parameters for text rendering
type textRenderConfig struct {
	renderConfig
	text          string
	contextWidth  int
	contextHeight int
	fontSize      float64
	alignment     gg.Align
}

// ImageConfig holds parameters for image rendering
type imageRenderConfig struct {
	renderConfig
	imagePath string
	width    float64
}

const (
	imagePosition = 0.025
	gap           = 0.05 // 5%

	defaultContextWidth  = 800
	defaultContextHeight = 200
	textVoxelSize        = 1.0
	textDepthOffset      = 2.0
	frontEmbedDepth      = 1.5

	usernameContextWidth  = 1000
	usernameContextHeight = 200
	usernameFontSize      = 48.0
	usernameZOffset       = 0.7

	yearContextWidth  = 800
	yearContextHeight = 200
	yearFontSize      = 56.0
	yearZOffset       = 0.4

	defaultImageWidth  = 9.2
	defaultImageScale  = 0.8
	imageLeftMargin    = 10.0
)

// Create3DText generates 3D text geometry for the username and year.
func Create3DText(username string, year string, innerWidth, baseHeight float64) ([]types.Triangle, error) {
	if username == "" {
		username = "anonymous"
	}

	usernameConfig := textRenderConfig{
		renderConfig: renderConfig{
			startX:     innerWidth*(2*gap) + defaultImageWidth*defaultImageScale,
			startY:     -textDepthOffset / 2,
			startZ:     baseHeight * usernameZOffset,
			voxelScale: textVoxelSize,
			depth:      frontEmbedDepth,
		},
		text:          username,
		contextWidth:  usernameContextWidth,
		contextHeight: usernameContextHeight,
		fontSize:      usernameFontSize,
		alignment:     gg.AlignLeft,
	}

	yearStartX := innerWidth*(1-gap) - (yearContextWidth*textVoxelSize*0.75/8)
	yearConfig := textRenderConfig{
		renderConfig: renderConfig{
			startX:     yearStartX,
			startY:     -textDepthOffset / 2,
			startZ:     baseHeight * yearZOffset,
			voxelScale: textVoxelSize * 0.75,
			depth:      frontEmbedDepth,
		},
		text:          year,
		contextWidth:  yearContextWidth,
		contextHeight: yearContextHeight,
		fontSize:      yearFontSize,
		alignment:     gg.AlignRight,
	}

	usernameTriangles, err := renderText(usernameConfig)
	if err != nil {
		return nil, err
	}

	yearTriangles, err := renderText(yearConfig)
	if err != nil {
		return nil, err
	}

	return append(usernameTriangles, yearTriangles...), nil
}

// renderText generates 3D geometry for the given text configuration.
func renderText(config textRenderConfig) ([]types.Triangle, error) {
	dc := gg.NewContext(config.contextWidth, config.contextHeight)

	// Get temporary font file
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

	dc.SetRGB(0, 0, 0)
	dc.Clear()
	dc.SetRGB(1, 1, 1)
	dc.DrawStringWrapped(config.text, 0, float64(config.contextHeight)/2, 0.0, 0.5, float64(config.contextWidth), 1.5, config.alignment)

	var triangles []types.Triangle

	for y := 0; y < config.contextHeight; y++ {
		for x := 0; x < config.contextWidth; x++ {
			if isPixelActive(dc, x, y) {
				xPos := config.startX + float64(x)*config.voxelScale/8
				zPos := config.startZ - float64(y)*config.voxelScale/8

				voxel, err := CreateCube(
					xPos,
					config.startY,
					zPos,
					config.voxelScale/2,
					config.depth,
					config.voxelScale/2,
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

// GenerateImageGeometry creates 3D geometry from the embedded logo image.
func GenerateImageGeometry(innerWidth, baseHeight float64) ([]types.Triangle, error) {
	// Get temporary image file
	imgPath, cleanup, err := getEmbeddedImage()
	if err != nil {
		return nil, err
	}

	config := imageRenderConfig{
		renderConfig: renderConfig{
			startX:     innerWidth * gap,
			startY:     -frontEmbedDepth / 2.0,
			startZ:     -0.85 * baseHeight,
			voxelScale: defaultImageScale,
			depth:      frontEmbedDepth,
		},
		imagePath: imgPath,
		width:    defaultImageWidth,
	}

	defer cleanup()

	return renderImage(config)
}

// renderImage generates 3D geometry for the given image configuration.
func renderImage(config imageRenderConfig) ([]types.Triangle, error) {
	reader, err := os.Open(config.imagePath)
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

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	scale := config.width / float64(width)

	var triangles []types.Triangle

	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			r, _, _, a := img.At(x, y).RGBA()
			if a > 32768 && r > 32768 {
				xPos := config.startX + float64(x)*config.voxelScale*scale
				zPos := config.startZ + float64(height-1-y)*config.voxelScale*scale

				voxel, err := CreateCube(
					xPos,
					config.startY,
					zPos,
					config.voxelScale*scale,
					config.depth,
					config.voxelScale*scale,
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
