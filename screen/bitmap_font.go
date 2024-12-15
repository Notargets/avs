package screen

import (
	"fmt"
	"image"
	colorlib "image/color"
	"image/png"
	"os"
	"runtime"

	"golang.org/x/image/math/fixed"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/google/uuid"
)

type String struct {
	VAO, VBO             uint32
	Text                 string
	ShaderProgram        uint32
	Position             mgl32.Vec2
	Color                [3]float32
	Texture              uint32
	StringType           RenderType
	FreezePosition       bool
	FrozenPositions      [4][4]float32 // Text position in projected world coordinates (clip space, vec4) after freezing
	polygonVertices      [4]mgl32.Vec2
	frozenPositionOffset int
}

func (scr *Screen) LoadFont(filePath string, fontSize float64) error {
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

func (scr *Screen) AddString(key Key, text string, x, y float32, color [3]float32, scale float64, centered, screenFixed bool) (newKey Key) {
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
				Text:           text,
				Position:       mgl32.Vec2{x, y},
				Color:          color,
				FreezePosition: false,
			}
			if screenFixed {
				str.StringType = FIXEDSTRING
			} else {
				str.StringType = STRING
			}
			str.ShaderProgram = str.addShader(scr)

			// Calculate text size
			scaledSize := fixed.Int26_6(scr.FontSize * scale * 64)
			textWidth := 0
			for _, ch := range text {
				glyphIndex := scr.Font.Index(ch)
				hMetric := scr.Font.HMetric(scaledSize, glyphIndex)
				textWidth += int(hMetric.AdvanceWidth >> 6)
			}

			textHeight := int32(scr.FontSize * scale)
			textureWidth := int32((textWidth + 3) & ^3) // Fixed alignment
			textureHeight := int32((textHeight + 3) & ^3)

			img := image.NewRGBA(image.Rect(0, 0, int(textureWidth), int(textureHeight)))
			ctx := freetype.NewContext()
			ctx.SetDPI(72)
			ctx.SetFont(scr.Font)
			ctx.SetFontSize(scr.FontSize * scale)
			ctx.SetClip(img.Bounds())
			ctx.SetDst(img)
			ctx.SetSrc(image.NewUniform(colorlib.RGBA{R: uint8(color[0] * 255), G: uint8(color[1] * 255), B: uint8(color[2] * 255), A: 255}))
			pt := freetype.Pt(0, int(ctx.PointToFixed(scr.FontSize*scale)>>6))
			_, err := ctx.DrawString(text, pt)
			if err != nil {
				fmt.Printf("Error drawing string: %v\n", err)
			}

			aspect := float32(textureWidth) / float32(textureHeight)
			width := float32(scr.XMax-scr.XMin) * float32(scale) / 10
			height := width / aspect

			// Calculate proper position and scale
			var posX, posY float32
			if centered {
				posX = x - float32(width)/2
				posY = y - float32(height)/2
			} else {
				posX = x
				posY = y
			}

			// Initialize polygon vertices for the 4 corners of the quad
			str.polygonVertices = [4]mgl32.Vec2{
				{posX, posY},                  // Bottom-left
				{posX + width, posY},          // Bottom-right
				{posX, posY + height},         // Top-left
				{posX + width, posY + height}, // Top-right
			}

			str.initializeVBO(img, textureWidth, textureHeight, color)

			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
	}
	return newKey
}

func (str *String) initializeVBO(img *image.RGBA, textureWidth, textureHeight int32, color [3]float32) {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, textureWidth, textureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	checkGLError("After TexImage2D")
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	str.Texture = texture

	uvs := [4]mgl32.Vec2{
		{0.0, 1.0}, {1.0, 1.0}, {0.0, 0.0}, {1.0, 0.0},
	}

	colors := [4][3]float32{
		color, color, color, color,
	}

	vertices := make([]float32, 0, 4*11)
	for i := 0; i < 4; i++ {
		vertices = append(vertices,
			str.polygonVertices[i].X(), str.polygonVertices[i].Y(),
			uvs[i].X(), uvs[i].Y(),
			colors[i][0], colors[i][1], colors[i][2],
			0, 0, 0, 1, // Placeholder for frozen position
		)
	}

	gl.GenBuffers(1, &str.VBO)
	gl.GenVertexArrays(1, &str.VAO)
	gl.BindVertexArray(str.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)
	offset := 0

	// **Position (location = 0)**
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 11*4, gl.PtrOffset(offset)) // Position (2 floats)
	gl.EnableVertexAttribArray(0)
	offset += 2 * 4 // Advance by 2 floats = 8 bytes

	// **UV (location = 1)**
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 11*4, gl.PtrOffset(offset)) // UV (2 floats)
	gl.EnableVertexAttribArray(1)
	offset += 2 * 4 // Advance by 2 floats = 8 bytes

	// **Color (location = 2)**
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 11*4, gl.PtrOffset(offset)) // Color (3 floats)
	gl.EnableVertexAttribArray(2)
	offset += 3 * 4 // Advance by 3 floats = 12 bytes

	// **Frozen Position (location = 3)**
	str.frozenPositionOffset = offset                                         // This is where frozen position begins
	gl.VertexAttribPointer(3, 4, gl.FLOAT, false, 11*4, gl.PtrOffset(offset)) // Frozen Position (4 floats)
	gl.EnableVertexAttribArray(3)

}

