package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/notargets/avs/screen"
)

func main() {
	//chart := screen.NewScreen(800, 600, 0, 10, -5, 5)
	chart := screen.NewScreen(800, 600, -10, 10, -5, 5)
	chart.SetBackgroundColor(0.1, 0.1, 0.1, 1.0)
	X := []float32{0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}
	Y := []float32{0.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, -1.0, 0.0, 1.0}

	linekey := chart.AddLine(uuid.Nil, X, Y, nil)
	_ = linekey

	//select {}
	i := 0
	for {
		time.Sleep(1 * time.Second)
		if i%2 == 0 {
			chart.SetBackgroundColor(0.1, 0.1, 0.1, 1.0)
		} else {
			chart.SetBackgroundColor(0.4, 0.4, 0.4, 1.0)
		}
		i++
	}
}

//	active := true
//	for i := 0; ; i++ {
//		if i%2 == 0 {
//			chart.SetBackgroundColor(0.1, 0.1, 0.1, 1.0)
//		} else {
//			chart.SetBackgroundColor(0.3, 0.3, 0.3, 1.0)
//		}
//
//		if i%3 == 1 {
//			active = !active
//			chart.SetObjectActive(lineKey, active) // Hide the line every 20 frames
//		}
//
//		time.Sleep(time.Second)
//	}
//}

//func makeExampleMesh() (trimesh graphics2D.TriMesh, xmin, xmax, ymin, ymax float64) {
//	xmin, xmax = -0.500, 2.500
//	ymin, ymax = 0.000, 1.000
//	var points = []graphics2D.Point{
//		{X: [2]float32{0.00, 0.00}},
//		{X: [2]float32{1.00, 0.00}},
//		{X: [2]float32{2.00, 0.00}},
//		{X: [2]float32{-0.50, 0.50}},
//		{X: [2]float32{0.50, 0.50}},
//		{X: [2]float32{1.50, 0.50}},
//		{X: [2]float32{2.50, 0.50}},
//		{X: [2]float32{0.00, 1.00}},
//		{X: [2]float32{1.00, 1.00}},
//		{X: [2]float32{2.00, 1.00}},
//	}
//	trimesh.Geometry = points
//
//	var triangles = []graphics2D.Triangle{
//		{Nodes: [3]int32{4, 3, 0}},
//		{Nodes: [3]int32{1, 4, 0}},
//		{Nodes: [3]int32{5, 4, 1}},
//		{Nodes: [3]int32{5, 1, 2}},
//		{Nodes: [3]int32{6, 5, 2}},
//		{Nodes: [3]int32{7, 3, 4}},
//		{Nodes: [3]int32{8, 7, 4}},
//		{Nodes: [3]int32{8, 4, 5}},
//		{Nodes: [3]int32{9, 8, 5}},
//		{Nodes: [3]int32{9, 5, 6}},
//	}
//	trimesh.Triangles = triangles
//
//	var data = [][]float32{
//		{0.00, 0.00, 5.00},
//		{0.00, 0.00, 0.00},
//		{0.00, 0.00, 0.00},
//		{0.00, 0.00, 0.00},
//		{0.00, 5.00, 0.00},
//		{0.00, 0.00, 0.00},
//		{0.00, 0.00, 0.00},
//		{5.00, 0.00, 0.00},
//		{0.00, 0.00, 0.00},
//		{0.00, 0.00, 0.00},
//	}
//	trimesh.Attributes = data
//	return
// }
