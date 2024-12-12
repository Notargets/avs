package screen

import (
	"fmt"
	"log"
	"runtime"
	"runtime/cgo"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/google/uuid"
)

type Key uuid.UUID

func NewKey() Key {
	return Key(uuid.New())
}

var (
	NEW = Key(uuid.Nil)
)

type Screen struct {
	Shaders          ShaderPrograms // Stores precompiled shaders for all graphics types
	Window           *glfw.Window
	FontTextureID    uint32 // Texture ID for the font atlas
	Objects          map[Key]Renderable
	RenderChannel    chan func()
	Scale            float32
	Position         [2]float32
	isDragging       bool
	lastX            float64
	lastY            float64
	projectionMatrix mgl32.Mat4
	ScreenWidth      int
	ScreenHeight     int
	XMin, XMax       float32
	YMin, YMax       float32
	PanSpeed         float32
	ZoomSpeed        float32
	ZoomFactor       float32
	PositionChanged  bool
	ScaleChanged     bool
	NeedsRedraw      bool
}

type Renderable struct {
	Active bool
	Object interface{} // Any object that has a Render method (e.g., Line, TriMesh)
}

func NewScreen(width, height int, xmin, xmax, ymin, ymax float32) *Screen {
	screen := &Screen{
		Shaders:       make(ShaderPrograms),
		Objects:       make(map[Key]Renderable),
		RenderChannel: make(chan func(), 100),
		isDragging:    false,
		ScreenWidth:   width,
		ScreenHeight:  height,
		XMin:          float32(xmin),
		XMax:          float32(xmax),
		YMin:          float32(ymin),
		YMax:          float32(ymax),
		PanSpeed:      1.0,
		ZoomSpeed:     1.0,
		ZoomFactor:    1.0,
		Scale:         0.9,
		Position:      [2]float32{0, 0},
		NeedsRedraw:   true,
	}

	// Launch the OpenGL thread
	go func() {
		runtime.LockOSThread()

		if err := glfw.Init(); err != nil {
			log.Fatalln("Failed to initialize glfw:", err)
		}

		window, err := glfw.CreateWindow(width, height, "Chart2D", nil, nil)
		if err != nil {
			panic(err)
		}
		// Get primary monitor video mode (used to get the screen dimensions)
		monitor := glfw.GetPrimaryMonitor()
		videoMode := monitor.GetVideoMode()

		// Calculate the position to center the window
		screenWidth := videoMode.Width
		screenHeight := videoMode.Height
		windowX := (screenWidth - width) / 2
		windowY := (screenHeight - height) / 2

		// Set the window position to the calculated coordinates
		window.SetPos(windowX, windowY)

		window.MakeContextCurrent()

		if err := gl.Init(); err != nil {
			log.Fatalln("Failed to initialize OpenGL context:", err)
		}
		gl.ClearColor(0.3, 0.3, 0.3, 1.0)

		// Store window reference
		screen.Window = window

		// Enable VSync
		glfw.SwapInterval(1)

		// Call the GL screen initialization
		gl.Viewport(0, 0, int32(width), int32(height))
		screen.updateProjectionMatrix()
		screen.SetCallbacks()

		// Force the first frame to render
		screen.PositionChanged = true
		screen.ScaleChanged = true

		// Start the event loop (OpenGL runs here)
		screen.EventLoop()
	}()

	return screen
}

func (scr *Screen) InitGLScreen(width, height int) {
	var err error
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	scr.Window, err = glfw.CreateWindow(width, height, "Chart2D", nil, nil)
	if err != nil {
		panic(err)
	}
	scr.Window.MakeContextCurrent()

	// Check if context is properly active
	if glfw.GetCurrentContext() == nil {
		log.Fatalln("GLFW Context is not current!")
	}

	// Initialize OpenGL function pointers
	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL context:", err)
	}

	// Check OpenGL version (optional, but useful for debugging)
	version := gl.GoStr(gl.GetString(gl.VERSION))
	if version == "" {
		log.Fatalln("OpenGL context not properly initialized")
	}
	fmt.Println("OpenGL version:", version)

	// Check for OpenGL errors
	checkGLError("glfw MakeContextCurrent")

	// Enable VSync (limit frame rate to refresh rate)
	glfw.SwapInterval(1)

	handle := cgo.NewHandle(scr)
	scr.Window.SetUserPointer(unsafe.Pointer(&handle))

	scr.Window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.mouseButtonCallback(w, button, action, mods)
	})

	scr.Window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.cursorPositionCallback(w, xpos, ypos)
	})

	scr.Window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.scrollCallback(w, xoff, yoff)
	})

	scr.Window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		handle := (*cgo.Handle)(w.GetUserPointer())
		scr := handle.Value().(*Screen)
		scr.resizeCallback(w, width, height)
	})

	return
}

func (scr *Screen) SetBackgroundColor(color [4]float32) {
	scr.RenderChannel <- func() {
		//gl.ClearColor(r, g, b, a)
		gl.ClearColor(color[0], color[1], color[2], color[3])
	}
}

func (scr *Screen) ChangeScale(scale float32) {
	scr.RenderChannel <- func() {
		scr.Scale = scale
		scr.ScaleChanged = true
	}
}

func (scr *Screen) ChangePosition(x, y float32) {
	scr.RenderChannel <- func() {
		scr.Position = [2]float32{x, y}
		scr.PositionChanged = true
	}
}

// checkGLError decodes OpenGL error codes into human-readable form and panics if an error occurs
func checkGLError(message string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		var errorMessage string
		switch err {
		case gl.INVALID_ENUM:
			errorMessage = "GL_INVALID_ENUM: An unacceptable value is specified for an enumerated argument."
		case gl.INVALID_VALUE:
			errorMessage = "GL_INVALID_VALUE: A numeric argument is out of range."
		case gl.INVALID_OPERATION:
			errorMessage = "GL_INVALID_OPERATION: The specified operation is not allowed in the current state."
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			errorMessage = "GL_INVALID_FRAMEBUFFER_OPERATION: The framebuffer object is not complete."
		case gl.OUT_OF_MEMORY:
			errorMessage = "GL_OUT_OF_MEMORY: There is not enough memory left to execute the command."
		case gl.STACK_UNDERFLOW:
			errorMessage = "GL_STACK_UNDERFLOW: An attempt has been made to perform an operation that would cause an internal stack to underflow."
		case gl.STACK_OVERFLOW:
			errorMessage = "GL_STACK_OVERFLOW: An attempt has been made to perform an operation that would cause an internal stack to overflow."
		default:
			errorMessage = fmt.Sprintf("Unknown OpenGL error code: 0x%X", err)
		}
		panic(fmt.Sprintf("OpenGL Error [%s]: %s (0x%X)", message, errorMessage, err))
	}
}
