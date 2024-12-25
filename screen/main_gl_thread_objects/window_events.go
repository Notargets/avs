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
	win.window.SetMouseButtonCallback(win.mouseButtonCallback)
	win.window.SetCursorPosCallback(win.cursorPositionCallback)
	win.window.SetScrollCallback(win.scrollCallback)
	win.window.SetSizeCallback(win.resizeCallback)
	win.window.SetFocusCallback(win.focusCallback)
}

func (win *Window) focusCallback(w *glfw.Window, focused bool) {
	if focused {
		// fmt.Printf("window: %v is now focused\n", win.windowIndex)
		win.MakeContextCurrent()
	}
}

func (win *Window) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton,
	action glfw.Action, mods glfw.ModifierKey) {
	switch button {
	case glfw.MouseButtonLeft:
		return
	case glfw.MouseButtonRight:
		if action == glfw.Press {
			win.isDragging = true
			win.lastX, win.lastY = w.GetCursorPos()
		} else if action == glfw.Release {
			win.isDragging = false
		}
	case glfw.MouseButtonMiddle:
		return
	default:
		return
	}
}

func (win *Window) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if win.isDragging {
		width, height := win.window.GetSize()

		// Calculate movement in world coordinates (pan logic)
		dx := float32(xpos-win.lastX) / float32(width) * (win.xMax - win.xMin) / win.scale
		dy := float32(ypos-win.lastY) / float32(height) * (win.yMax - win.yMin) / win.scale

		// setupVertices world position
		win.positionDelta[0] -= dx // X-axis pan
		win.positionDelta[1] += dy // Y-axis pan (
		// inverted since screen Y is inverted)

		// Mark the screen as needing a projection update
		win.positionChanged = true

		// setupVertices cursor tracking position
		win.lastX = xpos
		win.lastY = ypos
	}
}

func (win *Window) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// fmt.Printf("Scrolling window %v\n", win.windowIndex)
	// Adjust the zoom factor based on scroll input
	win.zoomFactor *= 1.0 + float32(yoff)*0.1*win.zoomSpeed

	// Constrain the zoom factor to prevent excessive zoom
	if win.zoomFactor < 0.1 {
		win.zoomFactor = 0.1
	}
	if win.zoomFactor > 10.0 {
		win.zoomFactor = 10.0
	}

	// Also adjust the scale value (legacy, previously working logic)
	win.scale *= 1.0 + float32(yoff)*0.1*win.zoomSpeed

	// Constrain the scale to prevent excessive zoom (if needed)
	if win.scale < 0.1 {
		win.scale = 0.1
	}
	if win.scale > 10.0 {
		win.scale = 10.0
	}

	// Flag that the scale has changed to trigger re-rendering
	win.scaleChanged = true
}

func (win *Window) resizeCallback(w *glfw.Window, width, height int) {
	// setupVertices screen dimensions
	win.width = uint32(width)
	win.height = uint32(height)

	// setupVertices OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Mark that a change occurred so the view is updated
	win.positionChanged = true
	win.scaleChanged = true
}

func (win *Window) SetBackgroundColor(screenColor color.Color) {
	fc := utils.ColorToFloat32(screenColor)
	gl.ClearColor(fc[0], fc[1], fc[2], fc[3])
}

func (win *Window) ChangeScale(scale float32) {
	win.scale = scale
	win.scaleChanged = true
}

func (win *Window) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		win.zoomSpeed = 1.0
		return
	}
	win.zoomSpeed = speed
}

func (win *Window) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		win.panSpeed = 1.0
		return
	}
	win.panSpeed = speed
}
