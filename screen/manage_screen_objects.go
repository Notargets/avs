/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"

	"github.com/notargets/avs/utils"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/notargets/avs/screen/main_gl_thread_object_actions"
)

type ObjectGroup []interface{}

func NewObjectGroup(object interface{}) ObjectGroup {
	return ObjectGroup{object}
}

type Renderable struct {
	Visible bool
	Window  Window      // The target window for rendering
	Objects ObjectGroup // Any object that has a Render method (e.g., Line,
	// TriMesh)
}

func NewRenderable(window Window, object interface{}) (rb *Renderable) {
	rb = &Renderable{
		Visible: true,
		Window:  window,
		Objects: NewObjectGroup(object),
	}
	return
}

func (rb *Renderable) Add(key utils.Key) {
	// An object group is append only by design
	rb.Objects = append(rb.Objects, key)
}

func (rb *Renderable) SetVisible(isVisible bool) {
	rb.Visible = isVisible
}

func (scr *Screen) fullScreenRender() {
	// Clear the screen before rendering
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, obj := range scr.Objects {
		renderObjList := obj.Objects
		for _, object := range renderObjList {
			switch renderObj := object.(type) {
			// switch renderObj := obj.Object.(type) {
			case *main_gl_thread_object_actions.Line:
				renderObj.Render()
			case *main_gl_thread_object_actions.String:
				renderObj.Render(scr.projectionMatrix, scr.WindowWidth, scr.WindowHeight, scr.XMin, scr.XMax, scr.YMin, scr.YMax)
			default:
				fmt.Printf("Unknown object type: %T\n", renderObj)
			}
		}
	}
	// Swap buffers to present the frame
	scr.Window.Window.SwapBuffers()
}
