package hikvision

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/interp"
)

func TestMove3D_Interpolation(t *testing.T) {
	s := interp.PiecewiseLinear{}
	s.Fit(
		[]float64{-1, 1},
		[]float64{0, 255},
	)
	res := s.Predict(0)
	fmt.Println(res)
}

func TestMove3D_Interpolation2(t *testing.T) {
	// width
	s := interp.PiecewiseLinear{}
	s.Fit(
		[]float64{-1, 1},
		[]float64{0, 640},
	)
	resWidth := s.Predict(0)
	fmt.Println(resWidth)

	// height
	h := 0.5
	inverted := h
	s.Fit(
		[]float64{-1, 1},
		[]float64{0, 480},
	)
	resHeight := s.Predict(float64(inverted))
	fmt.Println(resHeight)

	// to 3d range
	s.Fit(
		[]float64{0, 480},
		[]float64{0, 255},
	)
	height3d := s.Predict(resHeight)

	s.Fit(
		[]float64{0, 640},
		[]float64{0, 255},
	)
	width3d := s.Predict(resWidth)

	fmt.Println(width3d, height3d)
}
