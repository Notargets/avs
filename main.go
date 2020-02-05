package main

import (
    "fmt"
    "github.com/notargets/avs/chart2d"
    "math"
)

func main() {
    fmt.Println("Hello")
    cc := chart2d.NewChart2D(1800, 1200, 0, 2*math.Pi, -1, 1)
    cc.Plot()
}
