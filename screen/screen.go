package screen

import (
	"fmt"
	"log"
	"runtime"
	"runtime/cgo"
	"strings"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/google/uuid"
)

type RenderType uint16

const (
	LINE RenderType = iota
	TRIMESHEDGESUNICOLOR
	TRIMESHEDGES
	TRIMESHCONTOURS
	TRIMESHSMOOTH
	LINE3D
	TRIMESHEDGESUNICOLOR3D
	TRIMESHEDGES3D
	TRIMESHCONTOURS3D
	TRIMESHSMOOTH3D
)

type ShaderPrograms map[RenderType]uint32

type Screen struct {
	Shaders          ShaderPrograms // Stores precompiled shaders for all graphics types
	Window           *glfw.Window
	Objects          map[uuid.UUID]Renderable
	RenderChannel    chan func()
	Scale            float32
	Position         [2]float32
	isDragging       bool
	lastX            float64
	lastY            float64
	projectionMatrix mgl32.Mat4
	ScreenWidth      int
	ScreenHeight     int
	XMin, XMax       float32
	YMin, YMax       float32
	PanSpeed         float32
	ZoomSpeed        float32
	ZoomFactor       float32
	PositionChanged  bool
	ScaleChanged     bool
}

type Renderable struct {
	Active bool
	Object interface{} // Any object that has a Render method (e.g., Line, TriMesh)
}

func NewScreen(width, height int, xmin, xmax, ymin, ymax float32) *Screen {
	screen := &Screen{
		Shaders:       make(ShaderPrograms),
		Objects:       make(map[uuid.UUID]Renderable),
		RenderChannel: make(chan func(), 100),
		ScreenWidth:   width,
		ScreenHeight:  height,
		XMin:          float32(xmin),
		XMax:          float32(xmax),
		YMin:          float32(ymin),
		YMax:          float32(ymax),
		PanSpeed:      1.0,
		ZoomSpeed:     1.0,
		ZoomFactor:    1.0,
	}

	// Launch the OpenGL thread
	go func() {
		runtime.LockOSThread()

		if err := glfw.Init(); err != nil {
			log.Fatalln("Failed to initialize glfw:", err)
		}

		window, err := glfw.CreateWindow(width, height, "Chart2D", nil, nil)
		if err != nil {
			panic(err)
		}

		window.MakeContextCurrent()

		if err := gl.Init(); err != nil {
			log.Fatalln("Failed to initialize OpenGL context:", err)
		}
		gl.ClearColor(0.3, 0.3, 0.3, 1.0)

		// Store window reference
		screen.Window = window

		// Enable VSync
		glfw.SwapInterval(1)

		// Call the GL screen initialization
		screen.InitShaders()
		gl.Viewport(0, 0, int32(width), int32(height))
		screen.updateProjectionMatrix()

		// Force the first frame to render
		screen.PositionChanged = true
		screen.ScaleChanged = true

		// Start the event loop (OpenGL runs here)
		screen.EventLoop()
	}()

	return screen
}

func (scr *Screen) EventLoop() {
	for !scr.Window.ShouldClose() {
		// Process input events like key presses, mouse, etc.
		glfw.PollEvents()

		// Process any commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command()
		default:
			// No command to process
		}

		// Update the projection matrix if pan/zoom has changed
		if scr.PositionChanged || scr.ScaleChanged {
			scr.updateProjectionMatrix()
			scr.PositionChanged = false
			scr.ScaleChanged = false
		}

		// Clear the screen before rendering
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render all active objects (type-coerce and render)
		for _, renderable := range scr.Objects {
			if renderable.Active {
				switch obj := renderable.Object.(type) {
				case *Line:
					obj.Render(scr)
				case *TriMesh:
					//obj.Render(scr)
				case *TriMeshEdges:
					//obj.Render(scr)
				case *TriMeshContours:
					//obj.Render(scr)
				case *TriMeshSmooth:
					//obj.Render(scr)
				default:
					fmt.Printf("Unknown object type: %T\n", obj)
				}
			}
		}

		// Swap buffers to present the frame
		scr.Window.SwapBuffers()
	}
}

