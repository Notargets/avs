/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"
	"runtime"

	"github.com/notargets/avs/geometry"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/utils"
)

type Screen struct {
	RenderChannel chan Command
	DoneChan      chan struct{} // Re-usable synchronization channel
	drawWindow    *Window
	queues        *utils.RRQueues
}

type Command struct {
	queueID    int8 // primary queue
	subQueueID int8 // allows for finer slicing within a queue
	command    func()
}

var adminQueueID = int8(0)

func NewScreen(width, height uint32, xmin, xmax, ymin, ymax, scale float32,
	bgColor interface{}, position Position) (scr *Screen) {

	scr = &Screen{
		RenderChannel: make(chan Command, 100),
		DoneChan:      make(chan struct{}),
		queues:        utils.NewRRQueues(), // Queue 0 is the admin queue
	}

	go func() {
		runtime.LockOSThread()

		// Open a default window. User needs to getCurrentWindow before opening
		// a new window to return to the default, as the win pointer is not
		// exposed
		win := newWindow(width, height,
			xmin, xmax, ymin, ymax,
			scale, "Chart2D", bgColor, position)

		scr.SetDrawWindow(win) // Set default draw window

		if qID := scr.queues.AddQueue(); qID != win.windowIndex {
			panic("queueID doesn't match window index")
		}

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
	scale float32, title string, bgColor interface{},
	position Position) (win *Window) {

	scr.RenderChannel <- Command{adminQueueID, 0, func() {
		// fmt.Println("[newWindow] Inside New window")
		win = newWindow(width, height, xmin, xmax,
			ymin, ymax, scale, title, bgColor, position)
		scr.SetDrawWindow(win)
		if qID := scr.queues.AddQueue(); qID != win.windowIndex {
			panic("queueID doesn't match window index")
		}
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

	return
}

func (scr *Screen) NewLine(XY []float32, ColorInput interface{},
	rt ...utils.RenderType) (key utils.Key) {
	key = utils.NewKey()

	Colors := utils.GetColorArray(ColorInput, len(XY)/2)

	var win = scr.drawWindow
	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		// Create new line
		line := newLine(XY, Colors, win, rt...)
		win.newRenderable(key, line)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

	return
}
func (scr *Screen) ToggleVisible(win *Window, key utils.Key) {
	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		rb := win.GetObject(key)
		if rb.Visible {
			rb.Visible = false
		} else {
			rb.Visible = true
		}
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan
	return
}

func (scr *Screen) UpdateLine(win *Window, key utils.Key, XY, Colors []float32) {
	line := win.objects[key].Objects[0].(*Line)

	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		// Update line data
		if line.UniColor {
			line.setupVertices(XY, nil)
		} else {
			line.setupVertices(XY, Colors)
		}
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

}

func (scr *Screen) NewShadedVertexScalar(vs *geometry.VertexScalar, fMin,
	fMax float32) (key utils.Key) {
	key = utils.NewKey()

	var win = scr.drawWindow
	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		// Create new line
		shadedTris := newShadedVertexScalar(vs, win, fMin, fMax)
		win.newRenderable(key, shadedTris)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

	return
}

func (scr *Screen) UpdateShadedVertexScalar(win *Window, key utils.Key,
	vs *geometry.VertexScalar) {
	var (
		rb      *Renderable
		present bool
	)
	if rb, present = win.objects[key]; !present {
		panic("object not present")
	}
	// shadedVertexScalar := win.objects[key].Objects[0].(*ShadedVertexScalar)
	shadedVertexScalar := rb.Objects[0].(*ShadedVertexScalar)

	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		shadedVertexScalar.updateVertexScalarData(vs)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

}

func (scr *Screen) NewContourVertexScalar(vs *geometry.VertexScalar, fMin,
	fMax float32, numContours int) (key utils.Key) {
	key = utils.NewKey()

	var win = scr.drawWindow
	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		// Create new line
		contourTris := newContourVertexScalar(vs, win, fMin, fMax, numContours)
		win.newRenderable(key, contourTris)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

	return
}

func (scr *Screen) UpdateContourVertexScalar(win *Window, key utils.Key,
	vs *geometry.VertexScalar) {
	var (
		rb      *Renderable
		present bool
	)
	if rb, present = win.objects[key]; !present {
		panic("object not present")
	}
	// shadedVertexScalar := win.objects[key].Objects[0].(*ShadedVertexScalar)
	contourVertexScalar := rb.Objects[0].(*ContourVertexScalar)

	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		contourVertexScalar.updateVertexScalarData(vs)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
	<-scr.DoneChan

}

func (scr *Screen) NewPolyLine(XY []float32, ColorInput interface{}) (key utils.Key) {
	return scr.NewLine(XY, ColorInput, utils.POLYLINE)
}

func (scr *Screen) NewString(tf *assets.TextFormatter, x,
	y float32, text string) (key utils.Key) {
	key = utils.NewKey()

	if tf == nil {
		panic("textFormatter is nil")
	}

	var win = scr.drawWindow
	scr.RenderChannel <- Command{win.windowIndex, 0, func() {
		str := newString(tf, x, y, text, win)

		win.newRenderable(key, str)
		win.redraw()
		scr.DoneChan <- struct{}{}
	}}
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
		glfw.WaitEventsTimeout(0.001)

		win := getCurrentWindow()
		if win.shouldClose() {
			break
		}

		// High-level event categorization
		switch {
		// Handle channel commands
		case func() bool {
			select {
			case command := <-scr.RenderChannel:
				scr.queues.Enqueue(int(command.queueID), command.command)
				return true
			default:
				return false
			}
		}():
			// Channel command handled, continue

		// Handle state change
		case win.positionScaleChanged():
			scr.RenderChannel <- Command{win.windowIndex, 0, func() {
				win.redraw()
				win.resetPositionScaleTrackers()
			}}
		// Fallback case
		default:
			// Idle task or yield CPU
			runtime.Gosched()
		}
		// Dequeue and run a command
		if commandI := scr.queues.Dequeue(); commandI != nil {
			commandI.(func())()
		}
	}
}

func (scr *Screen) GetWorldSpaceCharHeight(win *Window, tf *assets.TextFormatter) (
	height float32) {
	return tf.GetWorldSpaceCharHeight(win.yMax-win.yMin, win.width, win.height)
}

func (scr *Screen) GetWorldSpaceCharWidth(win *Window, tf *assets.TextFormatter) (
	height float32) {
	return tf.GetWorldSpaceCharWidth(win.xMax-win.xMin, win.yMax-win.yMin,
		win.width, win.height)
}
