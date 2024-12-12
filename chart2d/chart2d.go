package chart2d

import (
	"github.com/google/uuid"
	"github.com/notargets/avs/screen"
)

type Chart2D struct {
	Scale      float32
	Position   [2]float32
	XMin, XMax float32
	YMin, YMax float32
	Screen     *screen.Screen
}

func NewChart2D(XMin, XMax, YMin, YMax float32, width, height int) *Chart2D {
	return &Chart2D{
		XMin:   XMin,
		XMax:   XMax,
		YMin:   YMin,
		YMax:   YMax,
		Screen: screen.NewScreen(width, height, XMin, XMax, YMin, YMax),
	}
}

func (chart *Chart2D) AddAxis(color [3]float32) {
	// Generate color array for 2 vertices per axis (X-axis and Y-axis)
	axisColor := []float32{
		color[0], color[1], color[2], // Color for (XMin, YMin)
		color[0], color[1], color[2], // Color for (XMax, YMin)
		color[0], color[1], color[2], // Color for (XMin, YMin)
		color[0], color[1], color[2], // Color for (XMin, YMax)
	}

	Xmid := (chart.XMax-chart.XMin)*.5 + chart.XMin
	Ymid := (chart.YMax-chart.YMin)*.5 + chart.YMin
	// **X-Axis** (1 line from (XMin, YMin) to (XMax, YMin))
	xAxisVertices := []float32{
		chart.XMin, chart.XMax,
		Ymid, Ymid,
	}
	// **Y-Axis** (1 line from (XMin, YMin) to (XMin, YMax))
	yAxisVertices := []float32{
		Xmid, Xmid,
		chart.YMin, chart.YMax,
	}
	//	fmt.Println(xAxisVertices, yAxisVertices, axisColor)
	//os.Exit(1)
	chart.Screen.AddLine(uuid.Nil, xAxisVertices, yAxisVertices, axisColor) // 2 points, so 2 * 3 = 6 colors
}
