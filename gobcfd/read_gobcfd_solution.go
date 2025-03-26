/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
)

type FieldMetadata struct {
	NumFields        int // How many fields are in the [][]float32
	FieldNames       []string
	SolutionMetadata map[string]interface{} // Fields like ReynoldsNumber,
	// gamma...
	GitVersion string
}

func (sfm *FieldMetadata) String() (str string) {
	str = fmt.Sprintf("NumFields: %d, FieldNames: %s",
		sfm.NumFields, sfm.FieldNames)
	for name, val := range sfm.SolutionMetadata {
		str += fmt.Sprintf("\t%s: %v\n", name, val)
	}
	return
}

type SingleFieldMetadata struct {
	Iterations int
	Time       float32
	Count      int // Number of fields
	Length     int // of each field, for skipping / readahead
}

func (sfm *SingleFieldMetadata) String() string {
	return fmt.Sprintf("Iterations: %d\n"+
		"Time: %f\n"+
		"Count: %d\n"+
		"Length: %d\n", sfm.Iterations, sfm.Time, sfm.Count, sfm.Length)
}

type SolutionReader struct {
	file    *os.File
	decoder *gob.Decoder
	MMD     *MeshMetadata
	FMD     *FieldMetadata       // Global field metadata
	SFMD    *SingleFieldMetadata // Current field metadata
	dataLoc int64                // beginning of field data
}

func (sr *SolutionReader) getCurPos() int64 {
	pos, err := sr.file.Seek(0, io.SeekCurrent)
	if err != nil {
		panic(err)
	}
	return pos
}

func (sr *SolutionReader) seekTo(pos int64) {
	pos, err := sr.file.Seek(pos, io.SeekStart)
	if err != nil {
		panic(err)
	}
}

func (sr *SolutionReader) gotoFieldData() {
	sr.seekTo(0)
	sr.decoder = gob.NewDecoder(sr.file)
	sr.readSolutionMetadata()
}

func (sr *SolutionReader) getFields() (fields map[string][]float32) {
	var firstError = true
	for {
		err := sr.decoder.Decode(&sr.SFMD)
		if err != nil {
			if errors.Is(err, io.EOF) {
				// It's an EOF error
				if firstError {
					firstError = false
					sr.gotoFieldData()
					continue
				} else {
					fmt.Printf("Unexpected EOF reading field data\n")
					panic(err)
				}
			} else {
				panic(err)
			}
		}
		break
	}
	err := sr.decoder.Decode(&fields)
	if err != nil {
		panic(err)
	}
	return
}

func NewSolutionReader(filename string) (sr *SolutionReader) {
	var err error
	sr = &SolutionReader{}
	sr.file, err = os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}
	sr.gotoFieldData()
	// We are now positioned at the beginning of the field data
	return
}

func (sr *SolutionReader) readSolutionMetadata() {
	var (
		err error
	)
	// Decode mesh metadata first
	if err = sr.decoder.Decode(&sr.MMD); err != nil {
		panic(err)
	}
	// Decode field metadata
	if err = sr.decoder.Decode(&sr.FMD); err != nil {
		panic(err)
	}
}
