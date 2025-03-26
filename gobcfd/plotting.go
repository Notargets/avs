/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package main

import (
	"log"
	"math"
	"time"

	"github.com/eiannone/keyboard"

	"github.com/notargets/avs/chart2d"
	"github.com/notargets/avs/geometry"
	"github.com/notargets/avs/utils"
)

func PlotMesh(gm geometry.TriMesh, quit <-chan struct{}) {
	defer func() {
		if err := keyboard.Close(); err != nil {
			log.Println("error closing keyboard:", err)
		}
	}()

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
	waitLoop(quit)
}

// waitLoop simulates a rendering loop running on the main thread.
func waitLoop(quit <-chan struct{}) {
	ticker := time.NewTicker(time.Second / 60) // 60 fps simulation
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			// do nothing
		}
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
