/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package readfiles

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/notargets/avs/geometry"
)

// From here: https://su2code.github.io/docs_v7/Mesh-File/
type SU2ElementType uint8

const (
	ELType_LINE          SU2ElementType = 3
	ELType_Triangle                     = 5
	ELType_Quadrilateral                = 9
	ELType_Tetrahedral                  = 10
	ELType_Hexahedral                   = 12
	ELType_Prism                        = 13
	ELType_Pyramid                      = 14
)

func ReadSU2(filename string, verbose bool) (tMesh geometry.TriMesh,
	BCEdges []*geometry.EdgeGroup) {
	var (
		file   *os.File
		err    error
		reader *bufio.Reader
	)
	if verbose {
		fmt.Printf("Reading SU2 file named: %s\n", filename)
	}
	if file, err = os.Open(filename); err != nil {
		panic(fmt.Errorf("unable to open file %s\n %s", filename, err))
	}
	defer file.Close()
	reader = bufio.NewReader(file)

	dimensionality := readNumber(reader)
	fmt.Printf("Read file with %d dimensional data...\n", dimensionality)
	tMesh = geometry.TriMesh{
		TriVerts: readElements(reader),
		XY:       readGeometry(reader),
	}
	BCEdges = readBCs(reader)
	return
}

func readBCs(reader *bufio.Reader) (BCEdges []*geometry.EdgeGroup) {
	var (
		nType  int
		v1, v2 int
		err    error
	)
	NBCs := readNumber(reader)
	BCEdges = make([]*geometry.EdgeGroup, NBCs)
	for n := 0; n < NBCs; n++ {
		label := readLabel(reader)
		nEdges := readNumber(reader)
		BCEdges[n] = geometry.NewEdgeGroup(label, nEdges)
		EdgeGroup := BCEdges[n]
		for i := 0; i < nEdges; i++ {
			line := getLine(reader)
			if _, err = fmt.Sscanf(line, "%d %d %d", &nType, &v1, &v2); err != nil {
				panic(err)
			}
			if SU2ElementType(nType) != ELType_LINE {
				panic("BCs should only contain line elements in 2D")
			}
			EdgeGroup.Edges[i] = [2]int64{int64(v1), int64(v2)}
		}
	}
	return
}

func readGeometry(reader *bufio.Reader) (XY []float32) {
	var (
		n    int
		x, y float64
		err  error
	)
	Nv := readNumber(reader)
	XY = make([]float32, 2*Nv)
	for i := 0; i < Nv; i++ {
		line := getLine(reader)
		if n, err = fmt.Sscanf(line, "%f %f", &x, &y); err != nil {
			panic(err)
		}
		if n != 2 {
			panic("unable to read coordinates")
		}
		XY[2*i], XY[2*i+1] = float32(x), float32(y)
	}
	return
}

func readElements(reader *bufio.Reader) (TriVerts [][3]int64) {
	var (
		n          int
		nType      int
		v1, v2, v3 int
		err        error
	)
	// EToV is K x 3
	K := readNumber(reader)
	TriVerts = make([][3]int64, K)
	for k := 0; k < K; k++ {
		line := getLine(reader)
		if n, err = fmt.Sscanf(line, "%d %d %d %d", &nType, &v1, &v2, &v3); err != nil {
			panic(err)
		}
		if n != 4 {
			panic("unable to read vertices")
		}
		if SU2ElementType(nType) != ELType_Triangle {
			panic("unable to deal with non-triangular elements right now")
		}
		TriVerts[k] = [3]int64{int64(v1), int64(v2), int64(v3)}
	}
	return
}

func getToken(reader *bufio.Reader) (token string) {
	var (
		line string
		err  error
	)
	line = getLineNoComments(reader)
	ind := strings.Index(line, "=")
	if ind < 0 {
		err = fmt.Errorf("badly formed input line [%s], should have an =", line)
		panic(err)
	}
	token = line[ind+1:]
	return
}

func readLabel(reader *bufio.Reader) (label string) {
	var (
		err error
	)
	token := getToken(reader)
	if _, err = fmt.Sscanf(token, "%s", &label); err != nil {
		err = fmt.Errorf("unable to read label from token: [%s]", token)
		panic(err)
	}
	label = strings.Trim(label, " ")
	return
}

func readNumber(reader *bufio.Reader) (num int) {
	var (
		err error
	)
	token := getToken(reader)
	if _, err = fmt.Sscanf(token, "%d", &num); err != nil {
		err = fmt.Errorf("unable to read number from token: [%s]", token)
		panic(err)
	}
	return
}

func getLineNoComments(reader *bufio.Reader) (line string) {
	var ()
	for {
		line = strings.Trim(getLine(reader), " ")
		// fmt.Printf("line = [%s]\n", line)
		ind := strings.Index(line, "%")
		if ind < 0 || ind != 0 {
			return
		}
	}
}

func getLine(reader *bufio.Reader) (line string) {
	var (
		err error
	)
	line, err = reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			err = fmt.Errorf("early end of file")
		}
		panic(err)
	}
	line = line[:len(line)-1] // Strip away the newline
	return
}
