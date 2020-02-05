package chart2d

import (
	"fmt"
	_ "image/png"
	"log"
	"math"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// GLFW event handling must run on the main OS thread
	//runtime.LockOSThread()
}

type Series struct {
	Xdata []float32
	Ydata []float32
	Gl    GlyphType
	Lt    LineType
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
}

func NewChart2D(w, h int, xmin, xmax, ymin, ymax float32) (cc *Chart2D) {
	cc = &Chart2D{}
	cc.Sc = NewScreen(w, h)
	cc.RmX = NewRangeMap(xmin, xmax, 0, 1)
	cc.RmY = NewRangeMap(ymin, ymax, 0, 1)
	cc.activeSeries = make(map[string]Series)
	cc.inputChan = make(chan *NewDataMsg, 1000)
	cc.stopChan = make(chan struct{})
	return
}

func (cc *Chart2D) StopPlot() {
	cc.stopChan <- struct{}{}
}

func (cc *Chart2D) AddSeries(name string, x, f []float32, gl GlyphType, lt LineType) (err error) {
	switch {
	case len(name) == 0 || len(f) == 0 || len(x) == 0:
		return fmt.Errorf("empty series")
	case len(x) != len(f):
		return fmt.Errorf("length of x data not equal to function data length")
	}
	s := Series{
		Xdata: x,
		Ydata: f,
		Gl:    gl,
		Lt:    lt,
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

func (cc *Chart2D) Plot() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(cc.Sc.Width, cc.Sc.Height, "Chart2D", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
	ticker := time.NewTicker(8 * time.Millisecond)
	for !window.ShouldClose() {
		select {
		case <-ticker.C:
			cc.processNewData()
			cc.drawGraph()
			window.SwapBuffers()
			glfw.PollEvents()
		case <-cc.stopChan:
			goto END
		}
	}
END:
	return
}

func (cc *Chart2D) drawGraph() {
	var (
		xmargin = 0.1
		ymargin = 0.1
	)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(-xmargin, 1+xmargin, -ymargin, 1+ymargin, 0, 2)

	drawAxes()
	gl.Color3f(1, 1, 1)

	for _, s := range cc.activeSeries {
		if s.Gl != NoGlyph {
			for i, x := range s.Xdata {
				f := s.Ydata[i]
				xc := cc.RmX.GetMappedCoordinate(x)
				yc := cc.RmY.GetMappedCoordinate(f)
				drawGlyph(xc, yc, s.Gl, cc.Sc.GetRatio())
			}
		}
		if s.Lt != NoLine {
			gl.Begin(gl.LINE_STRIP)
			for i, x := range s.Xdata {
				f := s.Ydata[i]
				xc := cc.RmX.GetMappedCoordinate(x)
				yc := cc.RmY.GetMappedCoordinate(f)
				gl.Vertex2f(xc, yc)
			}
			gl.End()
		}
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
	p = float32(math.Max(float64(p), float64(rg.pMin)))
	p = float32(math.Min(float64(p), float64(rg.pMax)))
	return p
}

func drawAxes() {
	// Y axis
	gl.Color3f(0, 1, 0)
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(-.03, 0)
	gl.Vertex2f(0, 0)
	gl.Vertex2f(0, 1)
	gl.Vertex2f(-0.03, 1)
	gl.End()
	// X axis
	gl.Color3f(1, 0, 0)
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(0, -.03)
	gl.Vertex2f(0, 0)
	gl.Vertex2f(1, 0)
	gl.Vertex2f(1, -.03)
	gl.End()
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
)

func drawGlyph(xc, yc float32, gt GlyphType, rat float32) {
	switch gt {
	case CircleGlyph:
		DrawCircle(xc, yc, 0.010, 6, rat)
	case XGlyph:
		DrawXGlyph(xc, yc, rat)
	case CrossGlyph:
		DrawCrossGlyph(xc, yc, rat)
	case StarGlyph:
		DrawXGlyph(xc, yc, rat)
		DrawCrossGlyph(xc, yc, rat)
	}
}

func DrawXGlyph(cx, cy, rat float32) {
	var (
		hWidth = float32(0.01)
	)
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx-hWidth*rat, cy-hWidth)
	gl.Vertex2f(cx+hWidth*rat, cy+hWidth)
	gl.End()
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx-hWidth*rat, cy+hWidth)
	gl.Vertex2f(cx+hWidth*rat, cy-hWidth)
	gl.End()
}

func DrawCrossGlyph(cx, cy, rat float32) {
	var (
		hWidth = float32(0.01)
	)
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx, cy-hWidth)
	gl.Vertex2f(cx, cy+hWidth)
	gl.End()
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx-hWidth*rat, cy)
	gl.Vertex2f(cx+hWidth*rat, cy)
	gl.End()
}

func DrawCircle(cx, cy, r float32, numSegments int, rat float32) {
	theta := 2 * math.Pi / float64(numSegments)
	tangentialFactor := math.Tan(theta) //calculate the tangential factor
	radialFactor := math.Cos(theta)     //calculate the radial factor
	gl.Begin(gl.LINE_LOOP)
	var x, y float32
	x = r
	for ii := 0; ii < numSegments; ii++ {
		gl.Vertex2f(x+cx, y+cy) //output vertex
		//calculate the tangential vector
		//remember, the radial vector is (x, y)
		//to get the tangential vector we flip those coordinates and negate one of them
		tx := float64(-y)
		ty := float64(x)
		//add the tangential vector
		x += float32(tx * tangentialFactor)
		y += float32(ty * tangentialFactor)
		//correct using the radial factor
		x = float32(float64(x) * radialFactor)
		y = float32(float64(y) * radialFactor)
	}
	gl.End()
}