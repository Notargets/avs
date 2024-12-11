package screen

import (
	"fmt"
	"log"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func (scr *Screen) EventLoop() {
	for !scr.Window.ShouldClose() {
		// Process input events like key presses, mouse, etc.
		glfw.WaitEvents()

		// Process any commands from the RenderChannel
		select {
		case command := <-scr.RenderChannel:
			command()
		default:
			// No command to process
		}

		// Update the projection matrix if pan/zoom has changed
		if scr.PositionChanged || scr.ScaleChanged {
			fmt.Println("Updating projection matrix...")
			scr.updateProjectionMatrix()
			scr.PositionChanged = false
			scr.ScaleChanged = false
		}

		// Clear the screen before rendering
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render all active objects (type-coerce and render)
		for _, renderable := range scr.Objects {
			if renderable.Active {
				switch obj := renderable.Object.(type) {
				case *Line:
					obj.Render(scr)
				case *TriMesh:
					//obj.Render(scr)
				case *TriMeshEdges:
					//obj.Render(scr)
				case *TriMeshContours:
					//obj.Render(scr)
				case *TriMeshSmooth:
					//obj.Render(scr)
				default:
					fmt.Printf("Unknown object type: %T\n", obj)
				}
			}
		}

		// Swap buffers to present the frame
		scr.Window.SwapBuffers()
	}
}

func (scr *Screen) fullScreenRender() {
	for _, obj := range scr.Objects {
		switch renderObj := obj.Object.(type) {
		case *Line:
			renderObj.Render(scr)
		case *TriMesh:
			//renderObj.Render(scr)
		case *TriMeshEdges:
			//renderObj.Render(scr)
		case *TriMeshContours:
			//renderObj.Render(scr)
		case *TriMeshSmooth:
			//renderObj.Render(scr)
		default:
			fmt.Printf("Unknown object type: %T\n", renderObj)
		}
	}
}

func (scr *Screen) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		scr.ZoomSpeed = 1.0
		return
	}
	scr.ZoomSpeed = speed
}
func (scr *Screen) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		scr.PanSpeed = 1.0
		return
	}
	scr.PanSpeed = speed
}

func (scr *Screen) SetCallbacks() {
	scr.Window.SetMouseButtonCallback(scr.mouseButtonCallback)
	scr.Window.SetCursorPosCallback(scr.cursorPositionCallback)
	scr.Window.SetScrollCallback(scr.scrollCallback)
	scr.Window.SetSizeCallback(scr.resizeCallback)

}

func (scr *Screen) updateProjectionMatrix() {
	fmt.Println("Updating projection matrix...")

	// Get the aspect ratio of the window
	aspectRatio := float32(scr.ScreenWidth) / float32(scr.ScreenHeight)

	// Determine world coordinate range based on zoom and position
	xRange := (scr.XMax - scr.XMin) / scr.ZoomFactor / scr.Scale
	yRange := (scr.YMax - scr.YMin) / scr.ZoomFactor / scr.Scale

	// Calculate the current center of the view
	centerX := (scr.XMin + scr.XMax) / 2.0
	centerY := (scr.YMin + scr.YMax) / 2.0

	// Adjust for the aspect ratio, but keep the world coordinates intact
	if aspectRatio > 1.0 {
		// Wide screen — stretch X
		xRange = yRange * aspectRatio
	} else {
		// Tall screen — stretch Y
		yRange = xRange / aspectRatio
	}

	// Use Position to adjust the camera's "pan" position in world space
	xmin := centerX - xRange/2.0 + scr.Position[0]
	xmax := centerX + xRange/2.0 + scr.Position[0]
	ymin := centerY - yRange/2.0 + scr.Position[1]
	ymax := centerY + yRange/2.0 + scr.Position[1]

	// Update the orthographic projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Send the updated projection matrix to all shaders
	for renderType, shaderProgram := range scr.Shaders {
		projectionUniform := gl.GetUniformLocation(shaderProgram, gl.Str("projection\x00"))
		if projectionUniform < 0 {
			fmt.Printf("Projection uniform not found for RenderType %v\n", renderType)
		} else {
			gl.UseProgram(shaderProgram)
			gl.UniformMatrix4fv(projectionUniform, 1, false, &scr.projectionMatrix[0])
		}
	}
}

func (scr *Screen) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonRight && action == glfw.Press {
		scr.isDragging = true
		scr.lastX, scr.lastY = w.GetCursorPos()
	} else if button == glfw.MouseButtonRight && action == glfw.Release {
		scr.isDragging = false
	}
}

func (scr *Screen) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if scr.isDragging {
		width, height := w.GetSize()

		// Calculate movement in world coordinates (pan logic)
		dx := float32(xpos-scr.lastX) / float32(width) * (scr.XMax - scr.XMin) / scr.Scale
		dy := float32(ypos-scr.lastY) / float32(height) * (scr.YMax - scr.YMin) / scr.Scale

		// Update world position
		scr.Position[0] -= dx // X-axis pan
		scr.Position[1] += dy // Y-axis pan (inverted since screen Y is inverted)

		// Mark the screen as needing a projection update
		scr.PositionChanged = true

		// Update cursor tracking position
		scr.lastX = xpos
		scr.lastY = ypos
	}
}

func (scr *Screen) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Adjust the zoom factor based on scroll input
	scr.ZoomFactor *= 1.0 + float32(yoff)*0.1*scr.ZoomSpeed

	// Constrain the zoom factor to prevent excessive zoom
	if scr.ZoomFactor < 0.1 {
		scr.ZoomFactor = 0.1
	}
	if scr.ZoomFactor > 10.0 {
		scr.ZoomFactor = 10.0
	}

	// Also adjust the scale value (legacy, previously working logic)
	scr.Scale *= 1.0 + float32(yoff)*0.1*scr.ZoomSpeed

	// Constrain the scale to prevent excessive zoom (if needed)
	if scr.Scale < 0.1 {
		scr.Scale = 0.1
	}
	if scr.Scale > 10.0 {
		scr.Scale = 10.0
	}

	// Flag that the scale has changed to trigger re-rendering
	scr.ScaleChanged = true
}

func (scr *Screen) resizeCallback(w *glfw.Window, width, height int) {
	// Update screen dimensions
	scr.ScreenWidth = width
	scr.ScreenHeight = height

	// Update OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Recompute the projection matrix to maintain the aspect ratio and world bounds
	scr.updateProjectionMatrix()

	// Mark that a change occurred so the view is updated
	scr.PositionChanged = true
	scr.ScaleChanged = true
}
