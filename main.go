package main

import (
	"fmt"
	"time"

	"github.com/notargets/avs/chart2d"
)

func main() {
	chart := chart2d.NewChart2D(1920, 1080, -10, 10, -5, 5) // World coordinates range from -10 to 10 in X, and -5 to 5 in Y
	window := chart.Init()

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

	chart.EventLoop(window)
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
