/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"

	"github.com/notargets/avs/screen/main_gl_thread_objects"

	"github.com/notargets/avs/utils"
)

type Screen struct {
	Window        SafeWindow // Concurrency safe so we can read in this thread
	Objects       map[utils.Key]*Renderable
	RenderChannel chan func()
}

type SafeWindow struct {
	value unsafe.Pointer
}

// Write sets a new value atomically
func (si *SafeWindow) Write(val *main_gl_thread_objects.Window) {
	atomic.StorePointer(&si.value, unsafe.Pointer(&val))
}

// Read atomically retrieves the variable's value.
// It returns an `int` type
func (si *SafeWindow) Read() *main_gl_thread_objects.Window {
	ptr := atomic.LoadPointer(&si.value)
	return *(*(*main_gl_thread_objects.Window))(ptr)
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor [4]float32, position main_gl_thread_objects.Position) *Screen {

	screen := &Screen{
		Objects:       make(map[utils.Key]*Renderable),
		RenderChannel: make(chan func(), 100),
	}
	// OpenGLReady is used to signal when OpenGL is fully initialized
	// Channel for synchronization
	doneChan := make(chan struct{})

	go func() {
		runtime.LockOSThread()

		// Open a default window. User needs to GetCurrentWindow before opening
		// a new Window to return to the default, as the win pointer is not
		// exposed
		screen.Window.Write(main_gl_thread_objects.NewWindow(width, height,
			xmin, xmax, ymin, ymax,
			scale, "Chart2D", screen.RenderChannel, bgColor, position))

		fmt.Println("[OpenGL] Initialization complete, signaling main thread.")
		doneChan <- struct{}{}

		// Start the event loop (OpenGL runs here)
		screen.EventLoop()
	}()
	// Wait for the OpenGL thread to signal readiness
	fmt.Println("[Main] Waiting for OpenGL initialization...")
	<-doneChan
	fmt.Println("[Main] OpenGL initialization complete, proceeding.")

	return screen
}

func (scr *Screen) Redraw() {
	scr.RenderChannel <- func() {}
}

func (scr *Screen) GetCurrentWindow() (win *main_gl_thread_objects.Window) {
	win = scr.Window.Read()
	return
}

func (scr *Screen) MakeContextCurrent(win *main_gl_thread_objects.Window) {
	scr.RenderChannel <- func() {
		scr.Window.Write(win)
		scr.Window.Read().MakeContextCurrent()
	}
}

func (scr *Screen) NewWindow(width, height uint32, xmin, xmax, ymin, ymax,
	scale float32, title string, bgColor [4]float32,
	position main_gl_thread_objects.Position) (win *main_gl_thread_objects.Window) {

	// Channel for synchronization
	fmt.Println("[NewWindow] Creating new Window")
	doneChan := make(chan struct{}, 2)
	scr.RenderChannel <- func() {
		fmt.Println("[NewWindow] Inside New Window")
		scr.Window.Write(main_gl_thread_objects.NewWindow(width, height, xmin, xmax,
			ymin, ymax, scale, title, scr.RenderChannel, bgColor, position))
		doneChan <- struct{}{}
	}
	<-doneChan

	win = scr.Window.Read()

	return
}
