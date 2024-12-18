package assets

import (
	"image/color"
)

type TextFormatter struct {
	Color        color.Color
	Centered     bool
	ScreenFixed  bool
	TypeFace     *OpenGLTypeFace
	WindowHeight uint32
	WindowWidth  uint32
}

func NewTextFormatter(fontBaseName, fontOptionName string, fontPitch, windowWidth, windowHeight int,
	color color.Color, centered, screenFixed bool, xRange, yRange float32) (tf *TextFormatter) {
	tf = &TextFormatter{
		Color:        color,
		Centered:     centered,
		ScreenFixed:  screenFixed,
		WindowWidth:  uint32(windowWidth),
		WindowHeight: uint32(windowHeight),
	}
	tf.TypeFace = NewOpenGLTypeFace(fontBaseName, fontOptionName, fontPitch, windowWidth, xRange, yRange)
	return
}
