// Concrete implementation of VertexCollection
// for traversing "loops"
package internal

type PointWithTriangleRing struct {
	root *PointWithTriangle
	len  int
}

func NewPointWithTriangleRing(root *PointWithTriangle) *PointWithTriangleRing {
	list := &PointWithTriangleRing{
		root: root,
		len:  1,
	}
	root.next = root
	root.prev = root
	root.list = list
	return list
}

func (r PointWithTriangleRing) Len() int {
	return r.len
}

func (r PointWithTriangleRing) front() *PointWithTriangle {
	return r.root
}

func (r PointWithTriangleRing) Prev(point *PointWithTriangle) *PointWithTriangle {
	return point.prev
}

func (r PointWithTriangleRing) Next(point *PointWithTriangle) *PointWithTriangle {
	return point.next
}

func (r PointWithTriangleRing) Do(f func(*PointWithTriangle) error) error {
	if r.Len() == 0 {
		return nil
	}
	for el := r.front(); el != nil; el = r.Next(el) {
		if err := f(el); err != nil {
			return err
		}
		if el.next == r.front() {
			break
		}
	}
	return nil
}

func (r *PointWithTriangleRing) insert(e, at *PointWithTriangle) {
	prev := at.prev
	e.prev = prev
	prev.next = e

	e.next = at
	at.prev = e
	e.list = r
	r.len++
}

func (r *PointWithTriangleRing) PushBack(e *PointWithTriangle) {
	r.insert(e, r.root)
}

func (r *PointWithTriangleRing) Remove(e *PointWithTriangle) {
	e.prev.next = e.next
	e.next.prev = e.prev
	// special case, when removing the root, change the root to be the sibling
	if e == r.root {
		r.root = e.next
	}
	e.next = nil
	e.prev = nil
	e.list = nil
	r.len--
}
