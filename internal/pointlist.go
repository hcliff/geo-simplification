// Concrete implementation of VertexCollection
// for traversing "lines"
package internal

type PointWithTriangleList struct {
	root PointWithTriangle // sentinal node
	len  int
}

func NewPointWithTriangleList() *PointWithTriangleList {
	list := PointWithTriangleList{}
	list.root.next = &list.root
	list.root.prev = &list.root
	return &list
}

func (l PointWithTriangleList) Len() int {
	return l.len
}

func (l PointWithTriangleList) front() *PointWithTriangle {
	return l.root.next
}

func (l *PointWithTriangleList) Prev(point *PointWithTriangle) *PointWithTriangle {
	if point.prev == nil || *point.prev == l.root {
		return nil
	}
	return point.prev
}

func (l *PointWithTriangleList) Next(point *PointWithTriangle) *PointWithTriangle {
	if point.next == nil || *point.next == l.root {
		return nil
	}
	return point.next
}

func (l *PointWithTriangleList) Do(f func(*PointWithTriangle) error) error {
	if l.Len() == 0 {
		return nil
	}
	for el := l.front(); el != nil; el = el.Next() {
		if err := f(el); err != nil {
			return err
		}
	}
	return nil
}

func (l *PointWithTriangleList) insert(e, at *PointWithTriangle) {
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	e.list = l
	l.len++
}

func (l *PointWithTriangleList) PushBack(e *PointWithTriangle) {
	l.insert(e, l.root.prev)
}

func (l *PointWithTriangleList) Remove(e *PointWithTriangle) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil
	l.len--
}
