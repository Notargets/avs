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
		in float v_scalar;              // Normalized scalar value from the vertex shader
		out vec4 outColor;              // Output fragment color
		
		layout(std140) uniform IsoData {
    		int numIsoContours;         // Number of iso-contours
    		float isoLevels[256];       // Iso-level values
		};

		uniform float isoThickness;     // Thickness of iso-contours

		// Hardcoded 5-point linear colormap
		const vec3 colormapPoints[5] = vec3[](
    		vec3(1.0, 0.0, 0.0), // Red
    		vec3(1.0, 1.0, 0.0), // Yellow
    		vec3(0.0, 1.0, 0.0), // Green
    		vec3(0.0, 1.0, 1.0), // Cyan
    		vec3(0.0, 0.0, 1.0)  // Blue
		);

		void main() {
    		vec3 baseColor = vec3(0.0);  // Default background color

    		// 1. Perform colormap interpolation based on v_scalar
    		for (int i = 0; i < 4; i++) {
        		float lowerBound = float(i) / 4.0;
        		float upperBound = float(i + 1) / 4.0;

        		if (v_scalar >= lowerBound && v_scalar <= upperBound) {
            		float t = (v_scalar - lowerBound) / (upperBound - lowerBound);
            		baseColor = mix(colormapPoints[i], colormapPoints[i + 1], t);
            		break;
        		}
    		}

    		// 2. Check if the scalar value is near any iso-level
    		vec4 color = vec4(baseColor, 1.0); // Default to base color
    		for (int i = 0; i < numIsoContours; i++) {
        		float distance = abs(v_scalar - isoLevels[i]);
        		if (distance < isoThickness) {
            		color = vec4(1.0, 1.0, 1.0, 1.0); // White for iso-lines
            		break;  // Exit the loop after finding a matching iso-level
        		}
    		}

    		outColor = color;  // Final output color
		}` + "\x00")

	shaderMap[utils.TRIMESHCONTOURS] = compileShaderProgram(vertexShader, fragmentShader)
}

type ContourVertexScalar struct {
	VAO, VBO             uint32 // OpenGL buffers: Vertex Array, Vertex Buffer
	ContourUBO           *IsoContourUBO
	ShaderProgram        uint32 // Shader program
	NumVertices          int32
	vertexData           []float32
	scalarMin, scalarMax float32
}

// NewContourVertexScalar creates and initializes the OpenGL buffers for a triangle mesh
func newContourVertexScalar(vs *geometry.VertexScalar, win *Window,
	fMin, fMax float32) *ContourVertexScalar {
	triMesh := &ContourVertexScalar{
		ShaderProgram: win.shaders[utils.TRIMESHCONTOURS],
		// Each vertex has 2 coords + 1 scalar
		NumVertices: int32(len(vs.TMesh.TriVerts) * 3), // Num tris x 3 verts
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
	triMesh.ContourUBO = newIsoContourUBO(isoLevels)

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
}

func newIsoContourUBO(levels []float32) *IsoContourUBO {
	ubo := &IsoContourUBO{
		IsoLevels:   levels,
		NumContours: len(levels),
	}

	gl.GenBuffers(1, &ubo.UBO)
	ubo.update()
	return ubo
}

func (ubo *IsoContourUBO) update() {
	// Calculate the buffer size: 4 bytes for the number of contours + 4 bytes per iso-level
	bufferSize := 4 + len(ubo.IsoLevels)*4
	data := make([]byte, bufferSize)

	// Pack data into the buffer
	offset := 0
	copy(data[offset:], utils.Int32ToBytes(int32(ubo.NumContours))) // Number of iso-contours
	offset += 4
	for _, level := range ubo.IsoLevels { // Iso-level values
		copy(data[offset:], utils.Float32ToBytes(level))
		offset += 4
	}

	// Upload data to the GPU
	gl.BindBuffer(gl.UNIFORM_BUFFER, ubo.UBO)
	gl.BufferData(gl.UNIFORM_BUFFER, len(data), gl.Ptr(data), gl.DYNAMIC_DRAW)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, 0, ubo.UBO)
}
