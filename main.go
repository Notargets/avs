package main

import (
	"github.com/notargets/avs/chart2d"
	"github.com/notargets/avs/screen"
)

// TODO: Same with the AddString, we need to bury the call inside Chart2D to get the mapping
// TODO: Should be no direct calls to Screen
func main() {
	var XMin, XMax, YMin, YMax float32
	style := 4
	switch style {
	case 4:
		XMin, XMax, YMin, YMax = -1.0, 1.0, -1.0, 1.0
	case 0:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -10.0, 10.0
	case 1:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -20.0, 20.0
	case 2:
		XMin, XMax, YMin, YMax = -20.0, 20.0, -10.0, 10.0
	}
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, 1000, 1000)
	chart.Screen.LoadFont("assets/fonts/Noto-Sans/static/NotoSans-Regular.ttf", 12)
	//chart.AddAxis(chart2d.Color{1., 1., 1.}, 11)

	//chart.Screen.AddString(screen.NEW, "0123456789012345678901234567890", 0.5, 0.5, [3]float32{1, 1, 1}, true, false)
	chart.Screen.AddString(screen.NEW, "This is an example of a title text string", 0.0, 0.0, [3]float32{1, 0, 0}, true, false)
	select {}
}
