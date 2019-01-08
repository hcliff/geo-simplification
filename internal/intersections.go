// helper package for determining if edges cross
package internal

import (
	"fmt"

	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s2"
)

// Factored out because of repeated confusion around correct fn to use
// H.C: while `EdgeOrVertexCrossing` may look tantilizing do _not_ use it
// it will identify two lines that share a vertex as crossing
func EdgesCross(a, b s2.Edge) bool {
	// CrossingSign reports whether the edge AB intersects the edge CD.
	// If AB crosses CD at a point that is interior to both edges, Cross is returned.
	// If any two vertices from different edges are the same it returns MaybeCross.
	// Otherwise it returns DoNotCross
	// Since this is linestring sharing vertices is fine
	switch s2.CrossingSign(a.V0, a.V1, b.V0, b.V1) {
	case s2.Cross:
		return true
	// If two edges share a vertex CrossingSign is MaybeCross
	// TODO: apply some logic to determine if one edge is a subset of another edge
	// this is crossing by our (and elasticsearches) definition
	case s2.MaybeCross:
		return false
	default:
		return false
	}
}

// This method provided exclusively for testing
// uses a more straightforward but slower approach, helpful when debugging
//
// using an rtree for efficient(ish) lookup identify if a polyline self intersects
//
// "R-trees are balanced, so maximum tree height is guaranteed to be logarithmic
// in the number of entries; however, good worst-case performance is not guaranteed.
// Instead, a number of rebalancing heuristics are applied that perform well in practice."
//
// https://en.wikipedia.org/wiki/R-tree
func PolylineSelfIntersects(polyline s2.Polyline) ([]s2.Edge, error) {
	if len(polyline) < 4 {
		return nil, nil
	}

	// the r-tree self balances, but constrain the # branches
	// tune these for "perf", these are sensible general numbers
	minBranchFactor := 25
	maxBranchFactor := 50
	rt := rtreego.NewTree(3, minBranchFactor, maxBranchFactor)

	boundableEdges := make([]rtreego.Spatial, len(polyline)-1)
	for i := range polyline[:len(polyline)-1] {
		edge := s2.Edge{polyline[i], polyline[i+1]}
		boundableEdge, err := newBoundableEdge(edge)
		if err != nil {
			return nil, fmt.Errorf("edge `%d`: %s", i, err.Error())
		}
		boundableEdges[i] = boundableEdge
		rt.Insert(boundableEdge)
	}

	for _, spatial := range boundableEdges {
		candidates := rt.SearchIntersect(spatial.Bounds())
		edge := spatial.(boundableEdge)
		for _, candidateI := range candidates {
			candidate := candidateI.(boundableEdge)
			// Do not compare ourselves to ourselves
			// assume that this polyline is unique
			if candidate == edge {
				continue
			}
			if EdgesCross(edge.Edge, candidate.Edge) {
				return []s2.Edge{edge.Edge, candidate.Edge}, nil
			}
		}
	}

	return nil, nil
}
