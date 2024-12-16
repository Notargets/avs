package chart2d

import (
	"github.com/notargets/avs/screen"
)

type Chart2D struct {
	Scale       float32
	Position    [2]float32
	XMin, XMax  float32
	YMin, YMax  float32
	Screen      *screen.Screen
	LineColor   Color
	ScreenColor Color
}

type Color [4]float32 // RGBA

func NewChart2D(XMin, XMax, YMin, YMax float32, width, height int) (chart *Chart2D) {
	chart = &Chart2D{
		XMin: XMin,
		XMax: XMax,
		YMin: YMin,
		YMax: YMax,
		//Screen:      screen.NewScreen(width, height, 0, 1, 0, 1),
		Screen:      screen.NewScreen(width, height, XMin, XMax, YMin, YMax, 0.95),
		LineColor:   Color{1, 1, 1, 1},
		ScreenColor: Color{0.18, 0.18, 0.18, 1.},
	}
	chart.Screen.SetBackgroundColor(chart.ScreenColor)
	return
}

func (chart *Chart2D) AddLine(X, Y []float32) {
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

func (chart *Chart2D) AddAxis(color Color, nSegs int) {
	var (
		xMin, xMax           = chart.XMin, chart.XMax
		yMin, yMax           = chart.YMin, chart.YMax
		xScale, yScale       = xMax - xMin, yMax - yMin
		xInc                 = xScale / float32(nSegs-1)
		yInc                 = yScale / float32(nSegs-1)
		xTickSize, yTickSize = 0.020 * xScale, 0.020 * yScale
		tickColor            = ScaleColor(color, 0.8)
		X                    = []float32{}
		Y                    = []float32{}
		C                    = []float32{}
	)
	if nSegs%2 == 0 {
		panic("nSegs must be odd")
	}
	// Generate color array for 2 vertices per axis (X-axis and Y-axis)

	X, Y, C = AddSegmentToLine(X, Y, C, xMin, 0, xMax, 0, color)
	X, Y, C = AddSegmentToLine(X, Y, C, 0, yMin, 0, yMax, color)

	colorTxt := [3]float32{color[0], color[1], color[2]}
	// Draw ticks along X axis
	var x, y = xMin, float32(0) // X axis is always drawn at Y = 0
	for i := 0; i < nSegs; i++ {
		if i == nSegs/2 {
			x = x + xInc
			continue
		}
		X, Y, C = AddSegmentToLine(X, Y, C, x, y, x, y-yTickSize, tickColor)
		chart.Screen.Printf(screen.NEW, x, y-2*yTickSize, colorTxt, xTickSize, true, false,
			"%4.1f", x)
		x = x + xInc
	}
	x = xMin + xScale/2.
	y = yMin
	for i := 0; i < nSegs; i++ {
		if i == nSegs/2 {
			y = y + yInc
			continue
		}
		X, Y, C = AddSegmentToLine(X, Y, C, x, y, x-xTickSize, y, tickColor)
		chart.Screen.Printf(screen.NEW, x-3*xTickSize, y, colorTxt, xTickSize, true, false,
			"%4.1f", y)
		y = y + yInc
	}
	//chart.Screen.ChangePosition(0.0, 0.0)
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
