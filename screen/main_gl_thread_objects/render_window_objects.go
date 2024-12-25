/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main_gl_thread_objects

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type currentWindowTracker struct {
	WindowIndex int
	Window      *Window
}

var currentWindow currentWindowTracker

func GetCurrentWindow() *Window {
	return currentWindow.Window
}

func (win *Window) SetCurrentWindow() (swapped bool, curWin *Window) {
	swapped = false
	curWin = GetCurrentWindow()
	if win != curWin {
		swapped = true
		win.MakeContextCurrent()
		win.window.Focus()
		currentWindow.WindowIndex = win.windowIndex
		currentWindow.Window = win
	}
	return
}

func (win *Window) FullScreenRender() {
	// Clear the screen before rendering
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, obj := range win.objects {
		renderObjList := obj.Objects
		for _, object := range renderObjList {
			switch renderObj := object.(type) {
			case *Line:
				renderObj.Render()
			case *String:
				renderObj.Render(win.projectionMatrix, win.width, win.height,
					win.xMin, win.xMax, win.yMin, win.yMax)
			default:
				fmt.Printf("Unknown object type: %T\n", renderObj)
			}
		}
	}
	// Swap buffers to present the frame
	win.SwapBuffers()
	glfw.PollEvents()
}
