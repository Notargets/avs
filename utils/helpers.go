package utils

import (
	"image/color"
	"math"
)

func ColorToFloat32(c color.Color) [4]float32 {
	r, g, b, a := c.RGBA()
	return [4]float32{
		float32(r) / 65535.0,
		float32(g) / 65535.0,
		float32(b) / 65535.0,
		float32(a) / 65535.0,
	}
}

func ClampNearZero(x, epsilon float32) float32 {
	if float32(math.Abs(float64(x))) < epsilon {
		return 0
	}
	return x
}

func AddSegmentToLine(X, Y, C []float32, X1, Y1, X2, Y2 float32, lineColor color.Color) (XX, YY, CC []float32) {
	XX = append(X, X1, X2)
	YY = append(Y, Y1, Y2)

	c := ColorToFloat32(lineColor)
	CC = append(C, c[0], c[1], c[2], c[0], c[1], c[2])
	return
}

func CalculateRightJustifiedTextOffset(yRight float32, charWidth float32) (deltaYLeft float32) {
	if yRight < 0 {
		deltaYLeft = charWidth * .5 // Minus sign is about 1/2 char
	}
	d := float32(math.Ceil(math.Log10(math.Abs(float64(yRight))))) + 1.1 // Add some for the decimal and the trailing zero
	//fmt.Printf("y: %v, CharWidth: %v, d: %v\n", y, charWidth, d)
	deltaYLeft += d * charWidth
	return
}

func GetSingleColorArray(Y []float32, singleColor color.Color) (colors []float32) {
	colors = make([]float32, len(Y)*3)
	r, g, b, _ := singleColor.RGBA() // Extract RGBA as uint32
	fc := [3]float32{
		float32(r) / 65535.0,
		float32(g) / 65535.0,
		float32(b) / 65535.0,
	}
	for i := range colors {
		colors[i*3] = fc[0]
		colors[i*3+1] = fc[1]
		colors[i*3+2] = fc[2]
	}
	return
}
