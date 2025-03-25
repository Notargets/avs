package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eiannone/keyboard"
)

type FlowFunction int

const (
	// Dummy constants for illustration.
	Density FlowFunction = iota
	XMomentum
	// ... other constants ...
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
		case char == 'q' || char == 'Q':
			fmt.Println("Quit command received. Exiting interactive loop.")
			close(quit)
			return
		}
	}
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
