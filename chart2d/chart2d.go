/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package chart2d

import (
	"github.com/notargets/avs/geometry"
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
	BGColor      interface{} // One of [4]float32, [3]float32, color.RGBA
}

type Color [4]float32 // RGBA

func NewChart2D(XMin, XMax, YMin, YMax float32, width, height int,
	lineColor, bgColor interface{}, scaleOpt ...float32) (chart *Chart2D) {
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
		XMax, YMin, YMax, scale, chart.BGColor, screen.AUTO)
	return
}

func (chart *Chart2D) AddLine(XY []float32, LineColor interface{},
	rt ...utils.RenderType) (key utils.Key) {
	return chart.Screen.NewLine(XY, LineColor, rt...)
}

func (chart *Chart2D) UpdateLine(win *screen.Window, key utils.Key,
	XY, Colors []float32) {
	chart.Screen.UpdateLine(win, key, XY, Colors)
	return
}

func (chart *Chart2D) Printf(formatter *assets.TextFormatter, x, y float32,
	format string, args ...interface{}) (key utils.Key) {
	return chart.Screen.Printf(formatter, x, y, format, args...)
}

func (chart *Chart2D) AddAxis(axisColor interface{}, tf *assets.TextFormatter,
	XLabel, YLabel string, xCoordOfYAxis, yCoordOfXAxis float32, nSegs int) (key utils.Key) {

	win := chart.Screen.GetCurrentWindow()
	key = chart.Screen.NewAxis(win, axisColor, tf, XLabel, YLabel, xCoordOfYAxis,
		yCoordOfXAxis, nSegs)
	return
}

func (chart *Chart2D) AddTriMesh(mesh geometry.TriMesh) (key utils.Key) {
	win := chart.Screen.GetCurrentWindow()
	key = chart.Screen.NewTriMesh(win, mesh)
	return
}

func (chart *Chart2D) AddShadedVertexScalar(vs *geometry.VertexScalar, fMin,
	fMax float32) (key utils.Key) {
	key = chart.Screen.NewShadedVertexScalar(vs, fMin, fMax)
	return
}

func (chart *Chart2D) UpdateShadedVertexScalar(win *screen.Window, key utils.Key,
	vs *geometry.VertexScalar) {
	chart.Screen.UpdateShadedVertexScalar(win, key, vs)
	return
}

func (chart *Chart2D) AddContourVertexScalar(vs *geometry.VertexScalar, fMin,
	fMax float32, numContours int) (key utils.Key) {
	key = chart.Screen.NewContourVertexScalar(vs, fMin, fMax, numContours)
	return
}

func (chart *Chart2D) UpdateContourVertexScalar(win *screen.Window,
	key utils.Key, vs *geometry.VertexScalar) {
	chart.Screen.UpdateContourVertexScalar(win, key, vs)
	return
}

func (chart *Chart2D) NewWindow(title string, scale float32,
	position screen.Position) (win *screen.Window) {

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

func (chart *Chart2D) SetDrawWindow(win *screen.Window) {
	chart.Screen.SetDrawWindow(win)
}

func (chart *Chart2D) GetCurrentWindow() (win *screen.Window) {
	if chart == nil {
		panic("chart is nil")
	}
	if chart.Screen == nil {
		panic("Screen is nil")
	}
	win = chart.Screen.GetCurrentWindow()
	return
}
