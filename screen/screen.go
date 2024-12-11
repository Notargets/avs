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

func (scr *Screen) SetCallbacks() {
	scr.Window.SetMouseButtonCallback(scr.mouseButtonCallback)
	scr.Window.SetCursorPosCallback(scr.cursorPositionCallback)
	scr.Window.SetScrollCallback(scr.scrollCallback)
}

func (scr *Screen) EventLoop() {
	for !scr.Window.ShouldClose() {
		// Process input events like key presses, mouse, etc.
		glfw.PollEvents()

		// Process any commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command()
		default:
			// No command to process
		}

		// Update the projection matrix if pan/zoom has changed
		if scr.PositionChanged || scr.ScaleChanged {
			fmt.Println("Updating projection matrix...")
			scr.updateProjectionMatrix()
			scr.PositionChanged = false
			scr.ScaleChanged = false
		}

		// Clear the screen before rendering
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render all active objects (type-coerce and render)
		for _, renderable := range scr.Objects {
			if renderable.Active {
				switch obj := renderable.Object.(type) {
				case *Line:
					obj.Render(scr)
				case *TriMesh:
					//obj.Render(scr)
				case *TriMeshEdges:
					//obj.Render(scr)
				case *TriMeshContours:
					//obj.Render(scr)
				case *TriMeshSmooth:
					//obj.Render(scr)
				default:
					fmt.Printf("Unknown object type: %T\n", obj)
				}
			}
		}

		// Swap buffers to present the frame
		scr.Window.SwapBuffers()
	}
}

func (scr *Screen) SetObjectActive(key uuid.UUID, active bool) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			renderable.Active = active
			scr.Objects[key] = renderable
		}
	}
}

func (scr *Screen) fullScreenRender() {
	for _, obj := range scr.Objects {
		switch renderObj := obj.Object.(type) {
		case *Line:
			renderObj.Render(scr)
		case *TriMesh:
			//renderObj.Render(scr)
		case *TriMeshEdges:
			//renderObj.Render(scr)
		case *TriMeshContours:
			//renderObj.Render(scr)
		case *TriMeshSmooth:
			//renderObj.Render(scr)
		default:
			fmt.Printf("Unknown object type: %T\n", renderObj)
		}
	}
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

func (scr *Screen) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		scr.ZoomSpeed = 1.0
		return
	}
	scr.ZoomSpeed = speed
}
func (scr *Screen) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		scr.PanSpeed = 1.0
		return
	}
	scr.PanSpeed = speed
}

func (scr *Screen) updateProjectionMatrix() {
	// Get the aspect ratio of the window
	aspectRatio := float32(scr.ScreenWidth) / float32(scr.ScreenHeight)

	// Calculate the world coordinates relative to the zoom and position
	var xRange, yRange float32
	if aspectRatio > 1.0 {
		// Landscape orientation
		xRange = (scr.XMax - scr.XMin) / scr.Scale
		yRange = xRange / aspectRatio
	} else {
		// Portrait orientation
		yRange = (scr.YMax - scr.YMin) / scr.Scale
		xRange = yRange * aspectRatio
	}

	// Apply position offset (world shift)
	xmin := scr.Position[0] - xRange/2.0
	xmax := scr.Position[0] + xRange/2.0
	ymin := scr.Position[1] - yRange/2.0
	ymax := scr.Position[1] + yRange/2.0

	// Update the orthographic projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Upload the new projection matrix to all shaders
	for renderType, shaderProgram := range scr.Shaders {
		projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
		if projectionUniform < 0 {
			fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
		} else {
			gl.UseProgram(shaderProgram)
			gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
		}
	}
}

func (scr *Screen) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonRight && action == glfw.Press {
		scr.isDragging = true
		scr.lastX, scr.lastY = w.GetCursorPos()
	} else if button == glfw.MouseButtonRight && action == glfw.Release {
		scr.isDragging = false
	}
}

func (scr *Screen) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if scr.isDragging {
		width, height := w.GetSize()

		// Calculate movement in world coordinates (pan logic)
		dx := float32(xpos-scr.lastX) / float32(width) * (scr.XMax - scr.XMin) / scr.Scale
		dy := float32(ypos-scr.lastY) / float32(height) * (scr.YMax - scr.YMin) / scr.Scale

		// Update world position
		scr.Position[0] -= dx // X-axis pan
		scr.Position[1] += dy // Y-axis pan (inverted since screen Y is inverted)

		// Mark the screen as needing a projection update
		scr.PositionChanged = true

		// Update cursor tracking position
		scr.lastX = xpos
		scr.lastY = ypos
	}
}

func (scr *Screen) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Calculate new zoom scale (increase/decrease)
	scaleChange := 1.0 + float32(yoff)*0.1*scr.ZoomSpeed
	newScale := scr.Scale * scaleChange

	// Constrain zoom level (avoid excessive zoom-in/out)
	if newScale < 0.1 {
		newScale = 0.1
	}
	if newScale > 10.0 {
		newScale = 10.0
	}

	// Update the scale
	scr.Scale = newScale

	// Mark screen to update projection matrix
	scr.ScaleChanged = true
}

func (scr *Screen) resizeCallback(w *glfw.Window, width, height int) {
	scr.ScreenWidth = width
	scr.ScreenHeight = height

	// Update OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Calculate the aspect ratio
	aspectRatio := float32(width) / float32(height)

	// Adjust world bounds based on aspect ratio
	if aspectRatio > 1.0 {
		// Landscape (widescreen) adjustment
		viewHeight := (scr.YMax - scr.YMin)
		viewWidth := viewHeight * aspectRatio
		centerX := (scr.XMax + scr.XMin) / 2.0
		scr.XMin = centerX - viewWidth/2.0
		scr.XMax = centerX + viewWidth/2.0
	} else {
		// Portrait (tall screen) adjustment
		viewWidth := (scr.XMax - scr.XMin)
		viewHeight := viewWidth / aspectRatio
		centerY := (scr.YMin + scr.YMax) / 2.0
		scr.YMin = centerY - viewHeight/2.0
		scr.YMax = centerY + viewHeight/2.0
	}

	// Update the projection matrix for the new window size
	scr.updateProjectionMatrix()

	// Mark the position and scale as changed so the event loop will force a re-render
	scr.PositionChanged = true
	scr.ScaleChanged = true
}
