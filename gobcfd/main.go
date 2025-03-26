package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/notargets/avs/geometry"

	"github.com/notargets/avs/screen"

	"github.com/notargets/avs/chart2d"

	"github.com/notargets/avs/utils"

	"github.com/eiannone/keyboard"
)

type GraphContext struct {
	activeMesh   utils.Key
	activeField  utils.Key
	activeChart  *chart2d.Chart2D
	activeWindow *screen.Window
	mu           sync.RWMutex
}

func (gc *GraphContext) SetActiveMesh(mesh utils.Key) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.activeMesh = mesh
}

func (gc *GraphContext) SetActiveField(k utils.Key) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.activeField = k
}
func (gc *GraphContext) SetActiveChart(chart *chart2d.Chart2D) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.activeChart = chart
}
func (gc *GraphContext) SetActiveWindow(window *screen.Window) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.activeWindow = window
}
func (gc *GraphContext) GetActiveField() utils.Key {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.activeField
}
func (gc *GraphContext) GetActiveMesh() utils.Key {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.activeMesh
}
func (gc *GraphContext) GetActiveChart() *chart2d.Chart2D {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.activeChart
}
func (gc *GraphContext) GetActiveWindow() *screen.Window {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.activeWindow
}

var (
	GC = GraphContext{}
	SR *SolutionReader
	GM geometry.TriMesh
)

func main() {
	// Command-line flags.
	meshFile := flag.String("mesh", "", "Path to mesh file (.gobcfd) [mandatory]")
	solutionFile := flag.String("solution", "", "Path to solution file (.gobcfd) [optional]")
	flag.Parse()

	// Validate mesh file.
	if *meshFile == "" {
		fmt.Fprintln(os.Stderr, "Error: mesh file is mandatory")
		flag.Usage()
		os.Exit(1)
	}
	if filepath.Ext(*meshFile) != ".gobcfd" {
		log.Fatalf("Mesh file must have a .gobcfd extension: got %s", *meshFile)
	}

	// Validate solution file if provided.
	if *solutionFile != "" {
		if filepath.Ext(*solutionFile) != ".gobcfd" {
			log.Fatalf("Solution file must have a .gobcfd extension: got %s", *solutionFile)
		}
		SR = NewSolutionReader(*solutionFile)
		fmt.Printf("Mesh Metadata from Solution File\n%s", SR.MMD.String())
		fmt.Printf("Solution Metadata\n%s", SR.FMD.String())
	}

	// Open the keyboard in main and defer its closure.
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := keyboard.Close(); err != nil {
			log.Println("error closing keyboard:", err)
		}
	}()

	// Channel to signal quitting.
	quit := make(chan struct{})

	// Launch the keyboard-driven interactive loop in a separate goroutine.
	go keyboardLoop(quit)

	// Read the mesh
	var err error
	var mmd MeshMetadata
	var BCXY map[string][][]float32
	mmd, GM, BCXY, err = ReadMesh(*meshFile)
	if err != nil {
		panic(err)
	}
	_ = BCXY
	fmt.Printf("Mesh Info =================\n%s", mmd.String())
	// Start the dummy rendering loop on the main thread.
	fmt.Println("Starting rendering pipeline on main thread...")
	PlotMesh(GM, quit)
	fmt.Println("Rendering pipeline terminated. Exiting application.")
}

// keyboardLoop listens for key events and triggers dummy callbacks.
func keyboardLoop(quit chan<- struct{}) {
	fmt.Println("Interactive command loop started:")
	fmt.Println(" - Use the up arrow to speed up the frame rate")
	fmt.Println(" - Use the down arrow to slow down the frame rate")
	fmt.Println(" - Press the space bar to toggle animation")
	fmt.Println(" - Press the m key to toggle mesh visibility")
	fmt.Println(" - Press 'q' to quit the app")

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Fatal(err)
		}

		switch {
		case key == keyboard.KeyArrowUp:
			increaseFrameRate()
		case key == keyboard.KeyArrowDown:
			decreaseFrameRate()
		case key == keyboard.KeySpace:
			toggleAnimation()
		case char == 'm' || char == 'M':
			toggleMeshVisible()
		case char == 'q' || char == 'Q':
			fmt.Println("Quit command received. Exiting interactive loop.")
			close(quit)
			return
		}
	}
}

func toggleMeshVisible() {
	fmt.Println("Toggling mesh visibility")
	GC.GetActiveChart().Screen.ToggleVisible(GC.GetActiveWindow(), GC.GetActiveMesh())
}

// Dummy callbacks:
func increaseFrameRate() {
	fmt.Println("Increasing frame rate (dummy callback).")
}

func decreaseFrameRate() {
	fmt.Println("Decreasing frame rate (dummy callback).")
}

func toggleAnimation() {
	if SR != nil {
		fmt.Println("Toggling animation")
		fields := SR.getFields()
		fmt.Printf("Single Field Metadata\n%s", SR.SFMD.String())
		name := SR.FMD.FieldNames[0]
		fmt.Printf("Reading %s\n", name)
		fMin, fMax := getFminFmax(fields[name])
		fmt.Printf("fMin = %f, fMax = %f\n", fMin, fMax)
		fMin, fMax = 1, 2.5
		if GC.GetActiveField().IsNil() {
			GC.SetActiveField(GC.GetActiveChart().AddShadedVertexScalar(
				&geometry.VertexScalar{
					TMesh:       &GM,
					FieldValues: fields[name],
				}, fMin, fMax))
		} else {
			GC.GetActiveChart().UpdateShadedVertexScalar(
				GC.GetActiveWindow(), GC.GetActiveField(),
				&geometry.VertexScalar{
					TMesh:       &GM,
					FieldValues: fields[name],
				})
		}
	} else {
		fmt.Println("No solution data")
	}
}

func getFminFmax(F []float32) (fMin, fMax float32) {
	for i, f := range F {
		if i == 0 {
			fMin = f
			fMax = f
		}
		if f > fMax {
			fMax = f
		}
		if f < fMin {
			fMin = f
		}
	}
	return
}
