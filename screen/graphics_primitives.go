package screen

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (scr *Screen) SetObjectActive(key Key, active bool) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			renderable.Active = active
			scr.Objects[key] = renderable
		}
	}
}

type Line struct {
	VAO, VBO, CBO uint32    // Vertex Array Object, Vertex Buffer Object, Color Buffer Object
	Vertices      []float32 // Flat list of vertex positions [x1, y1, x2, y2, ...]
	Colors        []float32 // Flat list of color data [r1, g1, b1, r2, g2, b2, ...]
	ShaderProgram uint32    // Shader program specific to this Line object
	LineType      RenderType
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
	case LINE:
		if len(X) > 0 && len(Y) > 0 && len(X)%2 != 0 {
			panic(fmt.Sprintf("Invalid vertex count for LINE: %d. Each line segment requires two points (X1, Y1, X2, Y2). Vertex count must be a multiple of 2.", len(X)))
		}
	case POLYLINE:
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
	if line.LineType == LINE {
		gl.DrawArrays(gl.LINES, 0, int32(len(line.Vertices)/2))
		checkGLError("After draw")
	} else if line.LineType == POLYLINE {
		gl.DrawArrays(gl.LINE_STRIP, 0, int32(len(line.Vertices)/2))
	}
	gl.BindVertexArray(0)
	checkGLError("After render")
}

func (scr *Screen) AddPolyLine(key Key, X, Y, Colors []float32) (newKey Key) {
	return scr.AddLine(key, X, Y, Colors, POLYLINE)
}

func (scr *Screen) AddLine(key Key, X, Y, Colors []float32, rt ...RenderType) (newKey Key) {
	if key == NEW {
		key = NewKey()
	}
	newKey = key

	var renderType = LINE
	if len(rt) != 0 {
		renderType = POLYLINE
	}

	// Send a command to create or update a line object
	scr.RenderChannel <- func() {
		var line *Line

		// Check if the object exists in the scene
		if existingRenderable, exists := scr.Objects[key]; exists {
			line = existingRenderable.Object.(*Line)
		} else {
			// Create new line
			line = &Line{LineType: renderType}
			line.ShaderProgram = line.addShader(scr)
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

func (line *Line) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[LINE]; !present {
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
	}
	return scr.Shaders[LINE]
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

//func (scr *Screen) InitShaders() {
//
//	i := 0
//	{
//		//for i := int(LINE); i <= int(TRIMESHSMOOTH3D); i++ {
//		switch RenderType(i) {
//		case LINE:
//			// Line Shaders
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec2 position;
//layout (location = 1) in vec3 color;
//uniform mat4 projection; // Add this line
//out vec3 fragColor;
//void main() {
//	gl_Position = projection * vec4(position, 0.0, 1.0); // Use projection
//	fragColor = color;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in vec3 fragColor;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragColor, 1.0);
//}
//` + "\x00"
//			scr.Shaders[LINE] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHEDGESUNICOLOR:
//			// TriMeshEdgesUniColor Shaders (for edges with uniform color)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//void main() {
//	gl_Position = vec4(position, 1.0);
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//out vec4 outColor;
//void main() {
//	outColor = vec4(0.0, 0.0, 0.0, 1.0); // Uniform black color for edges
//}
//` + "\x00"
//			scr.Shaders[TRIMESHEDGESUNICOLOR] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHEDGES:
//			// TriMeshEdges Shaders (for edges with per-vertex colors)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in vec4 color;
//out vec4 fragColor;
//void main() {
//	gl_Position = vec4(position, 1.0);
//	fragColor = color;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in vec4 fragColor;
//out vec4 outColor;
//void main() {
//	outColor = fragColor;
//}
//` + "\x00"
//			scr.Shaders[TRIMESHEDGES] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHCONTOURS:
//			// TriMeshContours Shaders (for contour shading)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in float scalar;
//out float fragValue;
//void main() {
//	gl_Position = vec4(position, 1.0);
//	fragValue = scalar;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in float fragValue;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragValue, fragValue, fragValue, 1.0);
//}
//` + "\x00"
//			scr.Shaders[TRIMESHCONTOURS] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHSMOOTH:
//			// TriMeshSmooth Shaders (for smooth shaded triangles)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in vec3 color;
//out vec3 fragColor;
//void main() {
//	gl_Position = vec4(position, 1.0);
//	fragColor = color;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in vec3 fragColor;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragColor, 1.0);
//}
//` + "\x00"
//			scr.Shaders[TRIMESHSMOOTH] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case LINE3D:
//			// 3D Line Shaders
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in vec3 color;
//out vec3 fragColor;
//uniform mat4 projection;
//void main() {
//	gl_Position = projection * vec4(position, 1.0);
//	fragColor = color;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in vec3 fragColor;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragColor, 1.0);
//}
//` + "\x00"
//			scr.Shaders[LINE3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHEDGESUNICOLOR3D:
//			// TriMeshEdgesUniColor3D Shaders
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//uniform mat4 projection;
//void main() {
//	gl_Position = projection * vec4(position, 1.0);
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//out vec4 outColor;
//void main() {
//	outColor = vec4(0.0, 0.0, 0.0, 1.0); // Black for edges
//}
//` + "\x00"
//			scr.Shaders[TRIMESHEDGESUNICOLOR3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHEDGES3D:
//			// TriMeshEdges3D Shaders (for edges with per-vertex colors)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in vec3 color;
//uniform mat4 projection;
//out vec3 fragColor;
//void main() {
//	gl_Position = projection * vec4(position, 1.0);
//	fragColor = color;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in vec3 fragColor;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragColor, 1.0);
//}
//` + "\x00"
//			scr.Shaders[TRIMESHEDGES3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHCONTOURS3D:
//			// TriMeshContours3D Shaders (for contour shading)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in float scalar;
//uniform mat4 projection;
//out float fragValue;
//void main() {
//	gl_Position = projection * vec4(position, 1.0);
//	fragValue = scalar;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in float fragValue;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragValue, fragValue, fragValue, 1.0);
//}
//` + "\x00"
//			scr.Shaders[TRIMESHCONTOURS3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//
//		case TRIMESHSMOOTH3D:
//			// TriMeshSmooth3D Shaders (for smooth shaded triangles)
//			var vertexShaderSource = `
//#version 450
//layout (location = 0) in vec3 position;
//layout (location = 1) in vec3 color;
//uniform mat4 projection;
//out vec3 fragColor;
//void main() {
//	gl_Position = projection * vec4(position, 1.0);
//	fragColor = color;
//}
//` + "\x00"
//
//			var fragmentShaderSource = `
//#version 450
//in vec3 fragColor;
//out vec4 outColor;
//void main() {
//	outColor = vec4(fragColor, 1.0);
//}
//` + "\x00"
//			scr.Shaders[TRIMESHSMOOTH3D] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
//		}
//	}
//	return
//}

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
