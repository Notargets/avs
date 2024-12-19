/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main

import (
	"image/color"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/chart2d"
)

// TODO: Alter the object management to add a top level map[WindowKey]map[ObjectKey]Renderable, where the Renderable is
// TODO: ... an Interface{} with the Methods: Add, Delete, Update, Show, Hide. The Add() will incorporate the ObjectKey
// TODO: ... into the object struct so that the Show/Hide functions can toggle the Visible in the Renderable
// TODO: ... implementation. This allows the event loop to query whether to draw or not before introspecting the object.
// TODO: ... The Delete() should cleanup any internal references, then delete the ObjectKey from the top level object
// TODO: ... map for the Window.
// TODO: Implement a map[WindowKey]Window such that windows can be created and separately managed. Create a "Default"
// TODO: ... Window at Scene creation time so that any Add() calls are put into the Default window context. If new
// TODO: ... windows are added to the Scene, the context within Scene's struct can be switched to a keyed windows and
// TODO: ... new Add() calls will be scoped to the "current" window. At some point, objects could be moved among
// TODO: ... windows.
// 012345678901234567890123456789012345678901234567890123456789012345678901234567
func main() {
	Test_Text()
}

func Test_Text() {
	width, height := 1920, 1080
	//width, height := 1000, 1000
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
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, width, height)
	tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{255, 255, 255, 255}, true, false)
	chart.AddAxis(color.RGBA{R: 255, G: 255, B: 255, A: 255}, tickText, 0, 11)

	DynamicText := assets.NewTextFormatter("NotoSans", "Regular", 36,
		color.RGBA{255, 0, 255, 255}, false, false)
	TitleText := assets.NewTextFormatter("NotoSans", "Bold", 64,
		color.RGBA{0, 255, 0, 255}, true, true)

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
	select {}
}
