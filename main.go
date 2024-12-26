/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main

import (
	"image/color"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/screen/gl_thread_objects"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/chart2d"
)

// TODO: Alter the object management to add a top level map[WindowKey]map[ObjectKey]Renderable, where the Renderable is
// TODO: ... an Interface{} with the Methods: Add, Delete, setupVertices, Show, Hide. The Add() will incorporate the ObjectKey
// TODO: ... into the object struct so that the Show/Hide functions can toggle the Visible in the Renderable
// TODO: ... implementation. This allows the event loop to query whether to draw or not before introspecting the object.
// TODO: ... The Delete() should cleanup any internal references, then delete the ObjectKey from the top level object
// TODO: ... map for the window.
// TODO: Implement a map[WindowKey]window such that windows can be created and separately managed. Create a "Default"
// TODO: ... window at Scene creation time so that any Add() calls are put into the Default window context. If new
// TODO: ... windows are added to the Scene, the context within Scene's struct can be switched to a keyed windows and
// TODO: ... new Add() calls will be scoped to the "current" window. At some point, objects could be moved among
// TODO: ... windows.
func main() {
	chart := TestText()
	Test2(chart)
	TestFunctionPlot(chart)
	select {}
}

func TestFunctionPlot(chart *chart2d.Chart2D) {
	// win := chart.NewWindow("Sin function", 0.9, gl_thread_objects.AUTO)
	win := chart.Screen.NewWindow(chart.WindowWidth, chart.WindowHeight,
		chart.XMin, chart.XMax, chart.YMin, chart.YMax, 0.9, "Sin Function",
		color.RGBA{46, 46, 46, 255}, gl_thread_objects.AUTO)

	tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{R: 255, G: 255, B: 255, A: 255}, true, false)
	chart.AddAxis(color.RGBA{R: 255, G: 255, B: 255, A: 255}, tickText, 0, 11)

	_ = win

}

func TestText() (chart *chart2d.Chart2D) {
	width, height := 1200, 760
	// width, height := 1000, 1000
	var XMin, XMax, YMin, YMax float32
	style := 2
	switch style {
	case 0:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -10.0, 10.0
	case 1:
		XMin, XMax, YMin, YMax = -20.0, 20.0, -10.0, 10.0
	case 2:
		XMin, XMax, YMin, YMax = -100.0, 100.0, -100.0, 100.0
	case 3:
		XMin, XMax, YMin, YMax = -5.0, -1.0, -1.0, 1.0
	case 4:
		XMin, XMax, YMin, YMax = -1.0, 1.0, -1.0, 1.0
	case 5:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -20.0, 20.0
	case 6:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -100.0, 100.0
	default:
		panic("No option here")
	}

	chart = chart2d.NewChart2D(XMin, XMax, YMin, YMax, width, height,
		color.RGBA{255, 255, 255, 255}, // Line Color Default
		color.RGBA{46, 46, 46, 255})    // BG color Default

	tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{R: 255, G: 255, B: 255, A: 255}, true, false)
	chart.AddAxis(color.RGBA{R: 255, G: 255, B: 255, A: 255}, tickText, 0, 11)

	DynamicText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{R: 255, B: 255, A: 255}, false, false)
	TitleText := assets.NewTextFormatter("NotoSans", "Bold", 36,
		color.RGBA{G: 255, A: 255}, true, true)

	titleHeight := chart.GetWorldSpaceCharHeight(TitleText)

	xRange := chart.XMax - chart.XMin
	_ = xRange
	yRange := chart.YMax - chart.YMin
	xpos := float32(0)
	ypos := chart.YMin + 0.5*yRange
	chart.Printf(DynamicText, xpos, ypos, "This is text that moves with the screen objects")
	ypos = chart.YMin + 0.4*yRange
	chart.Printf(DynamicText, xpos, ypos, "Pan and zoom with right mouse and scroll wheel")

	// Title
	ypos = 1.1*chart.YMax - titleHeight
	chart.Printf(TitleText, xpos, ypos, "This is an example of a title text string")
	// Add a 33% pad for the vertical line spacing between lines
	ypos = ypos - 1.33*titleHeight
	chart.Printf(TitleText, xpos, ypos, "Title text doesn't move with pan and zoom and remains the same size when window is resized")

	return
}

func Test2(chart *chart2d.Chart2D) {

	win1 := chart.GetCurrentWindow()

	win2 := chart.NewWindow("Second Window", 0.8*chart.Scale,
		gl_thread_objects.AUTO)

	chart.SetDrawWindow(win2)
	// Test text
	DynamicText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{R: 255, B: 255, A: 255}, false, false)
	xRange := chart.XMax - chart.XMin
	_ = xRange
	yRange := chart.YMax - chart.YMin
	xpos := float32(0)
	ypos := chart.YMin + 0.5*yRange
	chart.Printf(DynamicText, xpos, ypos, "window 2 Dynamic Text")

	// Title
	TitleText := assets.NewTextFormatter("NotoSans", "Bold", 36,
		color.RGBA{G: 255, A: 255}, true, true)

	titleHeight := chart.GetWorldSpaceCharHeight(TitleText)
	ypos = 0.6*chart.YMax - titleHeight
	chart.Printf(TitleText, 0, ypos,
		"Title 2 first line")
	// Add a 33% pad for the vertical line spacing between lines
	ypos = ypos - titleHeight
	chart.Printf(TitleText, 0, ypos, "Title 2 second line")

	// Draw in first window
	chart.SetDrawWindow(win1)

	chart.Printf(TitleText, 0, ypos, "Title 3 First window")

	X, Y, C := utils.AddSegmentToLine([]float32{}, []float32{}, []float32{},
		chart.XMin+0.25*xRange, chart.YMin+0.75*yRange,
		chart.XMin+0.5*xRange, chart.YMin+0.75*yRange,
		color.RGBA{255, 0, 0, 255})
	chart.AddLine(X, Y, C)

	// Draw in second window
	chart.SetDrawWindow(win2)

	chart.Printf(TitleText, 0, ypos-0.3*yRange, "Title 4 Second window")

	X, Y, C = utils.AddSegmentToLine([]float32{}, []float32{}, []float32{},
		chart.XMin+0.25*xRange, chart.YMin+0.75*yRange,
		chart.XMin+0.5*xRange, chart.YMin+0.75*yRange,
		color.RGBA{0, 255, 0, 255})
	chart.AddLine(X, Y, C)

	_, _ = win1, win2

}
