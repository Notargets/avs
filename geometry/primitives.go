/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package geometry

type Edge [2]int64

type EdgeGroup struct {
	GroupName string
	Edges     []Edge
}

func NewEdgeGroup(groupName string, nEdges int) *EdgeGroup {
	return &EdgeGroup{
		GroupName: groupName,
		Edges:     make([]Edge, nEdges),
	}
}
