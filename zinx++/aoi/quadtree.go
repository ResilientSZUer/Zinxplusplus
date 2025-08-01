package aoi

import "sync"

type Point struct {
	ObjID uint64
	X     float32
	Z     float32
}

func (p *Point) GetX() float32 {
	return p.X
}

func (p *Point) GetZ() float32 {
	return p.Z
}

func (p *Point) GetID() uint64 {
	return p.ObjID
}

type Rect struct {
	MinX float32
	MinZ float32
	MaxX float32
	MaxZ float32
}

func (r *Rect) ContainsPoint(p IPoint) bool {
	x, z := p.GetX(), p.GetZ()
	return x >= r.MinX && x < r.MaxX && z >= r.MinZ && z < r.MaxZ
}

func (r *Rect) Intersects(other *Rect) bool {
	return r.MinX < other.MaxX && r.MaxX > other.MinX &&
		r.MinZ < other.MaxZ && r.MaxZ > other.MinZ
}

func (r *Rect) Width() float32 {
	return r.MaxX - r.MinX
}

func (r *Rect) Height() float32 {
	return r.MaxZ - r.MinZ
}

type Quadtree struct {
	root *QuadtreeNode
	lock sync.RWMutex
}

func NewQuadtree(boundary Rect, capacity int, maxDepth int) *Quadtree {
	if capacity < 1 {
		capacity = 1
	}
	if maxDepth < 1 {
		maxDepth = 8
	}
	root := NewQuadtreeNode(boundary, capacity, 0, maxDepth)
	return &Quadtree{
		root: root,
	}
}

func (qt *Quadtree) Insert(p IPoint) bool {
	qt.lock.Lock()
	defer qt.lock.Unlock()
	return qt.root.insert(p)
}

func (qt *Quadtree) QueryRange(queryRange Rect) []uint64 {
	qt.lock.RLock()
	defer qt.lock.RUnlock()
	var results []uint64
	qt.root.queryRange(queryRange, &results)
	return results
}

func (qt *Quadtree) Remove(p IPoint) bool {
	qt.lock.Lock()
	defer qt.lock.Unlock()
	return qt.root.remove(p)
}

func (qt *Quadtree) Clear() {
	qt.lock.Lock()
	defer qt.lock.Unlock()

	newRoot := NewQuadtreeNode(qt.root.boundary, qt.root.capacity, 0, qt.root.maxDepth)
	qt.root = newRoot
}
