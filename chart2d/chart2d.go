package chart2d

import (
	"log"
	"math"
	"runtime"
	"runtime/cgo"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/notargets/avs/utils"
)

//type Series struct {
//	Vertices []float32 // Interleaved position (x, y) and color (r, g, b)
//}

type Chart2D struct {
	DataChan     chan Series // Channel for new data
	VAO          uint32      // Vertex Array Object
	VBO          uint32      // Vertex Buffer Object
	shader       uint32      // Shader program
	activeSeries []Series    // List of currently active series
	Scale        float32
	Position     [2]float32
	isDragging   bool    // Tracks whether the right mouse button is being held
	lastX        float64 // Last cursor X position
	lastY        float64 // Last cursor Y position
	ScreenWidth  int     // Current width of the screen
	ScreenHeight int     // Current height of the screen
}

func NewChart2D(width, height int) *Chart2D {
	return &Chart2D{
		DataChan:     make(chan Series, 100), // Buffer size can be adjusted
		isDragging:   false,
		lastX:        0,
		lastY:        0,
		Scale:        1.0,
		Position:     [2]float32{0.0, 0.0},
		ScreenWidth:  width,  // Set initial width
		ScreenHeight: height, // Set initial height
	}
}

func (cc *Chart2D) Init() *glfw.Window {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	window, err := glfw.CreateWindow(cc.ScreenWidth, cc.ScreenHeight, "Chart2D", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		panic(err)
	}

	handle := cgo.NewHandle(cc)
	window.SetUserPointer(unsafe.Pointer(handle))

	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.mouseButtonCallback(w, button, action, mods)
	})

	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.cursorPositionCallback(w, xpos, ypos)
	})

	window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.scrollCallback(w, xoff, yoff)
	})

	window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.resizeCallback(w, width, height)
	})

	cc.setupGLResources()
	cc.shader = cc.compileShaders()
	return window
}

func (cc *Chart2D) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonRight && action == glfw.Press {
		cc.isDragging = true
		cc.lastX, cc.lastY = w.GetCursorPos()
	} else if button == glfw.MouseButtonRight && action == glfw.Release {
		cc.isDragging = false
	}
}

func (cc *Chart2D) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if cc.isDragging {
		width, height := w.GetSize()
		dx := float32(xpos-cc.lastX) / float32(width) * 2
		dy := float32(ypos-cc.lastY) / float32(height) * 2
		cc.Position[0] += dx
		cc.Position[1] -= dy
		cc.lastX = xpos
		cc.lastY = ypos
	}
}

func (cc *Chart2D) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	cc.Scale += float32(yoff) * 0.1
	if cc.Scale < 0.1 {
		cc.Scale = 0.1
	}
	if cc.Scale > 10.0 {
		cc.Scale = 10.0
	}
}

func (cc *Chart2D) resizeCallback(w *glfw.Window, width, height int) {
	cc.ScreenWidth = width
	cc.ScreenHeight = height

	// Update the OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Recalculate the orthographic projection matrix
	aspectRatio := float32(width) / float32(height)
	cc.updateProjectionMatrix(aspectRatio)
}

func (cc *Chart2D) updateProjectionMatrix(aspectRatio float32) {
	// Adjust the orthographic projection to maintain the aspect ratio
	projection := mgl32.Ortho2D(-aspectRatio, aspectRatio, -1, 1)
	projectionUniform := gl.GetUniformLocation(cc.shader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])
}

func (cc *Chart2D) Render() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(cc.shader)

	// Calculate the projection and model matrices
	aspectRatio := float32(cc.ScreenWidth) / float32(cc.ScreenHeight)
	projection := mgl32.Ortho2D(-aspectRatio, aspectRatio, -1, 1)

	model := mgl32.Scale3D(cc.Scale, cc.Scale, 1).Mul4(
		mgl32.Translate3D(cc.Position[0], cc.Position[1], 0),
	)

	modelUniform := gl.GetUniformLocation(cc.shader, gl.Str("model\x00"))
	projectionUniform := gl.GetUniformLocation(cc.shader, gl.Str("projection\x00"))

	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	gl.BindVertexArray(cc.VAO)
	totalVertices := int32(len(cc.activeSeries) * 3)
	gl.DrawArrays(gl.TRIANGLES, 0, totalVertices)
	gl.BindVertexArray(0)
}

