/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package utils

import (
	"sync/atomic"
	"unsafe"

	"github.com/google/uuid"
)

type SafeInt struct {
	value unsafe.Pointer
}

// Write sets a new value atomically
func (si *SafeInt) Write(val int) {
	atomic.StorePointer(&si.value, unsafe.Pointer(&val))
}

// Read atomically retrieves the variable's value.
// It returns an `int` type
func (si *SafeInt) Read() int {
	ptr := atomic.LoadPointer(&si.value)
	return *(*int)(ptr)
}

type Key uuid.UUID

func NewKey() Key {
	return Key(uuid.New())
}

func (key Key) IsNil() bool {
	if key == Key(uuid.Nil) {
		return true
	}
	return false
}

type RenderType uint16

// These are ordered in the intended drawing/rendering order
const (
	LINE RenderType = iota
	POLYLINE
	LINE3D
	TRIMESHCONTOURS
	TRIMESHEDGESUNICOLOR
	TRIMESHEDGESUNICOLOR3D
	TRIMESHEDGES3D
	TRIMESHCONTOURS3D
	TRIMESHEDGES
	TRIMESHSMOOTH
	TRIMESHSMOOTH3D
	STRING
	FIXEDSTRING
)

func (r RenderType) String() string {
	switch r {
	case LINE:
		return "LINE"
	case POLYLINE:
		return "POLYLINE"
	case STRING:
		return "STRING"
	case FIXEDSTRING:
		return "FIXEDSTRING"
	case TRIMESHEDGESUNICOLOR:
		return "TRIMESHEDGESUNICOLOR"
	case TRIMESHEDGESUNICOLOR3D:
		return "TRIMESHEDGESUNICOLOR3D"
	case TRIMESHCONTOURS:
		return "TRIMESHCONTOURS"
	case TRIMESHSMOOTH:
		return "TRIMESHSMOOTH"
	case LINE3D:
		return "LINE3D"
	case TRIMESHEDGES3D:
		return "TRIMESHEDGES3D"
	case TRIMESHCONTOURS3D:
		return "TRIMESHCONTOURS3D"
	case TRIMESHSMOOTH3D:
		return "TRIMESHSMOOTH3D"
	default:
		return "Unknown"
	}
}
