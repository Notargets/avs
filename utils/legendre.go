package utils

type ScalarFunc func(float64) float64

func Legendre0(x float64) float64 { return 1 }
func Legendre1(x float64) float64 { return x }
func Legendre2(x float64) float64 { return 0.5 * (3*x*x - 1) }
func Legendre3(x float64) float64 { return 0.5 * x * (5*x*x - 3) }
func Legendre4(x float64) float64 { x2 := x * x; return 0.125 * (35*x2*x2 - 30*x2 + 3) }
func Legendre5(x float64) float64 { x2 := x * x; return 0.125 * x * (63*x2*x2 - 70*x2 + 15) }

func GetLegendrePoly(degree int) ScalarFunc {
	switch degree {
	case 0:
		return Legendre0
	case 1:
		return Legendre1
	case 2:
		return Legendre2
	case 3:
		return Legendre3
	case 4:
		return Legendre4
	case 5:
		return Legendre5
	}
	return nil
}
