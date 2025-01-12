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
	"github.com/notargets/avs/utils"
)

// Add triangle shader to shader map
func addTriangleMeshShader(shaderMap map[utils.RenderType]uint32) {
	var vertexShader = gl.Str(`
		#version 450
		layout (location = 0) in vec2 position;
		layout (location = 1) in float scalarValue;
		uniform mat4 projection;
		uniform vec3 colorMin;
		uniform vec3 colorMax;
		uniform float scalarMin;
		uniform float scalarMax;
		out vec3 fragColor;

		void main() {
			gl_Position = projection * vec4(position, 0.0, 1.0);

			// Interpolate scalar value into RGB color
			float t = clamp((scalarValue - scalarMin) / (scalarMax - scalarMin), 0.0, 1.0);
			fragColor = mix(colorMin, colorMax, t);
		}` + "\x00")

	var fragmentShader = gl.Str(`
		#version 450
		in vec3 fragColor;
		out vec4 outColor;
		void main() {
			outColor = vec4(fragColor, 1.0);
		}` + "\x00")

	shaderMap[utils.TRIMESHSMOOTH] = compileShaderProgram(vertexShader, fragmentShader)
}

// TriangleMesh represents a batch-rendered triangle mesh
type TriangleMesh struct {
	VAO, VBO, EBO uint32 // OpenGL buffers: Vertex Array, Vertex Buffer, Element Buffer
	ShaderProgram uint32 // Shader program
	VertexCount   int32  // Total number of vertices
	ElementCount  int32  // Total number of elements (indices)
}

// NewTriangleMesh creates and initializes the OpenGL buffers for a triangle mesh
func NewTriangleMesh(vs *geometry.VertexScalar, win *Window) *TriangleMesh {
	// Prepare packed vertex data and indices
	vertexData, indices := packTriangleMeshData(vs)
	triMesh := &TriangleMesh{
		ShaderProgram: win.shaders[utils.TRIMESHSMOOTH],
		VertexCount:   int32(len(vertexData) / 3),
		ElementCount:  int32(len(indices)),
	}

	// Generate and bind OpenGL buffers
	gl.GenVertexArrays(1, &triMesh.VAO)
	gl.GenBuffers(1, &triMesh.VBO)
	gl.GenBuffers(1, &triMesh.EBO)

	gl.BindVertexArray(triMesh.VAO)

	// Upload vertex data (positions + scalar values)
	gl.BindBuffer(gl.ARRAY_BUFFER, triMesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexData)*4, gl.Ptr(vertexData), gl.STATIC_DRAW)

	// Define vertex attributes
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 3*4, unsafe.Pointer(uintptr(0))) // Position (x, y)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 1, gl.FLOAT, false, 3*4, unsafe.Pointer(uintptr(2*4))) // Scalar value
	gl.EnableVertexAttribArray(1)

	// Upload indices
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, triMesh.EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.BindVertexArray(0)

	return triMesh
}

// Render the triangle mesh
func (triMesh *TriangleMesh) Render(colorMin, colorMax [3]float32, scalarMin, scalarMax float32) {
	setShaderProgram(triMesh.ShaderProgram)

	// Set uniforms for color range and scalar range
	gl.Uniform3fv(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("colorMin\x00")), 1, &colorMin[0])
	gl.Uniform3fv(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("colorMax\x00")), 1, &colorMax[0])
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("scalarMin\x00")), scalarMin)
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram, gl.Str("scalarMax\x00")), scalarMax)

	// Draw the mesh
	gl.BindVertexArray(triMesh.VAO)
	gl.DrawElements(gl.TRIANGLES, triMesh.ElementCount, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

// Helper function to pack vertex data and indices
func packTriangleMeshData(vs *geometry.VertexScalar) ([]float32, []uint32) {
	tMesh := vs.TMesh
	vertices := tMesh.XY
	fieldValues := vs.FieldValues

	if len(vertices)/2 != len(fieldValues) {
		panic("Vertex and scalar value counts do not match")
	}

	var vertexData []float32
	for i := 0; i < len(vertices)/2; i++ {
		vertexData = append(vertexData, vertices[2*i], vertices[2*i+1], fieldValues[i])
	}

	var indices []uint32
	for _, tri := range tMesh.TriVerts {
		indices = append(indices, uint32(tri[0]), uint32(tri[1]), uint32(tri[2]))
	}

	return vertexData, indices
}
