package graphics2D

import (
	"math"
	"testing"

	. "gopkg.in/check.v1"
)

var (
	drawpoly *Polygon
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	testPoly *Polygon
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpSuite(c *C) {
	/*
		Set up a polygon
	*/
	radius := 100.
	nSegs := 8
	radInc := 2 * math.Pi / float64(nSegs)
	geom := []Point{}
	var rads float64
	for i := 0; i < nSegs+1; i++ {
		x := float32(radius * math.Sin(rads))
		y := float32(radius * math.Cos(rads))
		geom = append(geom, Point{X: [2]float32{x, y}})
		rads += radInc
	}

	poly := NewPolygon(geom)
	s.testPoly = poly
	drawpoly = poly
}

func (s *TestSuite) TearDownSuite(c *C) {
}

func (s *TestSuite) TestLineIntersection(c *C) {
	line1 := NewLine(
		Point{X: [2]float32{-100, 0}},
		Point{X: [2]float32{100, 0}},
	)
	line2 := NewLine(
		Point{X: [2]float32{0, -200}},
		Point{X: [2]float32{0, 200}},
	)
	iPoint := LineLineIntersection(line1, line2)
	c.Assert(iPoint.X[0], Equals, float32(0))
	c.Assert(iPoint.X[1], Equals, float32(0))

	line2 = NewLine(
		Point{X: [2]float32{50, -200}},
		Point{X: [2]float32{50, 200}},
	)
	iPoint = LineLineIntersection(line1, line2)
	c.Assert(iPoint.X[0], Equals, float32(50))
	c.Assert(iPoint.X[1], Equals, float32(0))

	line1 = NewLine(
		Point{X: [2]float32{-100, 50}},
		Point{X: [2]float32{100, 50}},
	)
	iPoint = LineLineIntersection(line1, line2)
	c.Assert(iPoint.X[0], Equals, float32(50))
	c.Assert(iPoint.X[1], Equals, float32(50))

	c.Assert(line1.Box.PointInside(iPoint), Equals, true)
	iPoint.X[1] = 1000
	c.Assert(line1.Box.PointInside(iPoint), Equals, false)

	point := Point{X: [2]float32{0, 0}}
	outside := Point{X: [2]float32{3 * drawpoly.Box.XMax[0], 2 * drawpoly.Box.XMax[1]}}
	refLine := NewLine(point, outside)

	line := NewLine(drawpoly.Geometry[0], drawpoly.Geometry[1])
	iPoint = LineLineIntersection(refLine, line)
	c.Assert(line.Box.PointInside(iPoint), Equals, false)

	line = NewLine(drawpoly.Geometry[1], drawpoly.Geometry[2])
	iPoint = LineLineIntersection(refLine, line)
	c.Assert(line.Box.PointInside(iPoint), Equals, true)
}
func (s *TestSuite) TestPolyInclusion(c *C) {
	centroid := Point{X: [2]float32{0, 0}}
	c.Assert(s.testPoly.PointInside(centroid), Equals, true)

	point := Point{X: [2]float32{
		2 * s.testPoly.Box.XMax[0],
		2 * s.testPoly.Box.XMax[1],
	}}
	c.Assert(s.testPoly.PointInside(point), Equals, false)

	/*
		Barely inside
	*/
	point = Point{X: [2]float32{
		0.70 * s.testPoly.Box.XMax[0],
		0.70 * s.testPoly.Box.XMax[1],
	}}
	c.Assert(s.testPoly.PointInside(point), Equals, true)

	point = Point{X: [2]float32{
		0.72 * s.testPoly.Box.XMax[0],
		0.72 * s.testPoly.Box.XMax[1],
	}}
	c.Assert(s.testPoly.PointInside(point), Equals, false)
}
