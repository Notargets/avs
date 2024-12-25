/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package gl_thread_objects

import "github.com/notargets/avs/utils"

type ObjectGroup []interface{}

func NewObjectGroup(object interface{}) ObjectGroup {
	return ObjectGroup{object}
}

type Renderable struct {
	Visible bool
	Objects ObjectGroup // Any object that has a Render method (e.g., Line,
	// TriMesh)
}

func (rb *Renderable) Add(key utils.Key) {
	// An object group is append only by design
	rb.Objects = append(rb.Objects, key)
}

func (rb *Renderable) SetVisible(isVisible bool) {
	rb.Visible = isVisible
}
