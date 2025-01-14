/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
)

func GetColorArray(ColorAny interface{}, length int) (ColorArray []float32) {
	colorArrayLength := 3 * length
	switch c := ColorAny.(type) {
	case color.RGBA:
		ColorArray = make([]float32, colorArrayLength)
		ColorFloat := ColorToFloat32(c)
		for i := 0; i < colorArrayLength; i++ {
			ColorArray[i] = ColorFloat[i%3]
		}
	case [3]float32:
		// Expand the single color into an array to match X/Y
		ColorArray = make([]float32, colorArrayLength)
		for i := 0; i < colorArrayLength; i++ {
			ColorArray[i] = c[i%3]
		}
		return
	case [4]float32:
		// Expand the single color into an array to match X/Y
		ColorArray = make([]float32, colorArrayLength)
		for i := 0; i < colorArrayLength; i++ {
			ColorArray[i] = c[i%3]
		}
		return
	case []float32:
		if len(c) <= 4 { // Incoming color is a single RGB or RGBA float
			// Expand the single color into an array to match X/Y
			ColorArray = make([]float32, colorArrayLength)
			for i := 0; i < colorArrayLength; i++ {
				ColorArray[i] = c[i%3]
			}
			return
		} else if len(c) != length*3 {
			panic(fmt.Errorf("Length of input colors: %d is not equal to set"+
				" length: %d\n", len(c), length))
		} else {
			ColorArray = ColorAny.([]float32)
		}
	default:
		panic(fmt.Errorf("Unknown type: %T\n", ColorAny))
	}
	return
}

func ColorToFloat32(c color.RGBA) [4]float32 {
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

func AddSegmentToLine(XY []float32, X1, Y1, X2, Y2 float32) (XXYY []float32) {
	XXYY = append(XY, X1, Y1, X2, Y2)
	return
}

func CalculateRightJustifiedTextOffset(yRight float32, charWidth float32) (deltaYLeft float32) {
	if yRight < 0 {
		deltaYLeft = charWidth * .5 // Minus sign is about 1/2 char
	}
	if math.Abs(float64(yRight)) < 10. {
		yRight = 10.
	}
	d := float32(math.Ceil(math.Log10(math.Abs(float64(yRight))))) + 1.1 // add some for the decimal and the trailing zero
	// fmt.Printf("y: %v, CharWidth: %v, d: %v\n", y, charWidth, d)
	deltaYLeft += d * charWidth
	return
}

// Int32ToBytes converts an int32 to a byte slice in little-endian order.
func Int32ToBytes(val int32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, val)
	return buf.Bytes()
}

// Float32ToBytes converts a float32 to a byte slice in little-endian order.
func Float32ToBytes(val float32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, val)
	return buf.Bytes()
}
