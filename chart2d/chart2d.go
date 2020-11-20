package chart2d

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"runtime"
	"time"

	"github.com/notargets/avs/functions"

	"github.com/notargets/avs/utils"

	graphics2D "github.com/notargets/avs/geometry"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// GLFW event handling must run on the main OS thread
	//runtime.LockOSThread()
}

type Series struct {
	Xdata   []float32
	Ydata   []float32
	TriMesh *graphics2D.TriMesh
	Surface *functions.FSurface
	Vectors [][2]float64
	Gl      GlyphType
	Lt      LineType
	Co      *color.RGBA
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
		Xdata:   x,
		Ydata:   y,
		Vectors: vectors,
		Lt:      lt,
		Co:      &co,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) AddTriMesh(name string, Tris graphics2D.TriMesh, gl GlyphType, lt LineType, co color.RGBA) (err error) {
	s := Series{
		TriMesh: &Tris,
		Gl:      gl,
		Lt:      lt,
		Co:      &co,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) AddFunctionSurface(name string, fs functions.FSurface, lineType LineType, lineColor color.RGBA) (err error) {
	s := Series{
		Surface: &fs,
		Lt:      lineType,
		Co:      &lineColor,
	}
	cc.inputChan <- &NewDataMsg{name, s}
	return
}

func (cc *Chart2D) AddColorMap(cm *utils.ColorMap) {
	// Used to render data attributes
	cc.colormap = cm
}

func (cc *Chart2D) AddSeries(name string, xI, fI interface{}, gl GlyphType, lt LineType, co color.RGBA) (err error) {
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
		Xdata: x,
		Ydata: f,
		Gl:    gl,
		Lt:    lt,
		Co:    &co,
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
	gl.Ortho(-xmargin, 1+xmargin, -ymargin, 1+ymargin, -1, 2.0)

	drawAxes()

	gl.LineWidth(2)
	for _, s := range cc.activeSeries {
		gl.Color4ub(s.Co.R, s.Co.G, s.Co.B, s.Co.A)
		gl.PolygonOffset(1, 1)
		switch {
		case s.Vectors != nil:
			if s.Lt != NoLine {
				for i, x := range s.Xdata {
					y := s.Ydata[i]
					xc := cc.RmX.GetMappedCoordinate(x)
					yc := cc.RmY.GetMappedCoordinate(y)
					dx := float32(s.Vectors[i][0])
					dy := float32(s.Vectors[i][1])
					dxc := cc.RmX.GetMappedCoordinate(x + dx)
					dyc := cc.RmY.GetMappedCoordinate(y + dy)
					gl.Begin(gl.LINES)
					gl.Vertex2f(xc, yc)
					gl.Vertex2f(dxc, dyc)
					gl.End()
				}
			}
		case s.Surface != nil:
			drawVert := func(index int32, tmesh *graphics2D.TriMesh) {
				pt := tmesh.Geometry[index]
				xc := cc.RmX.GetMappedCoordinate(pt.X[0])
				yc := cc.RmY.GetMappedCoordinate(pt.X[1])
				gl.Vertex2f(xc, yc)
			}
			active := s.Surface.ActiveFunction
			tmesh := s.Surface.Tris
			switch {
			case len(s.Surface.Functions) == 0:
				panic("function surface has no data")
			case len(s.Surface.Functions[active]) == 0:
				panic("function surface has no data")
			case cc.colormap == nil:
				panic("empty colormap")
			}
			f := s.Surface.Functions[active]
			for _, tri := range s.Surface.Tris.Triangles {
				gl.Enable(gl.POLYGON_OFFSET_FILL)
				gl.Begin(gl.TRIANGLES)
				for _, vertIndex := range tri.Nodes {
					vValue := f[vertIndex]
					vertColor := cc.colormap.GetRGB(vValue)
					gl.Color4ub(vertColor.R, vertColor.G, vertColor.B, vertColor.A)
					drawVert(vertIndex, tmesh)
				}
				gl.End()
				gl.Disable(gl.POLYGON_OFFSET_FILL)
				gl.Color4ub(s.Co.R, s.Co.G, s.Co.B, s.Co.A)
				gl.Begin(gl.LINES)
				for _, vertIndex := range tri.Nodes {
					drawVert(vertIndex, tmesh)
				}
				drawVert(tri.Nodes[0], tmesh) // close the triangle
				gl.End()
			}
		case s.TriMesh != nil:
			if s.Gl != NoGlyph {
				for k, tri := range s.TriMesh.Triangles {
					for i, vertIndex := range tri.Nodes {
						if cc.colormap != nil && s.TriMesh.Attributes != nil {
							edgeValue := s.TriMesh.Attributes[k][i]
							edgeColor := cc.colormap.GetRGB(edgeValue)
							gl.Color4ub(edgeColor.R, edgeColor.G, edgeColor.B, edgeColor.A)
						}
						pt := s.TriMesh.Geometry[vertIndex]
						xc := cc.RmX.GetMappedCoordinate(pt.X[0])
						yc := cc.RmY.GetMappedCoordinate(pt.X[1])
						drawGlyph(xc, yc, s.Gl, cc.Sc.GetRatio())
					}
				}
			}
			if s.Lt != NoLine {
				for k, tri := range s.TriMesh.Triangles {
					for i, vertIndex := range tri.Nodes {
						gl.Begin(gl.LINES)
						if cc.colormap != nil && s.TriMesh.Attributes != nil {
							edgeValue := s.TriMesh.Attributes[k][i]
							edgeColor := cc.colormap.GetRGB(edgeValue)
							gl.Color4ub(edgeColor.R, edgeColor.G, edgeColor.B, edgeColor.A)
						}
						pt := s.TriMesh.Geometry[vertIndex]
						xc := cc.RmX.GetMappedCoordinate(pt.X[0])
						yc := cc.RmY.GetMappedCoordinate(pt.X[1])
						gl.Vertex2f(xc, yc)
						iplus := i + 1
						if i == 2 {
							iplus = 0
						}
						ptp := tri.Nodes[iplus]
						pt = s.TriMesh.Geometry[ptp]
						xc = cc.RmX.GetMappedCoordinate(pt.X[0])
						yc = cc.RmY.GetMappedCoordinate(pt.X[1])
						gl.Vertex2f(xc, yc)
						gl.End()
					}
				}
			}
		default:
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
	BoxGlyph
	TriangleGlyph
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
	case BoxGlyph:
		DrawBoxGlyph(xc, yc, rat)
	case TriangleGlyph:
		DrawTriangleGlyph(xc, yc, rat)
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

func DrawBoxGlyph(cx, cy, rat float32) {
	var (
		hWidth = float32(0.01)
		hWRat  = rat * hWidth
	)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2f(cx-hWRat, cy-hWidth)
	gl.Vertex2f(cx-hWRat, cy+hWidth)
	gl.Vertex2f(cx+hWRat, cy+hWidth)
	gl.Vertex2f(cx+hWRat, cy-hWidth)
	gl.End()
}

func DrawTriangleGlyph(cx, cy, rat float32) {
	var (
		hWidth = float32(0.01)
		hWRat  = rat * hWidth
	)
	gl.Begin(gl.LINE_LOOP)
	gl.Vertex2f(cx-hWRat, cy-hWidth)
	gl.Vertex2f(cx, cy+hWidth)
	gl.Vertex2f(cx+hWRat, cy-hWidth)
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

func ToFloat32Slice(A []float64) (R []float32) {
	R = make([]float32, len(A))
	for i, val := range A {
		R[i] = float32(val)
	}
	return R
}
