/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package utils

type ScalarFunc func(float64) float64

/*
Legendre polynomials

Transformed Legendre Polynomials:
  - transformed into the interval X: 0->1
  - normalized per: https://en.wikipedia.org/wiki/Discontinuous_Galerkin_method
    which multiplies each Pn by SQRT(2*n+1)
    P1: SQRT(3)
    P2: SQRT(5)
    P3: SQRT(7)
    P4: SQRT(9) = 3
    P5: SQRT(11)
*/
func Legendre0(x float64) float64 { return 1 }
func Legendre1(x float64) float64 { return x }
func Legendre2(x float64) float64 { return 0.5 * (3*x*x - 1) }
func Legendre3(x float64) float64 {
	return 0.5 * x * (5*x*x - 3)
}
func Legendre4(x float64) float64 {
	x2 := x * x
	return 0.125 * (35*x2*x2 - 30*x2 + 3)
}
func Legendre5(x float64) float64 {
	x2 := x * x
	return 0.125 * x * (63*x2*x2 - 70*x2 + 15)
}
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

func TLegendre1(x float64) float64 { x = 2*x - 1; return 1.73205080756887729352 * x }
func TLegendre2(x float64) float64 { x = 2*x - 1; return 2.23606797749978969640 * 0.5 * (3*x*x - 1) }
func TLegendre3(x float64) float64 {
	x = 2*x - 1
	return 2.64575131106459059050 * 0.5 * x * (5*x*x - 3)
}
func TLegendre4(x float64) float64 {
	x = 2*x - 1
	x2 := x * x
	return 3 * 0.125 * (35*x2*x2 - 30*x2 + 3)
}
func TLegendre5(x float64) float64 {
	x = 2*x - 1
	x2 := x * x
	return 3.31662479035539984911 * 0.125 * x * (63*x2*x2 - 70*x2 + 15)
}

func GetLegendrePolyT(degree int) ScalarFunc {
	switch degree {
	case 0:
		return Legendre0
	case 1:
		return TLegendre1
	case 2:
		return TLegendre2
	case 3:
		return TLegendre3
	case 4:
		return TLegendre4
	case 5:
		return TLegendre5
	}
	return nil
}
