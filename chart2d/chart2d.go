package chart2d

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Chart2D struct {
	Scene            *Screen      // Manages OpenGL screen contents
	DataChan         chan DataMsg // Channel for new data
	VAO              uint32       // Vertex Array Object
	VBO              uint32       // Vertex Buffer Object
	shader           uint32       // Shader program
	activeSeries     []Series     // List of currently active series
	Scale            float32
	Position         [2]float32
	isDragging       bool    // Tracks whether the right mouse button is being held
	lastX            float64 // Last cursor X position
	lastY            float64 // Last cursor Y position
	projectionMatrix mgl32.Mat4
	ScreenWidth      int // Current width of the screen
	ScreenHeight     int // Current height of the screen
	// Fields for World Coordinate Range**
	XMin, XMax      float32 // World X-range
	YMin, YMax      float32 // World Y-range
	PanSpeed        float32 // Speed of panning
	ZoomSpeed       float32 // Speed of zooming
	ZoomFactor      float32 // Factor controlling zoom (instead of scale)
	PositionChanged bool    // Tracks if position has changed
	ScaleChanged    bool    // Tracks if scale (zoom) has changed
}

func NewChart2D(width, height int, xmin, xmax, ymin, ymax float64) (cc *Chart2D) {
	cc = &Chart2D{
		DataChan:     make(chan DataMsg, 100), // Buffer size can be adjusted
		isDragging:   false,
		lastX:        0,
		lastY:        0,
		Scale:        1.0,
		Position:     [2]float32{0.0, 0.0},
		ScreenWidth:  width,  // Set initial width
		ScreenHeight: height, // Set initial height
		XMin:         float32(xmin),
		XMax:         float32(xmax),
		YMin:         float32(ymin),
		YMax:         float32(ymax),
		PanSpeed:     1.0,
		ZoomSpeed:    1.0,
		ZoomFactor:   1.0,
	}
	cc.Scene = cc.NewScreen(width, height)
	return
}

func (cc *Chart2D) UpdateSeries(newSeries Series) {
	cc.activeSeries = append(cc.activeSeries, newSeries)
	cc.updateVBO()
}

func (cc *Chart2D) updateVBO() {
	vertices := []float32{}
	for _, s := range cc.activeSeries {
		vertices = append(vertices, s.Vertices...)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, cc.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}
