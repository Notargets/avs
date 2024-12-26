/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"
	"image/color"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/utils"
)

type Screen struct {
	RenderChannel chan func()
	DoneChan      chan struct{} // Re-usable synchronization channel
	drawWindow    *Window
}

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor color.RGBA, position Position) (scr *Screen) {

	scr = &Screen{
		RenderChannel: make(chan func(), 100),
		DoneChan:      make(chan struct{}),
	}

	go func() {
		runtime.LockOSThread()

		// Open a default window. User needs to getCurrentWindow before opening
		// a new window to return to the default, as the win pointer is not
		// exposed
		win := newWindow(width, height,
			xmin, xmax, ymin, ymax,
			scale, "Chart2D", utils.ColorToFloat32(bgColor), position)

		scr.SetDrawWindow(win) // Set default draw window

		// fmt.Println("[OpenGL] Initialization complete, signaling main thread.")
		scr.DoneChan <- struct{}{}

		// Start the event loop (OpenGL runs here)
		scr.eventLoop()
	}()
	// Wait for the OpenGL thread to signal readiness
	// fmt.Println("[Main] Waiting for OpenGL initialization...")
	<-scr.DoneChan
	// fmt.Println("[Main] OpenGL initialization complete, proceeding.")

	return
}

func (scr *Screen) SetDrawWindow(drawWindow *Window) {
	scr.drawWindow = drawWindow
}

func (scr *Screen) GetCurrentWindow() (win *Window) {
	win = getCurrentWindow()
	return
}

func (scr *Screen) Redraw(win *Window) {
	win.redraw()
}

func (scr *Screen) NewWindow(width, height uint32, xmin, xmax, ymin, ymax,
	scale float32, title string, bgColor color.RGBA,
	position Position) (win *Window) {

	scr.RenderChannel <- func() {
		// fmt.Println("[newWindow] Inside New window")
		win = newWindow(width, height, xmin, xmax,
			ymin, ymax, scale, title, utils.ColorToFloat32(bgColor), position)
		scr.SetDrawWindow(win)
		scr.DoneChan <- struct{}{}
	}
	<-scr.DoneChan

	return
}

func (scr *Screen) NewLine(X, Y, Colors []float32,
	rt ...utils.RenderType) (key utils.Key) {
	key = utils.NewKey()

	scr.RenderChannel <- func() {
		// Create new line
		win := scr.drawWindow
		line := newLine(X, Y, Colors, win, rt...)
		win.newRenderable(key, line)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}
	<-scr.DoneChan

	return key
}

func (scr *Screen) NewPolyLine(X, Y, Colors []float32) (key utils.Key) {
	return scr.NewLine(X, Y, Colors, utils.POLYLINE)
}

func (scr *Screen) NewString(tf *assets.TextFormatter, x,
	y float32, text string) (key utils.Key) {
	key = utils.NewKey()

	if tf == nil {
		panic("textFormatter is nil")
	}

	scr.RenderChannel <- func() {
		win := scr.drawWindow
		str := newString(tf, x, y, text, win)

		win.newRenderable(key, str)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}
	<-scr.DoneChan

	return
}

func (scr *Screen) Printf(formatter *assets.TextFormatter, x, y float32,
	format string, args ...interface{}) (key utils.Key) {
	// Format the string using fmt.Sprintf
	text := fmt.Sprintf(format, args...)
	// Call newString with the formatted text
	return scr.NewString(formatter, x, y, text)
}

func (scr *Screen) eventLoop() {
	for {
		win := getCurrentWindow()
		if win.shouldClose() {
			break
		}
		// Wait for any input event (mouse, keyboard, resize, etc.)
		glfw.WaitEventsTimeout(0.02)

		// Process commands from the renderChannel
		select {
		case command := <-scr.RenderChannel:
			command() // Execute the command (
		default:
			// No command, continue
			break
		}

		// win = gl_thread_objects.getCurrentWindow()
		// setupVertices the projection matrix if pan/zoom has changed
		if win.positionScaleChanged() {
			win.redraw()
			win.resetPositionScaleTrackers()
		}

	}
}
