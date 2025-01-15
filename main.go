/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main

import (
	"fmt"
	"image/color"
	"math"
	_ "net/http/pprof"
	"time"

	"github.com/notargets/avs/geometry"

	"github.com/notargets/avs/readfiles"

	"github.com/notargets/avs/screen"

	"github.com/notargets/avs/utils"

	"github.com/notargets/avs/assets"

	"github.com/notargets/avs/chart2d"
)

// TODO: Alter the object management to add a top level map[WindowKey]map[ObjectKey]Renderable, where the Renderable is
// TODO: ... an Interface{} with the Methods: add, Delete, setupVertices, Show, Hide. The add() will incorporate the ObjectKey
// TODO: ... into the object struct so that the Show/Hide functions can toggle the Visible in the Renderable
// TODO: ... implementation. This allows the event loop to query whether to draw or not before introspecting the object.
// TODO: ... The Delete() should cleanup any internal references, then delete the ObjectKey from the top level object
// TODO: ... map for the window.
func main() {
	// TestConcurrency()
	// TestTriMeshCompareMeshes()
	TestVertexScalar()
	select {}
}

func TestVertexScalar() {
	tMesh, edges := readfiles.ReadGoCFDMesh("assets/wedge-order2.gcfd", true)
	XMin, XMax, YMin, YMax := getSurfaceRange(tMesh.XY, edges)
	XMin, XMax, YMin, YMax = getSquareBoundingBox(XMin, XMax, YMin, YMax)
	fmt.Printf("XMin, XMax, YMin, YMax: %f, %f, %f, %f\n", XMin, XMax, YMin,
		YMax)
	width, height := 1080, 1080
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, width, height,
		utils.WHITE, // Line Color Default
		utils.DARK)  // BG color Default
	var (
		first            = true
		Done             = false
		fI               []float32
		key              utils.Key
		win              *screen.Window
		vs               *geometry.VertexScalar
		gReader          *readfiles.GoCFDSolutionReader
		fMin, fMax, fAve float32
	)
	for !Done {
		if first {
			gReader = readfiles.NewGoCFDSolutionReader("assets/wedge-solution-order2.gcfd",
				true)
			fI, Done = gReader.GetField()
			vs = &geometry.VertexScalar{
				TMesh:       &tMesh,
				FieldValues: fI,
			}
			key = chart.AddContourVertexScalar(vs, 1.20, 2.0, 100)
			win = chart.GetCurrentWindow()
			first = false
		} else {
			vs.FieldValues, Done = gReader.GetField()
			chart.UpdateContourVertexScalar(win, key, vs)
		}
		fMin, fMax, fAve = getFRange(fI)
		fmt.Printf("Field step: %d, ", gReader.CurStep)
		fmt.Printf("FMin, FMax, FAve: %f, %f, %f\n", fMin, fMax, fAve)
		time.Sleep(200 * time.Millisecond)
	}
}

func getFRange(F []float32) (fMin, fMax, fAve float32) {
	var (
		fSum  float32
		count float32
	)
	fMin = F[0]
	fMax = fMin
	fSum = 0
	for _, f := range F {
		if f < fMin {
			fMin = f
		}
		if f > fMax {
			fMax = f
		}
		fSum += f
		count++
	}
	fAve = fSum / count
	return
}
func TestTriMeshCompareMeshes() {
	tMesh, edges := readfiles.ReadGoCFDMesh("assets/wedge-order0.gcfd", true)
	// tMesh, edges := readfiles.ReadGoCFDMesh("assets/meshfile.gcfd", true)
	// tMesh, edges := readfiles.ReadSU2Mesh("assets/nacaAirfoil-base.su2", true)
	// XMin, XMax, YMin, YMax := getRange(tMesh.XY)
	XMin, XMax, YMin, YMax := getSurfaceRange(tMesh.XY, edges)
	XMin, XMax, YMin, YMax = getSquareBoundingBox(XMin, XMax, YMin, YMax)
	fmt.Printf("XMin, XMax, YMin, YMax: %f, %f, %f, %f\n", XMin, XMax, YMin,
		YMax)
	width, height := 1080, 1080
	chart := chart2d.NewChart2D(XMin, XMax, YMin, YMax, width, height,
		utils.WHITE, // Line Color Default
		utils.DARK)  // BG color Default
	chart.AddTriMesh(tMesh)

	tMesh2, edges2 := readfiles.ReadSU2Mesh("assets/wedge.su2", true)
	chart.NewWindow("Original Mesh", chart.Scale, screen.AUTO)
	chart.AddTriMesh(tMesh2)
	_ = chart
	_ = edges
	_ = edges2
}

