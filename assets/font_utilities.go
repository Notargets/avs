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

func NewTextFormatter(fontBaseName, fontOptionName string, fontPitch, windowWidth int,
	color color.Color, centered, screenFixed bool, xRange, yRange float32) (tf *TextFormatter) {
	tf = &TextFormatter{
		Color:       color,
		Centered:    centered,
		ScreenFixed: screenFixed,
	}
	tf.TypeFace = NewOpenGLTypeFace(fontBaseName, fontOptionName, fontPitch, windowWidth, xRange, yRange)
	return
}
