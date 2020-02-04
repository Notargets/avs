package main

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

const (
	width, height = 1200, 800
)

var (
	shift = 0
	inc   = 0
	gt    = GlyphType(0)
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
		time.Sleep(1 * time.Millisecond)
		drawGraph(0, 1, width, height)
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
	p = (rg.pMax-rg.pMin)*(x-rg.xMin)/(rg.xMax-rg.xMin) + rg.pMin
	p = float32(math.Max(float64(p), float64(rg.pMin)))
	p = float32(math.Min(float64(p), float64(rg.pMax)))
	return p
}

func drawGraph(xrange, yrange float32, width, height int) {
	var (
		xmargin = 0.1
		ymargin = 0.1
	)
	if inc%height == 0 {
		shift += 1
		if shift%10 == 0 {
			gt = GlyphType(shift / 10 % 4)
			fmt.Printf("10x reached, shift = %d, gt = %d\n", shift, gt)
		}
	}
	xr := NewRange(0, 2*math.Pi, 0, 1)
	yr := NewRange(-1, 1, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(-xmargin, 1+xmargin, -ymargin, 1+ymargin, 0, 2)

	drawAxes(0, 0, 0, 0)
	gl.Color3f(1, 1, 1)
	//gl.Begin(gl.LINE_STRIP)
	size := 100
	for i := 0; i < size; i++ {
		frac := float32(i) / float32(size-1)
		xc := xr.GetProjection(frac * 2 * math.Pi)
		frac = float32(shift+i) / float32(size-1)
		yc := yr.GetProjection(float32(math.Sin(float64(frac * 2 * math.Pi))))
		//	gl.Vertex2f(xc, yc)
		//drawGlyph(xc, yc, CircleGlyph)
		//drawGlyph(xc, yc, XGlyph)
		//drawGlyph(xc, yc, CrossGlyph)
		//drawGlyph(xc, yc, StarGlyph)
		drawGlyph(xc, yc, gt)
	}
	//gl.End()
	inc += size
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

type GlyphType uint8

const (
	CircleGlyph GlyphType = iota
	XGlyph
	CrossGlyph
	StarGlyph
)

func drawGlyph(xc, yc float32, gt GlyphType) {
	switch gt {
	case CircleGlyph:
		DrawCircle(xc, yc, 0.010, 6)
	case XGlyph:
		DrawXGlyph(xc, yc)
	case CrossGlyph:
		DrawCrossGlyph(xc, yc)
	case StarGlyph:
		DrawXGlyph(xc, yc)
		DrawCrossGlyph(xc, yc)
	}
}

func DrawXGlyph(cx, cy float32) {
	var (
		hWidth = float32(0.01)
	)
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx-hWidth, cy-hWidth)
	gl.Vertex2f(cx+hWidth, cy+hWidth)
	gl.End()
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx-hWidth, cy+hWidth)
	gl.Vertex2f(cx+hWidth, cy-hWidth)
	gl.End()
}

func DrawCrossGlyph(cx, cy float32) {
	var (
		hWidth = float32(0.01)
	)
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx, cy-hWidth)
	gl.Vertex2f(cx, cy+hWidth)
	gl.End()
	gl.Begin(gl.LINE_STRIP)
	gl.Vertex2f(cx-hWidth, cy)
	gl.Vertex2f(cx+hWidth, cy)
	gl.End()
}

func DrawCircle(cx, cy, r float32, numSegments int) {
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