func getSurfaceRange(XY []float32, edges []*geometry.EdgeGroup) (xmin, xmax, ymin,
	ymax float32) {
	xmin = math.MaxFloat32
	xmax = -xmin
	ymin = xmin
	ymax = xmax
	for _, edgeGroup := range edges {
		// fmt.Printf("BC Group Name: [%s]\n", edgeGroup.GroupName)
		if edgeGroup.GroupName == "wall" {
			for _, edge := range edgeGroup.Edges {
				x1, y1 := XY[2*edge[0]], XY[2*edge[0]+1]
				x2, y2 := XY[2*edge[1]], XY[2*edge[1]+1]
				xmin = float32(math.Min(float64(xmin), float64(x1)))
				ymin = float32(math.Min(float64(ymin), float64(y1)))
				xmax = float32(math.Max(float64(xmax), float64(x1)))
				ymax = float32(math.Max(float64(ymax), float64(y1)))
				xmin = float32(math.Min(float64(xmin), float64(x2)))
				ymin = float32(math.Min(float64(ymin), float64(y2)))
				xmax = float32(math.Max(float64(xmax), float64(x2)))
				ymax = float32(math.Max(float64(ymax), float64(y2)))
			}
		}
	}
	return xmin, xmax, ymin, ymax
}

func getSquareBoundingBox(xMin, xMax, yMin, yMax float32) (xBMin,
	xBMax, yBMin, yBMax float32) {
	xRange := xMax - xMin
	yRange := yMax - yMin
	if yRange > xRange {
		yBMin = yMin
		yBMax = yMax
		xCent := xRange/2. + xMin
		xBMin = xCent - yRange/2.
		xBMax = xCent + yRange/2.
	} else {
		xBMin = xMin
		xBMax = xMax
		yCent := yRange/2. + yMin
		yBMin = yCent - xRange/2.
		yBMax = yCent + xRange/2.
	}
	return
}

func getRange(XY []float32) (xmin, xmax, ymin, ymax float32) {
	xmin = XY[0]
	xmax = XY[0]
	ymin = XY[1]
	ymax = XY[1]
	for i := 0; i < len(XY)/2; i++ {
		X := XY[2*i]
		Y := XY[2*i+1]
		if X < xmin {
			xmin = X
		}
		if X > xmax {
			xmax = X
		}
		if Y < ymin {
			ymin = Y
		}
		if Y > ymax {
			ymax = Y
		}
	}
	return xmin, xmax, ymin, ymax
}

func TestConcurrency() {
	chart := TestText()
	Test2(chart)
	doneChan := make(chan struct{})
	TestFunctionPlot(chart, doneChan, utils.RED, utils.BLACK)
	<-doneChan
	TestFunctionPlot(chart, doneChan, utils.GREEN, utils.BLUE)
	<-doneChan

	// chart := chart2d.NewChart2D(0, 1, -1, 1, 1920, 1080,
	// 	utils.WHITE, // Line Color Default
	// 	utils.DARK)  // BG color Default
	// TestFunctionPlot(chart)

	select {}
}

func TestFunctionPlot(chart *chart2d.Chart2D, doneChan chan struct{}, color1,
	color2 color.RGBA) {
	go func() {
		win := chart.Screen.NewWindow(chart.WindowWidth, chart.WindowHeight,
			0, 1, -1, 1, 0.5, "Sin Function",
			utils.DARK, screen.AUTO)

		tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
			utils.WHITE, true, false)
		chart.AddAxis(utils.WHITE, tickText, "X", "Y", 0, 0, 11)
		doneChan <- struct{}{}

		// Make a Sin function for plotting
		XY := make([]float32, 200)
		XY2 := make([]float32, 200)
		var (
			linekey, linekey2 utils.Key
			TwoPi             = float32(2. * math.Pi)
			x, xInc, t, tInc  float32
			iter              int
		)
		t = 0
		tInc = 0.05
		xInc = 1. / 100.
		for {
			x = 0
			for i := 0; i < 100; i++ {
				XY[2*i] = x
				XY[2*i+1] = float32(math.Sin(float64(x*TwoPi-t)) * math.Cos(
					float64(
						0.5*x*TwoPi-0.2*t)))
				XY2[2*i] = x
				XY2[2*i+1] = float32(math.Sin(float64(x*TwoPi-t)) +
					0.5*math.Cos(float64(2*x*TwoPi-0.5*t)))
				x += xInc
			}
			if iter == 0 {
				linekey = chart.AddLine(XY, color1, utils.POLYLINE)
				linekey2 = chart.AddLine(XY2, color2, utils.POLYLINE)
			} else {
				// chart.Screen.Redraw(win)
				chart.UpdateLine(win, linekey, XY, nil)
				chart.UpdateLine(win, linekey2, XY2, nil)
			}
			time.Sleep(time.Millisecond * 10)
			t += tInc
			iter++
			// if iter > 1 {
			// 	break
			// }
			_ = linekey
		}
	}()
}

