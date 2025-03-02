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

func addContourVertexScalarShader(shaderMap map[utils.RenderType]uint32) {
	var vertexShader = gl.Str(`
			#version 450
			layout (location = 0) in vec2 position;
			layout (location = 1) in float scalarValue;
			uniform mat4 projection;
			out float v_scalar;        // Interpolated scalar value
			out vec2 v_position;       // Vertex position

			void main() {
    			gl_Position = projection * vec4(position, 0.0, 1.0);
    			v_scalar = scalarValue;  // Pass the scalar value
    			v_position = position;   // Pass the position
		}` + "\x00")

	var geometryShader = gl.Str(`
			#version 450
			layout (triangles) in;
			layout (line_strip, max_vertices = 4) out;

			uniform IsoData {
    			int numIsoContours;         // Number of iso-contours
    			float isoLevels[256];       // Iso-level values
			};

			uniform float scalarMin;       // Minimum scalar value in the field
			uniform float scalarMax;       // Maximum scalar value in the field

			in float v_scalar[];            // Scalars passed from vertex shader
			out vec4 lineColor;             // Line color output

			const vec3 colormap[5] = vec3[](
    			vec3(0.0, 0.0, 1.0), // Blue
    			vec3(0.0, 1.0, 1.0), // Cyan
    			vec3(0.0, 1.0, 0.0), // Green
    			vec3(1.0, 1.0, 0.0), // Yellow
    			vec3(1.0, 0.0, 0.0)  // Red
			);

			void main() {
    			for (int isoIndex = 0; isoIndex < numIsoContours; isoIndex++) {
        			float isoLevel = isoLevels[isoIndex];

        			vec4 crossingPoints[2];
        			int crossingCount = 0;

        			// Check all edges for iso-level crossings
        			for (int edge = 0; edge < 3; edge++) {
            			int v1 = edge;
            			int v2 = (edge + 1) % 3;

            			float scalar1 = v_scalar[v1];
            			float scalar2 = v_scalar[v2];

            			// Check if iso-level crosses the edge
            			if ((scalar1 > isoLevel) != (scalar2 > isoLevel)) {
                			float t = (isoLevel - scalar1) / (scalar2 - scalar1); // Linear interpolation
                			crossingPoints[crossingCount++] = mix(gl_in[v1].gl_Position, gl_in[v2].gl_Position, t);
            			}
        			}

        			// Emit a line if exactly two crossings are found
        			if (crossingCount == 2) {
            			// Normalize the iso-level to [0, 1]
            			float normalizedIso = (isoLevel - scalarMin) / (scalarMax - scalarMin);
            			normalizedIso = clamp(normalizedIso, 0.0, 1.0);

            			// Determine colormap color
            			float t = normalizedIso * 4.0; // Map to [0, 4]
            			int index = int(floor(t));    // Lower index
            			float mixFactor = t - float(index); // Fractional part for interpolation
            			vec3 color = mix(colormap[index], colormap[index + 1], mixFactor);

            			// Pass color to fragment shader
            			lineColor = vec4(color, 1.0);

            			gl_Position = crossingPoints[0];
            			EmitVertex();

            			gl_Position = crossingPoints[1];
            			EmitVertex();

            			EndPrimitive();
        			}
    			}
		}` + "\x00")

	var fragmentShader = gl.Str(`
			#version 450
			in vec4 lineColor;  // Color passed from the geometry shader
			out vec4 outColor;

			void main() {
    			outColor = lineColor;
		}` + "\x00")

	shaderMap[utils.TRIMESHCONTOURS] = compileShaderProgram(vertexShader,
		fragmentShader, geometryShader)
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
	fMin, fMax float32, numContours int) *ContourVertexScalar {
	triMesh := &ContourVertexScalar{
		ShaderProgram: win.shaders[utils.TRIMESHCONTOURS],
		// Each vertex has 2 coords + 1 scalar
		NumVertices: int32(len(vs.TMesh.TriVerts) * 3), // Num tris x 3 verts
		scalarMin:   fMin,
		scalarMax:   fMax,
	}
	triMesh.vertexData = make([]float32, triMesh.NumVertices*3)

	// Create UBO for iso-contours
	fStep := (fMax - fMin) / float32(numContours-1)
	isoLevels := make([]float32, numContours)
	for i := 0; i < numContours; i++ {
		isoLevels[i] = fMin + float32(i)*fStep
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

	// Bind UBO for iso-levels
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
