package main

import (
    "fmt"
    "github.com/notargets/avs/chart2d"
    "math"
)

var (
    shift = 0
    inc   = 0
    gt    = chart2d.GlyphType(0)
)

func main() {
    fmt.Println("Hello")
    cc := chart2d.NewChart2D(1800, 1200, 0, 2*math.Pi, -1, 1)
    cc.SetGlyph(chart2d.XGlyph)
    x, f := getFunc(100, 1200)
    if err := cc.AddSeries("base", x, f); err != nil {
        panic(err)
    }
    cc.Plot()
}

func getFunc(size, Ht int) (x, f []float32){
    if inc%Ht == 0 {
        shift += 1
        if shift%10 == 0 {
            gt = chart2d.GlyphType(shift / 10 % 4)
            fmt.Printf("10x reached, shift = %d, gt = %d\n", shift, gt)
        }
    }
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