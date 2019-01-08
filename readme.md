# Geo Simplification

This work is based off [Jason Davies Excellent blog post](https://www.jasondavies.com/simplify/), it implements the Visvalingam line simplification algorithm in Golang. It's designed to work with the [official golang geo library](https://github.com/golang/geo).

It's been extended to support loops

# What doesn't work?

If your polygon has more than one loop this library will not work. This is due to issues with simplifying holes in polygons.

# Usage

## Reduce to a specific number of points
	import (
		geosimplification "github.com/hcliff/geo-simplification"
	)

	loop := []s2.Loop{
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.319325896602095, -73.14971923828125)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.32141999593439, -73.31451416015625)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.2313150803688, -73.4271240234375)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.24179856487036, -73.16619873046875)),
	}
	# keep removing points until we reach 4 points in the loop
	threshold := 0
	minPointsToKeep := 4
	avoidIntersections := true
	simplified, err := geosimplification.SimplifyLoop(loop, threshold, minPointsToKeep, avoidIntersections)

## Reduce to a specific level of detail
	import (
		geosimplification "github.com/hcliff/geo-simplification"
	)

	loop := []s2.Loop{
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.264856517201856, -73.32550048828125)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.319325896602095, -73.14971923828125)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.32141999593439, -73.31451416015625)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.2313150803688, -73.4271240234375)),
		s2.PointFromLatLng(s2.LatLngFromDegrees(40.24179856487036, -73.16619873046875)),
	}
	# remove points that don't provide much extra detail to the shape
	threshold := 0.001
	minPointsToKeep := 5
	avoidIntersections := true
	simplified, err := geosimplification.SimplifyLoop(loop, threshold, minPointsToKeep, avoidIntersections)