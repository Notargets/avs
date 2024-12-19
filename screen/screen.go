/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"
	"image/color"
	"log"
	"runtime"
	"runtime/cgo"
	"unsafe"

	"github.com/notargets/avs/screen/main_gl_thread_object_actions"

	"github.com/notargets/avs/utils"

	"golang.org/x/image/font"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Screen struct {
	Shaders          utils.ShaderPrograms // Stores precompiled shaders for all graphics types
	Window           *glfw.Window
	Font             font.Face // Using gltext font instead of raw OpenGL textures
	Objects          map[Key]Renderable
	RenderChannel    chan func()
	Scale            float32
	PositionDelta    [2]float32
	isDragging       bool
	lastX            float64
	lastY            float64
	projectionMatrix mgl32.Mat4
	WindowWidth      uint32
	WindowHeight     uint32
	XMin, XMax       float32
	YMin, YMax       float32
	PanSpeed         float32
	ZoomSpeed        float32
	ZoomFactor       float32
	PositionChanged  bool
	ScaleChanged     bool
	NeedsRedraw      bool
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32) *Screen {
	screen := &Screen{
		Shaders:       make(utils.ShaderPrograms),
		Objects:       make(map[Key]Renderable),
		RenderChannel: make(chan func(), 100),
		isDragging:    false,
		WindowWidth:   width,
		WindowHeight:  height,
		XMin:          float32(xmin),
		XMax:          float32(xmax),
		YMin:          float32(ymin),
		YMax:          float32(ymax),
		PanSpeed:      1.0,
		ZoomSpeed:     1.0,
		ZoomFactor:    1.0,
		Scale:         scale,
		PositionDelta: [2]float32{0, 0},
		NeedsRedraw:   true,
	}

	// Launch the OpenGL thread
	go func() {
		runtime.LockOSThread()

		if err := glfw.Init(); err != nil {
			log.Fatalln("Failed to initialize glfw:", err)
		}

		window, err := glfw.CreateWindow(int(width), int(height), "Chart2D", nil, nil)
		if err != nil {
			panic(err)
		}
		// Get primary monitor video mode (used to get the screen dimensions)
		monitor := glfw.GetPrimaryMonitor()
		videoMode := monitor.GetVideoMode()

		// Calculate the position to center the window
		screenWidth := videoMode.Width
		screenHeight := videoMode.Height
		windowX := (screenWidth - int(width)) / 2
		windowY := (screenHeight - int(height)) / 2

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
		log.Fatalln("OpenGL context not properly initializedFIXEDSTRING")
	}
	fmt.Println("OpenGL version:", version)

	// Check for OpenGL errors
	main_gl_thread_object_actions.CheckGLError("glfw MakeContextCurrent")

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

func (scr *Screen) SetBackgroundColor(screenColor color.Color) {

	scr.RenderChannel <- func() {
		//gl.ClearColor(r, g, b, a)
		fc := utils.ColorToFloat32(screenColor)
		gl.ClearColor(fc[0], fc[1], fc[2], fc[3])
	}
}

func (scr *Screen) ChangeScale(scale float32) {
	scr.RenderChannel <- func() {
		scr.Scale = scale
		scr.ScaleChanged = true
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
	aspectRatio := float32(scr.WindowWidth) / float32(scr.WindowHeight)

	// Determine world coordinate range based on zoom and position
	xRange := (scr.XMax - scr.XMin) / scr.ZoomFactor / scr.Scale
	yRange := (scr.YMax - scr.YMin) / scr.ZoomFactor / scr.Scale

	// Calculate the current center of the view
	centerX := (scr.XMin + scr.XMax) / 2.0
	centerY := (scr.YMin + scr.YMax) / 2.0

	// ** Key Change ** - Proper "squish" logic for X and Y
	if aspectRatio > 1.0 {
		// The screen is wider than it is tall, so "stretch" Y relative to X
		yRange = yRange / aspectRatio
	} else {
		// The screen is taller than it is wide, so "stretch" X relative to Y
		xRange = xRange * aspectRatio
	}

	// Use PositionDelta to adjust the camera's "pan" position in world space
	xmin := centerX - xRange/2.0 + scr.PositionDelta[0]
	xmax := centerX + xRange/2.0 + scr.PositionDelta[0]
	ymin := centerY - yRange/2.0 + scr.PositionDelta[1]
	ymax := centerY + yRange/2.0 + scr.PositionDelta[1]

	// Update the orthographic projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Send the updated projection matrix to all shaders
	for renderType, shaderProgram := range scr.Shaders {
		if renderType != utils.FIXEDSTRING {
			projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
			if projectionUniform < 0 {
				fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
			} else {
				gl.UseProgram(shaderProgram)
				gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
			}
		}
	}
}

func (scr *Screen) updateProjectionMatrixSquare() {
	// Get the aspect ratio of the window
	aspectRatio := float32(scr.WindowWidth) / float32(scr.WindowHeight)

	// Determine world coordinate range based on zoom and position
	xRange := (scr.XMax - scr.XMin) / scr.ZoomFactor / scr.Scale
	yRange := (scr.YMax - scr.YMin) / scr.ZoomFactor / scr.Scale

	// Calculate the current center of the view
	centerX := (scr.XMin + scr.XMax) / 2.0
	centerY := (scr.YMin + scr.YMax) / 2.0

	// Adjust for the aspect ratio, but keep the world coordinates intact
	if aspectRatio > 1.0 {
		// If the screen is wider than tall, we "stretch" xRange
		xRange = yRange * aspectRatio
	} else {
		// If the screen is taller than wide, we "stretch" yRange
		yRange = xRange / aspectRatio
	}

	// Use PositionDelta to adjust the camera's "pan" position in world space
	xmin := centerX - xRange/2.0 + scr.PositionDelta[0]
	xmax := centerX + xRange/2.0 + scr.PositionDelta[0]
	ymin := centerY - yRange/2.0 + scr.PositionDelta[1]
	ymax := centerY + yRange/2.0 + scr.PositionDelta[1]

	// Update the orthographic projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Send the updated projection matrix to all shaders
	for renderType, shaderProgram := range scr.Shaders {
		if renderType != utils.FIXEDSTRING {
			projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
			if projectionUniform < 0 {
				fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
			} else {
				gl.UseProgram(shaderProgram)
				gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
			}
		}
	}
}

func (scr *Screen) Redraw() {
	select {
	case scr.RenderChannel <- func() {}:
	default:
		// Channel is full, no need to push more redraws
	}
}
