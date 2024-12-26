/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"image/color"

	"github.com/notargets/avs/assets"
	"github.com/notargets/avs/utils"
)

func (scr *Screen) NewAxis(win *Window, axisColor color.Color,
	tf *assets.TextFormatter, yAxisLocation float32, nSegs int) (key utils.Key) {
	var (
		xMin, xMax           = win.xMin, win.xMax
		yMin, yMax           = win.yMin, win.yMax
		xScale, yScale       = xMax - xMin, yMax - yMin
		xInc                 = xScale / float32(nSegs-1)
		yInc                 = yScale / float32(nSegs-1)
		xTickSize, yTickSize = 0.020 * xScale, 0.020 * yScale
		tickColor            = axisColor
		X                    = make([]float32, 0)
		Y                    = make([]float32, 0)
		C                    = make([]float32, 0)
	)
	if nSegs%2 == 0 {
		panic("nSegs must be odd")
	}

	// Generate color array for 2 vertices per axis (X-axis and Y-axis)
	X, Y, C = utils.AddSegmentToLine(X, Y, C, xMin, 0, xMax, 0, axisColor)
	X, Y, C = utils.AddSegmentToLine(X, Y, C, yAxisLocation, yMin, yAxisLocation,
		yMax, axisColor)

	// Draw ticks along X axis
	var x, y = xMin, float32(0) // X Axis is always drawn at Y = 0
	for i := 0; i < nSegs; i++ {
		if x == yAxisLocation {
			x = x + xInc
			continue
		}
		X, Y, C = utils.AddSegmentToLine(X, Y, C, x, y, x, y-yTickSize, tickColor)
		x = utils.ClampNearZero(x, xScale/1000.)
		scr.Printf(tf, x, y-(scr.GetWorldSpaceCharHeight(win, tf)+yTickSize),
			"%4.1f", x)
		x = x + xInc
	}
	ptfY := *tf
	tfY := &ptfY
	tfY.Centered = false
	x = yAxisLocation
	y = yMin
	yTextDelta := utils.CalculateRightJustifiedTextOffset(yMin,
		scr.GetWorldSpaceCharWidth(win, tfY))
	for i := 0; i < nSegs; i++ {
		if i == nSegs/2 {
			y = y + yInc
			continue
		}
		X, Y, C = utils.AddSegmentToLine(X, Y, C, x, y, x-xTickSize, y, tickColor)
		y = utils.ClampNearZero(y, yScale/1000.)
		scr.Printf(tfY, x-yTextDelta, y, "%4.1f", y)
		y = y + yInc
	}
	// chart.Screen.ChangePosition(0.0, 0.0)
	return scr.NewLine(X, Y, C) // 2 points, so 2 * 3 = 6 colors
}
