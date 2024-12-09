package chart2d

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"runtime/cgo"
	"strings"
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
	// Fields for World Coordinate Range**
	XMin, XMax       float32 // World X-range
	YMin, YMax       float32 // World Y-range
	ProjectionMatrix mgl32.Mat4
	PanSpeed         float32 // Speed of panning
	ZoomSpeed        float32 // Speed of zooming
	ZoomFactor       float32 // Factor controlling zoom (instead of scale)
}

func NewChart2D(width, height int, xmin, xmax, ymin, ymax float64) *Chart2D {
	return &Chart2D{
		DataChan:     make(chan Series, 100), // Buffer size can be adjusted
		isDragging:   false,
		lastX:        0,
		lastY:        0,
		Scale:        1.0,
		Position:     [2]float32{0.0, 0.0},
		ScreenWidth:  width,  // Set initial width
		ScreenHeight: height, // Set initial height
		XMin:         float32(xmin),
		XMax:         float32(xmax),
		YMin:         float32(ymin),
		YMax:         float32(ymax),
		PanSpeed:     1.0,
		ZoomSpeed:    1.0,
		ZoomFactor:   1.0,
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

	// Setup OpenGL resources AFTER shader and VBO are ready
	cc.setupGLResources()

	// Compile shader and ensure it's ready for use
	cc.shader = cc.compileShaders()
	gl.UseProgram(cc.shader) // Activate the shader

	// Set the viewport and update projection
	gl.Viewport(0, 0, int32(cc.ScreenWidth), int32(cc.ScreenHeight))
	cc.updateProjectionMatrix()

	// Force a single render before the event loop to ensure something is drawn
	cc.Render()

	return window
}

func (cc *Chart2D) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		cc.ZoomSpeed = 1.0
		return
	}
	cc.ZoomSpeed = speed
}
func (cc *Chart2D) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		cc.PanSpeed = 1.0
		return
	}
	cc.PanSpeed = speed
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

		// Normalize pan speed relative to the world coordinates
		dx := float32(xpos-cc.lastX) / float32(width) * (cc.XMax - cc.XMin) * cc.PanSpeed
		dy := float32(ypos-cc.lastY) / float32(height) * (cc.YMax - cc.YMin) * cc.PanSpeed

		cc.Position[0] += dx
		cc.Position[1] -= dy

		cc.lastX = xpos
		cc.lastY = ypos
	}
}

func (cc *Chart2D) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Calculate the zoom factor
	zoomAmount := 1.0 + float32(yoff)*0.1*cc.ZoomSpeed
	cc.ZoomFactor *= zoomAmount

	if cc.ZoomFactor < 0.1 {
		cc.ZoomFactor = 0.1
	}
	if cc.ZoomFactor > 10.0 {
		cc.ZoomFactor = 10.0
	}

	// Update the projection matrix to reflect the new zoom
	cc.updateProjectionMatrix()
}

func (cc *Chart2D) resizeCallback(w *glfw.Window, width, height int) {
	cc.ScreenWidth = width
	cc.ScreenHeight = height

	gl.Viewport(0, 0, int32(width), int32(height))
	cc.updateProjectionMatrix()
}

func (cc *Chart2D) updateProjectionMatrix() {
	aspectRatio := float32(cc.ScreenWidth) / float32(cc.ScreenHeight)

	var xRange, yRange float32
	if aspectRatio > 1.0 {
		xRange = (cc.XMax - cc.XMin) / cc.ZoomFactor
		yRange = xRange / aspectRatio
	} else {
		yRange = (cc.YMax - cc.YMin) / cc.ZoomFactor
		xRange = yRange * aspectRatio
	}

	xCenter := (cc.XMin + cc.XMax) / 2.0
	yCenter := (cc.YMin + cc.YMax) / 2.0

	xmin := xCenter - xRange/2.0
	xmax := xCenter + xRange/2.0
	ymin := yCenter - yRange/2.0
	ymax := yCenter + yRange/2.0

	projection := mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	gl.UseProgram(cc.shader) // Activate the shader before accessing uniforms

	projectionUniform := gl.GetUniformLocation(cc.shader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	if err := gl.GetError(); err != 0 {
		fmt.Printf("OpenGL Error after setting projection: %d\n", err)
	}
}

func (cc *Chart2D) Render() {
	// Clear the screen
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Use the shader program
	gl.UseProgram(cc.shader)

	// Calculate the model matrix for pan/zoom
	model := mgl32.Translate3D(cc.Position[0], cc.Position[1], 0)

	// Get the uniform location for model matrix
	modelUniform := gl.GetUniformLocation(cc.shader, gl.Str("model\x00"))

	// Send the model matrix to the shader
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	// Bind the Vertex Array Object (VAO)
	gl.BindVertexArray(cc.VAO)

	// Calculate total number of vertices in the VBO
	totalVertices := int32(0)
	for _, series := range cc.activeSeries {
		totalVertices += int32(len(series.Vertices) / 5) // 2D position + 3 color components
	}

	// Draw all the triangles
	gl.DrawArrays(gl.TRIANGLES, 0, totalVertices)

	// Unbind VAO to prevent unintended modifications
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

	// Allocate an empty buffer of "potentially large" size (1MB here, but could be smaller)
	gl.BufferData(gl.ARRAY_BUFFER, 1024*1024, nil, gl.DYNAMIC_DRAW)

	// Set up vertex attributes for position (2 floats) and color (3 floats)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0)) // Position (x, y)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4)) // Color (r, g, b)
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (cc *Chart2D) compileShaders() uint32 {
	vertexShaderSource := `
#version 450

layout (location = 0) in vec2 position; // Vertex position
layout (location = 1) in vec3 color;    // Vertex color

uniform mat4 model;        // Model transformation (for pan, zoom, etc.)
uniform mat4 projection;   // Projection transformation (for screen-to-world transform)

out vec3 fragColor;        // Output color for the fragment shader

void main() {
    // Calculate final position using projection * model * position
    gl_Position = projection * model * vec4(position, 0.0, 1.0);
    fragColor = color;
}
` + "\x00"

	fragmentShaderSource := `
#version 450

in vec3 fragColor; // Interpolated color from the vertex shader
out vec4 color;    // Final color of the pixel

void main() {
    color = vec4(fragColor, 1.0); // Use the interpolated color as output
}
` + "\x00"

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexShaderSource)
	defer free()
	gl.ShaderSource(vertexShader, 1, csource, nil)
	gl.CompileShader(vertexShader)
	checkShaderCompileStatus(vertexShader) // Check for errors

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentShaderSource)
	defer free()
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	gl.CompileShader(fragmentShader)
	checkShaderCompileStatus(fragmentShader) // Check for errors

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)
	checkProgramLinkStatus(shaderProgram) // Check for errors

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}

func checkShaderCompileStatus(shader uint32) {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		fmt.Printf("SHADER COMPILE ERROR: %s\n", log)
		panic(log)
	}
}

func checkProgramLinkStatus(program uint32) {
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		fmt.Printf("PROGRAM LINK ERROR: %s\n", log)
		panic(log)
	}
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