func (scr *Screen) SetObjectActive(key uuid.UUID, active bool) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			renderable.Active = active
			scr.Objects[key] = renderable
		}
	}
}

func (scr *Screen) fullScreenRender() {
	for _, obj := range scr.Objects {
		switch renderObj := obj.Object.(type) {
		case *Line:
			renderObj.Render(scr)
		case *TriMesh:
			//renderObj.Render(scr)
		case *TriMeshEdges:
			//renderObj.Render(scr)
		case *TriMeshContours:
			//renderObj.Render(scr)
		case *TriMeshSmooth:
			//renderObj.Render(scr)
		default:
			fmt.Printf("Unknown object type: %T\n", renderObj)
		}
	}
}

func (scr *Screen) InitGLScreen(width, height int) {
	var err error
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	scr.Window, err = glfw.CreateWindow(width, height, "Chart2D", nil, nil)
	if err != nil {
		panic(err)
	}
	scr.Window.MakeContextCurrent()

	// Check if context is properly active
	if glfw.GetCurrentContext() == nil {
		log.Fatalln("GLFW Context is not current!")
	}

	// Initialize OpenGL function pointers
	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL context:", err)
	}

	// Check OpenGL version (optional, but useful for debugging)
	version := gl.GoStr(gl.GetString(gl.VERSION))
	if version == "" {
		log.Fatalln("OpenGL context not properly initialized")
	}
	fmt.Println("OpenGL version:", version)

	// Check for OpenGL errors
	checkGLError("glfw MakeContextCurrent")

	// Enable VSync (limit frame rate to refresh rate)
	glfw.SwapInterval(1)

	handle := cgo.NewHandle(scr)
	scr.Window.SetUserPointer(unsafe.Pointer(&handle))

	scr.Window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.mouseButtonCallback(w, button, action, mods)
	})

	scr.Window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.cursorPositionCallback(w, xpos, ypos)
	})

	scr.Window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.scrollCallback(w, xoff, yoff)
	})

	scr.Window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.resizeCallback(w, width, height)
	})

	return
}

func (scr *Screen) SetBackgroundColor(r, g, b, a float32) {
	scr.RenderChannel <- func() {
		gl.ClearColor(r, g, b, a)
	}
}

