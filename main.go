package main

import (
	"image/color"

	"github.com/notargets/avs/chart2d"
)

// TODO: Implement LINEKEY and TEXTKEY types with function pointers to return their base objects from the map
// TODO: ... LINEKEY, etc should be returned as object handles to enable UPDATE/DELETE/HIDE, etc
// TODO: Implement a "Destroy" window to enable multiple screen sessions within an app lifetime
//
// TODO: Fix world projection matrix to stretch world coordinates into the world bounds so that the left and right
// TODO: ... extrema are placed at the left and right boundaries of a windows when scale = 1. Right now, the world
// TODO: ... minX and maxX and Y coords are appearing well within the window boundaries. This is evident when the
// TODO: ... window aspect ratio is non unit
func main() {
	Test_Text()
}

func Test_Text() {
	width, height := 1920, 1080
	//width, height := 1000, 1000
	scale := float32(1.0)
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
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, scale, width, height)
	tickText := chart.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{255, 255, 255, 255}, true, false)
	chart.AddAxis(color.RGBA{R: 255, G: 255, B: 255, A: 255}, tickText, 0, 11)

	DynamicText := chart.NewTextFormatter("NotoSans", "Regular", 36,
		color.RGBA{255, 0, 255, 255}, false, false)
	TitleText := chart.NewTextFormatter("NotoSans", "Bold", 64,
		color.RGBA{0, 255, 0, 255}, true, true)

	xRange := chart.XMax - chart.XMin
	_ = xRange
	yRange := chart.YMax - chart.YMin
	xpos := float32(0)
	ypos := chart.YMin + 0.5*yRange
	chart.Printf(DynamicText, xpos, ypos, "This is text that moves with the screen objects")
	ypos = chart.YMin + 0.4*yRange
	chart.Printf(DynamicText, xpos, ypos, "Pan and zoom with right mouse and scroll wheel")
	ypos = chart.YMin + 1.05*yRange
	chart.Printf(TitleText, xpos, ypos, "This is an example of a title text string")
	ypos = chart.YMin + 1.00*yRange
	chart.Printf(TitleText, xpos, ypos, "Title text doesn't move with pan and zoom and remains the same size when window is resized")
	select {}
}
