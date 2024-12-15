package main

import (
	"github.com/notargets/avs/chart2d"
)

// TODO: Change the chart2D code to implement mapping from Chart world coords to 0->1, 0->1
// TODO: Same with the AddString, we need to bury the call inside Chart2D to get the mapping
// TODO: Should be no direct calls to Screen
// TODO: Add centered text (easy)
// TODO: Add screen coordinate text that doesn't use the projection matrix
// TODO: Fix the text string sizing to pad the bottom to allow for g and p chars with hanging bottoms
func main() {
	chart := chart2d.NewChart2D(-10, 10, -20, 20, 1000, 1000)
	chart.AddAxis(chart2d.Color{1., 1., 1.}, 11)
	chart.Screen.LoadFont("assets/fonts/Noto-Sans/static/NotoSans-Regular.ttf", 64)
	//chart.Screen.AddString(screen.NEW, "0123456789012345678901234567890", 0.5, 0.5, [3]float32{1, 1, 1}, 10.0, true, false)
	//chart.Screen.AddString(screen.NEW, "This is an example of a title text string", 0.5, 1.0, [3]float32{1, 1, 0}, 10.0, true, true)
	select {}
}
