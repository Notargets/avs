/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func (scr *Screen) EventLoop() {
	for !scr.Window.ShouldClose() {
		// Wait for any input event (mouse, keyboard, resize, etc.)
		glfw.WaitEventsTimeout(0.02)

		// Process commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command() // Execute the command (can include OpenGL calls)
			scr.NeedsRedraw = true
		default:
			// No command, continue
			break
		}

		// Update the projection matrix if pan/zoom has changed
		if scr.NeedsRedraw || scr.PositionChanged || scr.ScaleChanged {
			scr.updateProjectionMatrix()
			scr.PositionChanged = false
			scr.ScaleChanged = false
			scr.NeedsRedraw = false
			scr.fullScreenRender()
		}

	}
}

func (scr *Screen) SetCallbacks() {
	scr.Window.SetMouseButtonCallback(scr.mouseButtonCallback)
	scr.Window.SetCursorPosCallback(scr.cursorPositionCallback)
	scr.Window.SetScrollCallback(scr.scrollCallback)
	scr.Window.SetSizeCallback(scr.resizeCallback)

}

func (scr *Screen) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch button {
	case glfw.MouseButtonLeft:
		return
	case glfw.MouseButtonRight:
		if action == glfw.Press {
			scr.isDragging = true
			scr.lastX, scr.lastY = w.GetCursorPos()
		} else if action == glfw.Release {
			scr.isDragging = false
		}
	case glfw.MouseButtonMiddle:
		return
	default:
		return
	}
}

func (scr *Screen) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if scr.isDragging {
		width, height := w.GetSize()

		// Calculate movement in world coordinates (pan logic)
		dx := float32(xpos-scr.lastX) / float32(width) * (scr.XMax - scr.XMin) / scr.Scale
		dy := float32(ypos-scr.lastY) / float32(height) * (scr.YMax - scr.YMin) / scr.Scale

		// Update world position
		scr.PositionDelta[0] -= dx // X-axis pan
		scr.PositionDelta[1] += dy // Y-axis pan (inverted since screen Y is inverted)

		// Mark the screen as needing a projection update
		scr.PositionChanged = true

		// Update cursor tracking position
		scr.lastX = xpos
		scr.lastY = ypos
	}
}

func (scr *Screen) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Adjust the zoom factor based on scroll input
	scr.ZoomFactor *= 1.0 + float32(yoff)*0.1*scr.ZoomSpeed

	// Constrain the zoom factor to prevent excessive zoom
	if scr.ZoomFactor < 0.1 {
		scr.ZoomFactor = 0.1
	}
	if scr.ZoomFactor > 10.0 {
		scr.ZoomFactor = 10.0
	}

	// Also adjust the scale value (legacy, previously working logic)
	scr.Scale *= 1.0 + float32(yoff)*0.1*scr.ZoomSpeed

	// Constrain the scale to prevent excessive zoom (if needed)
	if scr.Scale < 0.1 {
		scr.Scale = 0.1
	}
	if scr.Scale > 10.0 {
		scr.Scale = 10.0
	}

	// Flag that the scale has changed to trigger re-rendering
	scr.ScaleChanged = true
}

func (scr *Screen) resizeCallback(w *glfw.Window, width, height int) {
	// Update screen dimensions
	scr.WindowWidth = uint32(width)
	scr.WindowHeight = uint32(height)

	// Update OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Recompute the projection matrix to maintain the aspect ratio and world bounds
	scr.updateProjectionMatrix()

	// Mark that a change occurred so the view is updated
	scr.PositionChanged = true
	scr.ScaleChanged = true
}
