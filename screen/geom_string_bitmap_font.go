/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"image"
	"unsafe"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/assets"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func addStringShaders(shaderMap map[utils.RenderType]uint32) {
	fragmentShaderSource := gl.Str(`
		#version 450
		in vec2 fragUV;
		in vec3 fragColor;
		uniform sampler2D fontTexture;
		out vec4 outColor;

		void main() {
			vec4 texColor = texture(fontTexture, fragUV);
			outColor = texColor * vec4(fragColor, texColor.a);
		}` + "\x00")

	vertexShaderSource := gl.Str(`
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
			}` + "\x00")
	shaderMap[utils.STRING] = compileShaderProgram(vertexShaderSource,
		fragmentShaderSource)
	CheckGLError("After String compileShaderProgram")

	vertexShaderSource = gl.Str(`
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
    				gl_Position = NDCposition; // Use the NDS position
                                               // directly  (clip space)
    			    fragUV = uv[gl_VertexID % 4]; // Select UV coordinate based
					                              // on gl_VertexID ( assumes quads)
    				fragColor = color;
				}` + "\x00")
	shaderMap[utils.FIXEDSTRING] = compileShaderProgram(vertexShaderSource,
		fragmentShaderSource)
	CheckGLError("After FixedString compileShaderProgram")
}

type String struct {
	VAO, VBO                    uint32
	Text                        string
	ShaderProgram               uint32
	Position                    mgl32.Vec2
	WindowWidth, WindowHeight   uint32
	Texture                     uint32
	StringType                  utils.RenderType
	polygonVertices             [4]mgl32.Vec2 // In world coordinates
	HostGPUBuffer               []float32
	InitializedFIXEDSTRING      bool
	textureImg                  *image.RGBA
	textureWidth, textureHeight uint32
	TextFormatter               *assets.TextFormatter
}

func newString(tf *assets.TextFormatter, x, y float32, text string,
	win *Window) (str *String) {

	str = &String{
		Text:                   text,
		Position:               mgl32.Vec2{x, y},
		TextFormatter:          tf,
		InitializedFIXEDSTRING: false,
		WindowWidth:            win.width,
		WindowHeight:           win.height,
	}
	if tf.ScreenFixed {
		str.StringType = utils.FIXEDSTRING
	} else {
		str.StringType = utils.STRING
	}
	str.ShaderProgram = win.shaders[str.StringType]
	return
}

