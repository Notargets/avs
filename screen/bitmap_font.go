package screen

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// LAL: Can generate a font here: https://snowb.org/

func (scr *Screen) LoadFontTexture(filePath string) error {
	// Step 1: Load image file
	imgFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to load font texture file %s: %v", filePath, err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("failed to decode image file %s: %v", filePath, err)
	}

	// Step 2: Convert image to RGBA format
	rgbaImg := image.NewRGBA(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			rgbaImg.Set(x, y, img.At(x, y))
		}
	}

	// Step 3: Generate OpenGL texture
	gl.GenTextures(1, &scr.FontTextureID)
	gl.BindTexture(gl.TEXTURE_2D, scr.FontTextureID)

	// Step 4: Upload texture data to the GPU
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,                            // Level
		gl.RGBA,                      // Internal format
		int32(rgbaImg.Bounds().Dx()), // Width
		int32(rgbaImg.Bounds().Dy()), // Height
		0,                            // Border
		gl.RGBA,                      // Data format
		gl.UNSIGNED_BYTE,             // Data type
		gl.Ptr(rgbaImg.Pix),          // Pixel data
	)

	// Step 5: Set texture parameters (wrapping, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)

	fmt.Println("Font texture loaded successfully.")
	return nil
}

type Character struct {
	VAO, VBO, UVBO, CBO uint32    // Vertex Array Object, Vertex Buffer Object, UV Buffer Object, Color Buffer Object
	Vertices            []float32 // Flat list of vertex positions [x1, y1, x2, y2, ...]
	UVs                 []float32 // Flat list of texture UVs [u1, v1, u2, v2, ...]
	Colors              []float32 // Flat list of color data [r1, g1, b1, r2, g2, b2, ...]
	ShaderProgram       uint32    // Shader program specific to this Character object
}

func (char *Character) Update(X, Y []float32, UVs []float32, Colors []float32) {
	// Validate the vertex count for quads
	if len(X) != 4 || len(Y) != 4 {
		panic(fmt.Sprintf("Invalid vertex count for Character. Expected 4 vertices, got %d", len(X)))
	}
	if len(UVs) != 8 {
		panic(fmt.Sprintf("Invalid UV count for Character. Expected 8 UVs, got %d", len(UVs)))
	}
	if len(Colors) != 12 {
		panic(fmt.Sprintf("Invalid color count for Character. Expected 12 colors (4 vertices x RGB), got %d", len(Colors)))
	}

	// Update the vertex positions
	char.Vertices = make([]float32, len(X)*2)
	for i := 0; i < len(X); i++ {
		char.Vertices[2*i] = X[i]
		char.Vertices[2*i+1] = Y[i]
	}

	// Update UV coordinates
	char.UVs = UVs

	// Update color data
	char.Colors = Colors

	// Upload vertex positions
	gl.BindVertexArray(char.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, char.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(char.Vertices)*4, gl.Ptr(char.Vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Upload UV coordinates
	gl.BindBuffer(gl.ARRAY_BUFFER, char.UVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(char.UVs)*4, gl.Ptr(char.UVs), gl.STATIC_DRAW)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)

	// Upload color data
	gl.BindBuffer(gl.ARRAY_BUFFER, char.CBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(char.Colors)*4, gl.Ptr(char.Colors), gl.STATIC_DRAW)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)
}

