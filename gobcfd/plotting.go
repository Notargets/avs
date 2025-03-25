/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package main

import (
	"math"

	"github.com/notargets/avs/chart2d"
	"github.com/notargets/avs/geometry"
	"github.com/notargets/avs/utils"
)

func PlotMesh(gm geometry.TriMesh) {
	var (
		xMin, xMax = float32(math.MaxFloat32), float32(-math.MaxFloat32)
		yMin, yMax = float32(math.MaxFloat32), float32(-math.MaxFloat32)
	)
	xMin, xMax, yMin, yMax = getMinMax(gm.XY, xMin, xMax, yMin, yMax)
	GC.SetActiveChart(
		chart2d.NewChart2D(xMin, xMax, yMin, yMax,
			1024, 1024, utils.WHITE, utils.BLACK))
	GC.SetActiveWindow(GC.GetActiveChart().GetCurrentWindow())
	// Create a vector field including the three vertices
	GC.SetActiveMesh(GC.GetActiveChart().AddTriMesh(gm))
	for {
	}
}

func getMinMax(XY []float32, xi, xa, yi, ya float32) (xMin, xMax, yMin, yMax float32) {
	var (
		x, y  float32
		lenXY = len(XY) / 2
	)
	for i := 0; i < lenXY; i++ {
		x, y = XY[i*2+0], XY[i*2+1]
		if i == 0 {
			xMin = xi
			xMax = xa
			yMin = yi
			yMax = ya
		} else {
			if x < xMin {
				xMin = x
			}
			if x > xMax {
				xMax = x
			}
			if y < yMin {
				yMin = y
			}
			if y > yMax {
				yMax = y
			}
		}
	}
	return
}
