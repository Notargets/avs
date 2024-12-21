/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"runtime"

	"github.com/notargets/avs/utils"
)

type Screen struct {
	Window        *Window
	Objects       map[utils.Key]*Renderable
	RenderChannel chan func()
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor [4]float32) *Screen {

	screen := &Screen{
		Objects:       make(map[utils.Key]*Renderable),
		RenderChannel: make(chan func(), 100),
	}
	// OpenGLReady is used to signal when OpenGL is fully initialized
	type OpenGLReady struct{}
	// Channel for synchronization
	initDone := make(chan OpenGLReady)

	go func(done chan OpenGLReady) {
		runtime.LockOSThread()

		screen.Window = NewWindow(width, height, xmin, xmax, ymin, ymax, scale,
			"Chart2D", screen.RenderChannel, bgColor, CENTER)

		// Notify the main thread that OpenGL is ready
		// fmt.Println("[OpenGL] Initialization complete, signaling main thread.")
		done <- OpenGLReady{}

		// Start the event loop (OpenGL runs here)
		screen.EventLoop()
	}(initDone)
	// Wait for the OpenGL thread to signal readiness
	// fmt.Println("[Main] Waiting for OpenGL initialization...")
	<-initDone
	// fmt.Println("[Main] OpenGL initialization complete, proceeding.")

	return screen
}

func (scr *Screen) Redraw() {
	scr.Window.Redraw()
}
