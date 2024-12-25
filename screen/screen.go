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
	DoneChan      chan struct{} // Re-usable synchronization channel
	drawWindow    *main_gl_thread_objects.Window
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor [4]float32, position main_gl_thread_objects.Position) (scr *Screen) {

	scr = &Screen{
		RenderChannel: make(chan func(), 100),
		DoneChan:      make(chan struct{}),
	}

	go func() {
		runtime.LockOSThread()

		// Open a default window. User needs to GetCurrentWindow before opening
		// a new window to return to the default, as the win pointer is not
		// exposed
		win := main_gl_thread_objects.NewWindow(width, height,
			xmin, xmax, ymin, ymax,
			scale, "Chart2D", scr.RenderChannel, bgColor, position)

		scr.SetDrawWindow(win) // Set default draw window

		// fmt.Println("[OpenGL] Initialization complete, signaling main thread.")
		scr.DoneChan <- struct{}{}

		// Start the event loop (OpenGL runs here)
		scr.EventLoop()
	}()
	// Wait for the OpenGL thread to signal readiness
	// fmt.Println("[Main] Waiting for OpenGL initialization...")
	<-scr.DoneChan
	// fmt.Println("[Main] OpenGL initialization complete, proceeding.")

	return
}

func (scr *Screen) SetDrawWindow(drawWindow *main_gl_thread_objects.Window) {
	scr.drawWindow = drawWindow
}

func (scr *Screen) GetCurrentWindow() (win *main_gl_thread_objects.Window) {
	win = main_gl_thread_objects.GetCurrentWindow()
	return
}

func (scr *Screen) Redraw(win *main_gl_thread_objects.Window) {
	win.Redraw()
}

func (scr *Screen) NewWindow(width, height uint32, xmin, xmax, ymin, ymax,
	scale float32, title string, bgColor [4]float32,
	position main_gl_thread_objects.Position) (win *main_gl_thread_objects.Window) {

	scr.RenderChannel <- func() {
		// fmt.Println("[NewWindow] Inside New window")
		main_gl_thread_objects.NewWindow(width, height, xmin, xmax,
			ymin, ymax, scale, title, scr.RenderChannel, bgColor, position)
		scr.DoneChan <- struct{}{}
	}
	<-scr.DoneChan
	win = scr.GetCurrentWindow()

	return
}
