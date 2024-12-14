package screen

import (
	"fmt"
	"image"
	colorlib "image/color"
	"image/draw"
	"image/png"
	"os"
	"runtime"

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
	fmt.Printf("Loading font from file: %s\n", filePath)
	printMemoryStats("Start")

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
		fmt.Println("[AddString] Starting to create text image")
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

			// Create the font image
			img := image.NewRGBA(image.Rect(0, 0, 512, 512))
			draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
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
			} else {
				SaveDebugImage(img, "debug_font.png")
			}

			// Create OpenGL Texture
			var texture uint32
			gl.GenTextures(1, &texture)
			gl.ActiveTexture(gl.TEXTURE0) // Activate texture unit 0
			gl.BindTexture(gl.TEXTURE_2D, texture)
			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 512, 512, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

			// Calculate proper position and scale
			width := float32(scr.XMax-scr.XMin) * float32(scale) / 10
			height := float32(scr.YMax-scr.YMin) * float32(scale) / 10
			posX := x
			posY := y

			// Create VAO and VBO for quad
			vertices := []float32{
				posX, posY, 0.0, 0.0, // Bottom-left  (position, UV)
				posX + width, posY, 1.0, 0.0, // Bottom-right (position, UV)
				posX, posY + height, 0.0, 1.0, // Top-left (position, UV)
				posX + width, posY + height, 1.0, 1.0, // Top-right (position, UV)
			}

			gl.GenVertexArrays(1, &str.VAO)
			gl.GenBuffers(1, &str.VBO)
			gl.BindVertexArray(str.VAO)

			// Bind VBO and load vertex data
			gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)
			gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

			// Set position attribute (2D position, first 2 floats of each vertex)
			gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
			gl.EnableVertexAttribArray(0)

			// Set UV attribute (2D UVs, next 2 floats of each vertex)
			gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
			gl.EnableVertexAttribArray(1)

			// Clean up
			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BindVertexArray(0)

			// Unbind the texture AFTER the VBO/VAO are set up
			gl.BindTexture(gl.TEXTURE_2D, 0)

			str.Texture = texture

			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
	}
	return key
}

func (str *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[STRING]; !present {
		vertexShaderSource := `
		#version 450
		layout (location = 0) in vec2 position;
		layout (location = 1) in vec2 uv;
		uniform mat4 projection;
		out vec2 fragUV;
		void main() {
			gl_Position = projection * vec4(position, 0.0, 1.0);
			fragUV = uv;
		}` + "\x00"

		fragmentShaderSource := `
		#version 450
		in vec2 fragUV;
		uniform sampler2D fontTexture;
		out vec4 outColor;
		void main() {
			vec4 texColor = texture(fontTexture, fragUV);
			outColor = texColor; // Display texture without multiplying it with another color
		}` + "\x00"

		shaderProgram := compileShaderProgram(vertexShaderSource, fragmentShaderSource)
		if shaderProgram == 0 {
			panic("Failed to compile shader program for STRING")
		}
		fmt.Println("[AddShader] Successfully created shader program for STRING")
		scr.Shaders[STRING] = shaderProgram
	}
	return scr.Shaders[STRING]
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

func (str *String) Render(scr *Screen) {
	gl.UseProgram(scr.Shaders[STRING])
	checkGLError("After UseProgram")

	// Enable depth testing
	gl.Enable(gl.DEPTH_TEST)
	checkGLError("After Enable DEPTH_TEST")

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

	// Set projection matrix
	projectionUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("projection\x00"))
	checkGLError("After GetUniformLocation")
	if projectionUniform < 0 {
		fmt.Println("[Render] Projection uniform not found!")
		panic("[Render] Projection uniform location returned -1")
	}
	gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
	checkGLError("After UniformMatrix4fv")

	// Activate texture unit and bind the texture
	gl.ActiveTexture(gl.TEXTURE0)
	checkGLError("After ActiveTexture")
	gl.BindTexture(gl.TEXTURE_2D, str.Texture)
	checkGLError("After BindTexture")

	// Bind the shader texture uniform to the correct texture unit
	textureUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("fontTexture\x00"))
	checkGLError("After GetUniformLocation for fontTexture")
	if textureUniform < 0 {
		fmt.Println("[Render] fontTexture uniform not found!")
		panic("[Render] fontTexture uniform location returned -1")
	}
	gl.Uniform1i(textureUniform, 0)
	checkGLError("After Uniform1i")

	// Bind the VAO (Vertex Array Object)
	gl.BindVertexArray(str.VAO)
	checkGLError("After BindVertexArray")

	// Draw the quad (two triangles)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	checkGLError("After DrawArrays")

	// Unbind VAO and texture to avoid side effects
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	checkGLError("After UnbindVertexArray and UnbindTexture")
}
