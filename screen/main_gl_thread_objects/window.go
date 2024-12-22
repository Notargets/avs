/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main_gl_thread_objects

import (
	"fmt"
	"image/color"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/notargets/avs/utils"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var windowCount int

func init() {
	windowCount = -1
}

type Position uint8

const (
	TOPLEFT Position = iota
	TOPRIGHT
	BOTTOMLEFT
	BOTTOMRIGHT
	CENTER
	AUTO
)

type Window struct {
	Window           *glfw.Window
	XMin, XMax       float32
	YMin, YMax       float32
	Scale            float32
	Width, Height    uint32
	NeedsRedraw      bool
	IsDragging       bool
	LastX, LastY     float64
	PositionChanged  bool
	PositionDelta    [2]float32
	ScaleChanged     bool
	ZoomFactor       float32
	ZoomSpeed        float32
	PanSpeed         float32
	RenderChannel    chan func()
	ProjectionMatrix mgl32.Mat4
	Shaders          map[utils.RenderType]uint32
}

func NewWindow(width, height uint32, xMin, xMax, yMin, yMax, scale float32,
	title string, renderChannel chan func(), bgColor [4]float32,
	position Position) (win *Window) {

	var (
		err error
	)

	win = &Window{
		Width:         width,
		Height:        height,
		XMin:          xMin,
		XMax:          xMax,
		YMin:          yMin,
		YMax:          yMax,
		Scale:         scale,
		RenderChannel: renderChannel,
		IsDragging:    false,
		PanSpeed:      1.,
		ZoomSpeed:     1.,
		ZoomFactor:    1.,
		PositionDelta: [2]float32{0, 0},
		ScaleChanged:  false,
		NeedsRedraw:   true,
		Shaders:       make(map[utils.RenderType]uint32),
	}
	// Launch the OpenGL thread
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize glfw:", err)
	}

	win.Window, err = glfw.CreateWindow(int(width), int(height), title, nil,
		nil)
	if err != nil {
		panic(err)
	}
	window := win.Window

	// Get primary monitor video mode (used to get the screen dimensions)
	monitor := glfw.GetPrimaryMonitor()
	videoMode := monitor.GetVideoMode()

	// Calculate the position to center the window
	screenWidth := videoMode.Width
	screenHeight := videoMode.Height

	// Put the window into a quadrant of the host window depending on window
	// number
	// fmt.Printf("Window Number: %d\n", windowCount.Read()+1)
	if position == AUTO {
		position = Position((windowCount + 1) % 4)
	}
	// fmt.Printf("Window Count+1 (current) = %d, Position = %d\n",
	// 	windowCount.Read()+1, position)
	var windowX, windowY int
	switch position {
	case TOPLEFT:
		windowX = screenWidth / 32
		windowY = screenHeight / 32
	case BOTTOMLEFT:
		windowX = screenWidth / 32
		windowY = screenHeight/2 + screenHeight/32
	case BOTTOMRIGHT:
		windowX = screenWidth/2 + screenWidth/32
		windowY = screenHeight/2 + screenHeight/32
	case TOPRIGHT:
		windowX = screenWidth/2 + screenWidth/32
		windowY = screenHeight / 32
	case CENTER:
		windowX = (screenWidth - int(width)) / 2
		windowY = (screenHeight - int(height)) / 2
	}

	// Set the window position to the calculated coordinates
	window.SetPos(windowX, windowY)

	window.MakeContextCurrent()

	if windowCount == -1 {
		if err := gl.Init(); err != nil {
			log.Fatalln("Failed to initialize OpenGL context:", err)
		}
	}
	windowCount++

	gl.ClearColor(bgColor[0], bgColor[1], bgColor[2], bgColor[3])
	gl.Clear(gl.COLOR_BUFFER_BIT)
	window.SwapBuffers()

	// Enable VSync
	glfw.SwapInterval(1)

	// Call the GL screen initialization
	gl.Viewport(0, 0, int32(width), int32(height))

	// For each object type in Screen, we need to load the shaders here
	AddStringShaders(win.Shaders)
	AddLineShader(win.Shaders)

	win.SetCallbacks()

	// Force the first frame to render
	win.PositionChanged = true
	win.ScaleChanged = true

	// Notify the main thread that OpenGL is ready
	return
}

func (win *Window) MakeContextCurrent() {
	win.Window.MakeContextCurrent()
}

func (win *Window) SetPos(windowX, windowY int) {
	// Set the window position to the calculated coordinates
	win.Window.SetPos(windowX, windowY)
}

func (win *Window) SetCallbacks() {
	win.Window.SetMouseButtonCallback(win.mouseButtonCallback)
	win.Window.SetCursorPosCallback(win.cursorPositionCallback)
	win.Window.SetScrollCallback(win.scrollCallback)
	win.Window.SetSizeCallback(win.resizeCallback)
}

func (win *Window) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton,
	action glfw.Action, mods glfw.ModifierKey) {
	switch button {
	case glfw.MouseButtonLeft:
		return
	case glfw.MouseButtonRight:
		if action == glfw.Press {
			win.IsDragging = true
			win.LastX, win.LastY = w.GetCursorPos()
		} else if action == glfw.Release {
			win.IsDragging = false
		}
	case glfw.MouseButtonMiddle:
		return
	default:
		return
	}
}