func (scr *Screen) AddCharacter(key Key, asciiChar rune, positionX, positionY, width, height float32, color [3]float32) (newKey Key) {
	if key == NEW {
		key = NewKey()
	}

	// Send a command to create or update a char object
	scr.RenderChannel <- func() {
		// Calculate UVs for the ASCII character in a 16x16 font atlas
		var charsPerRow, charsPerCol = 16, 16
		//atlasSize := float32(256) // Assuming the font atlas is 256x256
		charWidth := 1.0 / float32(charsPerRow)
		charHeight := 1.0 / float32(charsPerCol)

		charIndex := int(asciiChar) - 32 // ASCII 32 (' ') is the first printable character
		col := charIndex % charsPerRow
		row := charIndex / charsPerRow

		uMin := float32(col) * charWidth
		uMax := uMin + charWidth
		vMin := float32(row) * charHeight
		vMax := vMin + charHeight

		// Create UVs for the quad
		uvs := []float32{
			uMin, vMax, // Bottom-left
			uMax, vMax, // Bottom-right
			uMax, vMin, // Top-right
			uMin, vMin, // Top-left
		}

		// Create vertices for a quad at position (positionX, positionY)
		vertices := []float32{
			positionX, positionY,
			positionX + width, positionY,
			positionX + width, positionY + height,
			positionX, positionY + height,
		}

		// Color for each vertex
		colors := []float32{
			color[0], color[1], color[2],
			color[0], color[1], color[2],
			color[0], color[1], color[2],
			color[0], color[1], color[2],
		}

		// Generate a character object
		char := &Character{}
		char.addShader(scr)

		// Generate VAO, VBO, UVBO, CBO
		gl.GenVertexArrays(1, &char.VAO)
		gl.BindVertexArray(char.VAO)

		gl.GenBuffers(1, &char.VBO)
		gl.GenBuffers(1, &char.UVBO)
		gl.GenBuffers(1, &char.CBO)

		gl.BindVertexArray(0)

		// Upload character data
		char.Update(vertices, vertices, uvs, colors)

		// Store character in screen
		scr.Objects[key] = Renderable{
			Active: true,
			Object: char,
		}
	}

	return key
}

func (char *Character) Render(scr *Screen) {
	// Activate and bind the font texture
	gl.ActiveTexture(gl.TEXTURE0) // Activate texture unit 0
	gl.BindTexture(gl.TEXTURE_2D, scr.FontTextureID)

	// Use shader program
	gl.UseProgram(char.ShaderProgram)
	gl.BindVertexArray(char.VAO)

	// Upload the projection matrix
	projectionUniform := gl.GetUniformLocation(char.ShaderProgram, gl.Str("projection\x00"))
	if projectionUniform >= 0 {
		gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
	}

	// Bind the font texture to texture unit 0
	textureLocation := gl.GetUniformLocation(char.ShaderProgram, gl.Str("fontTexture\x00"))
	gl.Uniform1i(textureLocation, 0) // Texture unit 0

	// Draw character quad
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.BindVertexArray(0)
}

func (char *Character) addShader(scr *Screen) (shaderProgram uint32) {
	if _, present := scr.Shaders[CHARACTER]; !present {
		// Vertex Shader
		var vertexShaderSource = `
#version 450

layout (location = 0) in vec2 position;  // Position of the character
layout (location = 1) in vec2 uv;        // UV texture coordinates
layout (location = 2) in vec3 color;     // Color for the character

uniform mat4 projection;  // Projection matrix

out vec2 fragUV;         // Pass to fragment shader
out vec3 fragColor;      // Pass to fragment shader

void main() {
    gl_Position = projection * vec4(position, 0.0, 1.0); // Transform position
    fragUV = uv; // Pass the UVs to the fragment shader
    fragColor = color; // Pass the color to the fragment shader
}
` + "\x00"

		var fragmentShaderSource = `
#version 450

in vec2 fragUV;          // UV coordinates from vertex shader
in vec3 fragColor;       // Color from vertex shader

uniform sampler2D fontTexture; // Texture sampler for font

out vec4 outColor;

void main() {
    vec4 texColor = texture(fontTexture, fragUV); // Sample the font texture
    outColor = texColor * vec4(fragColor, 1.0);   // Multiply font color by vertex color
}
` + "\x00"
		scr.Shaders[CHARACTER] = compileShaderProgram(vertexShaderSource, fragmentShaderSource)
	}
	return scr.Shaders[CHARACTER]
}
