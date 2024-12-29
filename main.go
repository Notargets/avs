/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main

import (
	"image/color"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/notargets/avs/screen"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/chart2d"
)

// TODO: Alter the object management to add a top level map[WindowKey]map[ObjectKey]Renderable, where the Renderable is
// TODO: ... an Interface{} with the Methods: add, Delete, setupVertices, Show, Hide. The add() will incorporate the ObjectKey
// TODO: ... into the object struct so that the Show/Hide functions can toggle the Visible in the Renderable
// TODO: ... implementation. This allows the event loop to query whether to draw or not before introspecting the object.
// TODO: ... The Delete() should cleanup any internal references, then delete the ObjectKey from the top level object
// TODO: ... map for the window.
func main() {
	// chart := TestText()
	// Test2(chart)
	// TestFunctionPlot(chart)
	// Start pprof for profiling (port can be any free port)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	chart := chart2d.NewChart2D(0, 1, -1, 1, 1920, 1080,
		utils.WHITE, // Line Color Default
		utils.DARK)  // BG color Default

	TestFunctionPlot(chart)
	select {}
}

// TODO: Implement object sub-groups within ObjectGroup to enable
// TODO: ... nested objects - e.g. axis is a collection of text objs + line
// TODO: !!! Find the memory leak in the win.Redraw() path. Redrawing the same
// TODO: ... objects is leaking memory
func TestFunctionPlot(chart *chart2d.Chart2D) {
	// win := chart.Screen.NewWindow(chart.WindowWidth, chart.WindowHeight,
	// 	0, 1, -1, 1, 0.5, "Sin Function",
	// 	utils.DARK, screen.AUTO)

	win := chart.GetCurrentWindow()

	tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		utils.WHITE, true, false)

	// chart.AddAxis(utils.WHITE, tickText, "X", "Y", 0, 0, 11)

	// Make a Sin function for plotting
	X := make([]float32, 100)
	Y := make([]float32, 100)
	var (
		linekey          utils.Key
		TwoPi            = float32(2. * math.Pi)
		x, xInc, t, tInc float32
		iter             int
	)
	t = 0
	tInc = 0.05
	xInc = 1. / 100.
	for {
		if iter == 0 {
			chart.Printf(tickText, 0, 0, "Hello World 1")
			// chart.Printf(tickText, 0, 0.1, "Hello World 2")
			// chart.Printf(tickText, 0, 0.2, "Hello World 3")
			// chart.Printf(tickText, 0, 0.3, "Hello World 4")
			// chart.Printf(tickText, 0, 0.4, "Hello World 5")
			_ = tickText
			// chart.AddLine([]float32{0, 1}, []float32{0, 1}, utils.BLUE)
			x = 0
			for i := 0; i < 100; i++ {
				X[i] = x
				Y[i] = float32(math.Sin(float64(x*TwoPi - t)))
				x += xInc
			}
			// linekey = chart.AddLine(X, Y, utils.BLUE, utils.POLYLINE)
		} else {
			chart.Screen.Redraw(win)
			// chart.UpdateLine(win, linekey, X, Y, nil)
		}
		time.Sleep(time.Millisecond * 50)
		t += tInc
		iter++
		// if iter > 1 {
		// 	break
		// }
		_ = linekey
	}
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
		utils.WHITE, // Line Color Default
		utils.DARK)  // BG color Default

	tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		utils.WHITE, true, false)
	chart.AddAxis(utils.WHITE, tickText, "X", "Y", 0, 0, 11)

	DynamicText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		utils.RED, false, false)
	TitleText := assets.NewTextFormatter("NotoSans", "Bold", 36,
		utils.GREEN, true, true)

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
	// add a 33% pad for the vertical line spacing between lines
	ypos = ypos - 1.33*titleHeight
	chart.Printf(TitleText, xpos, ypos, "Title text doesn't move with pan and zoom and remains the same size when window is resized")

	return
}

func Test2(chart *chart2d.Chart2D) {

	win1 := chart.GetCurrentWindow()

	win2 := chart.NewWindow("Second Window", 0.8*chart.Scale,
		screen.AUTO)

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
	// add a 33% pad for the vertical line spacing between lines
	ypos = ypos - titleHeight
	chart.Printf(TitleText, 0, ypos, "Title 2 second line")

	// Draw in first window
	chart.SetDrawWindow(win1)

	chart.Printf(TitleText, 0, ypos, "Title 3 First window")

	X, Y := utils.AddSegmentToLine([]float32{}, []float32{},
		chart.XMin+0.25*xRange, chart.YMin+0.75*yRange,
		chart.XMin+0.5*xRange, chart.YMin+0.75*yRange)

	chart.AddLine(X, Y, utils.RED)

	// Draw in second window
	chart.SetDrawWindow(win2)

	chart.Printf(TitleText, 0, ypos-0.3*yRange, "Title 4 Second window")

	X, Y = utils.AddSegmentToLine([]float32{}, []float32{},
		chart.XMin+0.25*xRange, chart.YMin+0.75*yRange,
		chart.XMin+0.5*xRange, chart.YMin+0.75*yRange)
	chart.AddLine(X, Y, utils.GREEN)

	_, _ = win1, win2

}