func (win *Window) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if win.IsDragging {
		width, height := w.GetSize()

		// Calculate movement in world coordinates (pan logic)
		dx := float32(xpos-win.LastX) / float32(width) * (win.XMax - win.XMin) / win.Scale
		dy := float32(ypos-win.LastY) / float32(height) * (win.YMax - win.YMin) / win.Scale

		// setupVertices world position
		win.PositionDelta[0] -= dx // X-axis pan
		win.PositionDelta[1] += dy // Y-axis pan (
		// inverted since screen Y is inverted)

		// Mark the screen as needing a projection update
		win.PositionChanged = true

		// setupVertices cursor tracking position
		win.LastX = xpos
		win.LastY = ypos
	}
}

func (win *Window) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Adjust the zoom factor based on scroll input
	win.ZoomFactor *= 1.0 + float32(yoff)*0.1*win.ZoomSpeed

	// Constrain the zoom factor to prevent excessive zoom
	if win.ZoomFactor < 0.1 {
		win.ZoomFactor = 0.1
	}
	if win.ZoomFactor > 10.0 {
		win.ZoomFactor = 10.0
	}

	// Also adjust the scale value (legacy, previously working logic)
	win.Scale *= 1.0 + float32(yoff)*0.1*win.ZoomSpeed

	// Constrain the scale to prevent excessive zoom (if needed)
	if win.Scale < 0.1 {
		win.Scale = 0.1
	}
	if win.Scale > 10.0 {
		win.Scale = 10.0
	}

	// Flag that the scale has changed to trigger re-rendering
	win.ScaleChanged = true
}

func (win *Window) resizeCallback(w *glfw.Window, width, height int) {
	// setupVertices screen dimensions
	win.Width = uint32(width)
	win.Height = uint32(height)

	// setupVertices OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Recompute the projection matrix to maintain the aspect ratio and world bounds
	win.UpdateProjectionMatrix()

	// Mark that a change occurred so the view is updated
	win.PositionChanged = true
	win.ScaleChanged = true
}

func (win *Window) UpdateProjectionMatrix() {
	// Get the aspect ratio of the window
	aspectRatio := float32(win.Width) / float32(win.Height)

	// Determine world coordinate range based on zoom and position
	xRange := (win.XMax - win.XMin) / win.ZoomFactor / win.Scale
	yRange := (win.YMax - win.YMin) / win.ZoomFactor / win.Scale

	// Calculate the current center of the view
	centerX := (win.XMin + win.XMax) / 2.0
	centerY := (win.YMin + win.YMax) / 2.0

	// ** Key Change ** - Proper "squish" logic for X and Y
	if aspectRatio > 1.0 {
		// The screen is wider than it is tall, so "stretch" Y relative to X
		yRange = yRange / aspectRatio
	} else {
		// The screen is taller than it is wide, so "stretch" X relative to Y
		xRange = xRange * aspectRatio
	}

	// Use PositionDelta to adjust the camera's "pan" position in world space
	xmin := centerX - xRange/2.0 + win.PositionDelta[0]
	xmax := centerX + xRange/2.0 + win.PositionDelta[0]
	ymin := centerY - yRange/2.0 + win.PositionDelta[1]
	ymax := centerY + yRange/2.0 + win.PositionDelta[1]

	// calculate the orthographic projection matrix
	win.ProjectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Send the updated projection matrix to all shaders that share the world
	// view. FIXEDSTRING doesn't
	for renderType, shaderProgram := range win.Shaders {
		// win.ActiveShaders.Range(func(key, value interface{}) bool {
		// Type assertion for the key and value
		if renderType != utils.FIXEDSTRING {
			projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
			if projectionUniform < 0 {
				fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
			} else {
				gl.UseProgram(shaderProgram)
				gl.UniformMatrix4fv(projectionUniform, 1, false,
					&win.ProjectionMatrix[0])
			}
		}

	}
}

func (win *Window) SetBackgroundColor(screenColor color.Color) {
	doneChan := make(chan struct{})
	win.RenderChannel <- func() {
		fc := utils.ColorToFloat32(screenColor)
		gl.ClearColor(fc[0], fc[1], fc[2], fc[3])
		doneChan <- struct{}{}
	}
	<-doneChan
}

func (win *Window) ChangeScale(scale float32) {
	win.RenderChannel <- func() {
		win.Scale = scale
		win.ScaleChanged = true
	}
}

func (win *Window) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		win.ZoomSpeed = 1.0
		return
	}
	win.ZoomSpeed = speed
}

func (win *Window) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		win.PanSpeed = 1.0
		return
	}
	win.PanSpeed = speed
}

func (win *Window) SwapBuffers() {
	win.Window.SwapBuffers()
}

func (win *Window) Redraw() {
	select {
	case win.RenderChannel <- func() {}:
	default:
		// Channel is full, no need to push more redraws
	}
}
