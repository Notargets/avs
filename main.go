package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/chart2d"
)

func main() {
	cc := chart2d.NewChart2D(1800, 1200, 0, 1, -1, 1)
	col := utils.NewColorMap(0, 1, 1)
	//ff := make([]float32, 50)
	var x, f []float32
	for i := 0; i < 6; i++ {
		x, f = getFunc(i+1, 0, 1, utils.GetLegendrePoly(i))
		/*
			for i, val := range f {
				ff[i] += val
			}
		*/
		//x, f := getFunc(i+1, 0, 1, utils.GetLegendrePoly(i))
		name := "L" + strconv.Itoa(i)
		if err := cc.AddSeries(name, x, f, chart2d.GlyphType(i+1), chart2d.Solid, col.GetRGB(float32(i)/5)); err != nil {
			//if err := cc.AddSeries(name, x, f, 0, chart2d.Solid, col.GetRGB(float32(i)/5)); err != nil {
			panic(err)
		}
	}
	/*
		if err := cc.AddSeries("sum", x, ff, chart2d.BoxGlyph, chart2d.Solid, col.GetRGB(0.5)); err != nil {
			//if err := cc.AddSeries(name, x, f, 0, chart2d.Solid, col.GetRGB(float32(i)/5)); err != nil {
			panic(err)
		}
	*/
	go cc.Plot()
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	fmt.Println("Stopping Plot")
	cc.StopPlot()
	return
}

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
