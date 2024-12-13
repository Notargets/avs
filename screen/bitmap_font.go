package screen

import (
	"fmt"
	"os"

	"golang.org/x/image/math/fixed"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"

	"github.com/4ydx/gltext"
	v45 "github.com/4ydx/gltext/v4.5"
)

// LAL: Can generate a font here: https://snowb.org/

type String struct {
	Text          *v45.Text
	ShaderProgram uint32 // Shader program specific to this Line object
}

func (line *String) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[STRING]; !present {
		// Line Shaders
		var vertexShaderSource = `
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
		}
` + "\x00"

		var fragmentShaderSource = `
#version 450
		in vec2 fragUV;
		in vec3 fragColor;
		uniform sampler2D fontTexture;
		out vec4 outColor;
		void main() {
			vec4 texColor = texture(fontTexture, fragUV);
			outColor = texColor * vec4(fragColor, 1.0);
		}
` + "\x00"
		scr.Shaders[STRING] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
	}
	return scr.Shaders[STRING]
}

func (scr *Screen) LoadFont(filePath string, fontConfigPath string, fontSize int) error {
	var font *v45.Font

	// Try to load pre-saved font configuration
	config, err := gltext.LoadTruetypeFontConfig(fontConfigPath, "default_font")
	if err == nil {
		font, err = v45.NewFont(config)
		if err != nil {
			return fmt.Errorf("failed to load font from configuration: %v", err)
		}
		fmt.Println("Font loaded from font config...")
	} else {
		// Load TTF file and generate configuration
		fd, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to load font file %s: %v", filePath, err)
		}
		defer fd.Close()

		runeRanges := gltext.RuneRanges{
			{Low: 32, High: 126}, // ASCII characters
		}
		config, err = gltext.NewTruetypeFontConfig(fd, fixed.Int26_6(fontSize<<6), runeRanges, 128, 5)
		if err != nil {
			return fmt.Errorf("failed to create font config: %v", err)
		}
		font, err = v45.NewFont(config)
		if err != nil {
			return fmt.Errorf("failed to create font from config: %v", err)
		}
	}
	scr.Font = font
	return nil
}

func (scr *Screen) AddString(key Key, text string, x, y float32, color [3]float32, scale float32) (newKey Key) {
	if key == NEW {
		key = Key(uuid.New())
	}
	scr.RenderChannel <- func() {
		var str *v45.Text
		if existingRenderable, exists := scr.Objects[key]; exists {
			str = existingRenderable.Object.(*v45.Text)
		} else {
			str = v45.NewText(scr.Font, scale, scale)
			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
		str.SetString(text)
		str.SetColor(color)
		str.SetPosition(mgl32.Vec2{x, y})
	}
	return key
}

func (scr *Screen) Render(key Key) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			if str, ok := renderable.Object.(*v45.Text); ok {
				gl.UseProgram(scr.Shaders[STRING]) // Use the custom program
				projectionUniform := gl.GetUniformLocation(scr.Shaders[STRING], gl.Str("projection\x00"))
				if projectionUniform >= 0 {
					gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
				}
				str.Draw() // Use the existing logic for drawing text
			}
		}
	}
}
