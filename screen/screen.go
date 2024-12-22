/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"runtime"
	"sync/atomic"
	"unsafe"

	"github.com/notargets/avs/utils"
)

type Screen struct {
	// Window        *Window
	Window        SafeWindow
	Objects       map[utils.Key]*Renderable
	RenderChannel chan func()
}

type SafeWindow struct {
	value unsafe.Pointer
}

// Write sets a new value atomically
func (si *SafeWindow) Write(val *Window) {
	atomic.StorePointer(&si.value, unsafe.Pointer(&val))
}

// Read atomically retrieves the variable's value.
// It returns an `int` type
func (si *SafeWindow) Read() *Window {
	ptr := atomic.LoadPointer(&si.value)
	return *(*(*Window))(ptr)
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor [4]float32, position Position) *Screen {

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

		screen.Window.Write(NewWindow(width, height, xmin, xmax, ymin, ymax,
			scale, "Chart2D", screen.RenderChannel, bgColor, position))

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
	scr.Window.Read().Redraw()
}

func (scr *Screen) MakeContextCurrent(win *Window) {
	scr.Window.Write(win)
	scr.Window.Read().MakeContextCurrent()
}
