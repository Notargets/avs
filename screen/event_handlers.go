package screen

import (
	"fmt"
	"log"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

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
}

func (scr *Screen) updateProjectionMatrix() {
	// Get the aspect ratio of the window
	aspectRatio := float32(scr.ScreenWidth) / float32(scr.ScreenHeight)

	// Calculate the world coordinates relative to the zoom and position
	var xRange, yRange float32
	if aspectRatio > 1.0 {
		// Landscape orientation
		xRange = (scr.XMax - scr.XMin) / scr.Scale
		yRange = xRange / aspectRatio
	} else {
		// Portrait orientation
		yRange = (scr.YMax - scr.YMin) / scr.Scale
		xRange = yRange * aspectRatio
	}

	// Apply position offset (world shift)
	xmin := scr.Position[0] - xRange/2.0
	xmax := scr.Position[0] + xRange/2.0
	ymin := scr.Position[1] - yRange/2.0
	ymax := scr.Position[1] + yRange/2.0

	// Update the orthographic projection matrix
	scr.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)

	// Upload the new projection matrix to all shaders
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
	// Calculate new zoom scale (increase/decrease)
	scaleChange := 1.0 + float32(yoff)*0.1*scr.ZoomSpeed
	newScale := scr.Scale * scaleChange

	// Constrain zoom level (avoid excessive zoom-in/out)
	if newScale < 0.1 {
		newScale = 0.1
	}
	if newScale > 10.0 {
		newScale = 10.0
	}

	// Update the scale
	scr.Scale = newScale

	// Mark screen to update projection matrix
	scr.ScaleChanged = true
}

func (scr *Screen) resizeCallback(w *glfw.Window, width, height int) {
	scr.ScreenWidth = width
	scr.ScreenHeight = height

	// Update OpenGL viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Calculate the aspect ratio
	aspectRatio := float32(width) / float32(height)

	// Adjust world bounds based on aspect ratio
	if aspectRatio > 1.0 {
		// Landscape (widescreen) adjustment
		viewHeight := (scr.YMax - scr.YMin)
		viewWidth := viewHeight * aspectRatio
		centerX := (scr.XMax + scr.XMin) / 2.0
		scr.XMin = centerX - viewWidth/2.0
		scr.XMax = centerX + viewWidth/2.0
	} else {
		// Portrait (tall screen) adjustment
		viewWidth := (scr.XMax - scr.XMin)
		viewHeight := viewWidth / aspectRatio
		centerY := (scr.YMin + scr.YMax) / 2.0
		scr.YMin = centerY - viewHeight/2.0
		scr.YMax = centerY + viewHeight/2.0
	}

	// Update the projection matrix for the new window size
	scr.updateProjectionMatrix()

	// Mark the position and scale as changed so the event loop will force a re-render
	scr.PositionChanged = true
	scr.ScaleChanged = true
}
