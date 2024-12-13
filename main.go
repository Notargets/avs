package main

import (
	"github.com/notargets/avs/chart2d"
)

func main() {

	chart := chart2d.NewChart2D(-10, 10, -20, 20, 1000, 1000)
	chart.AddAxis(chart2d.Color{1., 1., 1.})
	select {}

}
