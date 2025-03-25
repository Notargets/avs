package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/notargets/avs/screen"

	"github.com/notargets/avs/chart2d"

	"github.com/notargets/avs/utils"

	"github.com/eiannone/keyboard"
)

type GraphContext struct {
	activeMesh   utils.Key
	activeChart  *chart2d.Chart2D
	activeWindow *screen.Window
	mu           sync.RWMutex
}

func (gc *GraphContext) SetActiveMesh(mesh utils.Key) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.activeMesh = mesh
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
	if *solutionFile != "" && filepath.Ext(*solutionFile) != ".gobcfd" {
		log.Fatalf("Solution file must have a .gobcfd extension: got %s", *solutionFile)
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
	mmd, gm, BCXY, err := ReadMesh(*meshFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Mesh Info =================\n%s", mmd.String())
	_, _ = gm, BCXY
	// Start the dummy rendering loop on the main thread.
	fmt.Println("Starting rendering pipeline on main thread...")
	// ch := chart2d.NewChart2D()
	go PlotMesh(gm)
	renderingLoop(quit)
	fmt.Println("Rendering pipeline terminated. Exiting application.")
}

// renderingLoop simulates a rendering loop running on the main thread.
func renderingLoop(quit <-chan struct{}) {
	ticker := time.NewTicker(time.Second / 60) // 60 fps simulation
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			// Here you would invoke your avs package rendering code.
		}
	}
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
	fmt.Println("Toggling animation (dummy callback).")
}