func (str *String) render(projectionMatrix mgl32.Mat4, windowWidth,
	windowHeight uint32, xMin, xMax, yMin, yMax float32) {
	setShaderProgram(str.ShaderProgram)

	// Draw the font into the image, calculate the polygon vertex bounds
	str.setupTextureMap(xMin, xMax, yMin, yMax)

	// Set the color
	textColor := utils.ColorToFloat32(str.TextFormatter.Color)

	str.loadHostBuffer(textColor, projectionMatrix, windowWidth, windowHeight)

	str.sendHostBufferToGPU()

	// Draw the quad after binding the texture to it
	// Bind the texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, str.Texture)
	CheckGLError("After BindTexture")

	// Bind the VAO and draw the polygon
	gl.BindVertexArray(str.VAO)
	CheckGLError("After BindVertexArray")

	// Enable Blending
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	CheckGLError("After BlendFunc")

	// Draw the quad (TRIANGLE_STRIP for simplicity)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	CheckGLError("After DrawArrays")

	// Clean up
	gl.Disable(gl.BLEND)
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (str *String) setupTextureMap(xMin, xMax, yMin, yMax float32) {
	// Load our texture map with the drawn text if not done already
	if str.textureImg == nil {
		var img *image.RGBA
		img = str.TextFormatter.TypeFace.RenderFontTextureImg(str.Text, str.TextFormatter.Color)
		str.textureImg = img
		str.textureWidth, str.textureHeight = uint32(str.textureImg.Bounds().Dx()), uint32(str.textureImg.Bounds().Dy())
	}

	if str.StringType == utils.STRING || !str.InitializedFIXEDSTRING {
		// This is done every time for STRING, only once for FIXEDSTRING
		// For STRING, this compensates for resize and pan via the projection matrix
		// For FIXEDSTRING, the projection is applied once to get to Screen / Pixel fixed coordinates
		str.calculatePolygonVertices(xMin, xMax, yMin, yMax)
	}
}

func (str *String) calculatePolygonVertices(xMin, xMax, yMin, yMax float32) {
	var (
		tf = str.TextFormatter
	)
	// setupVertices vertex coordinates
	x := str.Position.X()
	y := str.Position.Y()
	quadWidth, quadHeight := calculateQuadBounds(str.textureWidth, str.textureHeight,
		str.WindowWidth, str.WindowHeight,
		tf.TypeFace.FontDPI, xMax-xMin, yMax-yMin)

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

func (str *String) fixSTRINGAspectRatio(windowWidth, windowHeight uint32, ndc *[4]mgl32.Vec4, lenRow int) {
	// Transform the NDC coordinates to accommodate potential changing window dimensions, which will keep the text
	// ... visually the same size
	// First, retrieve the previous NDC coordinates from the GPU buffer
	for i := 0; i < 4; i++ {
		ndc[i][0] = str.HostGPUBuffer[i*lenRow]
		ndc[i][1] = str.HostGPUBuffer[i*lenRow+1]
		ndc[i][2] = str.HostGPUBuffer[i*lenRow+2]
		ndc[i][3] = str.HostGPUBuffer[i*lenRow+3]
	}
	// Compute the aspect ratio scaling factors for x and y
	Sx := float32(str.WindowWidth) / float32(windowWidth)
	Sy := float32(str.WindowHeight) / float32(windowHeight)

	// Calculate the center of the polygon in NDC coordinates
	var cR, cS float32
	for i := 0; i < 4; i++ {
		cR += ndc[i][0] // Sum of all r's
		cS += ndc[i][1] // Sum of all s's
	}
	cR /= 4 // Average of r-coordinates (center x)
	cS /= 4 // Average of s-coordinates (center y)

	// Apply the transformation to each of the 4 NDC vertices
	for i := 0; i < 4; i++ {
		// Extract the vertex
		r := ndc[i][0] // x-coordinate (r)
		s := ndc[i][1] // y-coordinate (s)
		t := ndc[i][2] // z-coordinate (t) - unchanged
		w := ndc[i][3] // w-coordinate (w) - unchanged

		// Apply the aspect ratio scaling to r and s
		newR := Sx*(r-cR) + cR
		newS := Sy*(s-cS) + cS

		// Store the updated vertex back into the buffer
		ndc[i] = mgl32.Vec4{newR, newS, t, w}
	}
	for i := 0; i < 4; i++ {
		str.HostGPUBuffer[i*lenRow] = ndc[i][0]
		str.HostGPUBuffer[i*lenRow+1] = ndc[i][1]
		str.HostGPUBuffer[i*lenRow+2] = ndc[i][2]
		str.HostGPUBuffer[i*lenRow+3] = ndc[i][3]
	}
}

func (str *String) loadHostBuffer(textColor [4]float32,
	projectionMatrix mgl32.Mat4, currentWidth, currentHeight uint32) {
	var (
		lenRow int
	)

	if str.StringType == utils.STRING {
		lenRow = 2 + 3
		lenV := 4 * (lenRow)
		if len(str.HostGPUBuffer) == 0 {
			str.HostGPUBuffer = make([]float32, lenV)
		}
		for i := 0; i < 4; i++ {
			str.HostGPUBuffer[i*lenRow] = str.polygonVertices[i][0]
			str.HostGPUBuffer[i*lenRow+1] = str.polygonVertices[i][1]
			str.HostGPUBuffer[i*lenRow+2] = textColor[0]
			str.HostGPUBuffer[i*lenRow+3] = textColor[1]
			str.HostGPUBuffer[i*lenRow+4] = textColor[2]
		}
	} else if str.StringType == utils.FIXEDSTRING {
		// Calculate fixed position in NDC coordinates once, via the initial projection matrix
		// This puts the location into fixed pixel coordinates mapped to the window via the ortho projection
		lenRow = 4 + 3
		lenV := 4 * (lenRow)
		var NDCVertexCoordinates [4]mgl32.Vec4
		if !str.InitializedFIXEDSTRING {
			// fmt.Printf("Rendering FIXED STRING...\n")
			for i := 0; i < 4; i++ {
				NDCVertexCoordinates[i] = projectionMatrix.Mul4x1(mgl32.Vec4{str.polygonVertices[i].X(), str.polygonVertices[i].Y(), 0.0, 1.0})
				NDCVertexCoordinates[i] = NDCVertexCoordinates[i].Mul(1.0 / NDCVertexCoordinates[i].W())
			}
			if len(str.HostGPUBuffer) == 0 {
				str.HostGPUBuffer = make([]float32, lenV)
			}
			for i := 0; i < 4; i++ {
				str.HostGPUBuffer[i*lenRow] = NDCVertexCoordinates[i][0]   // Clip space coordinates for fixed position
				str.HostGPUBuffer[i*lenRow+1] = NDCVertexCoordinates[i][1] // Clip space coordinates for fixed position
				str.HostGPUBuffer[i*lenRow+2] = NDCVertexCoordinates[i][2] // Clip space coordinates for fixed position
				str.HostGPUBuffer[i*lenRow+3] = NDCVertexCoordinates[i][3] // Clip space coordinates for fixed position
			}
			str.InitializedFIXEDSTRING = true // initialization is finished after this
		}
		// Load the current color for each vertex
		for i := 0; i < 4; i++ {
			str.HostGPUBuffer[i*lenRow+4] = textColor[0]
			str.HostGPUBuffer[i*lenRow+5] = textColor[1]
			str.HostGPUBuffer[i*lenRow+6] = textColor[2]
		}
		// Correct the vertex coordinates for a FIXEDSTRING if the window size has changed
		str.fixSTRINGAspectRatio(currentWidth, currentHeight, &NDCVertexCoordinates, lenRow)
	}
	// setupVertices string formatter window dimensions to match the current screen
	str.WindowWidth = currentWidth
	str.WindowHeight = currentHeight
}

func (str *String) sendHostBufferToGPU() {
	// Transmit the texture image into the GPU texture buffer
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	// fmt.Printf("Texture width: %d, Texture height: %d\n", str.textureWidth, str.textureHeight)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(str.textureWidth), int32(str.textureHeight),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(str.textureImg.Pix))
	CheckGLError("After TexImage2D")
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	str.Texture = texture

	// Transfer the vertex data into the VBO buffer. Use the VBA to identify the layout of the data
	// Generate VBO and VAO once
	gl.GenBuffers(1, &str.VBO)
	gl.GenVertexArrays(1, &str.VAO)

	// Bind VAO
	gl.BindVertexArray(str.VAO)
	// Bind VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)

	// Upload vertex data into the VBO as a flat array
	gl.BufferData(gl.ARRAY_BUFFER, len(str.HostGPUBuffer)*4, gl.Ptr(str.HostGPUBuffer), gl.STATIC_DRAW)
	CheckGLError("After VBO")

	// Load the flat array layout into the VBA
	// **positionDelta (location = 0)**
	var stride int32
	if str.StringType == utils.STRING {
		// stride = 4 * (2 + 2 + 3)
		stride = 4 * (2 + 3)
	} else {
		stride = 4 * (4 + 3)
	}
	offset := 0
	if str.StringType == utils.STRING {
		// Load the transformed vertex coordinates
		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(offset))) // positionDelta (2 floats)
		offset += 2 * 4                                                                        // Advance by 2 floats = 8 bytes
	} else if str.StringType == utils.FIXEDSTRING {
		// **Frozen positionDelta **
		// Load the NDC fixed vertex coordinates
		gl.VertexAttribPointer(0, 4, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(offset))) // Fixed positionDelta (4 floats)
		offset += 4 * 4                                                                        // Advance by 2 floats = 8 bytes
	}
	gl.EnableVertexAttribArray(0)
	// **Color **
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(offset))) // Color (3 floats)
	gl.EnableVertexAttribArray(1)
}

func calculateQuadBounds(textureWidth, textureHeight, windowWidth, windowHeight, fontDPI uint32,
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
