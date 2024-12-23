/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main_gl_thread_objects

import (
	"fmt"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/notargets/avs/utils"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var windowIndex int

func init() {
	windowIndex = -1
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
	Objects          map[utils.Key]*Renderable
	WindowIndex      int
	doneChannel      chan struct{}
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
		Shaders:       make(map[utils.RenderType]uint32),
		Objects:       make(map[utils.Key]*Renderable),
		doneChannel:   make(chan struct{}),
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

	// Get primary monitor video mode (used to get the screen dimensions)
	monitor := glfw.GetPrimaryMonitor()
	videoMode := monitor.GetVideoMode()

	// Calculate the position to center the window
	screenWidth := videoMode.Width
	screenHeight := videoMode.Height

	// Put the window into a quadrant of the host window depending on window
	// number
	// fmt.Printf("Window Number: %d\n", windowIndex.Read()+1)
	if position == AUTO {
		position = Position((windowIndex + 1) % 4)
	}
	// fmt.Printf("Window Count+1 (current) = %d, Position = %d\n",
	// 	windowIndex.Read()+1, position)
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
	win.SetPos(windowX, windowY)

	win.Window.MakeContextCurrent()

	if windowIndex == -1 {
		if err := gl.Init(); err != nil {
			log.Fatalln("Failed to initialize OpenGL context:", err)
		}
	}
	windowIndex++
	win.WindowIndex = windowIndex
	currentWindow.WindowIndex = windowIndex
	currentWindow.Window = win

	win.SetCallbacks()

	gl.ClearColor(bgColor[0], bgColor[1], bgColor[2], bgColor[3])
	gl.Clear(gl.COLOR_BUFFER_BIT)
	win.SwapBuffers()

	// Enable VSync
	glfw.SwapInterval(1)

	// Call the GL screen initialization
	gl.Viewport(0, 0, int32(width), int32(height))

	// For each object type in Screen, we need to load the shaders here
	AddStringShaders(win.Shaders)
	AddLineShader(win.Shaders)

	// Force the first frame to render
	win.PositionChanged = true
	win.ScaleChanged = true

	// Notify the main thread that OpenGL is ready
	return
}

func (win *Window) MakeContextCurrent() {
	win.RenderChannel <- func() {
		currentWindow.WindowIndex = win.WindowIndex
		currentWindow.Window = win
		win.Window.MakeContextCurrent()
	}
}

func (win *Window) NewRenderable(key utils.Key, object interface{}) (
	rb *Renderable) {
	rb = &Renderable{
		Visible: true,
		Objects: NewObjectGroup(object),
	}
	win.Objects[key] = rb
	return
}

func (win *Window) SetPos(windowX, windowY int) {
	// Set the window position to the calculated coordinates
	win.Window.SetPos(windowX, windowY)
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

func (win *Window) SwapBuffers() {
	win.Window.SwapBuffers()
}

func (win *Window) Redraw() {
	// Temporarily grab Current Context for the redraw if needed
	needReset, curWin := win.SetCurrentWindow()
	win.UpdateProjectionMatrix()
	win.FullScreenRender()
	if needReset {
		curWin.SetCurrentWindow()
	}
}
