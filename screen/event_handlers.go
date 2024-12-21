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
	for !scr.Window.Window.ShouldClose() {
		// Wait for any input event (mouse, keyboard, resize, etc.)
		glfw.WaitEventsTimeout(0.02)

		// Process commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command() // Execute the command (can include OpenGL calls)
			scr.Window.NeedsRedraw = true
		default:
			// No command, continue
			break
		}

		// setupVertices the projection matrix if pan/zoom has changed
		if scr.Window.NeedsRedraw || scr.Window.PositionChanged || scr.Window.ScaleChanged {
			scr.Window.UpdateProjectionMatrix()
			scr.Window.PositionChanged = false
			scr.Window.ScaleChanged = false
			scr.Window.NeedsRedraw = false
			scr.fullScreenRender()
		}

	}
}
