/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package chart2d

import (
	"image/color"

	"github.com/notargets/avs/screen/gl_thread_objects"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/screen"
)

type Chart2D struct {
	Scale        float32
	Position     [2]float32
	XMin, XMax   float32
	YMin, YMax   float32
	Screen       *screen.Screen
	WindowWidth  uint32
	WindowHeight uint32
	BGColor      color.RGBA
}

type Color [4]float32 // RGBA

func NewChart2D(XMin, XMax, YMin, YMax float32, width, height int,
	lineColor, bgColor color.RGBA, scaleOpt ...float32) (chart *Chart2D) {
	var scale float32
	if len(scaleOpt) == 0 {
		scale = 0.90 * float32(height) / float32(width)
	} else {
		scale = scaleOpt[0]
	}
	chart = &Chart2D{
		Scale:        scale,
		XMin:         XMin,
		XMax:         XMax,
		YMin:         YMin,
		YMax:         YMax,
		WindowWidth:  uint32(width),
		WindowHeight: uint32(height),
		BGColor:      bgColor,
	}
	chart.Screen = screen.NewScreen(uint32(width), uint32(height), XMin,
		XMax, YMin, YMax, scale, chart.BGColor, gl_thread_objects.AUTO)
	return
}

func (chart *Chart2D) AddLine(X, Y, Colors []float32,
	rt ...utils.RenderType) (key utils.Key) {
	return chart.Screen.NewLine(X, Y, Colors, rt...)
}
func (chart *Chart2D) Printf(formatter *assets.TextFormatter, x, y float32,
	format string, args ...interface{}) (key utils.Key) {
	return chart.Screen.Printf(formatter, x, y, format, args...)
}

func (chart *Chart2D) AddAxis(axisColor color.Color,
	tf *assets.TextFormatter, yAxisLocation float32, nSegs int) (key utils.Key) {
	var (
		xMin, xMax           = chart.XMin, chart.XMax
		yMin, yMax           = chart.YMin, chart.YMax
		xScale, yScale       = xMax - xMin, yMax - yMin
		xInc                 = xScale / float32(nSegs-1)
		yInc                 = yScale / float32(nSegs-1)
		xTickSize, yTickSize = 0.020 * xScale, 0.020 * yScale
		tickColor            = axisColor
		X                    = make([]float32, 0)
		Y                    = make([]float32, 0)
		C                    = make([]float32, 0)
	)
	if nSegs%2 == 0 {
		panic("nSegs must be odd")
	}

	// Generate color array for 2 vertices per axis (X-axis and Y-axis)
	X, Y, C = utils.AddSegmentToLine(X, Y, C, xMin, 0, xMax, 0, axisColor)
	X, Y, C = utils.AddSegmentToLine(X, Y, C, yAxisLocation, yMin, yAxisLocation,
		yMax, axisColor)

	// Draw ticks along X axis
	var x, y = xMin, float32(0) // X axis is always drawn at Y = 0
	for i := 0; i < nSegs; i++ {
		if x == yAxisLocation {
			x = x + xInc
			continue
		}
		X, Y, C = utils.AddSegmentToLine(X, Y, C, x, y, x, y-yTickSize, tickColor)
		x = utils.ClampNearZero(x, xScale/1000.)
		chart.Printf(tf, x, y-(chart.GetWorldSpaceCharHeight(tf)+yTickSize), "%4.1f", x)
		x = x + xInc
	}
	ptfY := *tf
	tfY := &ptfY
	tfY.Centered = false
	x = yAxisLocation
	y = yMin
	yTextDelta := utils.CalculateRightJustifiedTextOffset(yMin, chart.GetWorldSpaceCharWidth(tfY))
	for i := 0; i < nSegs; i++ {
		if i == nSegs/2 {
			y = y + yInc
			continue
		}
		X, Y, C = utils.AddSegmentToLine(X, Y, C, x, y, x-xTickSize, y, tickColor)
		y = utils.ClampNearZero(y, yScale/1000.)
		chart.Printf(tfY, x-yTextDelta, y, "%4.1f", y)
		y = y + yInc
	}
	// chart.Screen.ChangePosition(0.0, 0.0)
	return chart.AddLine(X, Y, C) // 2 points, so 2 * 3 = 6 colors
}

func (chart *Chart2D) NewWindow(title string, scale float32,
	position gl_thread_objects.Position) (win *gl_thread_objects.Window) {

	win = chart.Screen.NewWindow(chart.WindowWidth, chart.WindowHeight, chart.XMin,
		chart.XMax, chart.YMin, chart.YMax, scale, title,
		chart.BGColor, position)

	return
}

func (chart *Chart2D) GetWorldSpaceCharHeight(tf *assets.TextFormatter) (height float32) {
	return tf.GetWorldSpaceCharHeight(chart.YMax-chart.YMin, chart.WindowWidth, chart.WindowHeight)
}

func (chart *Chart2D) GetWorldSpaceCharWidth(tf *assets.TextFormatter) (height float32) {
	return tf.GetWorldSpaceCharWidth(chart.XMax-chart.XMin, chart.YMax-chart.YMin, chart.WindowWidth, chart.WindowHeight)
}

func (chart *Chart2D) SetDrawWindow(win *gl_thread_objects.Window) {
	chart.Screen.SetDrawWindow(win)
}

func (chart *Chart2D) GetCurrentWindow() (win *gl_thread_objects.Window) {
	win = chart.Screen.GetCurrentWindow()
	return
}
