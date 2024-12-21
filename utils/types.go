/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package utils

import "github.com/google/uuid"

type RenderType uint16

type Key uuid.UUID

func NewKey() Key {
	return Key(uuid.New())
}

var (
	NEW = Key(uuid.Nil)
)

const (
	LINE RenderType = iota
	POLYLINE
	STRING
	FIXEDSTRING
	TRIMESHEDGESUNICOLOR
	TRIMESHEDGES
	TRIMESHCONTOURS
	TRIMESHSMOOTH
	LINE3D
	TRIMESHEDGESUNICOLOR3D
	TRIMESHEDGES3D
	TRIMESHCONTOURS3D
	TRIMESHSMOOTH3D
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
