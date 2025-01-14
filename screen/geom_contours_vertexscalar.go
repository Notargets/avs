/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package screen

import (
	"unsafe"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/notargets/avs/geometry"
	utils "github.com/notargets/avs/utils"
)

func addContourVertexScalarShader(shaderMap map[utils.RenderType]uint32) {
	var vertexShader = gl.Str(`
		#version 450
		layout (location = 0) in vec2 position;
		layout (location = 1) in float scalarValue;
		uniform mat4 projection;
		uniform float scalarMin;
		uniform float scalarMax;
		out float v_scalar;

		void main() {
			gl_Position = projection * vec4(position, 0.0, 1.0);
			// Normalize scalarValue to [0, 1] and pass to the fragment shader
			v_scalar = clamp((scalarValue - scalarMin) / (scalarMax - scalarMin), 0.0, 1.0);
		}` + "\x00")

	var fragmentShader = gl.Str(`
		#version 450
		in float v_scalar;
		out vec4 outColor;

		layout(std140) uniform IsoData {
			int numIsoContours;         // Number of iso-contours
			float isoLevels[256];       // Iso-level values
			vec3 isoColors[256];        // Iso-contour colors
		};
		uniform float isoThickness;   // Thickness of iso-contours

		void main() {
			vec4 color = vec4(0.0);    // Default background color
			for (int i = 0; i < numIsoContours; i++) {
				float distance = abs(v_scalar - isoLevels[i]);
				if (distance < isoThickness) {
					color = vec4(isoColors[i], 1.0); // Assign the iso-color
					break;
				}
			}
			outColor = color;          // Output the fragment color
		}` + "\x00")

	shaderMap[utils.TRIMESHCONTOURS] = compileShaderProgram(vertexShader, fragmentShader)
}

type ContourVertexScalar struct {
	VAO, VBO             uint32 // OpenGL buffers: Vertex Array, Vertex Buffer
	ContourUBO           *IsoContourUBO
	ShaderProgram        uint32 // Shader program
	NumVertices          int32
	vertexData           []float32
	colorMin, colorMax   [3]float32
	scalarMin, scalarMax float32
}

// NewContourVertexScalar creates and initializes the OpenGL buffers for a triangle mesh
func newContourVertexScalar(vs *geometry.VertexScalar, win *Window,
	fMin, fMax float32) *ContourVertexScalar {
	triMesh := &ContourVertexScalar{
		ShaderProgram: win.shaders[utils.TRIMESHCONTOURS],
		// Each vertex has 2 coords + 1 scalar
		NumVertices: int32(len(vs.TMesh.TriVerts) * 3), // Num tris x 3 verts
		colorMin:    [3]float32{0, 0, 1},
		colorMax:    [3]float32{1, 0, 0},
		scalarMin:   fMin,
		scalarMax:   fMax,
	}
	triMesh.vertexData = make([]float32, triMesh.NumVertices*3)

	// Create UBO for iso-contours
	numContours := 20
	fStep := (triMesh.scalarMax - triMesh.scalarMin) / float32(numContours-1)
	isoLevels := make([]float32, numContours)
	for i := 0; i < numContours; i++ {
		isoLevels[i] = triMesh.scalarMin + float32(i)*fStep
	}
	isoColors := [][3]float32{
		{1.0, 0.0, 0.0}, // Red
		{0.0, 1.0, 0.0}, // Green
		{0.0, 0.0, 1.0}, // Blue
		{1.0, 1.0, 0.0}, // Yellow
		{1.0, 0.0, 1.0}, // Magenta
	}
	triMesh.ContourUBO = newIsoContourUBO(isoLevels, isoColors)

	// Generate and bind OpenGL buffers
	gl.GenVertexArrays(1, &triMesh.VAO)
	gl.GenBuffers(1, &triMesh.VBO)

	gl.BindVertexArray(triMesh.VAO)

	// Allocate Buffers
	gl.BindBuffer(gl.ARRAY_BUFFER, triMesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(triMesh.vertexData)*4,
		nil, gl.DYNAMIC_DRAW)

	// Define vertex attributes
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 3*4,
		unsafe.Pointer(uintptr(0))) // Position (x, y)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 1, gl.FLOAT, false, 3*4,
		unsafe.Pointer(uintptr(2*4))) // Scalar value
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	triMesh.updateVertexScalarData(vs)

	return triMesh
}

func (triMesh *ContourVertexScalar) updateVertexScalarData(vs *geometry.VertexScalar) {
	triMesh.vertexData = packVertexScalarData(vs)
	// Upload vertex data (positions + scalar values)
	gl.BindVertexArray(triMesh.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, triMesh.VBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(triMesh.vertexData)*4,
		gl.Ptr(triMesh.vertexData))
	gl.BindVertexArray(0)
}

func (triMesh *ContourVertexScalar) render() {
	setShaderProgram(triMesh.ShaderProgram)

	// Update scalar range uniforms
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("scalarMin\x00")), triMesh.scalarMin)
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("scalarMax\x00")), triMesh.scalarMax)
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("isoThickness\x00")), 0.2) // Example thickness

	// Bind UBO
	gl.BindBufferBase(gl.UNIFORM_BUFFER, 0, triMesh.ContourUBO.UBO)

	// Draw the mesh
	gl.BindVertexArray(triMesh.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, triMesh.NumVertices)
	gl.BindVertexArray(0)
}

type IsoContourUBO struct {
	UBO         uint32
	NumContours int
	IsoLevels   []float32
	IsoColors   [][3]float32
}

func newIsoContourUBO(levels []float32, colors [][3]float32) *IsoContourUBO {
	ubo := &IsoContourUBO{
		IsoLevels:   levels,
		IsoColors:   colors,
		NumContours: len(levels),
	}
	gl.GenBuffers(1, &ubo.UBO)
	ubo.update()
	return ubo
}

func (ubo *IsoContourUBO) update() {
	gl.BindBuffer(gl.UNIFORM_BUFFER, ubo.UBO)
	bufferSize := int32(4 + len(ubo.IsoLevels)*4 + len(ubo.IsoColors)*12)
	data := make([]byte, bufferSize)

	// Pack data
	offset := 0
	copy(data[offset:], utils.Int32ToBytes(int32(ubo.NumContours)))
	offset += 4
	for _, level := range ubo.IsoLevels {
		copy(data[offset:], utils.Float32ToBytes(level))
		offset += 4
	}
	for _, color := range ubo.IsoColors {
		copy(data[offset:], utils.Float32ToBytes(color[0]))
		offset += 4
		copy(data[offset:], utils.Float32ToBytes(color[1]))
		offset += 4
		copy(data[offset:], utils.Float32ToBytes(color[2]))
		offset += 4
	}

	// Upload data to the GPU
	gl.BufferData(gl.UNIFORM_BUFFER, len(data), gl.Ptr(data), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, 0, ubo.UBO)
}
