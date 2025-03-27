/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"sort"

	"github.com/notargets/avs/utils"
)

type ObjectGroup []interface{}

func newObjectGroup(object interface{}) ObjectGroup {
	return ObjectGroup{object}
}

// Len returns the number of objects.
func (og ObjectGroup) Len() int { return len(og) }

// Swap swaps two objects.
func (og ObjectGroup) Swap(i, j int) { og[i], og[j] = og[j], og[i] }

// Less defines the sort order based on type.
func (og ObjectGroup) Less(i, j int) bool {
	return typeOrder(og[i]) < typeOrder(og[j])
}

// typeOrder returns an integer order for each type.
// Lower values are drawn first.
func typeOrder(obj interface{}) int {
	switch obj.(type) {
	case *ShadedVertexScalar:
		return 3
	case *ContourVertexScalar:
		return 0
	case *Line:
		return 1
	case *String:
		return 2
	default:
		// Unknown or default types are rendered last.
		return 4
	}
}

type Renderable struct {
	Visible bool
	Objects ObjectGroup // Any object that has a render method (e.g., Line,
	// TriMesh)
	Type utils.RenderType
}

type RenderableMap map[utils.Key]*Renderable

func NewRenderableMap() RenderableMap {
	return make(RenderableMap)
}

func (rm RenderableMap) GetKeys() []utils.Key {
	// Build a slice of keys from the map.
	keys := make([]utils.Key, 0, len(rm))
	for key := range rm {
		keys = append(keys, key)
	}

	// Sort the keys by the corresponding Renderable's RenderType.
	sort.Slice(keys, func(i, j int) bool {
		return rm[keys[i]].Type < rm[keys[j]].Type
	})
	return keys
}
