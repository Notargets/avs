/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

func (scr *Screen) EventLoop() {
	for {
		win := scr.Window.Read()
		if win.Window.ShouldClose() {
			break
		}
		// Wait for any input event (mouse, keyboard, resize, etc.)
		glfw.WaitEventsTimeout(0.02)

		// Process commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command() // Execute the command (
			// can include OpenGL calls)
			(scr.Window.Read()).NeedsRedraw = true
		default:
			// No command, continue
			break
		}

		// setupVertices the projection matrix if pan/zoom has changed
		if win.NeedsRedraw || win.PositionChanged || win.ScaleChanged {
			win.UpdateProjectionMatrix()
			win.PositionChanged = false
			win.ScaleChanged = false
			win.NeedsRedraw = false
			scr.fullScreenRender(win)
		}

	}
}
