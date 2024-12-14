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
	VAO, VBO      uint32
	Text          string
	ShaderProgram uint32
	Position      mgl32.Vec2
	Color         [3]float32
	Texture       uint32
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

func (scr *Screen) AddString(key Key, text string, x, y float32, color [3]float32, scale float64) (newKey Key) {
	if key == NEW {
		key = Key(uuid.New())
	}
	newKey = key

	scr.RenderChannel <- func() {
		//fmt.Println("[AddString] Starting to create transparent text image")
		var str *String
		if object, present := scr.Objects[key]; present {
			str = object.Object.(*String)
		} else {
			str = &String{
				Text:     text,
				Position: mgl32.Vec2{x, y},
				Color:    color,
			}
			str.ShaderProgram = str.addShader(scr)

			// **Dynamically calculate texture size based on text dimensions**
			scaledSize := fixed.Int26_6(scr.FontSize * scale * 64) // Convert to fixed-point
			textWidth := 0
			for _, ch := range text {
				glyphIndex := scr.Font.Index(ch)
				hMetric := scr.Font.HMetric(scaledSize, glyphIndex)
				textWidth += int(hMetric.AdvanceWidth >> 6) // Convert to pixels
			}

			textHeight := int32(scr.FontSize * scale) // Text height directly from font size and scale

			// Add padding, then align to 4-byte boundaries
			textureWidth := int32((textWidth + 16 + 3) & ^3) // Align to next multiple of 4
			textureHeight := (textHeight + 16 + 3) & ^3      // Align to next multiple of 4

			// Ensure texture size is at least 1x1
			if textureWidth < 1 {
				textureWidth = 1
			}
			if textureHeight < 1 {
				textureHeight = 1
			}

			//fmt.Printf("[AddString] Calculated texture size: %dx%d (Width x Height)\n", textureWidth, textureHeight)

			// **Create transparent RGBA image**
			img := image.NewRGBA(image.Rect(0, 0, int(textureWidth), int(textureHeight)))
			for i := 0; i < len(img.Pix); i += 4 {
				img.Pix[i+0] = 0 // Red
				img.Pix[i+1] = 0 // Green
				img.Pix[i+2] = 0 // Blue
				img.Pix[i+3] = 0 // Alpha (fully transparent)
			}

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
				//} else {
				//	SaveDebugImage(img, "debug_text_with_transparency.png")
			}

			// Create OpenGL Texture
			var texture uint32
			gl.GenTextures(1, &texture)
			gl.BindTexture(gl.TEXTURE_2D, texture)
			//gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
			//gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, textureHeight, textureWidth, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, textureWidth, textureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
			gl.BindTexture(gl.TEXTURE_2D, 0)

			aspect := float32(textureWidth) / float32(textureHeight)
			width := float32(scr.XMax-scr.XMin) * float32(scale) / 10
			height := width / aspect

			// Calculate proper position and scale
			//width := float32(scr.XMax-scr.XMin) * float32(scale) / 10
			//height := float32(scr.YMax-scr.YMin) * float32(scale) / 10
			posX := x
			posY := y

			vertices := []float32{
				posX, posY, 0.0, 1.0, color[0], color[1], color[2], // Bottom-left
				posX + width, posY, 1.0, 1.0, color[0], color[1], color[2], // Bottom-right
				posX, posY + height, 0.0, 0.0, color[0], color[1], color[2], // Top-left
				posX + width, posY + height, 1.0, 0.0, color[0], color[1], color[2], // Top-right
			}

			// Create VAO and VBO for quad
			gl.GenVertexArrays(1, &str.VAO)
			gl.GenBuffers(1, &str.VBO)
			gl.BindVertexArray(str.VAO)

			gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)
			gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

			gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 7*4, gl.PtrOffset(0))
			gl.EnableVertexAttribArray(0)
			gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 7*4, gl.PtrOffset(2*4))
			gl.EnableVertexAttribArray(1)
			gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 7*4, gl.PtrOffset(4*4))
			gl.EnableVertexAttribArray(2)

			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BindVertexArray(0)

			str.Texture = texture
			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
	}
	return newKey
}

func (str *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[STRING]; !present {
		vertexShaderSource := `
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
}
` + "\x00"

		fragmentShaderSource := `
		#version 450
in vec2 fragUV;
in vec3 fragColor;
uniform sampler2D fontTexture;
out vec4 outColor;
void main() {
	vec4 texColor = texture(fontTexture, fragUV);
	outColor = texColor * vec4(fragColor, texColor.a); // Properly consider alpha
}` + "\x00"

		shaderProgram := compileShaderProgram(vertexShaderSource, fragmentShaderSource)
		if shaderProgram == 0 {
			panic("Failed to compile shader program for STRING")
		}
		//fmt.Println("[AddShader] Successfully created shader program for STRING")
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

	// Bind the projection matrix to the shader
	projectionUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("projection\x00"))
	checkGLError("After GetUniformLocation")
	if projectionUniform < 0 {
		fmt.Println("[Render] Projection uniform not found!")
		panic("[Render] Projection uniform location returned -1")
	}
	gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
	checkGLError("After UniformMatrix4fv")

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
