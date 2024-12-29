/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"
	"image/color"
	"unsafe"

	"github.com/notargets/avs/utils"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func addLineShader(shaderMap map[utils.RenderType]uint32) {
	// Line shaders
	var vertexShader = gl.Str(`
		#version 450
		layout (location = 0) in vec2 position;
		layout (location = 1) in vec3 color;
		uniform mat4 projection; // add this line
		out vec3 fragColor;
		void main() {
			gl_Position = projection * vec4(position, 0.0, 1.0); // Use projection
			fragColor = color;
		}` + "\x00")

	var fragmentShader = gl.Str(`
		#version 450
		in vec3 fragColor;
		out vec4 outColor;
		void main() {
			outColor = vec4(fragColor, 1.0);
		}` + "\x00")

	shaderMap[utils.LINE] = compileShaderProgram(vertexShader, fragmentShader)
	shaderMap[utils.POLYLINE] = compileShaderProgram(vertexShader,
		fragmentShader)
}

type Line struct {
	VAO, VBO, CBO uint32    // Vertex Array Object, Vertex Buffer Object, Color Buffer Object
	Vertices      []float32 // Flat list of vertex positions [x1, y1, x2, y2, ...]
	Colors        []float32 // Flat list of color data [r1, g1, b1, r2, g2, b2, ...]
	UniColor      bool      // Set if the line color is singular
	LineType      utils.RenderType
	ShaderProgram uint32 // Shader program specific to this Line object
}

func newLine(X, Y []float32, ColorInput interface{}, win *Window,
	rt ...utils.RenderType) (line *Line) {
	var renderType = utils.LINE

	if len(rt) != 0 {
		renderType = utils.POLYLINE
	}
	line = &Line{
		LineType:      renderType,
		ShaderProgram: win.shaders[renderType],
		Vertices:      make([]float32, len(X)*2),
		Colors:        make([]float32, len(X)*3),
	}
	// Determine if we're using a single color for this line
	var defaultColor [3]float32
	switch c := ColorInput.(type) {
	case color.RGBA:
		line.UniColor = true
		defaultColor = [3]float32{float32(c.R / 255), float32(c.G / 255), float32(c.B / 255)}
	case [4]float32:
		line.UniColor = true
		defaultColor = [3]float32{c[0], c[1], c[2]}
	case [3]float32:
		line.UniColor = true
		defaultColor = [3]float32{c[0], c[1], c[2]}
	case []float32:
		line.UniColor = false
		if len(c) != 3*len(X) {
			panic(fmt.Errorf("invalid color input: %v", c))
		}
		line.Colors = ColorInput.([]float32)
	}

	if line.UniColor {
		line.setupVertices(X, Y, nil, defaultColor)
	} else {
		line.setupVertices(X, Y, line.Colors)
	}
	return
}

func (line *Line) setupVertices(X, Y, Colors []float32, defaultColor ...[3]float32) {
	// Error check: Ensure X and Y are of the same length
	if len(X) > 0 && len(Y) > 0 && len(X) != len(Y) {
		panic("X and Y must have the same length if both are provided")
	}

	// Validate vertex count based on LineType
	switch line.LineType {
	case utils.LINE:
		if len(X) > 0 && len(Y) > 0 && len(X)%2 != 0 {
			panic(fmt.Sprintf("Invalid vertex count for LINE: %d. "+
				"Each line segment requires two points (X1, Y1, X2, "+
				"Y2). Vertex count must be a multiple of 2.", len(X)))
		}
	case utils.POLYLINE:
		if len(X) < 2 {
			panic(fmt.Sprintf("Invalid vertex count for POLYLINE: %d. "+
				"POLYLINE requires at least two vertices.", len(X)))
		}
	default:
		panic(fmt.Sprintf("Unsupported LineType: %v", line.LineType))
	}

	// Flatten X and Y into vertex array [x1, y1, x2, y2, ...]
	if len(X) > 0 && len(Y) > 0 {
		for i := 0; i < len(X); i++ {
			line.Vertices[2*i] = X[i]
			line.Vertices[2*i+1] = Y[i]
		}
	}

	// Update colors for each vertex
	if len(defaultColor) > 0 {
		for i := 0; i < len(Colors); i++ {
			line.Colors[i] = defaultColor[0][i%3]
		}
	} else if len(Colors) != 0 {
		for i := 0; i < len(Colors); i++ {
			line.Colors[i] = Colors[i]
		}
	}
}

func (line *Line) setupGPUBuffers() {
	gl.GenVertexArrays(1, &line.VAO)
	CheckGLError("After Gen VAO")
	gl.GenBuffers(1, &line.VBO)
	CheckGLError("After Gen VBO")
	gl.GenBuffers(1, &line.CBO)
	CheckGLError("After Gen CBO")

	gl.BindVertexArray(line.VAO)
	CheckGLError("After Bind VAO")

	gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
	CheckGLError("After Bind VBO")
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Vertices)*4, nil, gl.DYNAMIC_DRAW)
	CheckGLError("After Allocate VBO")
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, unsafe.Pointer(uintptr(0)))
	CheckGLError("After VAO set 1")
	gl.EnableVertexAttribArray(0)
	CheckGLError("After Enable VAO 1")

	gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
	CheckGLError("After Bind CBO")
	gl.BufferData(gl.ARRAY_BUFFER, len(line.Colors)*4, nil, gl.DYNAMIC_DRAW)
	CheckGLError("After Allocate CBO")
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, unsafe.Pointer(uintptr(0)))
	CheckGLError("After VAO set 2")
	gl.EnableVertexAttribArray(1)
	CheckGLError("After Enable VAO 2")

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	CheckGLError("After Unbind VBO")
	gl.BindVertexArray(0)
	CheckGLError("After Unbind VAO")
}

func (line *Line) loadGPUData() {
	// Upload vertex positions to GPU
	gl.BindVertexArray(line.VAO)
	CheckGLError("After Bind VAO")
	gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
	CheckGLError("After Bind VBO")
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(line.Vertices)*4, gl.Ptr(line.Vertices))
	CheckGLError("After Send Vertex Data")
	gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
	CheckGLError("After Bind CBO")
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(line.Colors)*4, gl.Ptr(line.Colors))
	CheckGLError("After Send Color Data")

	// Unbind the VAO to avoid unintended modifications
	gl.BindVertexArray(0)
	CheckGLError("After Unbind VAO")
	gl.BindBuffer(gl.ARRAY_BUFFER, 0) // Unbind VBO
	CheckGLError("After Unbind VBO")
}

// render draws the line using the shader program stored in Line
func (line *Line) render() {
	// Ensure shader program is active
	setShaderProgram(line.ShaderProgram)

	if line.VAO == 0 {
		line.setupGPUBuffers()
	}

	line.loadGPUData()

	gl.BindVertexArray(line.VAO)
	// Draw the line segments
	if line.LineType == utils.LINE {
		gl.DrawArrays(gl.LINES, 0, int32(len(line.Vertices)/2))
	} else if line.LineType == utils.POLYLINE {
		gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(line.Vertices)/2))
	}
	CheckGLError("After draw")
	gl.BindVertexArray(0)
	CheckGLError("After unbind VAO")
}
