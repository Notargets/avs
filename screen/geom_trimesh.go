/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"github.com/notargets/avs/geometry"
	"github.com/notargets/avs/utils"
)

func (scr *Screen) NewTriMesh(win *Window, mesh geometry.TriMesh) (key utils.Key) {
	var (
		nTris  = len(mesh.TriVerts)
		nLines = 3 * nTris
		XY     = make([]float32, nLines*4) // 4 coords per line
	)
	var lineNumber int
	for _, tri := range mesh.TriVerts {
		// fmt.Println("Tri Indices: [%d,%d,%d]\n", tri[0], tri[1], tri[2])
		for n := 0; n < 3; n++ {
			XY[lineNumber*4] = mesh.XY[2*tri[n]]
			XY[lineNumber*4+1] = mesh.XY[2*tri[n]+1]
			nNext := n + 1
			if n == 2 {
				nNext = 0
			}
			XY[lineNumber*4+2] = mesh.XY[2*tri[nNext]]
			XY[lineNumber*4+3] = mesh.XY[2*tri[nNext]+1]
			lineNumber++
		}
	}
	return scr.NewLine(XY, utils.WHITE)
}
