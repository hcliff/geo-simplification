// The heap contains every point in the polygon
// Popping the heap will give you the point with the smallest angle/area
// a.k.a: the least significant point
package internal

type PointWithTriangleHeap struct {
	indexed []*PointWithTriangle
}

func (h PointWithTriangleHeap) Len() int {
	return len(h.indexed)
}

// Our heap is sorted by area
func (h PointWithTriangleHeap) Less(i, j int) bool {
	return h.indexed[i].Area < h.indexed[j].Area
}

func (h PointWithTriangleHeap) Swap(i, j int) {
	// On removal the heap interface does Swap(0, len(heap)-1)
	// which if the heap is empty will trigger an index out of range error
	if i < 0 || j < 0 {
		return
	}
	h.indexed[i].HeapIndex = j
	h.indexed[j].HeapIndex = i
	h.indexed[i], h.indexed[j] = h.indexed[j], h.indexed[i]
}

func (heap *PointWithTriangleHeap) Push(value interface{}) {
	point := value.(*PointWithTriangle)
	// if there's nothing in the array this will be at the 0th position
	// off by one errors begone!
	point.HeapIndex = heap.Len()
	heap.indexed = append(heap.indexed, point)
}

func (heap *PointWithTriangleHeap) Pop() (tailI interface{}) {
	if heap.Len() == 0 {
		return nil
	}
	var tail *PointWithTriangle
	tail, heap.indexed = heap.indexed[heap.Len()-1], heap.indexed[:heap.Len()-1]

	tail.HeapIndex = -1
	return tail
}
