package chart2d

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (cc *Chart2D) Init() {
	// Setup OpenGL resources AFTER shader and VBO are ready
	cc.setupGLResources()

	// Compile shader and ensure it's ready for use
	cc.shader = cc.compileShaders()
	gl.UseProgram(cc.shader) // Activate the shader

	//// Set the viewport and update projection
	gl.Viewport(0, 0, int32(cc.ScreenWidth), int32(cc.ScreenHeight))
	cc.updateProjectionMatrix()

	// Force the first frame to render
	cc.PositionChanged = true
	cc.ScaleChanged = true
}

func (cc *Chart2D) Render() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(cc.shader)

	// Calculate the model matrix for pan/zoom
	model := mgl32.Translate3D(cc.Position[0], cc.Position[1], 0)

	// Get the uniform location for model and projection matrices
	modelUniform := gl.GetUniformLocation(cc.shader, gl.Str("model\x00"))
	projectionUniform := gl.GetUniformLocation(cc.shader, gl.Str("projection\x00"))

	// Send the model and projection matrices to the shader
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
	gl.UniformMatrix4fv(projectionUniform, 1, false, &cc.projectionMatrix[0])

	// Bind the Vertex Array Object (VAO)
	gl.BindVertexArray(cc.VAO)

	// Calculate total number of vertices in the VBO
	totalVertices := int32(0)
	for _, series := range cc.activeSeries {
		totalVertices += int32(len(series.Vertices) / 5)
	}

	// Draw all the triangles
	gl.DrawArrays(gl.TRIANGLES, 0, totalVertices)

	// Unbind VAO to prevent unintended modifications
	gl.BindVertexArray(0)
}

func (cc *Chart2D) EventLoop() {
	var window = cc.Scene.Window
	for !window.ShouldClose() {
		// Poll for events (mouse, keyboard, etc.)
		glfw.WaitEventsTimeout(0.016) // Wait for events but limit to ~60 FPS

		// Check for new data from the channel and update if available
		select {
		case newDataMsg := <-cc.DataChan:
			name := newDataMsg.Name
			_ = name
			newSeries := newDataMsg.Data
			cc.UpdateSeries(newSeries) // Add new series to active series
		default:
			// No data, continue to render if state changed
		}

		// Render only if position or scale has changed
		if cc.PositionChanged || cc.ScaleChanged {
			cc.Render()
			cc.PositionChanged = false
			cc.ScaleChanged = false
		}

		// Swap the buffers to show the new frame
		window.SwapBuffers()
	}
}

func (cc *Chart2D) updateProjectionMatrix() {
	// Get the aspect ratio of the window
	aspectRatio := float32(cc.ScreenWidth) / float32(cc.ScreenHeight)

	// Determine X and Y ranges for the orthographic projection
	var xRange, yRange float32
	if aspectRatio > 1.0 {
		// Screen is wider than tall, adjust Y range
		xRange = (cc.XMax - cc.XMin) / cc.ZoomFactor
		yRange = xRange / aspectRatio
	} else {
		// Screen is taller than wide, adjust X range
		yRange = (cc.YMax - cc.YMin) / cc.ZoomFactor
		xRange = yRange * aspectRatio
	}

	// Calculate the new world coordinate bounds, centered on the original world bounds
	xCenter := (cc.XMin + cc.XMax) / 2.0
	yCenter := (cc.YMin + cc.YMax) / 2.0

	xmin := xCenter - xRange/2.0
	xmax := xCenter + xRange/2.0
	ymin := yCenter - yRange/2.0
	ymax := yCenter + yRange/2.0

	// Cache the orthographic projection matrix
	cc.projectionMatrix = mgl32.Ortho2D(xmin, xmax, ymin, ymax)
}

func (cc *Chart2D) SetZoomSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Zoom speed must be positive, defaulting to 1.0")
		cc.ZoomSpeed = 1.0
		return
	}
	cc.ZoomSpeed = speed
}
func (cc *Chart2D) SetPanSpeed(speed float32) {
	if speed <= 0 {
		log.Println("Pan speed must be positive, defaulting to 1.0")
		cc.PanSpeed = 1.0
		return
	}
	cc.PanSpeed = speed
}

func (cc *Chart2D) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonRight && action == glfw.Press {
		cc.isDragging = true
		cc.lastX, cc.lastY = w.GetCursorPos()
	} else if button == glfw.MouseButtonRight && action == glfw.Release {
		cc.isDragging = false
	}
}

