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

func (chart *Chart2D) AddAxis(color [3]float32) {
	// Generate color array for 2 vertices per axis (X-axis and Y-axis)
	axisColor := []float32{
		color[0], color[1], color[2], // Color for (XMin, YMin)
		color[0], color[1], color[2], // Color for (XMax, YMin)
		color[0], color[1], color[2], // Color for (XMin, YMin)
		color[0], color[1], color[2], // Color for (XMin, YMax)
	}

	xAxisVertices := []float32{0, 1, 0, 0}
	yAxisVertices := []float32{0, 0, 1, 0}
	chart.Screen.AddLine(screen.NEW, xAxisVertices, yAxisVertices, axisColor) // 2 points, so 2 * 3 = 6 colors
}
