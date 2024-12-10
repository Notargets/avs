package main

import (
	"fmt"
	"time"

	graphics2D "github.com/notargets/avs/geometry"

	"github.com/notargets/avs/chart2d"
)

func main() {

	//trimesh, xmin, xmax, ymin, ymax := makeExampleMesh()
	//chart := chart2d.NewChart2D(1920, 1080, xmin, xmax, ymin, ymax)
	//colorMap := utils.NewColorMap(0, 11, 1)
	//chart.AddTriMesh("mesh", trimesh, chart2d.TriangleGlyph, 0.01, chart2d.Solid, colorMap.GetRGB(0))

	chart := chart2d.NewChart2D(1920, 1080, -10, 10, -5, 5) // World coordinates range from -10 to 10 in X, and -5 to 5 in Y
	chart.Init()

	// Generate some data
	count := 0
	go func() {
		for {
			count++
			fmt.Println("New Data... Count = ", count)
			chart.DataChan <- chart2d.DataMsg{"triangle", chart2d.Series{
				Vertices: []float32{
					-2.5, -2.5, 1.0, 0.0, 0.0, // Vertex 1
					2.5, -2.5, 0.0, 1.0, 0.0, // Vertex 2
					0.0, 2.5, 0.0, 0.0, 1.0, // Vertex 3
				}}}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	chart.EventLoop()
}

func makeExampleMesh() (trimesh graphics2D.TriMesh, xmin, xmax, ymin, ymax float64) {
	xmin, xmax = -0.500, 2.500
	ymin, ymax = 0.000, 1.000
	var points = []graphics2D.Point{
		{X: [2]float32{0.00, 0.00}},
		{X: [2]float32{1.00, 0.00}},
		{X: [2]float32{2.00, 0.00}},
		{X: [2]float32{-0.50, 0.50}},
		{X: [2]float32{0.50, 0.50}},
		{X: [2]float32{1.50, 0.50}},
		{X: [2]float32{2.50, 0.50}},
		{X: [2]float32{0.00, 1.00}},
		{X: [2]float32{1.00, 1.00}},
		{X: [2]float32{2.00, 1.00}},
	}
	trimesh.Geometry = points

	var triangles = []graphics2D.Triangle{
		{Nodes: [3]int32{4, 3, 0}},
		{Nodes: [3]int32{1, 4, 0}},
		{Nodes: [3]int32{5, 4, 1}},
		{Nodes: [3]int32{5, 1, 2}},
		{Nodes: [3]int32{6, 5, 2}},
		{Nodes: [3]int32{7, 3, 4}},
		{Nodes: [3]int32{8, 7, 4}},
		{Nodes: [3]int32{8, 4, 5}},
		{Nodes: [3]int32{9, 8, 5}},
		{Nodes: [3]int32{9, 5, 6}},
	}
	trimesh.Triangles = triangles

	var data = [][]float32{
		{0.00, 0.00, 5.00},
		{0.00, 0.00, 0.00},
		{0.00, 0.00, 0.00},
		{0.00, 0.00, 0.00},
		{0.00, 5.00, 0.00},
		{0.00, 0.00, 0.00},
		{0.00, 0.00, 0.00},
		{5.00, 0.00, 0.00},
		{0.00, 0.00, 0.00},
		{0.00, 0.00, 0.00},
	}
	trimesh.Attributes = data
	return
}

//func old_main() {
//	cc := chart2d.NewChart2D_old(1920, 1080, 0, 1, 0, 1, 10)
//	col := utils.NewColorMap(0, 1, 1)
//	//ff := make([]float32, 50)
//	//go cc.Plot()
//	var x, f []float32
//	for i := 0; i < 6; i++ {
//		//x, f = getFunc(i+1, 0, 1, utils.GetLegendrePoly(i))
//		x, f = getFunc(100, -1, 1, utils.GetLegendrePoly(i))
//		/*
//			for i, val := range f {
//				ff[i] += val
//			}
//		*/
//		//x, f := getFunc(i+1, 0, 1, utils.GetLegendrePoly(i))
//		name := "L" + strconv.Itoa(i)
//		if err := cc.AddSeries(name, x, f, chart2d.GlyphType(i+1), 0.01, chart2d.Solid, col.GetRGB(float32(i)/5)); err != nil {
//			//if err := cc.AddSeries(name, x, f, 0, chart2d.Solid, col.GetRGB(float32(i)/5)); err != nil {
//			panic(err)
//		}
//	}
//	/*
//		if err := cc.AddSeries("sum", x, ff, chart2d.BoxGlyph, chart2d.Solid, col.GetRGB(0.5)); err != nil {
//			//if err := cc.AddSeries(name, x, f, 0, chart2d.Solid, col.GetRGB(float32(i)/5)); err != nil {
//			panic(err)
//		}
//	*/
//	reader := bufio.NewReader(os.Stdin)
//	_, _ = reader.ReadString('\n')
//	fmt.Println("Stopping Plot")
//	//cc.StopPlot()
//}

func getFunc(size int, xmin, xmax float64, ff func(float64) float64) (x, f []float32) {
	x = make([]float32, size)
	f = make([]float32, size)
	xr := xmax - xmin
	for i := 0; i < size; i++ {
		frac := float64(i) / float64(size-1)
		xc := frac*xr + xmin
		yc := float32(ff(xc))
		x[i], f[i] = float32(xc), yc
	}
	return
}
