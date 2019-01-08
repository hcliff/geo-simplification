package geosimplification

import (
	"github.com/golang/geo/s2"
	"gitlab.com/hcliff/geo-simplification/internal"
)

func SimplifyLine(
	polyline s2.Polyline,
	threshold float64,
	minPointsToKeep int,
	avoidIntersections bool,
) (output s2.Polyline, err error) {
	// bail out if we don't have enough points
	if len(polyline) <= minPointsToKeep || len(polyline) <= 2 {
		return polyline[:], nil
	}

	pointList := internal.NewPointWithTriangleList()
	for i := range polyline {
		point := internal.NewPointWithTriangle(polyline[i])
		pointList.PushBack(point)
	}

	if err := internal.Visvalingam(
		pointList,
		threshold,
		minPointsToKeep,
		avoidIntersections,
	); err != nil {
		return nil, err
	}

	// Take the resulting linked list and build the lineString
	output = make(s2.Polyline, 0, pointList.Len())
	pointList.Do(func(point *internal.PointWithTriangle) error {
		output = append(output, point.Point)
		return nil
	})

	return output, nil
}

func SimplifyLoop(
	loop *s2.Loop,
	threshold float64,
	minPointsToKeep int,
	avoidIntersections bool,
) (output *s2.Loop, err error) {
	if err := loop.Validate(); err != nil {
		return nil, err
	}

	// We need the loop to be CW to work
	if loop.TurningAngle() < 0 {
		loop.Invert()
	}

	// Require 4 points to keep the loop valid
	// (double count start & finish)
	if minPointsToKeep < 4 {
		minPointsToKeep = 4
	}

	root := internal.NewPointWithTriangle(loop.Vertex(0))
	pointRing := internal.NewPointWithTriangleRing(root)
	for i := range loop.Vertices()[1:] {
		point := internal.NewPointWithTriangle(loop.Vertex(i + 1))
		pointRing.PushBack(point)
	}

	if err := internal.Visvalingam(pointRing, threshold, minPointsToKeep, avoidIntersections); err != nil {
		return nil, err
	}

	// Take the resulting linked list and build the lineString
	simplified := make([]s2.Point, 0, pointRing.Len())
	pointRing.Do(func(point *internal.PointWithTriangle) error {
		simplified = append(simplified, point.Point)
		return nil
	})

	return s2.LoopFromPoints(simplified), nil
}
