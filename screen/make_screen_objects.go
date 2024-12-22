/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"

	"github.com/notargets/avs/screen/main_gl_thread_objects"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/assets"
)

func (scr *Screen) NewLine(X, Y, Colors []float32,
	rt ...utils.RenderType) (key utils.Key) {
	key = utils.NewKey()

	// Create new line
	line := main_gl_thread_objects.NewLine(X, Y, Colors, scr.Window.Read().Shaders)

	scr.Objects[key] = NewRenderable(scr.Window.Read(), line)

	scr.Redraw()

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

	win := scr.Window.Read()
	str := main_gl_thread_objects.NewString(tf, x, y,
		text, win.Width, win.Height, win.Shaders)

	scr.Objects[key] = NewRenderable(win, str)

	scr.Redraw()

	return
}

func (scr *Screen) Printf(formatter *assets.TextFormatter, x, y float32,
	format string, args ...interface{}) (key utils.Key) {
	// Format the string using fmt.Sprintf
	text := fmt.Sprintf(format, args...)
	// Call NewString with the formatted text
	return scr.NewString(formatter, x, y, text)
}