func (cc *Chart2D) EventLoop(window *glfw.Window) {
	for !window.ShouldClose() {
		// Poll for events (mouse, keyboard, etc.)
		glfw.PollEvents()

		// Check for new data from the channel and update if available
		select {
		case newSeries := <-cc.DataChan:
			cc.UpdateSeries(newSeries) // Add new series to active series
		default:
			// No data, continue to render
		}

		// Render the current scene
		cc.Render()

		// Swap the buffers to show the new frame
		window.SwapBuffers()
	}
}

func (cc *Chart2D) setupGLResources() {
	gl.GenVertexArrays(1, &cc.VAO)
	gl.BindVertexArray(cc.VAO)

	gl.GenBuffers(1, &cc.VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cc.VBO)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0)) // X, Y
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4)) // R, G, B
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (cc *Chart2D) compileShaders() uint32 {
	vertexShaderSource := `
#version 450
layout (location = 0) in vec2 position;
layout (location = 1) in vec3 color;

uniform mat4 model;
uniform mat4 projection;

out vec3 fragColor;

void main() {
    gl_Position = projection * model * vec4(position, 0.0, 1.0);
    fragColor = color;
}` + "\x00"

	fragmentShaderSource := `
	#version 450

in vec3 fragColor;
out vec4 color;

void main() {
    color = vec4(fragColor, 1.0);
}` + "\x00"

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexShaderSource)
	defer free()
	gl.ShaderSource(vertexShader, 1, csource, nil)
	gl.CompileShader(vertexShader)

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentShaderSource)
	defer free()
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	gl.CompileShader(fragmentShader)

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}

func (cc *Chart2D) UpdateSeries(newSeries Series) {
	cc.activeSeries = append(cc.activeSeries, newSeries)
	cc.updateVBO()
}

func (cc *Chart2D) updateVBO() {
	vertices := []float32{}
	for _, s := range cc.activeSeries {
		vertices = append(vertices, s.Vertices...)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, cc.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func DrawGlyph(xc, yc float32, glyphType GlyphType, glyphSize float32) []float32 {
	switch glyphType {
	case CircleGlyph:
		return DrawCircle(xc, yc, glyphSize, 6)
		//DrawCircle(xc, yc, glyphSize, 6, rat)
	case XGlyph:
		//DrawXGlyph(xc, yc, rat)
	case CrossGlyph:
		return DrawCrossGlyph(xc, yc, glyphSize)
		//DrawCrossGlyph(xc, yc, rat)
	case StarGlyph:
		fallthrough
		//DrawXGlyph(xc, yc, rat)
		//DrawCrossGlyph(xc, yc, rat)
	case BoxGlyph:
		fallthrough
		//DrawBoxGlyph(xc, yc, rat)
	case TriangleGlyph:
		//DrawTriangleGlyph(xc, yc, rat)
		panic("unimplemented")
	}
	return []float32{}
}

func DrawCircle(cx, cy, r float32, segments int) []float32 {
	vertices := []float32{}
	theta := 2 * math.Pi / float64(segments)
	for i := 0; i < segments; i++ {
		x := cx + r*float32(math.Cos(float64(i)*theta))
		y := cy + r*float32(math.Sin(float64(i)*theta))
		vertices = append(vertices, x, y, 1.0, 0.0, 0.0)
	}
	return vertices
}

func DrawCrossGlyph(cx, cy, size float32) []float32 {
	return []float32{
		cx - size, cy, 1.0, 0.0, 0.0,
		cx + size, cy, 0.0, 1.0, 0.0,
		cx, cy - size, 0.0, 0.0, 1.0,
		cx, cy + size, 1.0, 1.0, 0.0,
	}
}

type Chart2D_old struct {
	Sc           *Screen
	RmX, RmY     *RangeMap
	activeSeries map[string]Series
	inputChan    chan *NewDataMsg
	stopChan     chan struct{}
	colormap     *utils.ColorMap
}

func (cc *Chart2D_old) StopPlot() {
	cc.stopChan <- struct{}{}
}

func (cc *Chart2D_old) processNewData() {
	for i := 0; i < len(cc.inputChan); i++ {
		msg := <-cc.inputChan
		cc.activeSeries[msg.Name] = msg.Data
	}
}
func NewChart2D_old(w, h int, xmin, xmax, ymin, ymax float32, chanDepth ...int) (cc *Chart2D_old) {
	cc = &Chart2D_old{}
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