func TestText() (chart *chart2d.Chart2D) {
	width, height := 1200, 760
	// width, height := 1000, 1000
	var XMin, XMax, YMin, YMax float32
	style := 2
	switch style {
	case 0:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -10.0, 10.0
	case 1:
		XMin, XMax, YMin, YMax = -20.0, 20.0, -10.0, 10.0
	case 2:
		XMin, XMax, YMin, YMax = -100.0, 100.0, -100.0, 100.0
	case 3:
		XMin, XMax, YMin, YMax = -5.0, -1.0, -1.0, 1.0
	case 4:
		XMin, XMax, YMin, YMax = -1.0, 1.0, -1.0, 1.0
	case 5:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -20.0, 20.0
	case 6:
		XMin, XMax, YMin, YMax = -10.0, 10.0, -100.0, 100.0
	default:
		panic("No option here")
	}

	chart = chart2d.NewChart2D(XMin, XMax, YMin, YMax, width, height,
		utils.WHITE, // Line Color Default
		utils.DARK)  // BG color Default

	tickText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		utils.WHITE, true, false)
	chart.AddAxis(utils.WHITE, tickText, "X", "Y", 0, 0, 11)

	DynamicText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		utils.RED, false, false)
	TitleText := assets.NewTextFormatter("NotoSans", "Bold", 36,
		utils.GREEN, true, true)

	titleHeight := chart.GetWorldSpaceCharHeight(TitleText)

	xRange := chart.XMax - chart.XMin
	_ = xRange
	yRange := chart.YMax - chart.YMin
	xpos := float32(0)
	ypos := chart.YMin + 0.5*yRange
	chart.Printf(DynamicText, xpos, ypos, "This is text that moves with the screen objects")
	ypos = chart.YMin + 0.4*yRange
	chart.Printf(DynamicText, xpos, ypos, "Pan and zoom with right mouse and scroll wheel")

	// Title
	ypos = 1.1*chart.YMax - titleHeight
	chart.Printf(TitleText, xpos, ypos, "This is an example of a title text string")
	// add a 33% pad for the vertical line spacing between lines
	ypos = ypos - 1.33*titleHeight
	chart.Printf(TitleText, xpos, ypos, "Title text doesn't move with pan and zoom and remains the same size when window is resized")

	return
}

func Test2(chart *chart2d.Chart2D) {

	win1 := chart.GetCurrentWindow()

	win2 := chart.NewWindow("Second Window", 0.8*chart.Scale,
		screen.AUTO)

	chart.SetDrawWindow(win2)
	// Test text
	DynamicText := assets.NewTextFormatter("NotoSans", "Regular", 24,
		color.RGBA{R: 255, B: 255, A: 255}, false, false)
	xRange := chart.XMax - chart.XMin
	_ = xRange
	yRange := chart.YMax - chart.YMin
	xpos := float32(0)
	ypos := chart.YMin + 0.5*yRange
	chart.Printf(DynamicText, xpos, ypos, "window 2 Dynamic Text")

	// Title
	TitleText := assets.NewTextFormatter("NotoSans", "Bold", 36,
		color.RGBA{G: 255, A: 255}, true, true)

	titleHeight := chart.GetWorldSpaceCharHeight(TitleText)
	ypos = 0.6*chart.YMax - titleHeight
	chart.Printf(TitleText, 0, ypos,
		"Title 2 first line")
	// add a 33% pad for the vertical line spacing between lines
	ypos = ypos - titleHeight
	chart.Printf(TitleText, 0, ypos, "Title 2 second line")

	// Draw in first window
	chart.SetDrawWindow(win1)

	chart.Printf(TitleText, 0, ypos, "Title 3 First window")

	XY := utils.AddSegmentToLine([]float32{},
		chart.XMin+0.25*xRange, chart.YMin+0.75*yRange,
		chart.XMin+0.5*xRange, chart.YMin+0.75*yRange)

	chart.AddLine(XY, utils.RED)

	// Draw in second window
	chart.SetDrawWindow(win2)

	chart.Printf(TitleText, 0, ypos-0.3*yRange, "Title 4 Second window")

	XY = utils.AddSegmentToLine([]float32{},
		chart.XMin+0.25*xRange, chart.YMin+0.75*yRange,
		chart.XMin+0.5*xRange, chart.YMin+0.75*yRange)
	chart.AddLine(XY, utils.GREEN)

	_, _ = win1, win2

}
