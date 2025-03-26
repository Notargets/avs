/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2025
 */

package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/notargets/avs/geometry"
)

// MeshMetadata is used to document mesh files
type MeshMetadata struct {
	Description     string // What type of method, e.g. Hybrid Lagrange/RT
	NDimensions     int    // Spatial dimensions
	Order           int    // Polynomial order
	NumElements     int    // Total number of elements in the mesh
	NumBaseElements int    // Number of elements, excluding the sub-element tris
	NumPerElement   int    // Elements are triangulated to approximate the poly
	LenXY           int    // Length of the XY coordinates in the mesh
	GitVersion      string
}

func (mmd *MeshMetadata) String() string {
	return fmt.Sprintf(
		"Description: %s\n"+
			"Number of dimensions: %d\n"+
			"Polynomial order: %d\n"+
			"Total number of elements: %d\n"+
			"Number of base elements: %d\n"+
			"Sub elements per base element: %d\n"+
			"Length of XY coordinates: %d\n"+
			"GitVersion: %s\n",
		mmd.Description, mmd.NDimensions, mmd.Order,
		mmd.NumElements, mmd.NumBaseElements,
		mmd.NumPerElement, mmd.LenXY, mmd.GitVersion)
}

// Function to write MeshMetadata and TriMesh sequentially
func WriteMesh(filename string, metadata *MeshMetadata,
	mesh geometry.TriMesh, BCXY map[string][][]float32) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)

	// Encode metadata first
	if err = encoder.Encode(metadata); err != nil {
		return err
	}

	// Encode mesh data
	if err = encoder.Encode(mesh); err != nil {
		return err
	}

	// Encode BCXY data
	if err = encoder.Encode(BCXY); err != nil {
		return err
	}

	return nil
}

// Function to read MeshMetadata and TriMesh sequentially
func ReadMesh(filename string) (md MeshMetadata, gm geometry.TriMesh,
	BCXY map[string][][]float32, err error) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)

	// Decode metadata
	if err = decoder.Decode(&md); err != nil {
		panic(err)
	}

	// Decode mesh data
	if err = decoder.Decode(&gm); err != nil {
		panic(err)
	}

	// Decode BCXY data
	if err = decoder.Decode(&BCXY); err != nil {
		panic(err)
	}
	return
}
