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

type Screen struct {
	Shaders          ShaderPrograms // Stores precompiled shaders for all graphics types
	Window           *glfw.Window
	Objects          map[uuid.UUID]Renderable
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
		Objects:       make(map[uuid.UUID]Renderable),
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
		Scale:         1.0,
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
		screen.InitShaders()
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

func (scr *Screen) SetBackgroundColor(r, g, b, a float32) {
	scr.RenderChannel <- func() {
		gl.ClearColor(r, g, b, a)
	}
}

func checkGLError(message string) {
	err := gl.GetError()
	if err != 0 {
		fmt.Printf("OpenGL Error [%s]: %d\n", message, err)
	}
}
