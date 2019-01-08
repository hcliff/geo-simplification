package internal

import (
	"container/heap"
	"math"

	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s2"
)

// Interface for traversing our shape that we
// can implement in ring & non-ring ways
// please see pointlist.go and pointring.go for concrete implementations
type VertexCollection interface {
	Remove(*PointWithTriangle)
	Len() int
	Do(f func(*PointWithTriangle) error) error
	Prev(p *PointWithTriangle) *PointWithTriangle
	Next(p *PointWithTriangle) *PointWithTriangle
}

type PointWithTriangle struct {
	Point s2.Point
	// The previous and next points in the shape
	// for a loop this can be cyclical
	prev, next *PointWithTriangle
	// The area this triangle occupies with
	// the triangle (point-1)(point)(point+1)
	Area      float64
	HeapIndex int
	// the bounding box of the triangle formed
	BBox *rtreego.Rect
	list VertexCollection
}

func NewPointWithTriangle(point s2.Point) *PointWithTriangle {
	t := PointWithTriangle{
		Point:     point,
		HeapIndex: -1,
	}
	return &t
}

// Defer to the list for next behaviour
func (p *PointWithTriangle) Next() *PointWithTriangle {
	if p.list == nil {
		return nil
	}
	return p.list.Next(p)
}

func (p *PointWithTriangle) Prev() *PointWithTriangle {
	if p.list == nil {
		return nil
	}
	return p.list.Prev(p)
}

func (p PointWithTriangle) Bounds() *rtreego.Rect {
	return p.BBox
}

// returns the first intersection it finds
// when removing `point`
// given the triangle abc, see if the vector ac intersects with
// the remainder of the linked list
func CreatesIntersection(
	rtree *rtreego.Rtree,
	point *PointWithTriangle,
) []s2.Edge {
	// special case, this is the start or end of a polyline
	// it's always preserved
	if point.Prev() == nil || point.Next() == nil {
		return nil
	}
	candidates := rtree.SearchIntersect(point.Bounds())
	// by remove the point `b` in `abc` we'd create a new edge `ac`
	proposedEdge := s2.Edge{point.Prev().Point, point.Next().Point}

	// for efficiency we only look at other edges that were in the points
	// bounding box. any outside are guaranteed not to intersect
	for _, candidateI := range candidates {
		candidate := candidateI.(*PointWithTriangle)
		// might hit ourselves
		if candidate == point {
			continue
		}

		// Check if the ab edge would intersect with our proposed edge
		if prev := candidate.Prev(); prev != nil {
			ab := s2.Edge{candidate.Point, prev.Point}
			if EdgesCross(proposedEdge, ab) {
				return []s2.Edge{proposedEdge, ab}
			}
		}

		// Check if the ac edge would intersect with our proposed edge
		if next := candidate.Next(); next != nil {
			bc := s2.Edge{candidate.Point, next.Point}
			if EdgesCross(proposedEdge, bc) {
				return []s2.Edge{proposedEdge, bc}
			}
		}
	}

	return nil
}

func TriangleArea(point *PointWithTriangle) float64 {
	if point.Prev() == nil || point.Next() == nil {
		return math.Inf(1)
	}

	// Note: under the covers this will defer to GirardArea if possible
	// no need to swap it out
	return s2.PointArea(point.Prev().Point, point.Point, point.Next().Point)
}

func TriangleBbox(point *PointWithTriangle) (*rtreego.Rect, error) {
	points := []s2.Point{point.Point}

	if prevPoint := point.Prev(); prevPoint != nil {
		points = append(points, prevPoint.Point)
	}
	if nextPoint := point.Next(); nextPoint != nil {
		points = append(points, nextPoint.Point)
	}

	return BuildRTreeRect(points...)
}

func Visvalingam(
	pointList VertexCollection,
	threshold float64,
	minPointsToKeep int,
	avoidIntersections bool,
) (err error) {

	minHeap := &PointWithTriangleHeap{}
	heap.Init(minHeap)

	// the r-tree self balances, but constrain the # branches
	// tune these for "perf", these are sensible general numbers
	// TODO: generate these based on the number of points
	minBranchFactor := 25
	maxBranchFactor := 50
	// Build a rtree
	rtree := rtreego.NewTree(3, minBranchFactor, maxBranchFactor)

	if err = pointList.Do(func(point *PointWithTriangle) error {
		// set the area and bounding box
		point.Area = TriangleArea(point)
		point.BBox, err = TriangleBbox(point)
		if err != nil {
			return err
		}
		// push it onto the heap
		heap.Push(minHeap, point)
		// add it to the rtree elements
		rtree.Insert(point)
		return nil
	}); err != nil {
		return err
	}

	maxArea := 0.0
	intersecting := []*PointWithTriangle{}
	// Pop the heap, because the heap maintains order by area
	// this means the point that forms the smallest area
	// will be removed front (tl;dr: most useless point removed first)
	for elementI := heap.Pop(minHeap); elementI != nil; elementI = heap.Pop(minHeap) {
		head := elementI.(*PointWithTriangle)

		// If the area of the current point is less than that of the previous point
		// to be eliminated, use the latters area instead. This ensures that the
		// current point cannot be eliminated without eliminating previously-
		// eliminated points.
		// H.C: this happens when you remove a point and the resulting triangle
		// has less area?
		if head.Area < maxArea {
			head.Area = maxArea
		} else {
			maxArea = head.Area
		}

		// if removing the node b in the triangle abc
		// would cause an intersection do not actually remove it
		if avoidIntersections && CreatesIntersection(rtree, head) != nil {
			intersecting = append(intersecting, head)
			continue
		}

		for _, intersectingElement := range intersecting {
			heap.Push(minHeap, intersectingElement)
		}
		intersecting = []*PointWithTriangle{}

		// If this area is greater than the threshold time to stop
		// removing points
		if head.Area >= threshold {
			break
		}

		prev := head.Prev()
		next := head.Next()

		// remove all trianges touched from the rtree
		rtree.Delete(head)
		if prev != nil {
			rtree.Delete(prev)
		}
		if next != nil {
			rtree.Delete(next)
		}

		// Remove our entry from the linked list
		pointList.Remove(head)

		// If we've reached the minimum number of points stop
		if pointList.Len() <= minPointsToKeep {
			break
		}

		// Since we dropped a point recompute the previous points area
		// and may have added back intersecting points
		// the heap will need to be rebuilt too
		if prev != nil {
			// Keep the heap up to date
			prev.Area = TriangleArea(prev)
			if prev.HeapIndex > -1 {
				heap.Fix(minHeap, prev.HeapIndex)
			}

			// keep the rtree up to date
			prev.BBox, err = TriangleBbox(prev)
			if err != nil {
				return err
			}
			rtree.Insert(prev)
		}

		// Since we dropped a point recompute the next points area
		// the heap will need to be rebuilt too
		if next != nil {
			// Keep the heap up to date
			next.Area = TriangleArea(next)
			if next.HeapIndex > -1 {
				heap.Fix(minHeap, next.HeapIndex)
			}

			// keep the rtree up to date
			next.BBox, err = TriangleBbox(next)
			if err != nil {
				return err
			}
			rtree.Insert(next)
		}
	}

	return nil
}
