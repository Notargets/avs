package functions

import graphics2D "github.com/notargets/avs/geometry"

type FSurface struct {
	Tris           *graphics2D.TriMesh
	Functions      [][]float32 // Dimensions: First [] is for which function, second [] is indexed data, same order as geometry
	ActiveFunction int         // Selector of the function to be plotted
}

func NewFSurface(tris *graphics2D.TriMesh, functions [][]float32, active int) *FSurface {
	return &FSurface{
		Tris:           tris,
		Functions:      functions,
		ActiveFunction: active,
	}
}
