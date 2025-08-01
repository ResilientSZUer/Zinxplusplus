package aoi

import "fmt"

const (
	nodeCapacityDefault = 4
	maxDepthDefault     = 8
)

type QuadtreeNode struct {
	boundary Rect
	points   []IPoint
	children [4]*QuadtreeNode
	isLeaf   bool
	capacity int
	depth    int
	maxDepth int
}

func NewQuadtreeNode(boundary Rect, capacity, depth, maxDepth int) *QuadtreeNode {
	if capacity <= 0 {
		capacity = nodeCapacityDefault
	}
	if maxDepth <= 0 {
		maxDepth = maxDepthDefault
	}
	return &QuadtreeNode{
		boundary: boundary,
		points:   make([]IPoint, 0, capacity),
		isLeaf:   true,
		capacity: capacity,
		depth:    depth,
		maxDepth: maxDepth,
	}
}

func (n *QuadtreeNode) insert(p IPoint) bool {

	if !n.boundary.ContainsPoint(p) {
		return false
	}

	if n.isLeaf {

		if len(n.points) < n.capacity || n.depth == n.maxDepth {

			for _, existingPoint := range n.points {
				if existingPoint.GetID() == p.GetID() {

					return false
				}
			}
			n.points = append(n.points, p)
			return true
		}

		n.subdivide()
	}

	if n.children[0].insert(p) {
		return true
	}
	if n.children[1].insert(p) {
		return true
	}
	if n.children[2].insert(p) {
		return true
	}
	if n.children[3].insert(p) {
		return true
	}

	fmt.Printf("[QuadtreeNode] Error: Point %v could not be inserted into any child of node %v\n", p, n.boundary)
	return false
}

func (n *QuadtreeNode) subdivide() {
	if !n.isLeaf {
		return
	}

	n.isLeaf = false

	x := n.boundary.MinX
	z := n.boundary.MinZ
	hw := n.boundary.Width() / 2
	hh := n.boundary.Height() / 2
	nextDepth := n.depth + 1

	n.children[0] = NewQuadtreeNode(Rect{x, z, x + hw, z + hh}, n.capacity, nextDepth, n.maxDepth)

	n.children[1] = NewQuadtreeNode(Rect{x + hw, z, x + hw + hw, z + hh}, n.capacity, nextDepth, n.maxDepth)

	n.children[2] = NewQuadtreeNode(Rect{x, z + hh, x + hw, z + hh + hh}, n.capacity, nextDepth, n.maxDepth)

	n.children[3] = NewQuadtreeNode(Rect{x + hw, z + hh, x + hw + hw, z + hh + hh}, n.capacity, nextDepth, n.maxDepth)

	oldPoints := n.points
	n.points = nil

	for _, p := range oldPoints {
		inserted := false
		for i := 0; i < 4; i++ {
			if n.children[i].insert(p) {
				inserted = true
				break
			}
		}
		if !inserted {

			fmt.Printf("[QuadtreeNode] Warning: Point %v failed to reinsert during subdivide of node %v\n", p, n.boundary)

		}
	}
}

func (n *QuadtreeNode) queryRange(queryRange Rect, results *[]uint64) {

	if !n.boundary.Intersects(&queryRange) {
		return
	}

	if n.isLeaf {
		for _, p := range n.points {

			px, pz := p.GetX(), p.GetZ()
			if px >= queryRange.MinX && px < queryRange.MaxX && pz >= queryRange.MinZ && pz < queryRange.MaxZ {
				*results = append(*results, p.GetID())
			}
		}
		return
	}

	for i := 0; i < 4; i++ {
		if n.children[i] != nil {
			n.children[i].queryRange(queryRange, results)
		}
	}
}

func (n *QuadtreeNode) remove(p IPoint) bool {

	if !n.boundary.ContainsPoint(p) {
		return false
	}

	if n.isLeaf {
		found := false
		newPoints := make([]IPoint, 0, len(n.points))
		for _, existingPoint := range n.points {

			if existingPoint.GetID() == p.GetID() {
				found = true
			} else {
				newPoints = append(newPoints, existingPoint)
			}
		}
		if found {
			n.points = newPoints
		}
		return found
	}

	removed := false
	for i := 0; i < 4; i++ {
		if n.children[i] != nil && n.children[i].remove(p) {
			removed = true

			break
		}
	}
	return removed
}