func (cc *Chart2D) cursorPositionCallback(w *glfw.Window, xpos, ypos float64) {
	if cc.isDragging {
		width, height := w.GetSize()

		// Normalize pan speed relative to the world coordinates
		dx := float32(xpos-cc.lastX) / float32(width) * (cc.XMax - cc.XMin) * cc.PanSpeed
		dy := float32(ypos-cc.lastY) / float32(height) * (cc.YMax - cc.YMin) * cc.PanSpeed

		cc.Position[0] += dx
		cc.Position[1] -= dy

		// Indicate that the position has changed
		cc.PositionChanged = true

		// Update last cursor position
		cc.lastX = xpos
		cc.lastY = ypos
	}
}

func (cc *Chart2D) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	// Adjust zoom factor based on scroll input
	cc.ZoomFactor *= 1.0 + float32(yoff)*0.1*cc.ZoomSpeed

	// Constrain zoom factor to a safe range
	if cc.ZoomFactor < 0.1 {
		cc.ZoomFactor = 0.1
	}
	if cc.ZoomFactor > 10.0 {
		cc.ZoomFactor = 10.0
	}

	// Indicate that the scale has changed
	cc.ScaleChanged = true

	// Update the projection matrix
	cc.updateProjectionMatrix()
}

func (cc *Chart2D) resizeCallback(w *glfw.Window, width, height int) {
	cc.ScreenWidth = width
	cc.ScreenHeight = height

	gl.Viewport(0, 0, int32(width), int32(height))

	aspectRatio := float32(width) / float32(height)

	if aspectRatio > 1.0 {
		viewHeight := (cc.YMax - cc.YMin)
		viewWidth := viewHeight * aspectRatio
		centerX := (cc.XMax + cc.XMin) / 2.0
		cc.XMin = centerX - viewWidth/2.0
		cc.XMax = centerX + viewWidth/2.0
	} else {
		viewWidth := (cc.XMax - cc.XMin)
		viewHeight := viewWidth / aspectRatio
		centerY := (cc.YMin + cc.YMax) / 2.0
		cc.YMin = centerY - viewHeight/2.0
		cc.YMax = centerY + viewHeight/2.0
	}

	// Update the projection matrix with new dimensions
	cc.updateProjectionMatrix()
}

func (cc *Chart2D) setupGLResources() {
	gl.GenVertexArrays(1, &cc.VAO)
	gl.BindVertexArray(cc.VAO)

	gl.GenBuffers(1, &cc.VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cc.VBO)

	// Allocate an empty buffer of "potentially large" size (1MB here, but could be smaller)
	gl.BufferData(gl.ARRAY_BUFFER, 1024*1024, nil, gl.DYNAMIC_DRAW)

	// Set up vertex attributes for position (2 floats) and color (3 floats)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0)) // Position (x, y)
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4)) // Color (r, g, b)
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (cc *Chart2D) compileShaders() uint32 {
	vertexShaderSource := `
#version 450

layout (location = 0) in vec2 position; // Vertex position
layout (location = 1) in vec3 color;    // Vertex color

uniform mat4 model;        // Model transformation (for pan, zoom, etc.)
uniform mat4 projection;   // Projection transformation (for screen-to-world transform)

out vec3 fragColor;        // Output color for the fragment shader

void main() {
    // Calculate final position using projection * model * position
    gl_Position = projection * model * vec4(position, 0.0, 1.0);
    fragColor = color;
}
` + "\x00"

	fragmentShaderSource := `
#version 450

in vec3 fragColor; // Interpolated color from the vertex shader
out vec4 color;    // Final color of the pixel

void main() {
    color = vec4(fragColor, 1.0); // Use the interpolated color as output
}
` + "\x00"

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexShaderSource)
	defer free()
	gl.ShaderSource(vertexShader, 1, csource, nil)
	gl.CompileShader(vertexShader)
	checkShaderCompileStatus(vertexShader) // Check for errors

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentShaderSource)
	defer free()
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	gl.CompileShader(fragmentShader)
	checkShaderCompileStatus(fragmentShader) // Check for errors

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)
	checkProgramLinkStatus(shaderProgram) // Check for errors

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}

func checkShaderCompileStatus(shader uint32) {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		fmt.Printf("SHADER COMPILE ERROR: %s\n", log)
		panic(log)
	}
}

func checkProgramLinkStatus(program uint32) {
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		fmt.Printf("PROGRAM LINK ERROR: %s\n", log)
		panic(log)
	}
}
