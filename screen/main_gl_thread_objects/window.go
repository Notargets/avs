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
	window           *glfw.Window
	xMin, xMax       float32
	yMin, yMax       float32
	scale            float32
	width, height    uint32
	isDragging       bool
	lastX, lastY     float64
	positionChanged  bool
	positionDelta    [2]float32
	scaleChanged     bool
	zoomFactor       float32
	zoomSpeed        float32
	panSpeed         float32
	renderChannel    chan func()
	projectionMatrix mgl32.Mat4
	shaders          map[utils.RenderType]uint32
	objects          map[utils.Key]*Renderable
	windowIndex      int
	doneChannel      chan struct{}
}

func NewWindow(width, height uint32, xMin, xMax, yMin, yMax, scale float32,
	title string, renderChannel chan func(), bgColor [4]float32,
	position Position) (win *Window) {

	var (
		err error
	)

	win = &Window{
		width:         width,
		height:        height,
		xMin:          xMin,
		xMax:          xMax,
		yMin:          yMin,
		yMax:          yMax,
		scale:         scale,
		renderChannel: renderChannel,
		isDragging:    false,
		panSpeed:      1.,
		zoomSpeed:     1.,
		zoomFactor:    1.,
		positionDelta: [2]float32{0, 0},
		scaleChanged:  false,
		shaders:       make(map[utils.RenderType]uint32),
		objects:       make(map[utils.Key]*Renderable),
		doneChannel:   make(chan struct{}),
	}
	// Launch the OpenGL thread
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize glfw:", err)
	}

	win.window, err = glfw.CreateWindow(int(width), int(height), title, nil,
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
	// fmt.Printf("window Number: %d\n", windowIndex.Read()+1)
	if position == AUTO {
		position = Position((windowIndex + 1) % 4)
	}
	// fmt.Printf("window Count+1 (current) = %d, Position = %d\n",
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

	win.window.MakeContextCurrent()

	if windowIndex == -1 {
		if err := gl.Init(); err != nil {
			log.Fatalln("Failed to initialize OpenGL context:", err)
		}
	}
	windowIndex++
	win.windowIndex = windowIndex
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
	AddStringShaders(win.shaders)
	AddLineShader(win.shaders)

	// Force the first frame to render
	win.positionChanged = true
	win.scaleChanged = true

	// Notify the main thread that OpenGL is ready
	return
}

func (win *Window) MakeContextCurrent() {
	win.renderChannel <- func() {
		currentWindow.WindowIndex = win.windowIndex
		currentWindow.Window = win
		win.window.MakeContextCurrent()
	}
}

func (win *Window) NewRenderable(key utils.Key, object interface{}) (
	rb *Renderable) {
	rb = &Renderable{
		Visible: true,
		Objects: NewObjectGroup(object),
	}
	win.objects[key] = rb
	return
}

func (win *Window) SetPos(windowX, windowY int) {
	// Set the window position to the calculated coordinates
	win.window.SetPos(windowX, windowY)
}

func (win *Window) UpdateProjectionMatrix() {
	// Get the aspect ratio of the window
	aspectRatio := float32(win.width) / float32(win.height)

	// Determine world coordinate range based on zoom and position
	xRange := (win.xMax - win.xMin) / win.zoomFactor / win.scale
	yRange := (win.yMax - win.yMin) / win.zoomFactor / win.scale

	// Calculate the current center of the view
	centerX := (win.xMin + win.xMax) / 2.0
	centerY := (win.yMin + win.yMax) / 2.0

	// ** Key Change ** - Proper "squish" logic for X and Y
	if aspectRatio > 1.0 {
		// The screen is wider than it is tall, so "stretch" Y relative to X
		yRange = yRange / aspectRatio
	} else {
		// The screen is taller than it is wide, so "stretch" X relative to Y
		xRange = xRange * aspectRatio
	}

	// Use positionDelta to adjust the camera's "pan" position in world space
	xmin := centerX - xRange/2.0 + win.positionDelta[0]
	xmax := centerX + xRange/2.0 + win.positionDelta[0]
	ymin := centerY - yRange/2.0 + win.positionDelta[1]
	ymax := centerY + yRange/2.0 + win.positionDelta[1]

	// calculate the orthographic projection matrix
	win.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Send the updated projection matrix to all shaders that share the world
	// view. FIXEDSTRING doesn't
	for renderType, shaderProgram := range win.shaders {
		// win.ActiveShaders.Range(func(key, value interface{}) bool {
		// Type assertion for the key and value
		if renderType != utils.FIXEDSTRING {
			projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
			if projectionUniform < 0 {
				fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
			} else {
				gl.UseProgram(shaderProgram)
				gl.UniformMatrix4fv(projectionUniform, 1, false,
					&win.projectionMatrix[0])
			}
		}

	}
}

func (win *Window) SwapBuffers() {
	win.window.SwapBuffers()
}

func (win *Window) Redraw() {
	win.SetCurrentWindow()
	win.UpdateProjectionMatrix()
	win.FullScreenRender()
}

func (win *Window) PositionScaleChanged() bool {
	if win.positionChanged || win.scaleChanged {
		return true
	} else {
		return false
	}
}

func (win *Window) ResetPositionScaleTrackers() {
	win.positionChanged = false
	win.scaleChanged = false
}

func (win *Window) ShouldClose() bool {
	return win.window.ShouldClose()
}
