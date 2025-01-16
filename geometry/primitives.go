/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package geometry

type EdgeXY [4]float32

type EdgeGroup struct {
	GroupName string
	EdgeXYs   []EdgeXY
}

func NewEdgeGroup(groupName string, nEdges int) *EdgeGroup {
	return &EdgeGroup{
		GroupName: groupName,
		EdgeXYs:   make([]EdgeXY, nEdges),
	}
}
