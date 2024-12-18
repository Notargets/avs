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
	var XMin, XMax, YMin, YMax float32
	style := 4
	switch style {
	case 5:
		XMin, XMax, YMin, YMax = -100.0, 100.0, -100.0, 100.0
	case 4:
		XMin, XMax, YMin, YMax = -1.0, 1.0, -1.0, 1.0
	case 3:
		XMin, XMax, YMin, YMax = -5.0, -1.0, -1.0, 1.0
	case 0:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -10.0, 10.0
	case 1:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -20.0, 20.0
	case 2:
		XMin, XMax, YMin, YMax = -20.0, 20.0, -10.0, 10.0
	default:
		panic("No option here")
	}
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, 0.50, 1920, 1080)
	tickText := chart.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{255, 255, 255, 255}, true, false)
	chart.AddAxis(color.RGBA{R: 255, G: 255, B: 255, A: 255}, tickText, -1, 11)

	DynamicText := chart.NewTextFormatter("NotoSans", "Regular", 36,
		color.RGBA{255, 0, 255, 255}, false, false)
	TitleText := chart.NewTextFormatter("NotoSans", "Bold", 64,
		color.RGBA{0, 255, 0, 255}, true, true)

	chart.Printf(DynamicText, 0.0, 0.5, "This is text that moves with the screen objects")
	chart.Printf(DynamicText, 0.0, 0.4, "Pan and zoom with right mouse and scroll wheel")
	chart.Printf(TitleText, 0., 1.090, "This is an example of a title text string")
	chart.Printf(TitleText, 0.0, 0.985, "Title text doesn't move with pan and zoom and remains the same size when window is resized")
	select {}
}
