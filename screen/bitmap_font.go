package screen

import (
	"fmt"
	"image"
	"image/color"
	"runtime"

	"github.com/notargets/avs/assets"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

type String struct {
	VAO, VBO        uint32
	Text            string
	ShaderProgram   uint32
	Position        mgl32.Vec2
	Texture         uint32
	StringType      RenderType
	polygonVertices [4]mgl32.Vec2
	TextFormatter   *assets.TextFormatter
}

func printMemoryStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("[%s] Memory Usage: Alloc = %v MB, TotalAlloc = %v MB, Sys = %v MB, NumGC = %v\n",
		label, m.Alloc/1024/1024, m.TotalAlloc/1024/1024, m.Sys/1024/1024, m.NumGC)
}

func (scr *Screen) Printf(formatter *assets.TextFormatter, x, y float32, format string, args ...interface{}) (newKey Key) {
	// Format the string using fmt.Sprintf
	text := fmt.Sprintf(format, args...)

	// Call AddString with the formatted text
	newKey = scr.AddString(NEW, formatter, x, y, text)

	return newKey
}

func (scr *Screen) NewTextFormatter(fontBaseName, fontOptionName string, fontPitch int, fontColor color.Color,
	centered, screenFixed bool) (tf *assets.TextFormatter) {
	tf = assets.NewTextFormatter(fontBaseName, fontOptionName, fontPitch, scr.ScreenWidth, fontColor,
		centered, screenFixed, scr.XMax-scr.XMin, scr.YMax-scr.YMin)
	return
}

func (scr *Screen) AddString(key Key, textFormatter *assets.TextFormatter, x, y float32, text string) (newKey Key) {
	if key == NEW {
		key = Key(uuid.New())
	}
	newKey = key

	if textFormatter == nil {
		panic("textFormatter is nil")
	}

	scr.RenderChannel <- func() {
		var str *String
		if object, present := scr.Objects[key]; present {
			str = object.Object.(*String)
		} else {
			str = &String{
				Text:          text,
				Position:      mgl32.Vec2{x, y},
				TextFormatter: textFormatter,
			}
			if str.TextFormatter.ScreenFixed {
				str.StringType = FIXEDSTRING
			} else {
				str.StringType = STRING
			}
			str.ShaderProgram = str.addShader(scr)

			img, textureWidth, textureHeight, quadWidth, quadHeight := str.TextFormatter.TypeFace.RenderFontTextureImg(text, str.TextFormatter.Color)

			// **Step 4: Calculate proper position and scale**
			var posX, posY float32
			if textFormatter.Centered {
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
			c := ColorToFloat32(textFormatter.Color)
			str.initializeVBO(scr, img, textureWidth, textureHeight, c)

			// Store the string in the screen objects
			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
	}
	return newKey
}

func (str *String) initializeVBO(scr *Screen, img *image.RGBA, textureWidth, textureHeight int, color [4]float32) {
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
