package chart2d

import (
	"fmt"

	"github.com/notargets/avs/screen"
)

type Chart2D struct {
	Scale       float32
	Position    [2]float32
	XMin, XMax  float32
	YMin, YMax  float32
	Screen      *screen.Screen
	LineColor   Color
	AxisColor   Color
	ScreenColor Color
}

type Color [4]float32 // RGBA

func NewChart2D(XMin, XMax, YMin, YMax float32, width, height int) (chart *Chart2D) {
	chart = &Chart2D{
		XMin:        XMin,
		XMax:        XMax,
		YMin:        YMin,
		YMax:        YMax,
		Screen:      screen.NewScreen(width, height, 0, 1, 0, 1),
		LineColor:   Color{1, 1, 1, 1},
		AxisColor:   Color{1, 1, 1, 1},
		ScreenColor: Color{0.18, 0.18, 0.18, 1.},
	}
	chart.Screen.SetBackgroundColor(chart.ScreenColor)
	return
}

func (chart *Chart2D) AddLine(X, Y []float32) {
	chart.TransformXYToUnit(X, Y)

	chart.Screen.AddPolyLine(screen.NEW, X, Y, chart.GetSingleColorArray(Y, chart.LineColor))
}

func (chart *Chart2D) GetSingleColorArray(Y []float32, color Color) (colors []float32) {
	colors = make([]float32, len(Y)*3)
	for i := range colors {
		colors[i*3] = color[0]
		colors[i*3+1] = color[1]
		colors[i*3+2] = color[2]
	}
	return
}

// TransformXYToUnit normalizes the X and Y arrays into the [0,1] range
func (chart *Chart2D) TransformXYToUnit(X, Y []float32) (err error) {
	if len(X) != len(Y) {
		return fmt.Errorf("X and Y must have the same length, got X: %d, Y: %d", len(X), len(Y))
	}
	if chart.XMin == chart.XMax {
		return fmt.Errorf("Invalid X-axis bounds: XMin (%.2f) cannot be equal to XMax (%.2f)", chart.XMin, chart.XMax)
	}
	if chart.YMin == chart.YMax {
		return fmt.Errorf("Invalid Y-axis bounds: YMin (%.2f) cannot be equal to YMax (%.2f)", chart.YMin, chart.YMax)
	}

	xRange := chart.XMax - chart.XMin
	yRange := chart.YMax - chart.YMin
	for i := 0; i < len(X); i++ {
		X[i] = (X[i] - chart.XMin) / xRange
		Y[i] = (Y[i] - chart.YMin) / yRange

		// Clamp values to [0,1] in case of small floating-point errors
		if X[i] < 0 {
			X[i] = 0
		} else if X[i] > 1 {
			X[i] = 1
		}

		if Y[i] < 0 {
			Y[i] = 0
		} else if Y[i] > 1 {
			Y[i] = 1
		}
	}
	return
}

func (chart *Chart2D) AddAxis(color Color) {
	var (
		nSegs     = 11
		ticksize  = float32(0.015)
		x, y      = float32(0), float32(0)
		inc       = 1. / float32(nSegs-1)
		tickColor = ScaleColor(chart.AxisColor, 0.8)
		X         = make([]float32, 0)
		Y         = make([]float32, 0)
		C         = make([]float32, 0)
	)
	// Generate color array for 2 vertices per axis (X-axis and Y-axis)

	X, Y, C = AddSegmentToLine(X, Y, C, 0, 0, 1, 0, chart.AxisColor)
	X, Y, C = AddSegmentToLine(X, Y, C, 0, 0, 0, 1, chart.AxisColor)

	colorTxt := [3]float32{chart.AxisColor[0], chart.AxisColor[1], chart.AxisColor[2]}
	for i := 0; i < nSegs; i++ {
		X, Y, C = AddSegmentToLine(X, Y, C, x, y, x, y-ticksize, tickColor)
		chart.Screen.Printf(screen.NEW, x, y-2*ticksize, colorTxt, 0.35, true, false,
			"%4.1f", x)
		x = x + inc
	}
	x = 0.
	for i := 0; i < nSegs; i++ {
		X, Y, C = AddSegmentToLine(X, Y, C, x, y, x-ticksize, y, tickColor)
		chart.Screen.Printf(screen.NEW, x-3*ticksize, y, colorTxt, 0.35, true, false,
			"%4.1f", y)
		y = y + inc
	}
	chart.Screen.ChangePosition(0.0, 0.0)
	chart.Screen.AddLine(screen.NEW, X, Y, C) // 2 points, so 2 * 3 = 6 colors
}

func ScaleColor(color Color, scale float32) (scaled Color) {
	for i, f := range color {
		scaled[i] = f * scale
		if scaled[i] > 1. {
			scaled[i] = 1.
		}
	}
	return
}

func AddSegmentToLine(X, Y, C []float32, X1, Y1, X2, Y2 float32, color Color) (XX, YY, CC []float32) {
	XX = append(X, X1, X2)
	YY = append(Y, Y1, Y2)
	CC = append(C, color[0], color[1], color[2], color[0], color[1], color[2])
	return
}

func ColorArray(X, Y []float32, color Color) (colorArray []float32) {
	if len(X) != len(Y) {
		panic("X and Y must have the same length")
	}
	colorArray = make([]float32, 3*len(X))
	for i := 0; i < len(X); i++ {
		colorArray[i*3] = color[0]
		colorArray[i*3+1] = color[1]
		colorArray[i*3+2] = color[2]
	}
	return
}
