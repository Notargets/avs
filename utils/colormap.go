package utils

import (
    "fmt"
    graphics2D "github.com/notargets/avs/geometry"
    "image/color"
    "math"
)

type ColorMap struct {
    MinVal, MaxVal float32
    AlphaVal       uint8
}

func NewColorMap(MinVal, MaxVal float32, AlphaVal float32) (cm *ColorMap) {
    return &ColorMap{MinVal, MaxVal, uint8(AlphaVal * 255)}
}

func (cm *ColorMap) GetRGBString(val float32) (rgbaStr string) {
    rgba := cm.GetRGB(val)
    rgbaStr = fmt.Sprintf("rgba(%d,%d,%d,%0.2f)",
        rgba.R, rgba.G, rgba.B, float32(rgba.A)/255.)
    return rgbaStr
}
func (cm *ColorMap) GetRGB(val float32) (rgba color.RGBA) {
    // From: www.particleincell.com/2014/colormap/
    // Normalize
    nval := (val - cm.MinVal) / (cm.MaxVal - cm.MinVal)
    a := float64((1 - nval) * 4)
    X := math.Floor(a)
    Y := uint8(math.Floor(255 * (a - X)))
    switch int(X) {
    case 0:
        return color.RGBA{255, Y, 0, cm.AlphaVal}
    case 1:
        return color.RGBA{255 - Y, 255, 0, cm.AlphaVal}
    case 2:
        return color.RGBA{0, 255, Y, cm.AlphaVal}
    case 3:
        return color.RGBA{0, 255 - Y, 255, cm.AlphaVal}
    case 4:
        return color.RGBA{0, 0, 255, cm.AlphaVal}
    }
    return color.RGBA{}
}
func (cm *ColorMap) GetDrawing(box *graphics2D.BoundingBox) (qm *graphics2D.QuadMesh) {
    width, height := box.XMax[0]-box.XMin[0], box.XMax[1]-box.XMin[1]
    /*
    	Produce a quadmesh that displays the color values left/right, min/max
    */
    qm = &graphics2D.QuadMesh{}
    qm.Dimensions = [2]int{100, 2}
    size := qm.Dimensions[1] * qm.Dimensions[0]
    qm.Geometry = make([]graphics2D.Point, size)
    qm.Attributes = make([][]float32, 1)
    qm.Attributes[0] = make([]float32, size)

    sInc := (cm.MaxVal - cm.MinVal) / float32(qm.Dimensions[0]-1)
    xInc := width / float32(qm.Dimensions[0]-1)
    yInc := height / float32(qm.Dimensions[1]-1)
    Y := box.XMin[1]
    for j := 0; j < qm.Dimensions[1]; j++ {
        X := box.XMin[0]
        S := cm.MinVal
        for i := 0; i < qm.Dimensions[0]; i++ {
            //qm.Geometry[i+qm.Dimensions[0]*j] = Point{[2]float32{X, Y}}
            qm.Geometry[i+qm.Dimensions[0]*j] = *graphics2D.NewPoint(X, Y)
            qm.Attributes[0][i+qm.Dimensions[0]*j] = S
            X += xInc
            S += sInc
        }
        Y += yInc
    }
    qm.Box = graphics2D.NewBoundingBox(qm.Geometry)
    return qm
}
