/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"runtime"

	"github.com/notargets/avs/screen/main_gl_thread_objects"
)

type Screen struct {
	RenderChannel chan func()
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor [4]float32, position main_gl_thread_objects.Position) *Screen {

	screen := &Screen{
		RenderChannel: make(chan func(), 100),
	}

	// Channel for synchronization
	doneChan := make(chan struct{})
	go func() {
		runtime.LockOSThread()

		// Open a default window. User needs to GetCurrentWindow before opening
		// a new Window to return to the default, as the win pointer is not
		// exposed
		main_gl_thread_objects.NewWindow(width, height,
			xmin, xmax, ymin, ymax,
			scale, "Chart2D", screen.RenderChannel, bgColor, position)

		// fmt.Println("[OpenGL] Initialization complete, signaling main thread.")
		doneChan <- struct{}{}

		// Start the event loop (OpenGL runs here)
		screen.EventLoop()
	}()
	// Wait for the OpenGL thread to signal readiness
	// fmt.Println("[Main] Waiting for OpenGL initialization...")
	<-doneChan
	// fmt.Println("[Main] OpenGL initialization complete, proceeding.")

	return screen
}

func (scr *Screen) Redraw() {
	scr.RenderChannel <- func() {}
}

func (scr *Screen) GetCurrentWindow() (win *main_gl_thread_objects.Window) {
	win = main_gl_thread_objects.CurrentWindow.Window
	return
}

func (scr *Screen) MakeContextCurrent(win *main_gl_thread_objects.Window) {
	doneChan := make(chan struct{})
	scr.RenderChannel <- func() {
		win.MakeContextCurrent()
		doneChan <- struct{}{}
	}
	<-doneChan
}

func (scr *Screen) NewWindow(width, height uint32, xmin, xmax, ymin, ymax,
	scale float32, title string, bgColor [4]float32,
	position main_gl_thread_objects.Position) (win *main_gl_thread_objects.Window) {

	// Channel for synchronization
	// fmt.Println("[NewWindow] Creating new Window")
	doneChan := make(chan struct{}, 2)
	scr.RenderChannel <- func() {
		// fmt.Println("[NewWindow] Inside New Window")
		main_gl_thread_objects.NewWindow(width, height, xmin, xmax,
			ymin, ymax, scale, title, scr.RenderChannel, bgColor, position)
		doneChan <- struct{}{}
	}
	<-doneChan

	win = main_gl_thread_objects.CurrentWindow.Window

	return
}
