package screen

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/notargets/avs/utils"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/notargets/avs/assets"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (scr *Screen) SetObjectActive(key Key, active bool) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			renderable.Active = active
			scr.Objects[key] = renderable
		}
	}
}

func (scr *Screen) NewLine(key Key, X, Y, Colors []float32, rt ...utils.RenderType) (newKey Key) {
	if key == NEW {
		key = NewKey()
	}
	newKey = key

	var renderType = utils.LINE
	if len(rt) != 0 {
		renderType = utils.POLYLINE
	}

	// Send a command to create or update a line object
	scr.RenderChannel <- func() {
		var line *Line

		// Check if the object exists in the scene
		if existingRenderable, exists := scr.Objects[key]; exists {
			line = existingRenderable.Object.(*Line)
		} else {
			// Create new line
			line = &Line{LineType: renderType}
			line.ShaderProgram = line.addShader(scr)
			scr.Objects[key] = Renderable{
				Active: true,
				Object: line,
			}

			gl.GenVertexArrays(1, &line.VAO)
			gl.BindVertexArray(line.VAO)

			gl.GenBuffers(1, &line.VBO)
			gl.BindBuffer(gl.ARRAY_BUFFER, line.VBO)
			gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
			gl.EnableVertexAttribArray(0)

			gl.GenBuffers(1, &line.CBO)
			gl.BindBuffer(gl.ARRAY_BUFFER, line.CBO)
			gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
			gl.EnableVertexAttribArray(1)

			gl.BindVertexArray(0)
		}

		// Update vertex positions and color
		line.Update(X, Y, Colors)
	}

	return key
}

func (scr *Screen) NewPolyLine(key Key, X, Y, Colors []float32) (newKey Key) {
	return scr.NewLine(key, X, Y, Colors, utils.POLYLINE)
}

func (scr *Screen) NewString(key Key, textFormatter *assets.TextFormatter, x, y float32, text string) (newKey Key) {
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
				Text:                   text,
				Position:               mgl32.Vec2{x, y},
				TextFormatter:          textFormatter,
				initializedFIXEDSTRING: false,
				WindowWidth:            scr.WindowWidth,
				WindowHeight:           scr.WindowHeight,
			}
			//fmt.Printf("In NewString: ScreenFixed = %v\n", str.TextFormatter.ScreenFixed)
			if str.TextFormatter.ScreenFixed {
				str.StringType = utils.FIXEDSTRING
			} else {
				str.StringType = utils.STRING
			}
			str.ShaderProgram = str.addShader(scr)

			// Store the string in the screen objects
			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
			}
		}
	}
	return newKey
}

func (scr *Screen) Printf(formatter *assets.TextFormatter, x, y float32, format string, args ...interface{}) (newKey Key) {
	// Format the string using fmt.Sprintf
	text := fmt.Sprintf(format, args...)

	// Call NewString with the formatted text
	newKey = scr.NewString(NEW, formatter, x, y, text)

	return newKey
}

func (scr *Screen) GetWorldSpaceCharHeight(tf *assets.TextFormatter) (charHeight float32) {
	// Implement a scale factor to reduce the polygon size commensurate with the dynamic DPI scaling, relative to the
	// standard 72 DPI of the Opentype package
	//worldPerPixel := (scr.YMax - scr.YMin) / float32(scr.WindowHeight)
	worldPerPixel := (scr.YMax - scr.YMin) / float32(scr.WindowHeight)
	screenRatio := float32(scr.WindowHeight) / float32(scr.WindowWidth)
	pixelHeight := tf.TypeFace.FontHeight
	//fmt.Printf("pitch: %v, pixelHeight: %v, DPI: %v\n", tf.TypeFace.FontPitch,
	//	pixelHeight, tf.TypeFace.FontDPI)
	// Height includes the inter-line height, so divide by 1.5
	charHeight = (worldPerPixel) * float32(pixelHeight) * float32(72) / float32(tf.TypeFace.FontDPI) * screenRatio / 1.5
	return
}

func (scr *Screen) GetWorldSpaceCharWidth(tf *assets.TextFormatter) (charWidth float32) {
	charHeight := scr.GetWorldSpaceCharHeight(tf)
	// Scale the height by the world aspect ratio to get the width
	charWidth = charHeight * (scr.XMax - scr.XMin) / (scr.YMax - scr.YMin)
	return
}

func (scr *Screen) NewTextFormatter(fontBaseName, fontOptionName string, fontPitch int, fontColor color.Color,
	centered, screenFixed bool) (tf *assets.TextFormatter) {
	tf = assets.NewTextFormatter(fontBaseName, fontOptionName, fontPitch,
		int(scr.WindowWidth),
		fontColor, centered, screenFixed, scr.XMax-scr.XMin, scr.YMax-scr.YMin)
	return
}

func compileShaderProgram(vertexSource, fragmentSource string) uint32 {
	// Compile vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexSource) // Unpack both pointer and cleanup function
	gl.ShaderSource(vertexShader, 1, csource, nil)
	defer free() // Defer cleanup to release the C string memory
	gl.CompileShader(vertexShader)

	// Check for vertex shader compile errors
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
		fmt.Printf("Vertex Shader Compile Error: %s\n", log)
	}

	// Compile fragment shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentSource) // Unpack again for fragment shader
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	defer free() // Defer cleanup to release the C string memory
	gl.CompileShader(fragmentShader)

	// Check for fragment shader compile errors
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		fmt.Printf("Fragment Shader Compile Error: %s\n", log)
	}

	// Link the shader program
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	// Check for linking errors
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		fmt.Printf("Shader Link Error: %s\n", log)
	}

	// Clean up the compiled shaders after linking
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}
