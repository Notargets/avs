/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/google/uuid"
	"github.com/notargets/avs/screen/main_gl_thread_object_actions"
)

type Key uuid.UUID

func NewKey() Key {
	return Key(uuid.New())
}

var (
	NEW = Key(uuid.Nil)
)

type Renderable struct {
	Active bool
	Object interface{}  // Any object that has a Render method (e.g., Line, TriMesh)
	Window *glfw.Window // The target window for rendering
}

func (scr *Screen) SetObjectActive(key Key, active bool, window *glfw.Window) {
	scr.RenderChannel <- func() {
		if renderable, exists := scr.Objects[key]; exists {
			renderable.Active = active
			renderable.Window = window
		}
	}
	scr.Redraw()
}

func (scr *Screen) fullScreenRender() {
	// Clear the screen before rendering
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, obj := range scr.Objects {
		switch renderObj := obj.Object.(type) {
		case *main_gl_thread_object_actions.Line:
			renderObj.Render(scr.projectionMatrix)
		case *main_gl_thread_object_actions.String:
			renderObj.Render(scr.projectionMatrix, scr.WindowWidth, scr.WindowHeight, scr.XMin, scr.XMax, scr.YMin, scr.YMax)
		default:
			fmt.Printf("Unknown object type: %T\n", renderObj)
		}
	}
	// Swap buffers to present the frame
	scr.Window.SwapBuffers()
}
