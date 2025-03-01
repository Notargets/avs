/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package geometry

type TriMesh struct {
	XY       []float32  // X1,Y1,X2,Y2...XImax,YImax, "packed" node coordinates
	TriVerts [][3]int64 // Every corner index specified for each of Kx3 tris
}

func NewTriMesh(XY []float32, Verts [][3]int64) TriMesh {
	return TriMesh{XY, Verts}
}

type VertexScalar struct {
	TMesh       *TriMesh  // Geometry, triangle vertex locations
	FieldValues []float32 // {F1,F2,F3,F4,F5...} Same order as coordinates
}
