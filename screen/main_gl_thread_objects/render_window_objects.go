/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main_gl_thread_objects

import (
	"fmt"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type currentWindowTracker struct {
	WindowIndex int
	Window      *Window
}

var CurrentWindow currentWindowTracker

func (win *Window) FullScreenRender() {
	// Clear the screen before rendering
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, obj := range win.Objects {
		renderObjList := obj.Objects
		for _, object := range renderObjList {
			switch renderObj := object.(type) {
			case *Line:
				renderObj.Render()
			case *String:
				renderObj.Render(win.ProjectionMatrix, win.Width, win.Height,
					win.XMin, win.XMax, win.YMin, win.YMax)
			default:
				fmt.Printf("Unknown object type: %T\n", renderObj)
			}
		}
	}
	// Swap buffers to present the frame
	win.SwapBuffers()
}