func (str *String) uploadFrozenPositions(scr *Screen) {
	for i, v := range str.polygonVertices {
		projected := scr.projectionMatrix.Mul4x1(mgl32.Vec4{v.X(), v.Y(), 0.0, 1.0})
		str.FrozenPositions[i] = [4]float32{projected.X(), projected.Y(), projected.Z(), projected.W()}
	}
	str.FreezePosition = true

	frozenPositionData := make([]float32, 16)
	for i := 0; i < 4; i++ {
		copy(frozenPositionData[i*4:], str.FrozenPositions[i][:])
	}

	for i := 0; i < 4; i++ {
		fmt.Printf("Frozen Position Vertex %d: [%.4f, %.4f, %.4f, %.4f]\n", i,
			str.FrozenPositions[i][0],
			str.FrozenPositions[i][1],
			str.FrozenPositions[i][2],
			str.FrozenPositions[i][3])
	}

	for i, v := range str.FrozenPositions {
		if v[3] == 0 {
			fmt.Printf("[uploadFrozenPositions] Error: w = 0 for vertex %d\n", i)
		}
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, str.frozenPositionOffset, len(frozenPositionData)*4, gl.Ptr(frozenPositionData))
}

func (str *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[STRING]; !present {
		vertexShaderSource := `
		#version 450
		layout (location = 0) in vec2 position;
		layout (location = 1) in vec2 uv;
		layout (location = 2) in vec3 color;
		layout (location = 3) in vec4 frozenPosition;
		uniform mat4 projection;
		uniform bool freezePosition;

		out vec2 fragUV;
		out vec3 fragColor;

		void main() {
			vec4 finalPosition;
			if (freezePosition) {
    			if (frozenPosition.w != 0.0) {
        			finalPosition = vec4(frozenPosition.xyz / frozenPosition.w, 1.0);
    			} else {
        			finalPosition = frozenPosition;
    			}
			} else {
    			finalPosition = projection * vec4(position, 0.0, 1.0);
			}
			gl_Position = finalPosition;

			fragUV = uv;
			fragColor = color;
		}` + "\x00"

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

		shaderProgram := compileShaderProgram(vertexShaderSource, fragmentShaderSource)
		scr.Shaders[STRING] = shaderProgram
	}
	return scr.Shaders[STRING]
}

func (str *String) Render(scr *Screen) {
	gl.UseProgram(scr.Shaders[STRING])
	checkGLError("After UseProgram")

	// Check if the active program matches
	var activeProgram int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &activeProgram)
	if uint32(activeProgram) != scr.Shaders[STRING] {
		fmt.Printf("[Render] Shader program mismatch! Active: %d, Expected: %d\n", activeProgram, scr.Shaders[STRING])
		panic("[Render] Shader program is not active as expected")
	}

	if scr.Shaders[STRING] == 0 {
		fmt.Println("[Render] Shader program handle is 0. Possible compilation/linking failure.")
		panic("[Render] Shader program handle is 0")
	}

	// Set the freezePosition uniform (1 for true, 0 for false)
	freezeUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("freezePosition\x00"))
	if freezeUniform < 0 {
		fmt.Println("[Render] freezePosition uniform not found!")
		panic("[Render] freezePosition uniform location returned -1")
	}
	gl.Uniform1i(freezeUniform, boolToInt(str.FreezePosition)) // Send 1 if FreezePosition is true, 0 if false
	checkGLError("After Uniform1i(freezePosition)")

	// Bind the projection matrix to the shader
	projectionUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("projection\x00"))
	checkGLError("After GetUniformLocation")
	if projectionUniform < 0 {
		fmt.Println("[Render] Projection uniform not found!")
		panic("[Render] Projection uniform location returned -1")
	}
	gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
	checkGLError("After UniformMatrix4fv")

	// **Call uploadFrozenPositions for FIXEDSTRING if it hasn't been done yet**
	if str.StringType == FIXEDSTRING && !str.FreezePosition {
		fmt.Println("[Render] Uploading frozen positions...")
		str.uploadFrozenPositions(scr)
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

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
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
