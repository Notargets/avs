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
	tf *assets.TextFormatter, XLabel, YLabel string, yCoordOfXAxis, xCoordOfYAxis float32,
	nSegs int) (key utils.Key) {

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
	// Horizontal X Axis line
	X, Y, C = utils.AddSegmentToLine(X, Y, C, xMin, yCoordOfXAxis, xMax,
		yCoordOfXAxis, axisColor)
	// VerticaL Y Axis line
	X, Y, C = utils.AddSegmentToLine(X, Y, C, xCoordOfYAxis, yMin, xCoordOfYAxis,
		yMax, axisColor)

	// X Axis
	var x, y = xMin, yCoordOfXAxis
	if yMin == 0 {
		y = yMin
	}
	for i := 0; i < nSegs; i++ {
		if x == xCoordOfYAxis {
			x = x + xInc
			continue
		}
		X, Y, C = utils.AddSegmentToLine(X, Y, C, x, y, x, y-yTickSize, tickColor)
		x = utils.ClampNearZero(x, xScale/1000.)
		scr.Printf(tf, x, y-(scr.GetWorldSpaceCharHeight(win, tf)+yTickSize),
			"%4.1f", x)
		x = x + xInc
	}
	// X Axis label
	scr.Printf(tf, xMax+yTickSize, yCoordOfXAxis, "%s",
		XLabel)

	// Y Axis
	ptfY := *tf
	tfY := &ptfY
	tfY.Centered = false
	x = xCoordOfYAxis
	y = yMin
	yTextDelta := utils.CalculateRightJustifiedTextOffset(yMin,
		scr.GetWorldSpaceCharWidth(win, tfY))
	for i := 0; i < nSegs; i++ {
		if yMin != 0 {
			if i == nSegs/2 {
				y = y + yInc
				continue
			}
		}
		X, Y, C = utils.AddSegmentToLine(X, Y, C, x, y, x-xTickSize, y, tickColor)
		y = utils.ClampNearZero(y, yScale/1000.)
		scr.Printf(tfY, x-yTextDelta, y, "%4.1f", y)
		y = y + yInc
	}
	// Y Axis Label
	scr.Printf(tf, xCoordOfYAxis+xTickSize, yMax, "%s",
		YLabel)
	return scr.NewLine(X, Y, C) // 2 points, so 2 * 3 = 6 colors
}
