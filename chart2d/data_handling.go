package chart2d

import (
	"image/color"

	"github.com/notargets/avs/functions"
	graphics2D "github.com/notargets/avs/geometry"
)

type Series struct {
	Vertices  []float32
	Xdata     []float32
	Ydata     []float32
	TriMesh   *graphics2D.TriMesh
	Surface   *functions.FSurface
	Vectors   [][2]float64
	Glyph     GlyphType
	GlyphSize float32
	Linetype  LineType
	Color     *color.RGBA
}

type DataMsg struct {
	Name string
	Data Series
}

type GlyphType uint8

const (
	NoGlyph GlyphType = iota
	CircleGlyph
	XGlyph
	CrossGlyph
	StarGlyph
	BoxGlyph
	TriangleGlyph
)

const (
	NoLine LineType = iota
	Solid
	Dashed
)

func (cc *Chart2D) AddTriMesh(name string, Tris graphics2D.TriMesh, Glyph GlyphType, GlyphSize float32, lt LineType, co color.RGBA) (err error) {
	s := Series{
		TriMesh:   &Tris,
		Glyph:     Glyph,
		GlyphSize: GlyphSize,
		Linetype:  lt,
		Color:     &co,
	}
	cc.DataChan <- DataMsg{name, s}
	return
}

//func (cc *Chart2D_old) AddVectors(name string, Geom []graphics2D.Point, vectors [][2]float64, lt LineType, co color.RGBA) (err error) {
//	var (
//		x, y = make([]float32, len(Geom)), make([]float32, len(Geom))
//	)
//	for i, pt := range Geom {
//		x[i] = pt.X[0]
//		y[i] = pt.X[1]
//	}
//	s := Series{
//		Xdata:    x,
//		Ydata:    y,
//		Vectors:  vectors,
//		Linetype: lt,
//		Color:    &co,
//	}
//	cc.inputChan <- &DataMsg{name, s}
//	return
//}
//
//func (cc *Chart2D_old) AddTriMesh(name string, Tris graphics2D.TriMesh, Glyph GlyphType, GlyphSize float32, lt LineType, co color.RGBA) (err error) {
//	s := Series{
//		TriMesh:   &Tris,
//		Glyph:     Glyph,
//		GlyphSize: GlyphSize,
//		Linetype:  lt,
//		Color:     &co,
//	}
//	cc.inputChan <- &DataMsg{name, s}
//	return
//}
//
//func (cc *Chart2D_old) AddFunctionSurface(name string, fs functions.FSurface, lineType LineType, lineColor color.RGBA) (err error) {
//	s := Series{
//		Surface:  &fs,
//		Linetype: lineType,
//		Color:    &lineColor,
//	}
//	cc.inputChan <- &DataMsg{name, s}
//	return
//}
//
//func (cc *Chart2D_old) AddColorMap(cm *utils.ColorMap) {
//	// Used to render data attributes
//	cc.colormap = cm
//}
//
//func (cc *Chart2D_old) AddSeries(name string, xI interface{}, fI interface{}, glyphType GlyphType, glyphSize float32, lineType LineType, co color.RGBA) (err error) {
//	var (
//		x, f []float32
//	)
//	switch xI.(type) {
//	case []float32:
//		x = xI.([]float32)
//		f = fI.([]float32)
//	case []float64:
//		x = toFloat32Slice(xI.([]float64))
//		f = toFloat32Slice(fI.([]float64))
//	}
//	switch {
//	case len(name) == 0 || len(f) == 0 || len(x) == 0:
//		return fmt.Errorf("empty series")
//	case len(x) != len(f):
//		return fmt.Errorf("length of x data not equal to function data length")
//	}
//	s := Series{
//		Xdata:     x,
//		Ydata:     f,
//		Glyph:     glyphType,
//		GlyphSize: glyphSize,
//		Linetype:  lineType,
//		Color:     &co,
//	}
//	cc.inputChan <- &DataMsg{name, s}
//	return
//}

func toFloat32Slice(A []float64) (R []float32) {
	R = make([]float32, len(A))
	for i, val := range A {
		R[i] = float32(val)
	}
	return R
}
