package chart2d

import (
	"log"
	"runtime"
	"time"

	graphics2D "github.com/notargets/avs/geometry"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var (
	gl_LINES               = 0
	gl_TRIANGLES           = 0
	gl_LINE_STRIP          = 0
	gl_LINE_LOOP           = 0
	gl_COLOR_BUFFER_BIT    = 0
	gl_DEPTH_BUFFER_BIT    = 0
	gl_POLYGON_OFFSET_FILL = 0
	gl_PROJECTION          = 0 // dummy
)

func gl_Init(...interface{}) (err *error) { return nil }
func gl_Clear(...interface{})             {}
func gl_ClearColor(...interface{})        {}
func gl_MatrixMode(...interface{})        {}
func gl_LoadIdentity(...interface{})      {}
func gl_Ortho(...interface{})             {}
func gl_Color4ub(...interface{})          {}
func gl_PolygonOffset(...interface{})     {}
func gl_Begin(...interface{})             {}
func gl_Vertex2f(...interface{})          {}
func gl_Color3f(...interface{})           {}
func gl_End(...interface{})               {}
func gl_Enable(...interface{})            {}
func gl_Disable(...interface{})           {}
func gl_LineWidth(...interface{})         {}

func (cc *Chart2D_old) Plot() {
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

	if err := gl_Init(); err != nil {
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
func (cc *Chart2D_old) drawGraph() {
	var (
		xmargin = 0.1
		ymargin = 0.1
	)
	gl_ClearColor(1.0, 1.0, 1.0, 1.0)
	gl_Clear(gl_COLOR_BUFFER_BIT | gl_DEPTH_BUFFER_BIT)
	gl_MatrixMode(gl_PROJECTION)
	gl_LoadIdentity()
	gl_Ortho(-xmargin, 1+xmargin, -ymargin, 1+ymargin, -1, 2.0)

	drawAxes()

	gl_LineWidth(2)
	for _, s := range cc.activeSeries {
		gl_Color4ub(s.Color.R, s.Color.G, s.Color.B, s.Color.A)
		gl_PolygonOffset(1, 1)
		switch {
		case s.Vectors != nil:
			if s.Linetype != NoLine {
				for i, x := range s.Xdata {
					y := s.Ydata[i]
					xc := cc.RmX.GetMappedCoordinate(x)
					yc := cc.RmY.GetMappedCoordinate(y)
					dx := float32(s.Vectors[i][0])
					dy := float32(s.Vectors[i][1])
					dxc := cc.RmX.GetMappedCoordinate(x + dx)
					dyc := cc.RmY.GetMappedCoordinate(y + dy)
					gl_Begin(gl_LINES)
					gl_Vertex2f(xc, yc)
					gl_Vertex2f(dxc, dyc)
					gl_End()
				}
			}
		case s.Surface != nil:
			getXY := func(index int32, tmesh *graphics2D.TriMesh) (xc, yc float32) {
				pt := tmesh.Geometry[index]
				xc = cc.RmX.GetMappedCoordinate(pt.X[0])
				yc = cc.RmY.GetMappedCoordinate(pt.X[1])
				return
			}
			drawVert := func(index int32, tmesh *graphics2D.TriMesh) {
				gl_Vertex2f(getXY(index, tmesh))
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
			gl_Enable(gl_POLYGON_OFFSET_FILL)
			for _, tri := range s.Surface.Tris.Triangles {
				gl_Begin(gl_TRIANGLES)
				for _, vertIndex := range tri.Nodes {
					vValue := f[vertIndex]
					vertColor := cc.colormap.GetRGB(vValue)
					gl_Color4ub(vertColor.R, vertColor.G, vertColor.B, vertColor.A)
					drawVert(vertIndex, tmesh)
				}
				gl_End()
			}
			if s.Linetype != NoLine {
				for _, tri := range s.Surface.Tris.Triangles {
					gl_Disable(gl_POLYGON_OFFSET_FILL)
					gl_Color4ub(s.Color.R, s.Color.G, s.Color.B, s.Color.A)
					gl_Begin(gl_LINES)
					for _, vertIndex := range tri.Nodes {
						drawVert(vertIndex, tmesh)
					}
					drawVert(tri.Nodes[0], tmesh) // close the triangle
					gl_End()
				}
			}
		case s.TriMesh != nil:
			getXY := func(index int32, tmesh *graphics2D.TriMesh) (xc, yc float32) {
				pt := tmesh.Geometry[index]
				xc = cc.RmX.GetMappedCoordinate(pt.X[0])
				yc = cc.RmY.GetMappedCoordinate(pt.X[1])
				return
			}
			drawVert := func(index int32, tmesh *graphics2D.TriMesh) {
				gl_Vertex2f(getXY(index, tmesh))
			}
			drawGlyphL := func(index int32, tmesh *graphics2D.TriMesh) {
				xc, yc := getXY(index, tmesh)
				DrawGlyph(xc, yc, s.Glyph, s.GlyphSize)
			}
			if s.Glyph != NoGlyph {
				for k, tri := range s.TriMesh.Triangles {
					for i, vertIndex := range tri.Nodes {
						if cc.colormap != nil && s.TriMesh.Attributes != nil {
							edgeValue := s.TriMesh.Attributes[k][i]
							edgeColor := cc.colormap.GetRGB(edgeValue)
							gl_Color4ub(edgeColor.R, edgeColor.G, edgeColor.B, edgeColor.A)
						}
						drawGlyphL(vertIndex, s.TriMesh)
					}
				}
			}
			if s.Linetype != NoLine {
				for k, tri := range s.TriMesh.Triangles {
					for i, vertIndex := range tri.Nodes {
						gl_Begin(gl_LINES)
						if cc.colormap != nil && s.TriMesh.Attributes != nil {
							edgeValue := s.TriMesh.Attributes[k][i]
							edgeColor := cc.colormap.GetRGB(edgeValue)
							gl_Color4ub(edgeColor.R, edgeColor.G, edgeColor.B, edgeColor.A)
						}
						drawVert(vertIndex, s.TriMesh)
						iplus := i + 1
						if i == 2 {
							iplus = 0
						}
						ptp := tri.Nodes[iplus]
						drawVert(ptp, s.TriMesh)
						gl_End()
					}
				}
			}
		default:
			if s.Glyph != NoGlyph {
				for i, x := range s.Xdata {
					f := s.Ydata[i]
					xc := cc.RmX.GetMappedCoordinate(x)
					yc := cc.RmY.GetMappedCoordinate(f)
					DrawGlyph(xc, yc, s.Glyph, s.GlyphSize)
				}
			}
			if s.Linetype != NoLine {
				gl_Begin(gl_LINE_STRIP)
				for i, x := range s.Xdata {
					f := s.Ydata[i]
					xc := cc.RmX.GetMappedCoordinate(x)
					yc := cc.RmY.GetMappedCoordinate(f)
					gl_Vertex2f(xc, yc)
				}
				gl_End()
			}
		}
	}
}

func drawAxes() {
	// Y axis
	gl_Color3f(0, 1, 0)
	gl_Begin(gl_LINE_STRIP)
	gl_Vertex2f(-.03, 0)
	gl_Vertex2f(0, 0)
	gl_Vertex2f(0, 1)
	gl_Vertex2f(-0.03, 1)
	gl_End()
	// X axis
	gl_Color3f(1, 0, 0)
	gl_Begin(gl_LINE_STRIP)
	gl_Vertex2f(0, -.03)
	gl_Vertex2f(0, 0)
	gl_Vertex2f(1, 0)
	gl_Vertex2f(1, -.03)
	gl_End()
}

func DrawXGlyph(cx, cy, rat float32) {
	gl_Begin(gl_LINE_STRIP)
	gl_Vertex2f(cx-0.01*rat, cy-0.01)
	gl_Vertex2f(cx+0.01*rat, cy+0.01)
	gl_End()
	gl_Begin(gl_LINE_STRIP)
	gl_Vertex2f(cx-0.01*rat, cy+0.01)
	gl_Vertex2f(cx+0.01*rat, cy-0.01)
	gl_End()
}

//func DrawCrossGlyph(cx, cy, rat float32) {
//	gl_Begin(gl_LINE_STRIP)
//	gl_Vertex2f(cx, cy-0.01)
//	gl_Vertex2f(cx, cy+0.01)
//	gl_End()
//	gl_Begin(gl_LINE_STRIP)
//	gl_Vertex2f(cx-0.01*rat, cy)
//	gl_Vertex2f(cx+0.01*rat, cy)
//	gl_End()
//}

func DrawBoxGlyph(cx, cy, rat float32) {
	gl_Begin(gl_LINE_LOOP)
	gl_Vertex2f(cx-0.01*rat, cy-0.01)
	gl_Vertex2f(cx-0.01*rat, cy+0.01)
	gl_Vertex2f(cx+0.01*rat, cy+0.01)
	gl_Vertex2f(cx+0.01*rat, cy-0.01)
	gl_End()
}

func DrawTriangleGlyph(cx, cy, rat float32) {
	gl_Begin(gl_LINE_LOOP)
	gl_Vertex2f(cx-0.01*rat, cy-0.01)
	gl_Vertex2f(cx, cy+0.01)
	gl_Vertex2f(cx+0.01*rat, cy-0.01)
	gl_End()
}

//func DrawCircle(cx, cy, r float32, numSegments int, rat float32) {
//	theta := 2 * math.Pi / float64(numSegments)
//	tangentialFactor := math.Tan(theta)
//	radialFactor := math.Cos(theta)
//	gl_Begin(gl_LINE_LOOP)
//	var x, y float32
//	x = r
//	for ii := 0; ii < numSegments; ii++ {
//		gl_Vertex2f(x+cx, y+cy)
//		tx := float64(-y)
//		ty := float64(x)
//		x += float32(tx * tangentialFactor)
//		y += float32(ty * tangentialFactor)
//		x = float32(float64(x) * radialFactor)
//		y = float32(float64(y) * radialFactor)
//	}
//	gl_End()
//}

func DrawAxes() {
	gl_Color3f(0, 1, 0)
	gl_Begin(gl_LINE_STRIP)
	gl_Vertex2f(-.03, 0)
	gl_Vertex2f(0, 0)
	gl_Vertex2f(0, 1)
	gl_Vertex2f(-0.03, 1)
	gl_End()
	gl_Color3f(1, 0, 0)
	gl_Begin(gl_LINE_STRIP)
	gl_Vertex2f(0, -.03)
	gl_Vertex2f(0, 0)
	gl_Vertex2f(1, 0)
	gl_Vertex2f(1, -.03)
	gl_End()
}

//func DrawGlyph(xc, yc float32, glyphType GlyphType, glyphSize, rat float32) {
//	switch glyphType {
//	case CircleGlyph:
//		DrawCircle(xc, yc, glyphSize, 6, rat)
//	case XGlyph:
//		DrawXGlyph(xc, yc, rat)
//	case CrossGlyph:
//		DrawCrossGlyph(xc, yc, rat)
//	case StarGlyph:
//		DrawXGlyph(xc, yc, rat)
//		DrawCrossGlyph(xc, yc, rat)
//	case BoxGlyph:
//		DrawBoxGlyph(xc, yc, rat)
//	case TriangleGlyph:
//		DrawTriangleGlyph(xc, yc, rat)
//	}
//}
