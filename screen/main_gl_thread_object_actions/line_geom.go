package main_gl_thread_object_actions

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/notargets/avs/utils"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type Line struct {
	VAO, VBO, CBO uint32    // Vertex Array Object, Vertex Buffer Object, Color Buffer Object
	Vertices      []float32 // Flat list of vertex positions [x1, y1, x2, y2, ...]
	Colors        []float32 // Flat list of color data [r1, g1, b1, r2, g2, b2, ...]
	LineType      utils.RenderType
	ShaderProgram uint32 // Shader program specific to this Line object
}

func (line *Line) Update(X, Y, Colors []float32, defaultColor ...[3]float32) {
	// Error check: Ensure X and Y are of the same length
	if len(X) > 0 && len(Y) > 0 && len(X) != len(Y) {
		panic("X and Y must have the same length if both are provided")
	}
	if len(Colors) != 0 && len(Colors) != 3*len(X) {
		panic("Colors must have 3*length(X) if any are provided, one RGB each vertex")
	}

	// Validate vertex count based on LineType
	switch line.LineType {
	case utils.LINE:
		if len(X) > 0 && len(Y) > 0 && len(X)%2 != 0 {
			panic(fmt.Sprintf("Invalid vertex count for LINE: %d. Each line segment requires two points (X1, Y1, X2, Y2). Vertex count must be a multiple of 2.", len(X)))
		}
	case utils.POLYLINE:
		if len(X) < 2 {
			panic(fmt.Sprintf("Invalid vertex count for POLYLINE: %d. POLYLINE requires at least two vertices.", len(X)))
		}
	default:
		panic(fmt.Sprintf("Unsupported LineType: %v", line.LineType))
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
	var colorToUse [3]float32 = [3]float32{1.0, 1.0, 1.0} // Default color is white
	if len(defaultColor) > 0 {
		colorToUse = defaultColor[0]
	}

	// Error check: Ensure Colors array is a multiple of 3 (RGB per vertex)
	if len(Colors) > 0 && len(Colors)%3 != 0 {
		panic(fmt.Sprintf("Invalid color count: %d. Color array must be a multiple of 3 (R, G, B per vertex).", len(Colors)))
	}

	// Create colors for each vertex
	if len(Colors) > 0 {
		line.Colors = make([]float32, len(Colors))
		copy(line.Colors, Colors)
	} else {
		// Assign the default color to each vertex
		numVertices := len(X)
		line.Colors = make([]float32, numVertices*3)
		for i := 0; i < numVertices; i++ {
			line.Colors[3*i] = colorToUse[0]   // R
			line.Colors[3*i+1] = colorToUse[1] // G
			line.Colors[3*i+2] = colorToUse[2] // B
		}
	}

	// Upload vertex positions to GPU
	gl.BindVertexArray(line.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Vertices)*4, gl.Ptr(line.Vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Upload color data to GPU
	gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Colors)*4, gl.Ptr(line.Colors), gl.STATIC_DRAW)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)

	// Unbind the VAO to avoid unintended modifications
	gl.BindVertexArray(0)
}

// Render draws the line using the shader program stored in Line
func (line *Line) Render(projectionMatrix mgl32.Mat4) {
	// Ensure shader program is active
	gl.UseProgram(line.ShaderProgram)
	gl.BindVertexArray(line.VAO)

	// Upload the projection matrix
	projectionUniform := gl.GetUniformLocation(line.ShaderProgram, gl.Str("projection\x00"))
	if projectionUniform >= 0 {
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projectionMatrix[0])
	} else {
		fmt.Println("Projection matrix uniform not found in Line shader")
	}

	// Draw the line segments
	if line.LineType == utils.LINE {
		gl.DrawArrays(gl.LINES, 0, int32(len(line.Vertices)/2))
		CheckGLError("After draw")
	} else if line.LineType == utils.POLYLINE {
		gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(line.Vertices)/2))
	}
	gl.BindVertexArray(0)
	CheckGLError("After render")
}

func (line *Line) AddShader() (shaderProgram uint32) {
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
	shaderProgram = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
	return
}
