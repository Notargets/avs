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
	x, f := getFunc(100, 1200)
	if err := cc.AddSeries("base", x, f, gt, chart2d.Solid); err != nil {
		panic(err)
	}
	go cc.Plot()
	for {
		time.Sleep(16 * time.Millisecond)
		if inc%1200 == 0 {
			shift += 1
			if shift%10 == 0 {
				gt = chart2d.GlyphType(shift / 10 % 5)
				if gt == 0 {
					gt = 1
				}
				fmt.Printf("10x reached, shift = %d, gt = %d\n", shift, gt)
			}
		}
		x, f = getFunc(100, 1200)
		if err := cc.AddSeries("base", x, f, gt, chart2d.Solid); err != nil {
			panic(err)
		}
	}
}

func getFunc(size, Ht int) (x, f []float32) {
	x = make([]float32, size)
	f = make([]float32, size)
	for i := 0; i < size; i++ {
		frac := float32(i) / float32(size-1)
		xc := frac * 2 * math.Pi
		frac = float32(shift+i) / float32(size-1)
		yc := float32(math.Sin(float64(frac * 2 * math.Pi)))
		x[i], f[i] = xc, yc
	}
	inc += size
	return
}
