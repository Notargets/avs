package screen

import (
	"fmt"
	"image"
	colorlib "image/color"
	"image/png"
	"math"
	"os"
	"runtime"

	"golang.org/x/image/math/fixed"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/google/uuid"
)

type String struct {
	VAO, VBO        uint32
	Text            string
	ShaderProgram   uint32
	Position        mgl32.Vec2
	Color           [3]float32
	Texture         uint32
	StringType      RenderType
	polygonVertices [4]mgl32.Vec2
}

func (scr *Screen) LoadFont(filePath string, fontSize float32) error {
	//fmt.Printf("Loading font from file: %s\n", filePath)
	//printMemoryStats("Start")

	fontBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read font file: %v", err)
	}

	ft, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	scr.Font = ft
	scr.FontSize = fontSize
	return nil
}

func printMemoryStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("[%s] Memory Usage: Alloc = %v MB, TotalAlloc = %v MB, Sys = %v MB, NumGC = %v\n",
		label, m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

func (scr *Screen) Printf(key Key, x, y float32, color [3]float32, scale float32, centered, screenFixed bool, format string, args ...interface{}) (newKey Key) {
	// Format the string using fmt.Sprintf
	text := fmt.Sprintf(format, args...)

	// Call AddString with the formatted text
	newKey = scr.AddString(key, text, x, y, color, scale, centered, screenFixed)

	return newKey
}

func (scr *Screen) AddString(key Key, text string, x, y float32, color [3]float32, scale float32, centered, screenFixed bool) (newKey Key) {
	if key == NEW {
		key = Key(uuid.New())
	}
	newKey = key

	scr.RenderChannel <- func() {
		var str *String
		if object, present := scr.Objects[key]; present {
			str = object.Object.(*String)
		} else {
			str = &String{
				Text:     text,
				Position: mgl32.Vec2{x, y},
				Color:    color,
			}
			if screenFixed {
				str.StringType = FIXEDSTRING
			} else {
				str.StringType = STRING
			}
			str.ShaderProgram = str.addShader(scr)

			img, textureWidth, textureHeight, quadWidth, quadHeight := scr.calculateFontBitmapSize(text, color, scale, centered)

			// **Step 4: Calculate proper position and scale**
			var posX, posY float32
			if centered {
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

			// Initialize the vertex buffer object (VBO)
			str.initializeVBO(scr, img, textureWidth, textureHeight, color)

			// Store the string in the screen objects
			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
	}
	return newKey
}

func (scr *Screen) calculateFontBitmapSize(text string, color [3]float32, scale float32, centered bool) (img *image.RGBA, textureWidth, textureHeight int32, quadWidth, quadHeight float32) {
	// Calculate text width and height using FreeType context
	scaledSize := fixed.Int26_6(scr.FontSize * scale * 64) // Scale font size for 26.6 fixed-point format
	textWidth := 0

	// Create a FreeType context
	img = image.NewRGBA(image.Rect(0, 0, 1, 1)) // Dummy image for context
	ctx := freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(scr.Font)
	ctx.SetFontSize(float64(scr.FontSize * scale))
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)

	// Calculate total width of the text using HMetric
	lastCharRightEdge := 0
	for i, ch := range text {
		glyphIndex := scr.Font.Index(ch)
		hMetric := scr.Font.HMetric(scaledSize, glyphIndex)
		textWidth += int(hMetric.AdvanceWidth >> 6) // Total advance width

		if i == len(text)-1 { // Handle the last character specially
			bounds := scr.Font.Bounds(scaledSize)
			rightEdge := textWidth + (bounds.Max.X.Round() >> 6)
			if rightEdge > textWidth {
				lastCharRightEdge = rightEdge - textWidth
			}
		}
	}

	// Add the extra width for the last character's overflow
	textWidth += lastCharRightEdge

	// Calculate font height from font ascent and descent
	fontMetrics := scr.Font.Bounds(scaledSize)
	ascent := fontMetrics.Max.Y.Round()
	descent := -fontMetrics.Min.Y.Round()
	textHeight := ascent + descent

	// Ensure texture dimensions are aligned to 4-byte boundaries
	textureWidth = int32((textWidth + 3) & ^3)   // Align width to 4-byte boundary
	textureHeight = int32((textHeight + 3) & ^3) // Align height to 4-byte boundary

	// Create an image of the proper size to hold the full text
	img = image.NewRGBA(image.Rect(0, 0, int(textureWidth), int(textureHeight)))
	ctx = freetype.NewContext()
	ctx.SetDPI(72)
	ctx.SetFont(scr.Font)
	ctx.SetFontSize(float64(scr.FontSize * scale))
	ctx.SetClip(img.Bounds())
	ctx.SetDst(img)
	ctx.SetSrc(image.NewUniform(colorlib.RGBA{R: uint8(color[0] * 255), G: uint8(color[1] * 255), B: uint8(color[2] * 255), A: 255}))

	// Draw the string at the proper position (start at the baseline)
	pt := freetype.Pt(0, ascent) // Baseline is ascent from the top of the image
	_, err := ctx.DrawString(text, pt)
	if err != nil {
		fmt.Printf("Error drawing string: %v\n", err)
	}

	// Calculate quad width and height from texture dimensions
	xRange := scr.XMax - scr.XMin
	yRange := scr.YMax - scr.YMin

	// **Step 1: Calculate initial quad size from raw text dimensions**
	quadWidth = float32(textureWidth) / float32(xRange) * scale
	quadHeight = float32(textureHeight) / float32(yRange) * scale

	// **Step 2: Stretch the larger dimension to maintain aspect ratio**
	if xRange > yRange {
		quadWidth *= (xRange / yRange) // Stretch width
	} else {
		quadHeight *= (yRange / xRange) // Stretch height
	}

	// **Step 3: Apply inverse scale correction to keep the total size constant**
	scaleFactor := float32(math.Min(float64(xRange), float64(yRange))) / float32(math.Max(float64(xRange), float64(yRange)))
	quadWidth *= scaleFactor
	quadHeight *= scaleFactor

	return
}

func (str *String) initializeVBO(scr *Screen, img *image.RGBA, textureWidth, textureHeight int32, color [3]float32) {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	//fmt.Printf("Texture width: %d, Texture height: %d\n", textureWidth, textureHeight)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, textureWidth, textureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	checkGLError("After TexImage2D")
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	str.Texture = texture
	var uv = [4][2]float32{
		{0, 1},
		{1, 1},
		{0, 0},
		{1, 0},
	}
	var vertices []float32
	switch str.StringType {
	case STRING:
		lenRow := 2 + 2 + 3
		lenV := 4 * (lenRow)
		vertices = make([]float32, lenV)
		for i := 0; i < 4; i++ {
			vertices[i*lenRow] = str.polygonVertices[i][0]
			vertices[i*lenRow+1] = str.polygonVertices[i][1]
			vertices[i*lenRow+2] = uv[i][0]
			vertices[i*lenRow+3] = uv[i][1]
			vertices[i*lenRow+4] = color[0]
			vertices[i*lenRow+5] = color[1]
			vertices[i*lenRow+6] = color[2]
		}
	case FIXEDSTRING:
		var projected [4]mgl32.Vec4
		for i := 0; i < 4; i++ {
			projected[i] = scr.projectionMatrix.Mul4x1(mgl32.Vec4{str.polygonVertices[i].X(), str.polygonVertices[i].Y(), 0.0, 1.0})
			projected[i] = projected[i].Mul(1.0 / projected[i].W())
		}
		lenRow := 2 + 2 + 3 + 4
		lenV := 4 * (lenRow)
		vertices = make([]float32, lenV)
		for i := 0; i < 4; i++ {
			vertices[i*lenRow] = str.polygonVertices[i][0]
			vertices[i*lenRow+1] = str.polygonVertices[i][1]
			vertices[i*lenRow+2] = uv[i][0]
			vertices[i*lenRow+3] = uv[i][1]
			vertices[i*lenRow+4] = color[0]
			vertices[i*lenRow+5] = color[1]
			vertices[i*lenRow+6] = color[2]
			for j := 0; j < 4; j++ {
				vertices[i*lenRow+7+j] = projected[i][j]
			}
		}
	}

	// Generate VBO and VAO once
	gl.GenBuffers(1, &str.VBO)
	gl.GenVertexArrays(1, &str.VAO)

	// Bind VAO
	gl.BindVertexArray(str.VAO)

	// Bind VBO
	gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)

	// Upload vertex data
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	checkGLError("After VBO")

	// **Setup Vertex Attributes**
	offset := 0

	// **PositionDelta (location = 0)**
	var stride int32
	if str.StringType == STRING {
		stride = 4 * (2 + 2 + 3)
	} else {
		stride = 4 * (2 + 2 + 3 + 4)
	}
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // PositionDelta (2 floats)
	gl.EnableVertexAttribArray(0)
	offset += 2 * 4 // Advance by 2 floats = 8 bytes

	// **UV (location = 1)**
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // UV (2 floats)
	gl.EnableVertexAttribArray(1)
	offset += 2 * 4 // Advance by 2 floats = 8 bytes

	// **Color (location = 2)**
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // Color (3 floats)
	gl.EnableVertexAttribArray(2)
	offset += 3 * 4 // Advance by 3 floats = 12 bytes

	// **Frozen PositionDelta (location = 3)**
	if str.StringType == FIXEDSTRING {
		gl.VertexAttribPointer(3, 4, gl.FLOAT, false, stride, gl.PtrOffset(offset)) // Fixed PositionDelta (4 floats)
		gl.EnableVertexAttribArray(3)
	}

}

func (str *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[str.StringType]; !present {
		var vertexShaderSource string
		switch str.StringType {
		case STRING:
			vertexShaderSource = `
				#version 450
				layout (location = 0) in vec2 position;
				layout (location = 1) in vec2 uv;
				layout (location = 2) in vec3 color;
				uniform mat4 projection; // <-- projection matrix
				out vec2 fragUV;
				out vec3 fragColor;
				void main() {
    				gl_Position = projection * vec4(position, 0.0, 1.0); // Apply projection matrix here
    				fragUV = uv;
    				fragColor = color;
				}` + "\x00"
		case FIXEDSTRING:
			vertexShaderSource = `
				#version 450
				layout (location = 0) in vec2 position;
				layout (location = 1) in vec2 uv;
				layout (location = 2) in vec3 color;
				layout (location = 3) in vec4 fixedPosition;

				out vec2 fragUV;
				out vec3 fragColor;

				void main() {
    				gl_Position = fixedPosition; // Use the fixed position directly (clip space)
    				fragUV = uv;
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
	if str.StringType == STRING {
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

func SaveDebugImage(img *image.RGBA, filename string) {
	if img == nil {
		fmt.Println("[SaveDebugImage] Image is nil, nothing to save.")
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("[SaveDebugImage] Failed to create image file: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		fmt.Printf("[SaveDebugImage] Failed to save image as PNG: %v\n", err)
	} else {
		fmt.Printf("[SaveDebugImage] Image saved as '%s'.\n", filename)
	}
}
