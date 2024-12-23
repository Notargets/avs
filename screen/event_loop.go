/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/notargets/avs/screen/main_gl_thread_objects"
)

func (scr *Screen) EventLoop() {
	for {
		win := main_gl_thread_objects.GetCurrentWindow()
		if win.Window.ShouldClose() {
			break
		}
		// Wait for any input event (mouse, keyboard, resize, etc.)
		glfw.WaitEventsTimeout(0.02)

		// Process commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command() // Execute the command (
		default:
			// No command, continue
			break
		}

		win = main_gl_thread_objects.GetCurrentWindow()
		// setupVertices the projection matrix if pan/zoom has changed
		if win.PositionChanged || win.ScaleChanged {
			win.Redraw()
			win.PositionChanged = false
			win.ScaleChanged = false
		}

	}
}
