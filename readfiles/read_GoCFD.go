/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package readfiles

import (
	"encoding/binary"
	"fmt"
	"io"
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
	fmt.Printf("Number of Coordinates: %d\n", lenXYCoords)
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
		for i, _ := range BCEdges[n].EdgeXYs {
			binary.Read(file, binary.LittleEndian, &BCEdges[n].EdgeXYs[i])
		}
	}
	return
}

func ReadGoCFDSolution(path string, verbose bool) (fI []float32) {
	var (
		file  *os.File
		err   error
		lenFi int64
	)

	if verbose {
		fmt.Printf("Reading GoCFD solution file named: %s\n", path)
	}

	if file, err = os.Open(path); err != nil {
		panic(fmt.Errorf("unable to open file %s\n %s", path, err))
	}
	defer file.Close()

	// For now just read the first field
	binary.Read(file, binary.LittleEndian, &lenFi)
	fmt.Printf("Number of Field Elements: %d\n", lenFi)
	fI = make([]float32, lenFi)
	binary.Read(file, binary.LittleEndian, &fI)
	return
}

type GoCFDSolutionReader struct {
	file         *os.File
	currentField []float32
	CurStep      int
	StepsTotal   int
	lenField     int
}

func NewGoCFDSolutionReader(path string, verbose bool) (gcfdReader *GoCFDSolutionReader) {
	var (
		err    error
		isDone bool
	)
	gcfdReader = &GoCFDSolutionReader{
		lenField: -1,
	}
	if verbose {
		fmt.Printf("Reading GoCFD solution file named: %s\n", path)
	}
	if gcfdReader.file, err = os.Open(path); err != nil {
		panic(fmt.Errorf("unable to open file %s\n %s", path, err))
	}
	// Iterate through file to find the total number of time steps within
	var lenFieldFile int64
	for {
		err = binary.Read(gcfdReader.file, binary.LittleEndian, &lenFieldFile)
		if gcfdReader.lenField == -1 {
			gcfdReader.lenField = int(lenFieldFile)
			fmt.Printf("Length of field: %d\n", gcfdReader.lenField)
			gcfdReader.currentField = make([]float32, gcfdReader.lenField)
		}
		switch {
		case err == io.EOF:
			isDone = true
			break
		case err != nil:
			panic(err)
		case int(lenFieldFile) != gcfdReader.lenField:
			panic(fmt.Errorf("read garbage field length", lenFieldFile))
		}
		if isDone {
			break
		}
		gcfdReader.file.Seek(int64(gcfdReader.lenField)*4, io.SeekCurrent)
		gcfdReader.StepsTotal++
	}
	_, err = gcfdReader.file.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Number of Entries Per Field: %d\n", gcfdReader.lenField)
	fmt.Printf("Number of Fields: %d\n", gcfdReader.StepsTotal)
	return
}

func (gcfdReader *GoCFDSolutionReader) GetField() (fI []float32, end bool) {
	var (
		nEntriesFile int64
	)
	binary.Read(gcfdReader.file, binary.LittleEndian, &nEntriesFile)
	if int(nEntriesFile) != gcfdReader.lenField {
		panic(fmt.Errorf("read garbage field length %d", nEntriesFile))
	}
	binary.Read(gcfdReader.file, binary.LittleEndian, &gcfdReader.currentField)
	gcfdReader.CurStep++
	if gcfdReader.CurStep == gcfdReader.StepsTotal {
		end = true
	}
	fI = gcfdReader.currentField
	return
}