func (scr *Screen) InitShaders() {

	i := 0
	{
		//for i := int(LINE); i <= int(TRIMESHSMOOTH3D); i++ {
		switch RenderType(i) {
		case LINE:
			// Line Shaders
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec2 position;
layout (location = 1) in vec3 color;
uniform mat4 projection; // Add this line
out vec3 fragColor;
void main() {
	gl_Position = projection * vec4(position, 0.0, 1.0); // Use projection
	fragColor = color;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in vec3 fragColor;
out vec4 outColor;
void main() {
	outColor = vec4(fragColor, 1.0);
}
` + "\x00"
			scr.Shaders[LINE] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHEDGESUNICOLOR:
			// TriMeshEdgesUniColor Shaders (for edges with uniform color)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
void main() {
	gl_Position = vec4(position, 1.0);
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
out vec4 outColor;
void main() {
	outColor = vec4(0.0, 0.0, 0.0, 1.0); // Uniform black color for edges
}
` + "\x00"
			scr.Shaders[TRIMESHEDGESUNICOLOR] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHEDGES:
			// TriMeshEdges Shaders (for edges with per-vertex colors)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in vec4 color;
out vec4 fragColor;
void main() {
	gl_Position = vec4(position, 1.0);
	fragColor = color;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in vec4 fragColor;
out vec4 outColor;
void main() {
	outColor = fragColor;
}
` + "\x00"
			scr.Shaders[TRIMESHEDGES] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHCONTOURS:
			// TriMeshContours Shaders (for contour shading)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in float scalar;
out float fragValue;
void main() {
	gl_Position = vec4(position, 1.0);
	fragValue = scalar;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in float fragValue;
out vec4 outColor;
void main() {
	outColor = vec4(fragValue, fragValue, fragValue, 1.0);
}
` + "\x00"
			scr.Shaders[TRIMESHCONTOURS] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHSMOOTH:
			// TriMeshSmooth Shaders (for smooth shaded triangles)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
out vec3 fragColor;
void main() {
	gl_Position = vec4(position, 1.0);
	fragColor = color;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in vec3 fragColor;
out vec4 outColor;
void main() {
	outColor = vec4(fragColor, 1.0);
}
` + "\x00"
			scr.Shaders[TRIMESHSMOOTH] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case LINE3D:
			// 3D Line Shaders
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
out vec3 fragColor;
uniform mat4 projection;
void main() {
	gl_Position = projection * vec4(position, 1.0);
	fragColor = color;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in vec3 fragColor;
out vec4 outColor;
void main() {
	outColor = vec4(fragColor, 1.0);
}
` + "\x00"
			scr.Shaders[LINE3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHEDGESUNICOLOR3D:
			// TriMeshEdgesUniColor3D Shaders
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
uniform mat4 projection;
void main() {
	gl_Position = projection * vec4(position, 1.0);
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
out vec4 outColor;
void main() {
	outColor = vec4(0.0, 0.0, 0.0, 1.0); // Black for edges
}
` + "\x00"
			scr.Shaders[TRIMESHEDGESUNICOLOR3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHEDGES3D:
			// TriMeshEdges3D Shaders (for edges with per-vertex colors)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
uniform mat4 projection;
out vec3 fragColor;
void main() {
	gl_Position = projection * vec4(position, 1.0);
	fragColor = color;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in vec3 fragColor;
out vec4 outColor;
void main() {
	outColor = vec4(fragColor, 1.0);
}
` + "\x00"
			scr.Shaders[TRIMESHEDGES3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHCONTOURS3D:
			// TriMeshContours3D Shaders (for contour shading)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in float scalar;
uniform mat4 projection;
out float fragValue;
void main() {
	gl_Position = projection * vec4(position, 1.0);
	fragValue = scalar;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in float fragValue;
out vec4 outColor;
void main() {
	outColor = vec4(fragValue, fragValue, fragValue, 1.0);
}
` + "\x00"
			scr.Shaders[TRIMESHCONTOURS3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)

		case TRIMESHSMOOTH3D:
			// TriMeshSmooth3D Shaders (for smooth shaded triangles)
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
uniform mat4 projection;
out vec3 fragColor;
void main() {
	gl_Position = projection * vec4(position, 1.0);
	fragColor = color;
}
` + "\x00"

			var fragmentShaderSource = `
#version 450
in vec3 fragColor;
out vec4 outColor;
void main() {
	outColor = vec4(fragColor, 1.0);
}
` + "\x00"
			scr.Shaders[TRIMESHSMOOTH3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
		}
	}
	return
}

func compileShaderProgram(vertexSource, fragmentSource string) uint32 {
	// Compile vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexSource) // Unpack both pointer and cleanup function
	gl.ShaderSource(vertexShader, 1, csource, nil)
	defer free() // Defer cleanup to release the C string memory
	gl.CompileShader(vertexShader)

	// Check for vertex shader compile errors
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
		fmt.Printf("Vertex Shader Compile Error: %s\n", log)
	}

	// Compile fragment shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentSource) // Unpack again for fragment shader
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	defer free() // Defer cleanup to release the C string memory
	gl.CompileShader(fragmentShader)

	// Check for fragment shader compile errors
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		fmt.Printf("Fragment Shader Compile Error: %s\n", log)
	}

	// Link the shader program
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	// Check for linking errors
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		fmt.Printf("Shader Link Error: %s\n", log)
	}

	// Clean up the compiled shaders after linking
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}

type Line struct {
	VAO, VBO, CBO uint32    // Vertex Array Object, Vertex Buffer Object, Color Buffer Object
	Vertices      []float32 // Flat list of vertex positions [x1, y1, x2, y2, ...]
	Colors        []float32 // Flat list of color data [r1, g1, b1, r2, g2, b2, ...]
	ShaderProgram uint32    // Shader program specific to this Line object
}

func (line *Line) Update(X, Y, Colors []float32, defaultColor ...[3]float32) {
	if len(X) > 0 && len(Y) > 0 && len(X) != len(Y) {
		panic("X and Y must have the same length if both are provided")
	}

	// Flatten X and Y into vertex array [x1, y1, x2, y2, ...]
	if len(X) > 0 && len(Y) > 0 {
		line.Vertices = make([]float32, len(X)*2)
		for i := 0; i < len(X); i++ {
			line.Vertices[2*i] = X[i]
			line.Vertices[2*i+1] = Y[i]
		}
	}

	// Default color logic
	var colorToUse [3]float32 = [3]float32{1.0, 1.0, 1.0}
	if len(defaultColor) > 0 {
		colorToUse = defaultColor[0]
	}

	// Create colors for each vertex
	if len(Colors) > 0 {
		if len(Colors)%3 != 0 {
			panic("Colors array must be a multiple of 3 (R, G, B per vertex)")
		}
		line.Colors = make([]float32, len(Colors))
		copy(line.Colors, Colors)
	} else {
		numVertices := len(X)
		line.Colors = make([]float32, numVertices*3)
		for i := 0; i < numVertices; i++ {
			line.Colors[3*i] = colorToUse[0]   // R
			line.Colors[3*i+1] = colorToUse[1] // G
			line.Colors[3*i+2] = colorToUse[2] // B
		}
	}

	// Upload vertex positions
	gl.BindVertexArray(line.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Vertices)*4, gl.Ptr(line.Vertices), gl.STATIC_DRAW)

	// Upload color data
	gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Colors)*4, gl.Ptr(line.Colors), gl.STATIC_DRAW)
	gl.BindVertexArray(0)
}

// Render draws the line using the shader program stored in Line
func (line *Line) Render(scr *Screen) {
	// Ensure shader program is active
	gl.UseProgram(line.ShaderProgram)
	gl.BindVertexArray(line.VAO)

	// Upload the projection matrix
	projectionUniform := gl.GetUniformLocation(line.ShaderProgram, gl.Str("projection\x00"))
	if projectionUniform >= 0 {
		gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
	} else {
		fmt.Println("Projection matrix uniform not found in Line shader")
	}

	// Draw the line segments
	gl.DrawArrays(gl.LINES, 0, int32(len(line.Vertices)/2))
	gl.BindVertexArray(0)
}

func checkGLError(message string) {
	err := gl.GetError()
	if err != 0 {
		fmt.Printf("OpenGL Error [%s]: %d\n", message, err)
	}
}
func (scr *Screen) AddLine(key uuid.UUID, X, Y, Colors []float32) uuid.UUID {
	if key == uuid.Nil {
		key = uuid.New()
	}

	// Send a command to create or update a line object
	scr.RenderChannel <- func() {
		var line *Line

		// Check if the object exists in the scene
		if existingRenderable, exists := scr.Objects[key]; exists {
			line = existingRenderable.Object.(*Line)
		} else {
			// Create new line
			line = &Line{ShaderProgram: scr.Shaders[LINE]}
			scr.Objects[key] = Renderable{
				Active: true,
				Object: line,
			}

			gl.GenVertexArrays(1, &line.VAO)
			gl.BindVertexArray(line.VAO)

			gl.GenBuffers(1, &line.VBO)
			gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
			gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
			gl.EnableVertexAttribArray(0)

			gl.GenBuffers(1, &line.CBO)
			gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
			gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
			gl.EnableVertexAttribArray(1)

			gl.BindVertexArray(0)
		}

		// Update vertex positions and color
		line.Update(X, Y, Colors)
	}

	return key
}

func (scr *Screen) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		scr.ZoomSpeed = 1.0
		return
	}
	scr.ZoomSpeed = speed
}
func (scr *Screen) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		scr.PanSpeed = 1.0
		return
	}
	scr.PanSpeed = speed
}

