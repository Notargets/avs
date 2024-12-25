/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package assets

import (
	"image/color"
)

type TextFormatter struct {
	Color       color.Color
	Centered    bool
	ScreenFixed bool
	TypeFace    *OpenGLTypeFace
}

func NewTextFormatter(fontBaseName, fontOptionName string, fontPitch uint32, color color.Color,
	centered, screenFixed bool) (tf *TextFormatter) {
	tf = &TextFormatter{
		Color:       color,
		Centered:    centered,
		ScreenFixed: screenFixed,
	}
	tf.TypeFace = NewOpenGLTypeFace(fontBaseName, fontOptionName, fontPitch)
	return
}

func (tf *TextFormatter) GetWorldSpaceCharHeight(yRange float32, windowWidth, windowHeight uint32) (charHeight float32) {
	// Implement a scale factor to reduce the polygon size commensurate with the dynamic DPI scaling, relative to the
	// standard 72 DPI of the Opentype package
	worldPerPixel := yRange / float32(windowHeight)
	screenRatio := float32(windowHeight) / float32(windowWidth)
	pixelHeight := tf.TypeFace.FontHeight
	// fmt.Printf("pitch: %v, pixelHeight: %v, DPI: %v\n", tf.TypeFace.FontPitch, pixelHeight, tf.TypeFace.FontDPI)
	// height includes the inter-line height, so divide by 1.5
	charHeight = (worldPerPixel) * float32(pixelHeight) * float32(72) / float32(tf.TypeFace.FontDPI) * screenRatio / 1.5
	return
}

func (tf *TextFormatter) GetWorldSpaceCharWidth(xRange, yRange float32, windowWidth, windowHeight uint32) (charWidth float32) {
	charHeight := tf.GetWorldSpaceCharHeight(yRange, windowWidth, windowHeight)
	// ccale the height by the world aspect ratio to get the width
	charWidth = charHeight * xRange / yRange
	return
}
