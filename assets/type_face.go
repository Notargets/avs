package assets

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type OpenGLTypeFace struct {
	Face         font.Face // Equivalent to a Font object plus metadata like DPI, etc
	FontFilePath string    // Path to the TTF font file
	FontPitch    uint32    // Font size in "Pitch", eg: 12 Pitch font
	FontHeight   uint32    // Pixel height of font, calculated
	FontDPI      uint32    // Dynamically calculated to ensure quality at all sizes
}

func NewOpenGLTypeFace(fontBaseName, fontOptionName string, fontPitch int, windowWidth int, XRange, YRange float32) (tf *OpenGLTypeFace) {
	tf = &OpenGLTypeFace{
		FontFilePath: FontOptionsMap[fontBaseName][fontOptionName],
		FontPitch:    uint32(fontPitch),
		FontDPI:      calculateDynamicDPI(uint32(fontPitch)),
	}
	if len(tf.FontFilePath) == 0 {
		panic("font_file_path is empty, check your font basename and option name in the asset map")
	}

	// Read the font file bytes
	fontBytes, err := os.ReadFile(tf.FontFilePath)
	if err != nil {
		err = fmt.Errorf("failed to read font file: %v", err)
		panic(err)
	}

	// Parse the font
	ttf, err := opentype.Parse(fontBytes)
	if err != nil {
		err = fmt.Errorf("failed to parse font: %v", err)
		panic(err)
	}

	// Create a font face with the desired font size and DPI
	tf.Face, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    float64(fontPitch),  // Set the font pitch as the font size
		DPI:     float64(tf.FontDPI), // Default DPI
		Hinting: font.HintingFull,    // Enable hinting for accurate measurements
	})
	if err != nil {
		err = fmt.Errorf("failed to create font face: %v", err)
		return
	}

	// Calculate the maximum pixel height of all characters, including downstroke
	tf.FontHeight, err = calculateFontHeight(tf.Face)
	if err != nil {
		err = fmt.Errorf("failed to calculate font height: %v", err)
		return
	}

	return
}

// generateHash generates a simple hash using FNV for a filename and font pitch
func generateHash(filename string, pitch int) uint64 {
	// Create a new FNV-1a hash
	h := fnv.New64a()

	// Write the filename and pitch to the hash
	h.Write([]byte(fmt.Sprintf("%s|%d", filename, pitch)))

	// Return the 64-bit hash as a uint64
	return h.Sum64()
}

func (tf *OpenGLTypeFace) RenderFontTextureImg(text string, fontColor color.Color) (img *image.RGBA) {
	var (
		err error
	)
	// Create an image of the proper size to hold the full text
	img, err = tf.drawText(text, fontColor, color.RGBA{R: 0, G: 0, B: 0, A: 0})
	if err != nil {
		panic(err)
	}
	//SaveDebugImage(img, "debug_image.png")
	//fmt.Printf("Text Width: %d, Height %d\n", textureWidth, textureHeight)
	return
}

// DrawText draws the provided text onto an image using OpenType and returns the dimensions of the image and the image itself
func (tf *OpenGLTypeFace) drawText(text string, fontColor, bgColor color.Color) (*image.RGBA, error) {

	// Calculate the pixel dimensions for the text
	textWidth := calculateStringPixelWidth(tf.Face, text)
	textHeight := tf.FontHeight

	// Create an image to draw the text on (width = textWidth, height = textHeight)
	img := image.NewRGBA(image.Rect(0, 0, int(textWidth), int(textHeight)))

	// Fill the image with the background color
	draw.Draw(img, img.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	// Get the ascent from face metrics to position the baseline correctly
	metrics := tf.Face.Metrics()
	ascent := metrics.Ascent.Round() // Convert fixed.Int26_6 to pixels

	// Set up the font drawer
	drawer := font.Drawer{
		Dst:  img,                                       // Destination image
		Src:  image.NewUniform(fontColor),               // Color to use for text
		Face: tf.Face,                                   // Font face
		Dot:  fixed.Point26_6{X: 0, Y: fixed.I(ascent)}, // Baseline position (start at ascent)
	}

	// Draw the text
	drawer.DrawString(text)

	return img, nil
}

// SaveDebugImage saves an image as a PNG file with the specified filename and logs the result
func SaveDebugImage(img *image.RGBA, filename string) {
	if img == nil {
		fmt.Println("[SaveDebugImage] Image is nil, nothing to save.")
		return
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("[SaveDebugImage] Failed to create image file: %v\n", err)
		return
	}
	defer file.Close()

	// Encode the image as PNG
	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("[SaveDebugImage] Failed to save image as PNG: %v\n", err)
	} else {
		fmt.Printf("[SaveDebugImage] Image successfully saved as %s\n", filename)
	}
}

func calculateDynamicDPI(fontPitch uint32) uint32 {
	switch {
	case fontPitch <= 12:
		return 512
	case fontPitch <= 24: // No need for fontPitch > 12 because fontPitch > 12 is implied
		return 256
	case fontPitch <= 36: // No need for fontPitch > 24 because fontPitch > 24 is implied
		return 128
	default:
		return 96 // If none of the above conditions are met, fallback to 96 DPI
	}
}

// calculateFontHeight computes the maximum pixel height for the font using face.Metrics()
func calculateFontHeight(face font.Face) (uint32, error) {
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()
	totalHeight := ascent + descent
	return uint32(totalHeight), nil
}

func calculateStringPixelWidth(fontFace font.Face, text string) uint32 {
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
	return uint32(totalWidth.Round()) // Convert fixed.Int26_6 to integer pixels
}