func (scr *Screen) updateProjectionMatrix() {
	// Get the aspect ratio of the window
	aspectRatio := float32(scr.ScreenWidth) / float32(scr.ScreenHeight)

	// Determine world coordinate range (XMin/XMax and YMin/YMax) based on zoom and position
	var xRange, yRange float32
	if aspectRatio > 1.0 {
		// Landscape view
		xRange = (scr.XMax - scr.XMin) / scr.ZoomFactor
		yRange = xRange / aspectRatio
	} else {
		// Portrait view
		yRange = (scr.YMax - scr.YMin) / scr.ZoomFactor
		xRange = yRange * aspectRatio
	}

	// Adjust for position (pan offset)
	xmin := scr.Position[0] - xRange/2.0
	xmax := scr.Position[0] + xRange/2.0
	ymin := scr.Position[1] - yRange/2.0
	ymax := scr.Position[1] + yRange/2.0

	// Update the projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Upload the new projection matrix to all shaders
	for renderType, shaderProgram := range scr.Shaders {
		projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
		if projectionUniform < 0 {
			fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
		} else {
			gl.UseProgram(shaderProgram)
			gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
		}
	}
}

func (scr *Screen) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonRight && action == glfw.Press {
		scr.isDragging = true
		scr.lastX, scr.lastY = w.GetCursorPos()
	} else if button == glfw.MouseButtonRight && action == glfw.Release {
		scr.isDragging = false
	}
}

