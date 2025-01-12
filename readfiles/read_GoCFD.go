/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package readfiles

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/notargets/avs/geometry"
)

func ReadGoCFDMesh(path string, verbose bool) (tMesh geometry.TriMesh,
	BCEdges []*geometry.EdgeGroup) {
	var (
		file        *os.File
		err         error
		nDimensions int64
		lenTriVerts int64
		lenXYCoords int64
	)

	if verbose {
		fmt.Printf("Reading GoCFD mesh file named: %s\n", path)
	}

	if file, err = os.Open(path); err != nil {
		panic(fmt.Errorf("unable to open file %s\n %s", path, err))
	}
	defer file.Close()

	// Read the dimensions and data sizes
	binary.Read(file, binary.LittleEndian, &nDimensions)
	binary.Read(file, binary.LittleEndian, &lenTriVerts)
	fmt.Printf("Number of Dimensions: %d\n", nDimensions)
	fmt.Printf("Number of Tri Elements: %d\n", lenTriVerts/3)

	// Read triVerts array
	triVerts := make([]int64, lenTriVerts)
	binary.Read(file, binary.LittleEndian, &triVerts)

	// Read XY coordinates
	binary.Read(file, binary.LittleEndian, &lenXYCoords)
	xy := make([]float64, lenXYCoords*2)
	binary.Read(file, binary.LittleEndian, &xy)

	xy32 := make([]float32, lenXYCoords*2)
	for i := range xy {
		xy32[i] = float32(xy[i])
	}

	verts := make([][3]int64, lenTriVerts/3)
	for i := range verts {
		verts[i][0] = triVerts[3*i]
		verts[i][1] = triVerts[3*i+1]
		verts[i][2] = triVerts[3*i+2]
	}
	tMesh = *geometry.NewTriMesh(xy32, verts)

	var nBCs int64
	binary.Read(file, binary.LittleEndian, &nBCs)
	BCEdges = make([]*geometry.EdgeGroup, nBCs)
	fmt.Printf("Number of BCs: %d\n", nBCs)
	// Read boundary conditions
	for n := 0; n < int(nBCs); n++ {
		var fString [16]byte
		binary.Read(file, binary.LittleEndian, &fString)
		bcName := strings.TrimRight(string(fString[:]), "\x00 ")
		var bcLen int64
		binary.Read(file, binary.LittleEndian, &bcLen)
		BCEdges[n] = geometry.NewEdgeGroup(bcName, int(bcLen))
		for i := range BCEdges[n].Edges {
			binary.Read(file, binary.LittleEndian, &BCEdges[n].Edges[i])
		}
	}
	return
}
