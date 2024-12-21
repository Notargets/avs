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
	"sync"

	"github.com/notargets/avs/screen/main_gl_thread_objects"

	"github.com/notargets/avs/utils"

	"golang.org/x/image/font"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Screen struct {
	Window        Window
	Font          font.Face // Using gltext font instead of raw OpenGL textures
	Objects       map[utils.Key]*Renderable
	RenderChannel chan func()
	// ActiveShaders sync.Map[utils.RenderType]uint32 // We store a pointer to the
	ActiveShaders sync.Map // We store a pointer to the
	// package shader
	// vars
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
		Objects:       make(map[utils.Key]*Renderable),
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
	// OpenGLReady is used to signal when OpenGL is fully initialized
	type OpenGLReady struct{}
	// Channel for synchronization
	initDone := make(chan OpenGLReady)

	go func(done chan OpenGLReady) {
		runtime.LockOSThread()

		// Launch the OpenGL thread
		if err := glfw.Init(); err != nil {
			log.Fatalln("Failed to initialize glfw:", err)
		}

		window, err := glfw.CreateWindow(int(width), int(height), "Chart2D", nil, nil)
		if err != nil {
			panic(err)
		}
		screen.Window = Window{window}
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

		// Enable VSync
		glfw.SwapInterval(1)

		// Call the GL screen initialization
		gl.Viewport(0, 0, int32(width), int32(height))

		// For each object type in Screen, we need to load the shaders here
		main_gl_thread_objects.AddStringShaders()
		main_gl_thread_objects.AddLineShader()

		screen.SetCallbacks()

		// Force the first frame to render
		screen.PositionChanged = true
		screen.ScaleChanged = true

		// Notify the main thread that OpenGL is ready
		fmt.Println("[OpenGL] Initialization complete, signaling main thread.")
		done <- OpenGLReady{}

		// Start the event loop (OpenGL runs here)
		screen.EventLoop()
	}(initDone)
	// Wait for the OpenGL thread to signal readiness
	// fmt.Println("[Main] Waiting for OpenGL initialization...")
	<-initDone
	// fmt.Println("[Main] OpenGL initialization complete, proceeding.")

	return screen
}

func (scr *Screen) SetBackgroundColor(screenColor color.Color) {

	scr.RenderChannel <- func() {
		// gl.ClearColor(r, g, b, a)
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

	// calculate the orthographic projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Send the updated projection matrix to all shaders that share the world
	// view. FIXEDSTRING doesn't
	scr.ActiveShaders.Range(func(key, value interface{}) bool {
		// Type assertion for the key and value
		renderType, okKey := key.(utils.RenderType)
		shaderProgram, okValue := value.(uint32)

		if !okKey || !okValue {
			fmt.Println("[Error] Type assertion failed for ActiveShaders Range")
			return true // Continue iterating despite the error
		}

		if renderType != utils.FIXEDSTRING {
			projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
			if projectionUniform < 0 {
				fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
			} else {
				gl.UseProgram(shaderProgram)
				gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
			}
		}

		return true // Continue to the next item
	})

}

func (scr *Screen) Redraw() {
	select {
	case scr.RenderChannel <- func() {}:
	default:
		// Channel is full, no need to push more redraws
	}
}
