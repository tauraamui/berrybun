package utils

import (
	"math"
)

func CombineNumbers(a float64, b float64) int {
	return int(math.Round(math.Pow((a+b), 2)+(3*a)+b) / 2)
}

func SplitNumbers(n int) (int, int) {
	c := int(math.Round(math.Sqrt(8*float64(n)+1)-1) / 2)
	a := n - c*(c+1)/2
	b := c*(c+3)/2 - n
	return a, b
}
