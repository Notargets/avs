package chart2d

import (
	"log"
	"runtime"
	"runtime/cgo"
	"unsafe"

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
	Shaders ShaderPrograms // Stores precompiled shaders for all graphics types in the RenderType list
	Window  *glfw.Window
	Objects map[uuid.UUID]interface{}
	// Insert initialization of OGL 4.5 shader definitions here and compile and save the shader programs
}

func (cc *Chart2D) NewScreen(width, height int) (scr *Screen) {
	scr = &Screen{
		Shaders: make(ShaderPrograms),
	}
	scr.InitGLScreen(cc, width, height)
	scr.InitShaders()
	return
}

func (scr *Screen) InitGLScreen(cc *Chart2D, width, height int) {
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

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Enable VSync (limit frame rate to refresh rate)
	glfw.SwapInterval(1)

	handle := cgo.NewHandle(cc)
	scr.Window.SetUserPointer(unsafe.Pointer(handle))

	scr.Window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.mouseButtonCallback(w, button, action, mods)
	})

	scr.Window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.cursorPositionCallback(w, xpos, ypos)
	})

	scr.Window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.scrollCallback(w, xoff, yoff)
	})

	scr.Window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		handle := cgo.Handle(w.GetUserPointer())
		cc := handle.Value().(*Chart2D)
		cc.resizeCallback(w, width, height)
	})

	return
}

func (scr *Screen) InitShaders() {

	for i := int(LINE); i <= int(TRIMESHSMOOTH3D); i++ {
		switch RenderType(i) {
		case LINE:
			// Line Shaders
			var vertexShaderSource = `
#version 450
layout (location = 0) in vec2 position;
layout (location = 1) in vec3 color;
out vec3 fragColor;
void main() {
	gl_Position = vec4(position, 0.0, 1.0);
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
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexSource)
	gl.ShaderSource(vertexShader, 1, csource, nil)
	gl.CompileShader(vertexShader)
	free()

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentSource)
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	gl.CompileShader(fragmentShader)
	free()

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}

type Line struct {
	VAO, VBO, CBO uint32    // Vertex Array Object, Vertex Buffer Object, Color Buffer Object
	Vertices      []float32 // Flat list of vertex positions [x1, y1, x2, y2, ...]
	Colors        []float32 // Flat list of color data [r1, g1, b1, r2, g2, b2, ...]
}

func (scr *Screen) AddLine(key uuid.UUID, X, Y, Colors []float32, defaultColor ...[3]float32) uuid.UUID {
	var line *Line

	// Check if the line already exists
	if key != uuid.Nil {
		// Try to retrieve existing line from the screen object map
		existingLine, exists := scr.Objects[key]
		if exists {
			line = existingLine.(*Line)
		} else {
			// If no object exists for this key, create a new Line
			key = uuid.New()
			line = &Line{}
			scr.Objects[key] = line
		}
	} else {
		// Create a new Line if no key is provided
		key = uuid.New()
		line = &Line{}
		scr.Objects[key] = line
	}

	// Initialize OpenGL resources only if this is a new line
	if line.VAO == 0 {
		// Generate VAO
		gl.GenVertexArrays(1, &line.VAO)
		gl.BindVertexArray(line.VAO)

		// Generate VBO for vertex positions
		gl.GenBuffers(1, &line.VBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)

		// Define position layout (location = 0 in shader)
		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.Ptr(nil))
		gl.EnableVertexAttribArray(0)

		// Generate CBO for vertex colors
		gl.GenBuffers(1, &line.CBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)

		// Define color layout (location = 1 in shader)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.Ptr(nil))
		gl.EnableVertexAttribArray(1)

		// Unbind the VAO
		gl.BindVertexArray(0)
	}

	// Call the `Update` method to upload the vertex and color data
	line.Update(X, Y, Colors, defaultColor...)

	return key
}

// Update updates the line's vertex positions (X, Y) and vertex colors on the GPU
func (line *Line) Update(X, Y, Colors []float32, defaultColor ...[3]float32) {
	if len(X) > 0 && len(Y) > 0 && len(X) != len(Y) {
		panic("X and Y must have the same length if both are provided")
	}

	// Update vertex positions if X and Y are provided
	if len(X) > 0 && len(Y) > 0 {
		// Create the vertex positions as [x1, y1, x2, y2, ...]
		line.Vertices = make([]float32, len(X)*2)
		for i := 0; i < len(X); i++ {
			line.Vertices[2*i] = X[i]
			line.Vertices[2*i+1] = Y[i]
		}

		// Bind the VAO
		gl.BindVertexArray(line.VAO)

		// Update vertex positions in the VBO
		gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(line.Vertices)*4, gl.Ptr(line.Vertices), gl.STATIC_DRAW)
	}

	// Determine default color if Colors are not provided
	var colorToUse [3]float32 = [3]float32{1.0, 1.0, 1.0} // Default white color
	if len(defaultColor) > 0 {
		colorToUse = defaultColor[0]
	}

	// Update vertex colors if Colors are provided or generate default colors
	if len(Colors) > 0 {
		if len(Colors)%3 != 0 {
			panic("Colors array must be a multiple of 3 (R, G, B per vertex)")
		}
		line.Colors = make([]float32, len(Colors))
		copy(line.Colors, Colors)
	} else {
		// Generate color data using the default color
		numVertices := len(X)
		line.Colors = make([]float32, numVertices*3) // RGB for each vertex
		for i := 0; i < numVertices; i++ {
			line.Colors[3*i] = colorToUse[0]   // R
			line.Colors[3*i+1] = colorToUse[1] // G
			line.Colors[3*i+2] = colorToUse[2] // B
		}
	}

	// Bind the VAO
	gl.BindVertexArray(line.VAO)

	// Update color data in the CBO
	gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Colors)*4, gl.Ptr(line.Colors), gl.STATIC_DRAW)

	// Unbind the VAO
	gl.BindVertexArray(0)
}

// Render draws the line using the specified shader program
func (line *Line) Render(shaderProgram uint32) {
	gl.UseProgram(shaderProgram)
	gl.BindVertexArray(line.VAO)
	gl.DrawArrays(gl.LINES, 0, int32(len(line.Vertices)/2))
	gl.BindVertexArray(0)
}

type TriMeshEdgesUniColor struct {
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
