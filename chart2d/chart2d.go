package chart2d

import (
	"fmt"
	"image/color"
	"math"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/screen"
)

type Chart2D struct {
	Scale       float32
	Position    [2]float32
	XMin, XMax  float32
	YMin, YMax  float32
	Screen      *screen.Screen
	LineColor   color.Color
	ScreenColor color.Color
}

type Color [4]float32 // RGBA

func NewChart2D(XMin, XMax, YMin, YMax, scale float32, width, height int) (chart *Chart2D) {
	chart = &Chart2D{
		XMin:        XMin,
		XMax:        XMax,
		YMin:        YMin,
		YMax:        YMax,
		Screen:      screen.NewScreen(width, height, XMin, XMax, YMin, YMax, scale),
		LineColor:   color.RGBA{255, 255, 255, 255},
		ScreenColor: color.RGBA{46, 46, 46, 255},
	}
	chart.Screen.SetBackgroundColor(chart.ScreenColor)
	return
}

func (chart *Chart2D) NewTextFormatter(fontBaseName, fontOptionName string, fontPitch int, fontColor color.Color,
	centered, screenFixed bool) (tf *assets.TextFormatter) {
	tf = chart.Screen.NewTextFormatter(fontBaseName, fontOptionName, fontPitch, fontColor, centered, screenFixed)
	return
}

func (chart *Chart2D) Printf(formatter *assets.TextFormatter, x, y float32, format string, args ...interface{}) {
	chart.Screen.Printf(formatter, x, y, format, args...)
}

func (chart *Chart2D) GetWorldSpaceCharHeight(tf *assets.TextFormatter) (height float32) {
	return chart.Screen.GetWorldSpaceCharHeight(tf)
}

func (chart *Chart2D) AddLine(X, Y []float32) {
	chart.Screen.NewPolyLine(screen.NEW, X, Y, chart.GetSingleColorArray(Y, chart.LineColor))
}

func (chart *Chart2D) GetSingleColorArray(Y []float32, singleColor color.Color) (colors []float32) {
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

func (chart *Chart2D) AddAxis(axisColor color.Color, tf *assets.TextFormatter, yAxisLocation float32, nSegs int) {
	var (
		xMin, xMax           = chart.XMin, chart.XMax
		yMin, yMax           = chart.YMin, chart.YMax
		xScale, yScale       = xMax - xMin, yMax - yMin
		xInc                 = xScale / float32(nSegs-1)
		yInc                 = yScale / float32(nSegs-1)
		xTickSize, yTickSize = 0.020 * xScale, 0.020 * yScale
		tickColor            = axisColor
		X                    = []float32{}
		Y                    = []float32{}
		C                    = []float32{}
	)
	if nSegs%2 == 0 {
		panic("nSegs must be odd")
	}

	// Generate color array for 2 vertices per axis (X-axis and Y-axis)
	X, Y, C = AddSegmentToLine(X, Y, C, xMin, 0, xMax, 0, axisColor)
	X, Y, C = AddSegmentToLine(X, Y, C, yAxisLocation, yMin, yAxisLocation, yMax, axisColor)

	// Draw ticks along X axis
	var x, y = xMin, float32(0) // X axis is always drawn at Y = 0
	for i := 0; i < nSegs; i++ {
		if x == yAxisLocation {
			x = x + xInc
			continue
		}
		X, Y, C = AddSegmentToLine(X, Y, C, x, y, x, y-yTickSize, tickColor)
		x = clampNearZero(x, xScale/1000.)
		chart.Printf(tf, x, y-(chart.GetWorldSpaceCharHeight(tf)+yTickSize), "%4.1f", x)
		x = x + xInc
	}
	ptfY := *tf
	tfY := &ptfY
	tfY.Centered = false
	x = yAxisLocation
	y = yMin
	yTextDelta := calculateYTickTextOffset(yMin, chart.GetWorldSpaceCharHeight(tfY))
	for i := 0; i < nSegs; i++ {
		if i == nSegs/2 {
			y = y + yInc
			continue
		}
		X, Y, C = AddSegmentToLine(X, Y, C, x, y, x-xTickSize, y, tickColor)
		y = clampNearZero(y, yScale/1000.)
		chart.Printf(tfY, x-yTextDelta, y, "%4.1f", y)
		y = y + yInc
	}
	//chart.Screen.ChangePosition(0.0, 0.0)
	chart.Screen.NewLine(screen.NEW, X, Y, C) // 2 points, so 2 * 3 = 6 colors
}

func calculateYTickTextOffset(y float32, charWidth float32) (delta float32) {
	if y < 0 {
		delta = charWidth * .5 // Minus sign is about 1/2 char
	}
	d := float32(math.Ceil(math.Log10(math.Abs(float64(y))))) + 1.1 // Add some for the decimal and the trailing zero
	fmt.Printf("y: %v, CharWidth: %v, d: %v\n", y, charWidth, d)
	delta += d * charWidth
	return
}

func clampNearZero(x, epsilon float32) float32 {
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

func ColorToFloat32(c color.Color) [3]float32 {
	r, g, b, _ := c.RGBA() // Extract RGBA as uint32
	return [3]float32{
		float32(r) / 65535.0,
		float32(g) / 65535.0,
		float32(b) / 65535.0,
	}
}
