package chart2d

type Screen struct {
	Width, Height int
	Ratio         float32
}

func NewScreen(w, h int) *Screen {
	return &Screen{
		Width:  w,
		Height: h,
		Ratio:  float32(h) / float32(w),
	}
}

func (sc *Screen) GetRatio() (rat float32) {
	return sc.Ratio
}

type RangeMap struct {
	xMin, xMax float32
	pMin, pMax float32
}

func NewRangeMap(xMin, xMax float32, pMin, pMax float32) *RangeMap {
	return &RangeMap{
		xMin, xMax,
		pMin, pMax,
	}
}

func (rg *RangeMap) GetMappedCoordinate(x float32) (p float32) {
	p = (rg.pMax-rg.pMin)*(x-rg.xMin)/(rg.xMax-rg.xMin) + rg.pMin
	return p
}

type LineType uint8
