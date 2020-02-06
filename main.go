package main

import (
	"fmt"
	"math"
	"time"

	"github.com/notargets/avs/chart2d"
)

var (
	shift = 0
	inc   = 0
	gt    = chart2d.XGlyph
)

func main() {
	fmt.Println("Hello")
	cc := chart2d.NewChart2D(1800, 1200, 0, 2*math.Pi, -1, 1)
	x, f := getFunc(1000, 1200, math.Sin)
	if err := cc.AddSeries("sin", x, f, 0, chart2d.Solid, chart2d.NewColor(1, 1, 1)); err != nil {
		panic(err)
	}
	go cc.Plot()
	var iters, growInc int
	growInc = 1
	for {
		iters++
		time.Sleep(16 * time.Millisecond)
		if inc%1200 == 0 {
			shift += growInc
			if shift%10 == 0 {
				gt = chart2d.GlyphType(shift / 10 % 5)
				if gt == 0 {
					gt = 1
				}
				fmt.Printf("10x reached, shift = %d, gt = %d\n", shift, gt)
				growInc += 1
			}
		}
		x, f = getFunc(1000, 1200, math.Sin)
		if err := cc.AddSeries("cos", x, f, 0, chart2d.Solid, chart2d.NewColor(0.7, 0.4, 0.7)); err != nil {
			panic(err)
		}
		if iters == 1000 {
			goto END
		}
	}
END:
	fmt.Println("Stopping Plot")
	cc.StopPlot()
	return
}

func getFunc(size, Ht int, ff func(float64) float64) (x, f []float32) {
	x = make([]float32, size)
	f = make([]float32, size)
	for i := 0; i < size; i++ {
		frac := float32(i) / float32(size-1)
		xc := frac * 2 * math.Pi
		frac = float32(shift+i) / float32(size-1)
		//yc := float32(math.Sin(float64(frac * 2 * math.Pi)))
		yc := float32(ff(float64(frac * 2 * math.Pi)))
		x[i], f[i] = xc, yc
	}
	inc += size
	return
}
