package chart2d

import (
	"fmt"
	"image/color"
	_ "image/png"

	"github.com/notargets/avs/functions"

	"github.com/notargets/avs/utils"

	graphics2D "github.com/notargets/avs/geometry"
)

func init() {
	// GLFW event handling must run on the main OS thread
	//runtime.LockOSThread()
}

type Series struct {
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

type NewDataMsg struct {
	Name string
	Data Series
}

type Chart2D struct {
	Sc           *Screen
	RmX, RmY     *RangeMap
	activeSeries map[string]Series
	inputChan    chan *NewDataMsg
	stopChan     chan struct{}
	colormap     *utils.ColorMap
}

func NewChart2D(w, h int, xmin, xmax, ymin, ymax float32, chanDepth ...int) (cc *Chart2D) {
	cc = &Chart2D{}
	cc.Sc = NewScreen(w, h)
	cc.RmX = NewRangeMap(xmin, xmax, 0, 1)
	cc.RmY = NewRangeMap(ymin, ymax, 0, 1)
	cc.activeSeries = make(map[string]Series)
	if len(chanDepth) != 0 {
		cc.inputChan = make(chan *NewDataMsg, chanDepth[0])
	} else {
		cc.inputChan = make(chan *NewDataMsg, 1)
	}
	cc.stopChan = make(chan struct{})
	return
}

func (cc *Chart2D) StopPlot() {
	cc.stopChan <- struct{}{}
}

func (cc *Chart2D) AddVectors(name string, Geom []graphics2D.Point, vectors [][2]float64, lt LineType, co color.RGBA) (err error) {
	var (
		x, y = make([]float32, len(Geom)), make([]float32, len(Geom))
	)
	for i, pt := range Geom {
		x[i] = pt.X[0]
		y[i] = pt.X[1]
	}
	s := Series{
		Xdata:    x,
		Ydata:    y,
		Vectors:  vectors,
		Linetype: lt,
		Color:    &co,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) AddTriMesh(name string, Tris graphics2D.TriMesh, Glyph GlyphType, GlyphSize float32, lt LineType, co color.RGBA) (err error) {
	s := Series{
		TriMesh:   &Tris,
		Glyph:     Glyph,
		GlyphSize: GlyphSize,
		Linetype:  lt,
		Color:     &co,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) AddFunctionSurface(name string, fs functions.FSurface, lineType LineType, lineColor color.RGBA) (err error) {
	s := Series{
		Surface:  &fs,
		Linetype: lineType,
		Color:    &lineColor,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) AddColorMap(cm *utils.ColorMap) {
	// Used to render data attributes
	cc.colormap = cm
}

func (cc *Chart2D) AddSeries(name string, xI interface{}, fI interface{}, glyphType GlyphType, glyphSize float32, lineType LineType, co color.RGBA) (err error) {
	var (
		x, f []float32
	)
	switch xI.(type) {
	case []float32:
		x = xI.([]float32)
		f = fI.([]float32)
	case []float64:
		x = ToFloat32Slice(xI.([]float64))
		f = ToFloat32Slice(fI.([]float64))
	}
	switch {
	case len(name) == 0 || len(f) == 0 || len(x) == 0:
		return fmt.Errorf("empty series")
	case len(x) != len(f):
		return fmt.Errorf("length of x data not equal to function data length")
	}
	s := Series{
		Xdata:     x,
		Ydata:     f,
		Glyph:     glyphType,
		GlyphSize: glyphSize,
		Linetype:  lineType,
		Color:     &co,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) processNewData() {
	for i := 0; i < len(cc.inputChan); i++ {
		msg := <-cc.inputChan
		cc.activeSeries[msg.Name] = msg.Data
	}
}

type Screen struct {
	Width, Height int
	Ratio         float32
}

func NewScreen(w, h int) *Screen {
	return &Screen{
		Width:  w,
		Height: h,
		Ratio:  float32(h) / float32(w),
	}
}

func (sc *Screen) GetRatio() (rat float32) {
	return sc.Ratio
}

type RangeMap struct {
	xMin, xMax float32
	pMin, pMax float32
}

func NewRangeMap(xMin, xMax float32, pMin, pMax float32) *RangeMap {
	return &RangeMap{
		xMin, xMax,
		pMin, pMax,
	}
}

func (rg *RangeMap) GetMappedCoordinate(x float32) (p float32) {
	p = (rg.pMax-rg.pMin)*(x-rg.xMin)/(rg.xMax-rg.xMin) + rg.pMin
	//p = float32(math.Max(float64(p), float64(rg.pMin)))
	//p = float32(math.Min(float64(p), float64(rg.pMax)))
	return p
}

type LineType uint8

const (
	NoLine LineType = iota
	Solid
	Dashed
)

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

func ToFloat32Slice(A []float64) (R []float32) {
	R = make([]float32, len(A))
	for i, val := range A {
		R[i] = float32(val)
	}
	return R
}
