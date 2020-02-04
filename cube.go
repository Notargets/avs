package main

import (
	_ "image/png"
	"log"
	"math"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)


const (
	width, height = 1600, 1200
)

var (
	shift = 0
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func run() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(width, height, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	for !window.ShouldClose() {
	    time.Sleep(1*time.Millisecond)
		drawGraph(0,1,width, height)
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

type Range struct {
	xMin, xMax float32
	pMin, pMax float32
}

func NewRange(xMin, xMax float32, pMin, pMax float32) *Range {
	return &Range{
		xMin, xMax,
		pMin, pMax,
	}
}

func (rg *Range) GetProjection(x float32) (p float32) {
    p = (rg.pMax-rg.pMin)*(x - rg.xMin) / (rg.xMax-rg.xMin) + rg.pMin
    p = float32(math.Max(float64(p), float64(rg.pMin)))
	p = float32(math.Min(float64(p), float64(rg.pMax)))
	return p
}

func drawGraph(xrange, yrange float32, width, height int) {
    var (
    	xmargin = 0.1
    	ymargin = 0.1
	)
    xr := NewRange(0, 2*math.Pi, 0, 1)
	yr := NewRange(-1, 1, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(-xmargin, 1+xmargin, -ymargin, 1+ymargin, 0, 2)

	drawAxes(0,0,0,0)
	gl.Color3f(1, 1, 1)
	gl.Begin(gl.LINE_STRIP)
	size := 100
	for i := 0; i<size; i++ {
		frac := float32(i)/float32(size-1)
		xc := xr.GetProjection(frac*2*math.Pi)
		frac = float32(shift + i)/float32(size-1)
		yc := yr.GetProjection(float32(math.Sin(float64(frac*2*math.Pi))))
		gl.Vertex2f(xc, yc)
	}
	gl.End()
	shift += 1
}

func drawAxes(xmin, xmax, ymin, ymax float32) {
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

