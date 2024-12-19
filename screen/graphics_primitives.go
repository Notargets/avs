package screen

import (
	"fmt"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/notargets/avs/screen/main_gl_thread_object_actions"

	"github.com/notargets/avs/utils"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/notargets/avs/assets"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (scr *Screen) SetObjectActive(key Key, active bool, window *glfw.Window) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			renderable.Active = active
			renderable.Window = window
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
		var line *main_gl_thread_object_actions.Line

		// Check if the object exists in the scene
		if existingRenderable, exists := scr.Objects[key]; exists {
			line = existingRenderable.Object.(*main_gl_thread_object_actions.Line)
		} else {
			// Create new line
			line = &main_gl_thread_object_actions.Line{LineType: renderType}
			if shader, present := scr.Shaders[utils.LINE]; !present {
				line.ShaderProgram = line.AddShader()
			} else {
				line.ShaderProgram = shader
			}
			scr.Shaders[utils.LINE] = line.ShaderProgram
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
		var str *main_gl_thread_object_actions.String
		if object, present := scr.Objects[key]; present {
			str = object.Object.(*main_gl_thread_object_actions.String)
		} else {
			str = &main_gl_thread_object_actions.String{
				Text:                   text,
				Position:               mgl32.Vec2{x, y},
				TextFormatter:          textFormatter,
				InitializedFIXEDSTRING: false,
				WindowWidth:            scr.WindowWidth,
				WindowHeight:           scr.WindowHeight,
				ShaderProgram:          math.MaxUint32,
			}
			//fmt.Printf("In NewString: ScreenFixed = %v\n", str.TextFormatter.ScreenFixed)
			if str.TextFormatter.ScreenFixed {
				str.StringType = utils.FIXEDSTRING
			} else {
				str.StringType = utils.STRING
			}

			if shader, present1 := scr.Shaders[str.StringType]; !present1 {
				str.ShaderProgram = str.AddShader()
			} else {
				str.ShaderProgram = shader
			}

			// Store the string in the screen objects
			scr.Objects[key] = Renderable{
				Active: true,
				Object: str,
				Window: scr.Window,
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
