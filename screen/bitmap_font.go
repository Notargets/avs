package screen

import (
	"fmt"
	"image"
	"runtime"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/assets"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type String struct {
	VAO, VBO                    uint32
	Text                        string
	ShaderProgram               uint32
	Position                    mgl32.Vec2
	Texture                     uint32
	StringType                  utils.RenderType
	polygonVertices             [4]mgl32.Vec2
	projectedVertices           []float32
	initializedFIXEDSTRING      bool
	textureImg                  *image.RGBA
	textureWidth, textureHeight int
	TextFormatter               *assets.TextFormatter
}

func printMemoryStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("[%s] Memory Usage: Alloc = %v MB, TotalAlloc = %v MB, Sys = %v MB, NumGC = %v\n",
		label, m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

func (str *String) setupTextureMap(scr *Screen) {
	var (
		tf = str.TextFormatter
	)

	// Load our texture map with the drawn text if not done already
	if str.textureImg == nil {
		var img *image.RGBA
		img = str.TextFormatter.TypeFace.RenderFontTextureImg(str.Text, str.TextFormatter.Color)
		str.textureImg = img
		str.textureWidth, str.textureHeight = str.textureImg.Bounds().Dx(), str.textureImg.Bounds().Dy()
	}

	// Update vertex coordinates for STRING, FIXEDSTRING only does this once
	if str.StringType == utils.STRING || !str.initializedFIXEDSTRING {
		x := str.Position.X()
		y := str.Position.Y()
		quadWidth, quadHeight := calculateQuadBounds(str.textureWidth, str.textureHeight, scr.WindowWidth, scr.WindowHeight,
			tf.TypeFace.FontDPI, scr.XMax-scr.XMin, scr.YMax-scr.YMin)

		// **Step 4: Calculate proper position and scale**
		var posX, posY float32
		if str.TextFormatter.Centered {
			posX = x - quadWidth/2
			posY = y - quadHeight/2
		} else {
			posX = x
			posY = y
		}

		// **Step 5: Initialize polygon vertices for the 4 corners of the quad**
		str.polygonVertices = [4]mgl32.Vec2{
			{posX, posY},                          // Bottom-left
			{posX + quadWidth, posY},              // Bottom-right
			{posX, posY + quadHeight},             // Top-left
			{posX + quadWidth, posY + quadHeight}, // Top-right
		}
	}

	// Set the color
	c := ColorToFloat32(str.TextFormatter.Color)
	if str.StringType == utils.STRING {
		lenRow := 2 + 3
		lenV := 4 * (lenRow)
		if len(str.projectedVertices) == 0 {
			str.projectedVertices = make([]float32, lenV)
		}
		for i := 0; i < 4; i++ {
			str.projectedVertices[i*lenRow] = str.polygonVertices[i][0]
			str.projectedVertices[i*lenRow+1] = str.polygonVertices[i][1]
			str.projectedVertices[i*lenRow+2] = c[0]
			str.projectedVertices[i*lenRow+3] = c[1]
			str.projectedVertices[i*lenRow+4] = c[2]
		}
	} else if str.StringType == utils.FIXEDSTRING {
		// Calculate fixed position in NDC coordinates once, via the initial projection matrix
		// This puts the location into fixed pixel coordinates mapped to the window via the ortho projection
		lenRow := 4 + 3
		lenV := 4 * (lenRow)
		if !str.initializedFIXEDSTRING {
			//fmt.Printf("Rendering FIXED STRING...\n")
			var NDCVertexCoordinates [4]mgl32.Vec4
			for i := 0; i < 4; i++ {
				NDCVertexCoordinates[i] = scr.projectionMatrix.Mul4x1(mgl32.Vec4{str.polygonVertices[i].X(), str.polygonVertices[i].Y(), 0.0, 1.0})
				NDCVertexCoordinates[i] = NDCVertexCoordinates[i].Mul(1.0 / NDCVertexCoordinates[i].W())
			}
			if len(str.projectedVertices) == 0 {
				str.projectedVertices = make([]float32, lenV)
			}
			// Color is updated below
			for i := 0; i < 4; i++ {
				str.projectedVertices[i*lenRow] = NDCVertexCoordinates[i][0]   // Clip space coordinates for fixed position
				str.projectedVertices[i*lenRow+1] = NDCVertexCoordinates[i][1] // Clip space coordinates for fixed position
				str.projectedVertices[i*lenRow+2] = NDCVertexCoordinates[i][2] // Clip space coordinates for fixed position
				str.projectedVertices[i*lenRow+3] = NDCVertexCoordinates[i][3] // Clip space coordinates for fixed position
			}
			str.initializedFIXEDSTRING = true // initialization is finished after this
		}
		// Update the color fields every time, in case the color is changed
		for i := 0; i < 4; i++ {
			str.projectedVertices[i*lenRow+4] = c[0]
			str.projectedVertices[i*lenRow+5] = c[1]
			str.projectedVertices[i*lenRow+6] = c[2]
		}
	}

	str.loadGPUData(scr, str.textureImg, str.textureWidth, str.textureHeight, c)
}

func (str *String) loadGPUData(scr *Screen, img *image.RGBA, textureWidth, textureHeight int, color [4]float32) {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	//fmt.Printf("Texture width: %d, Texture height: %d\n", textureWidth, textureHeight)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(textureWidth), int32(textureHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	checkGLError("After TexImage2D")
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	str.Texture = texture

	// Generate VBO and VAO once
	gl.GenBuffers(1, &str.VBO)
	gl.GenVertexArrays(1, &str.VAO)

	// Bind VAO
	gl.BindVertexArray(str.VAO)

	// Bind VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)

	// Upload vertex data
	gl.BufferData(gl.ARRAY_BUFFER, len(str.projectedVertices)*4, gl.Ptr(str.projectedVertices), gl.STATIC_DRAW)
	checkGLError("After VBO")

	// **Setup Vertex Attributes**

	// **PositionDelta (location = 0)**
	var stride int32
	if str.StringType == utils.STRING {
		//stride = 4 * (2 + 2 + 3)
		stride = 4 * (2 + 3)
	} else {
		stride = 4 * (4 + 3)
	}
	offset := 0
	if str.StringType == utils.STRING {
		// Load the transformed vertex coordinates
		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // PositionDelta (2 floats)
		offset += 2 * 4                                                             // Advance by 2 floats = 8 bytes
	} else if str.StringType == utils.FIXEDSTRING {
		// **Frozen PositionDelta **
		// Load the NDC fixed vertex coordinates
		gl.VertexAttribPointer(0, 4, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // Fixed PositionDelta (4 floats)
		offset += 4 * 4                                                             // Advance by 2 floats = 8 bytes
	}
	gl.EnableVertexAttribArray(0)

	// **Color **
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // Color (3 floats)
	gl.EnableVertexAttribArray(1)
	offset += 3 * 4 // Advance by 3 floats = 12 bytes

}

func calculateQuadBounds(textureWidth, textureHeight, windowWidth, windowHeight, fontDPI int,
	xRange, yRange float32) (quadWidth, quadHeight float32) {
	// Calculate the width of the text string in window coordinates based on the fact that the xRange corresponds
	// with the window width
	// First, percentage of width covered by the text pixels:
	windowPercent := float32(textureWidth) / float32(windowWidth)
	bitmapAspectRatio := float32(textureHeight) / float32(textureWidth)
	// Now how much world space this represents
	worldSpaceWidth := windowPercent * xRange
	worldSpaceHeight := bitmapAspectRatio * worldSpaceWidth
	// Now correct the worldSpaceHeight to remove the stretch factor of the ortho transform
	var ratio float32
	ratio = yRange / xRange
	worldSpaceHeight *= ratio
	// Implement a scale factor to reduce the polygon size commensurate with the dynamic DPI scaling, relative to the
	// standard 72 DPI of the Opentype package
	scaleFromDPI := 72 / float32(fontDPI)
	winRatio := float32(1.)
	// Correct for image scaling when window is resized and width is < height
	if windowWidth < windowHeight {
		winRatio = float32(windowWidth) / float32(windowHeight)
	}
	quadWidth = winRatio * scaleFromDPI * worldSpaceWidth
	quadHeight = winRatio * scaleFromDPI * worldSpaceHeight
	return
}

func (str *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[str.StringType]; !present {
		var vertexShaderSource string
		switch str.StringType {
		case utils.STRING:
			//fmt.Printf("Adding shader: %s\n", str.StringType)
			vertexShaderSource = `
			#version 450

			layout (location = 0) in vec2 position; // Position input from the vertex buffer
			layout (location = 1) in vec3 color;    // Color input from the vertex buffer

			uniform mat4 projection; // Projection matrix

			// Constant array of vec2 representing the UV coordinates for 4 vertices
			const vec2 uv[4] = vec2[4](
    			vec2(0.0, 1.0), // Top-left
    			vec2(1.0, 1.0), // Top-right
    			vec2(0.0, 0.0), // Bottom-left
    			vec2(1.0, 0.0)  // Bottom-right
			);

			out vec2 fragUV;
			out vec3 fragColor;

			void main() {
    			gl_Position = projection * vec4(position, 0.0, 1.0); // Apply projection matrix
    			fragUV = uv[gl_VertexID % 4]; // Select UV coordinate based on gl_VertexID (assumes quads)
    			fragColor = color;
			}` + "\x00"
		case utils.FIXEDSTRING:
			//fmt.Printf("Adding shader: %s\n", str.StringType)
			vertexShaderSource = `
				#version 450
				layout (location = 0) in vec4 NDCposition;
				layout (location = 1) in vec3 color;

			    // Constant array of vec2 representing the UV coordinates for 4 vertices
			    const vec2 uv[4] = vec2[4](
    			    vec2(0.0, 1.0), // Top-left
    			    vec2(1.0, 1.0), // Top-right
    			    vec2(0.0, 0.0), // Bottom-left
    			    vec2(1.0, 0.0)  // Bottom-right
			    );

				out vec2 fragUV;
				out vec3 fragColor;

				void main() {
    				gl_Position = NDCposition; // Use the NDS position directly (clip space)
    			    fragUV = uv[gl_VertexID % 4]; // Select UV coordinate based on gl_VertexID (assumes quads)
    				fragColor = color;
				}` + "\x00"
		default:
			panic(fmt.Errorf("unknown shader type %v", str.StringType))
		}

		fragmentShaderSource := `
		#version 450
		in vec2 fragUV;
		in vec3 fragColor;
		uniform sampler2D fontTexture;
		out vec4 outColor;

		void main() {
			vec4 texColor = texture(fontTexture, fragUV);
			outColor = texColor * vec4(fragColor, texColor.a);
		}` + "\x00"

		scr.Shaders[str.StringType] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
		checkGLError("After compileShaderProgram")
	}
	return scr.Shaders[str.StringType]
}

func (str *String) Render(scr *Screen) {
	str.setupTextureMap(scr)

	//fmt.Printf("Rendering %s\n", str.StringType)
	gl.UseProgram(scr.Shaders[str.StringType])
	checkGLError("After UseProgram")

	// Check if the active program matches
	var activeProgram int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &activeProgram)
	if uint32(activeProgram) != scr.Shaders[str.StringType] {
		fmt.Printf("[Render] Shader program mismatch! Active: %d, Expected: %d\n", activeProgram, scr.Shaders[str.StringType])
		panic("[Render] Shader program is not active as expected")
	}

	if scr.Shaders[str.StringType] == 0 {
		fmt.Println("[Render] Shader program handle is 0. Possible compilation/linking failure.")
		panic("[Render] Shader program handle is 0")
	}

	// Bind the projection matrix to the shader
	if str.StringType == utils.STRING {
		projectionUniform := gl.GetUniformLocation(scr.Shaders[str.StringType], gl.Str("projection\x00"))
		checkGLError("After GetUniformLocation")
		if projectionUniform < 0 {
			fmt.Printf("[Render] Projection uniform not found for String Type: %s!", str.StringType.String())
			panic("[Render] Projection uniform location returned -1")
		}
		gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
		checkGLError("After UniformMatrix4fv")
	}

	// Bind the texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, str.Texture)
	checkGLError("After BindTexture")

	// Bind the VAO and draw the polygon
	gl.BindVertexArray(str.VAO)
	checkGLError("After BindVertexArray")

	// Enable Blending
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	checkGLError("After BlendFunc")

	// Draw the quad (TRIANGLE_STRIP for simplicity)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	checkGLError("After DrawArrays")

	// Clean up
	gl.Disable(gl.BLEND)
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}
