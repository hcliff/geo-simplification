package internal

import (
	"errors"

	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s2"
)

// helper to construct the arguments in the order rtreego expects them
func minMax(nums ...float64) (min, max float64) {
	if len(nums) == 0 {
		return 0, 0
	}
	min, max = nums[0], nums[0]
	for _, num := range nums[1:] {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}
	return min, max
}

// helper to construct the arguments in the order rtreego expects them
func minAndDistance(nums ...float64) (min, diff float64) {
	min, max := minMax(nums...)
	return min, max - min
}

func BuildRTreeRect(points ...s2.Point) (*rtreego.Rect, error) {
	x := make([]float64, len(points))
	y := make([]float64, len(points))
	z := make([]float64, len(points))
	for i, point := range points {
		x[i] = point.X
		y[i] = point.Y
		z[i] = point.Z
	}

	minX, xDistance := minAndDistance(x...)
	minY, yDistance := minAndDistance(y...)
	minZ, zDistance := minAndDistance(z...)
	if xDistance == 0 && yDistance == 0 && zDistance == 0 {
		return nil, errors.New("invalid edge, identicle vertices")
	}

	// Colinear points are fine, but rTree doesn't support them, add some fudge
	if xDistance == 0 {
		xDistance = 0.0001
	}
	if yDistance == 0 {
		yDistance = 0.0001
	}
	if zDistance == 0 {
		zDistance = 0.0001
	}

	min := rtreego.Point{minX, minY, minZ}
	return rtreego.NewRect(min, []float64{xDistance, yDistance, zDistance})
}

// Wrap s2.Edge to provide bounding box information about it
// assumed to be immutable
type boundableEdge struct {
	s2.Edge
	// Cached for efficiency
	rect *rtreego.Rect
}

// fufill the rtreego.Spatial interface
func (p boundableEdge) Bounds() *rtreego.Rect {
	return p.rect
}

func newBoundableEdge(edge s2.Edge) (boundableEdge, error) {
	rect, err := BuildRTreeRect(edge.V0, edge.V1)
	if err != nil {
		return boundableEdge{}, err
	}

	return boundableEdge{
		Edge: edge,
		rect: rect,
	}, nil
}