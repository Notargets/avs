package opengl_rendering

import (
	"github.com/go-gl/gl/v4.5-core/gl"
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
	// Insert initialization of OGL 4.5 shader definitions here and compile and save the shader programs
}

func NewScreen() (scr *Screen) {
	scr = &Screen{
		Shaders: make(ShaderPrograms),
	}

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
	return scr
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

// NewLine initializes a new Line object, creates GPU buffers, and sets up the VAO
func NewLine() *Line {
	line := &Line{}

	// Generate VAO
	gl.GenVertexArrays(1, &line.VAO)
	gl.BindVertexArray(line.VAO)

	// Generate VBO for vertex positions
	gl.GenBuffers(1, &line.VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, 0, nil, gl.STATIC_DRAW) // Initial empty buffer

	// Define position layout (location = 0 in shader)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Generate CBO for vertex colors
	gl.GenBuffers(1, &line.CBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
	gl.BufferData(gl.ARRAY_BUFFER, 0, nil, gl.STATIC_DRAW) // Initial empty buffer

	// Define color layout (location = 1 in shader)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)

	// Unbind the VAO
	gl.BindVertexArray(0)

	return line
}

// Add adds the line geometry (X and F) to the GPU
func (line *Line) Add(X, F []float32) {
	// Ensure X and F have the same length
	if len(X) != len(F) {
		panic("X and F must have the same length")
	}

	// Create the vertex positions as [x1, y1, x2, y2, ...]
	line.Vertices = make([]float32, len(X)*2)
	for i := 0; i < len(X); i++ {
		line.Vertices[2*i] = X[i]
		line.Vertices[2*i+1] = F[i]
	}

	// For simplicity, we'll use a default color of white for all vertices (can be customized later)
	line.Colors = make([]float32, len(X)*3) // RGB for each vertex
	for i := 0; i < len(X); i++ {
		line.Colors[3*i] = 1.0   // R
		line.Colors[3*i+1] = 1.0 // G
		line.Colors[3*i+2] = 1.0 // B
	}

	// Bind the VAO
	gl.BindVertexArray(line.VAO)

	// Update vertex positions in the VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Vertices)*4, gl.Ptr(line.Vertices), gl.STATIC_DRAW)

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