func (scr *Screen) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if scr.isDragging {
		width, height := w.GetSize()

		// Calculate delta in screen space
		dx := float32(xpos-scr.lastX) / float32(width) * (scr.XMax - scr.XMin) / scr.ZoomFactor
		dy := float32(ypos-scr.lastY) / float32(height) * (scr.YMax - scr.YMin) / scr.ZoomFactor

		// Update world position relative to the screen movement
		scr.Position[0] -= dx
		scr.Position[1] += dy

		// Flag that position has changed
		scr.PositionChanged = true

		// Update last cursor position
		scr.lastX = xpos
		scr.lastY = ypos
	}
}

func (scr *Screen) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Adjust the zoom factor based on scroll input
	scr.ZoomFactor *= 1.0 + float32(yoff)*0.1*scr.ZoomSpeed

	// Constrain the zoom factor to prevent excessive zoom
	if scr.ZoomFactor < 0.1 {
		scr.ZoomFactor = 0.1
	}
	if scr.ZoomFactor > 10.0 {
		scr.ZoomFactor = 10.0
	}

	// Flag that the zoom has changed
	scr.ScaleChanged = true
}

func (scr *Screen) resizeCallback(w *glfw.Window, width, height int) {
	scr.ScreenWidth = width
	scr.ScreenHeight = height

	// Update OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Calculate the aspect ratio
	aspectRatio := float32(width) / float32(height)

	// Adjust world bounds based on aspect ratio
	if aspectRatio > 1.0 {
		// Landscape (widescreen) adjustment
		viewHeight := (scr.YMax - scr.YMin)
		viewWidth := viewHeight * aspectRatio
		centerX := (scr.XMax + scr.XMin) / 2.0
		scr.XMin = centerX - viewWidth/2.0
		scr.XMax = centerX + viewWidth/2.0
	} else {
		// Portrait (tall screen) adjustment
		viewWidth := (scr.XMax - scr.XMin)
		viewHeight := viewWidth / aspectRatio
		centerY := (scr.YMin + scr.YMax) / 2.0
		scr.YMin = centerY - viewHeight/2.0
		scr.YMax = centerY + viewHeight/2.0
	}

	// Update the projection matrix for the new window size
	scr.updateProjectionMatrix()

	// Mark the position and scale as changed so the event loop will force a re-render
	scr.PositionChanged = true
	scr.ScaleChanged = true
}

type TriMesh struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshEdges struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshContours struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshSmooth struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}

type Line3D struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshEdgesUniColor3D struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshEdges3D struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshContours3D struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
type TriMeshSmooth3D struct {
	// Insert VAO, VBO, CBO as needed to save the current GPU contents for reuse for this type
}
