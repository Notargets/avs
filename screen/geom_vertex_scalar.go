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

// Add shaded triangle mesh shader
func addShadedTriMeshShader(shaderMap map[utils.RenderType]uint32) {
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

// ShadedTriMesh represents a batch-rendered triangle mesh
type ShadedTriMesh struct {
	VAO, VBO      uint32 // OpenGL buffers: Vertex Array, Vertex Buffer
	ShaderProgram uint32 // Shader program
	NumVertices   int32
	vertexData    []float32
}

// NewShadedTriMesh creates and initializes the OpenGL buffers for a triangle mesh
func NewShadedTriMesh(vs *geometry.VertexScalar, win *Window) *ShadedTriMesh {
	triMesh := &ShadedTriMesh{
		ShaderProgram: win.shaders[utils.TRIMESHSMOOTH],
		// Each vertex has 2 coords + 1 scalar
		NumVertices: int32(len(vs.TMesh.TriVerts) * 3), // Num tris x 3 verts
	}
	triMesh.vertexData = make([]float32, triMesh.NumVertices*3)
	triMesh.UpdateTriMeshData(vs)

	// Generate and bind OpenGL buffers
	gl.GenVertexArrays(1, &triMesh.VAO)
	gl.GenBuffers(1, &triMesh.VBO)

	gl.BindVertexArray(triMesh.VAO)

	// Upload vertex data (positions + scalar values)
	gl.BindBuffer(gl.ARRAY_BUFFER, triMesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(triMesh.vertexData)*4,
		gl.Ptr(triMesh.vertexData), gl.STATIC_DRAW)

	// Define vertex attributes
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 3*4,
		unsafe.Pointer(uintptr(0))) // Position (x, y)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 1, gl.FLOAT, false, 3*4,
		unsafe.Pointer(uintptr(2*4))) // Scalar value
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	return triMesh
}

func (triMesh *ShadedTriMesh) UpdateTriMeshData(vs *geometry.VertexScalar) {
	triMesh.vertexData = packShadedTriMeshData(vs)
}

// Render the triangle mesh
func (triMesh *ShadedTriMesh) Render(colorMin, colorMax [3]float32, scalarMin,
	scalarMax float32) {
	setShaderProgram(triMesh.ShaderProgram)

	// Set uniforms for color range and scalar range
	gl.Uniform3fv(gl.GetUniformLocation(triMesh.ShaderProgram,
		gl.Str("colorMin\x00")), 1, &colorMin[0])
	gl.Uniform3fv(gl.GetUniformLocation(triMesh.ShaderProgram,
		gl.Str("colorMax\x00")), 1, &colorMax[0])
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram,
		gl.Str("scalarMin\x00")), scalarMin)
	gl.Uniform1f(gl.GetUniformLocation(triMesh.ShaderProgram,
		gl.Str("scalarMax\x00")), scalarMax)

	// Draw the mesh
	gl.BindVertexArray(triMesh.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, triMesh.NumVertices)
	gl.BindVertexArray(0)
}

// Helper function to pack vertex data
func packShadedTriMeshData(vs *geometry.VertexScalar) []float32 {
	tMesh := vs.TMesh
	coordinates := tMesh.XY
	fieldValues := vs.FieldValues

	numVertices := len(tMesh.TriVerts) * 3
	// Pre-allocate memory for vertex data
	vertexData := make([]float32,
		numVertices*3) // Each vertex has 2 coords + 1 scalar

	// Pack vertex data
	var vert int
	for _, triVert := range tMesh.TriVerts {
		for n := 0; n < 3; n++ {
			vertexData[vert*3+0] = coordinates[triVert[n]*2]   // x
			vertexData[vert*3+1] = coordinates[triVert[n]*2+1] // y
			vertexData[vert*3+2] = fieldValues[vert]           // scalar
			vert++
		}
	}

	return vertexData
}
