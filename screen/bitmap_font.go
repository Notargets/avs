package screen

import (
	"fmt"
	"image"
	colorlib "image/color"
	"image/draw"
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

func (line *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[STRING]; !present {
		vertexShaderSource := `
		#version 450
		layout (location = 0) in vec2 position;
		layout (location = 1) in vec2 uv;
		layout (location = 2) in vec3 color;
		uniform mat4 projection;
		out vec2 fragUV;
		out vec3 fragColor;
		void main() {
			gl_Position = projection * vec4(position, 0.0, 1.0);
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
			outColor = texColor * vec4(fragColor, 1.0);
		}` + "\x00"

		scr.Shaders[STRING] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
	}
	return scr.Shaders[STRING]
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
		var str *String
		if renderable, exists := scr.Objects[key]; exists {
			str = renderable.Object.(*String)
		} else {
			str = &String{
				Text:          text,
				Position:      mgl32.Vec2{x, y},
				Color:         color,
				ShaderProgram: str.addShader(scr),
			}
		}
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
		}

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 512, 512, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		str.Texture = texture
		scr.Objects[key] = Renderable{
			Active: true,
			Object: str,
		}
	}
	return key
}

func (str *String) Render(scr *Screen) {
	scr.RenderChannel <- func() {
		gl.UseProgram(scr.Shaders[STRING])
		checkGLError("String: After bind texture")

		// Enable blending for transparency
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		checkGLError("String: After blend")

		// Disable depth testing to ensure text is not obscured
		gl.Disable(gl.DEPTH_TEST)
		checkGLError("String: After disable depth text")

		projectionUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("projection\x00"))
		if projectionUniform >= 0 {
			gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
			checkGLError("String: After UniformMatrix4fv")
		}

		gl.ActiveTexture(gl.TEXTURE0)
		textureUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("fontTexture\x00"))
		gl.Uniform1i(textureUniform, 0)
		checkGLError("String: After Uniform1i")
		gl.BindTexture(gl.TEXTURE_2D, str.Texture)
		checkGLError("String: After binftexture")

		// Render the quad for the text
		gl.GenVertexArrays(1, &str.VAO)
		checkGLError("String: After GenVertexArrays")
		gl.BindVertexArray(str.VAO)
		checkGLError("String: After BindVertexArrays")

		// Position and UV coordinates for the text quad
		vertices := []float32{
			str.Position[0], str.Position[1], 0.0, 0.0,
			str.Position[0] + 1.0, str.Position[1], 1.0, 0.0,
			str.Position[0], str.Position[1] + 1.0, 0.0, 1.0,
			str.Position[0] + 1.0, str.Position[1] + 1.0, 1.0, 1.0,
		}

		gl.GenBuffers(1, &str.VBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, str.VBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
		checkGLError("String: 1")

		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(0)
		checkGLError("String: 2")

		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
		gl.EnableVertexAttribArray(1)
		checkGLError("String: 3")

		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		checkGLError("String: DrawArrays")

		// Cleanup
		gl.DisableVertexAttribArray(0)
		gl.DisableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
		gl.DeleteBuffers(1, &str.VBO)
		gl.DeleteVertexArrays(1, &str.VAO)
		checkGLError("String: Cleanup")

		// Unbind texture
		gl.BindTexture(gl.TEXTURE_2D, 0)
		checkGLError("String: Unbind")

		// Restore OpenGL state
		gl.Enable(gl.DEPTH_TEST)
		gl.Disable(gl.BLEND)
		checkGLError("String: Enable Depth_test, disable blend")
	}
}
