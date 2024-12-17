package main

import (
	"image/color"

	"github.com/notargets/avs/chart2d"
)

// TODO: Same with the AddString, we need to bury the call inside Chart2D to get the mapping
// TODO: Should be no direct calls to Screen
func main() {
	Test1()
}

func Test1() {

	//TODO: Re-calculate the stretch ratio for FIXEDSTRING text types to make Resize not stretch the text poly
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
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, 0.87, 1000, 1000)
	chart.AddAxis(color.RGBA{R: 255, G: 255, B: 255, A: 255}, -1, 11)

	titleText1 := chart.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{255, 0, 255, 255}, true, false)
	titleText2 := chart.NewTextFormatter("NotoSans", "Bold", 36,
		color.RGBA{0, 255, 0, 255}, true, true)

	chart.Printf(titleText1, 0.5, 0.5, "This is text that moves")
	chart.Printf(titleText2, 0., 1.075, "This is an example of a title text string")
	select {}
}
