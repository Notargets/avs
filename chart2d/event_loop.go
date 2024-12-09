package chart2d

//
//import (
//	"log"
//
//	"github.com/go-gl/gl/v4.5-core/gl"
//	"github.com/go-gl/glfw/v3.3/glfw"
//)
//
//func Plot() {
//	if err := glfw.Init(); err != nil {
//		log.Fatalln("failed to initialize glfw:", err)
//	}
//	defer glfw.Terminate()
//
//	glfw.WindowHint(glfw.ContextVersionMajor, 4)
//	glfw.WindowHint(glfw.ContextVersionMinor, 5)
//	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
//	window, err := glfw.CreateWindow(1920, 1080, "Chart2D with OpenGL 4.5", nil, nil)
//	if err != nil {
//		panic(err)
//	}
//	window.MakeContextCurrent()
//
//	err = gl.Init()
//	version := gl.GoStr(gl.GetString(gl.VERSION))
//	renderer := gl.GoStr(gl.GetString(gl.RENDERER))
//	vendor := gl.GoStr(gl.GetString(gl.VENDOR))
//	log.Printf("OpenGL version: %s, Renderer: %s, Vendor: %s", version, renderer, vendor)
//	if err != nil {
//		panic(err)
//	}
//
//	chart := NewChart2D()
//	chart.Init()
//
//	var lastX, lastY float64
//	var isDragging bool
//
//	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
//		if button == glfw.MouseButtonRight && action == glfw.Press {
//			isDragging = true
//			lastX, lastY = w.GetCursorPos()
//		} else if button == glfw.MouseButtonRight && action == glfw.Release {
//			isDragging = false
//		}
//	})
//
//	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
//		if isDragging {
//			dx := float32(xpos - lastX)
//			dy := float32(ypos - lastY)
//			chart.Position[0] += dx / 100
//			chart.Position[1] -= dy / 100
//			lastX = xpos
//			lastY = ypos
//		}
//	})
//
//	window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
//		chart.Scale += float32(yoff) * 0.1
//		if chart.Scale < 0.1 {
//			chart.Scale = 0.1
//		}
//	})
//
//	// This is the actual event loop
//	for !window.ShouldClose() {
//		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
//		chart.Render()
//		window.SwapBuffers()
//		glfw.PollEvents()
//	}
//}
