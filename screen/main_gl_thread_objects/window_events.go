/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main_gl_thread_objects

import (
	"image/color"
	"log"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/notargets/avs/utils"
)

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
		if action == glfw.Press {
			win.MakeContextCurrent()
		}
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
