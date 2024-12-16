package assets

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// LoadFont loads a TTF/OTF font file from the given file path, returns a font face,
// the max pixel height (including downstroke), and an error if applicable.
func LoadFont(filePath string, fontPitch float32, dpi int) (face font.Face, pixelHeight int, err error) {
	// Read the font file bytes
	fontBytes, err := os.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("failed to read font file: %v", err)
		return
	}

	// Parse the font
	ttf, err := opentype.Parse(fontBytes)
	if err != nil {
		err = fmt.Errorf("failed to parse font: %v", err)
		return
	}

	// Create a font face with the desired font size and DPI
	face, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    float64(fontPitch), // Set the font pitch as the font size
		DPI:     float64(dpi),       // Default DPI
		Hinting: font.HintingFull,   // Enable hinting for accurate measurements
	})
	if err != nil {
		err = fmt.Errorf("failed to create font face: %v", err)
		return
	}

	// Calculate the maximum pixel height of all characters, including downstroke
	pixelHeight, err = calculateFontHeight(face)
	if err != nil {
		err = fmt.Errorf("failed to calculate font height: %v", err)
		return
	}

	return
}

func CalculateDynamicFontDPI(fontPitch float32) int {
	switch {
	case fontPitch <= 12:
		return 512
	case fontPitch > 12 && fontPitch <= 24:
		return 256
	case fontPitch > 24 && fontPitch <= 36:
		return 128
	case fontPitch > 36:
		return 96
	default:
		return 96 // Default DPI if none of the above conditions are met
	}
}

// calculateFontHeight computes the maximum pixel height for the font using face.Metrics()
func calculateFontHeight(face font.Face) (int, error) {
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()
	totalHeight := ascent + descent
	return totalHeight, nil
}

// CalculateStringPixelWidth calculates the pixel width of a string
func CalculateStringPixelWidth(fontFace font.Face, text string) int {
	var totalWidth fixed.Int26_6
	for i, char := range text {
		advance, ok := fontFace.GlyphAdvance(char)
		if ok {
			totalWidth += advance
		}
		if i < len(text)-1 {
			nextChar := rune(text[i+1])
			kern := fontFace.Kern(char, nextChar)
			totalWidth += kern
		}
	}
	return totalWidth.Round() // Convert fixed.Int26_6 to integer pixels
}

// DrawText draws the provided text onto an image using OpenType and returns the dimensions of the image and the image itself
func DrawText(face font.Face, text string, fontColor [4]float32, bgColor [4]float32) (int, int, *image.RGBA, error) {
	// Calculate the pixel dimensions for the text
	textWidth := CalculateStringPixelWidth(face, text)
	textHeight, err := calculateFontHeight(face)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to calculate font height: %v", err)
	}

	// Create an image to draw the text on (width = textWidth, height = textHeight)
	img := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))

	// Convert background color from [4]float64 to image color.RGBA
	bg := color.RGBA{
		R: uint8(bgColor[0] * 255),
		G: uint8(bgColor[1] * 255),
		B: uint8(bgColor[2] * 255),
		A: uint8(bgColor[3] * 255),
	}

	// Fill the image with the background color
	draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	// Convert font color from [4]float64 to image color.RGBA
	fontCol := color.RGBA{
		R: uint8(fontColor[0] * 255),
		G: uint8(fontColor[1] * 255),
		B: uint8(fontColor[2] * 255),
		A: uint8(fontColor[3] * 255),
	}

	// Get the ascent from face metrics to position the baseline correctly
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round() // Convert fixed.Int26_6 to pixels

	// Set up the font drawer
	drawer := font.Drawer{
		Dst:  img,                                       // Destination image
		Src:  image.NewUniform(fontCol),                 // Color to use for text
		Face: face,                                      // Font face
		Dot:  fixed.Point26_6{X: 0, Y: fixed.I(ascent)}, // Baseline position (start at ascent)
	}

	// Draw the text
	drawer.DrawString(text)

	return textWidth, textHeight, img, nil
}
