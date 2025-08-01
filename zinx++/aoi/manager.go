package aoi

import (
	"fmt"
	"sync"

	"zinxplusplus/ziface"
)

type AoiManager struct {
	quadtree *Quadtree
	objMap   map[uint64]IPoint
	mapLock  sync.RWMutex
}

func NewQuadtreeAoiManager(minX, maxX, minZ, maxZ float32, capacity int, maxDepth int) ziface.IAoiManager {
	boundary := Rect{
		MinX: minX,
		MinZ: minZ,
		MaxX: maxX,
		MaxZ: maxZ,
	}
	qt := NewQuadtree(boundary, capacity, maxDepth)
	return &AoiManager{
		quadtree: qt,
		objMap:   make(map[uint64]IPoint),
	}
}

func (m *AoiManager) GetSurroundingObjectIDs(x, z float32) []uint64 {

	viewRange := float32(50.0)

	queryRect := Rect{
		MinX: x - viewRange,
		MinZ: z - viewRange,
		MaxX: x + viewRange,
		MaxZ: z + viewRange,
	}

	m.mapLock.RLock()
	defer m.mapLock.RUnlock()

	return m.quadtree.QueryRange(queryRect)
}

func (m *AoiManager) AddObjectToGridByPos(objID uint64, x, z float32) error {
	p := &Point{ObjID: objID, X: x, Z: z}

	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	if _, ok := m.objMap[objID]; ok {
		return fmt.Errorf("object %d already exists in AOI manager", objID)
	}

	if !m.quadtree.Insert(p) {
		return fmt.Errorf("failed to insert object %d into quadtree (maybe out of bounds?)", objID)
	}

	m.objMap[objID] = p

	return nil
}

func (m *AoiManager) RemoveObjectFromGridByPos(objID uint64, x, z float32) error {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	existingPoint, ok := m.objMap[objID]
	if !ok {
		return fmt.Errorf("object %d not found in AOI manager", objID)
	}

	if !m.quadtree.Remove(existingPoint) {

		fmt.Printf("[AOIManager] Warning: Failed to remove object %d from quadtree\n", objID)

	}

	delete(m.objMap, objID)

	return nil
}

func (m *AoiManager) UpdateObjectPos(objID uint64, oldX, oldZ, newX, newZ float32) error {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	existingPoint, ok := m.objMap[objID]
	if !ok {
		return fmt.Errorf("object %d not found for update", objID)
	}

	if !m.quadtree.Remove(existingPoint) {
		fmt.Printf("[AOIManager] Warning: Failed to remove object %d from old position (%f, %f) during update\n", objID, oldX, oldZ)

	}

	newPoint := &Point{ObjID: objID, X: newX, Z: newZ}
	m.objMap[objID] = newPoint

	if !m.quadtree.Insert(newPoint) {

		fmt.Printf("[AOIManager] Error: Failed to insert object %d into new position (%f, %f) during update\n", objID, newX, newZ)

		delete(m.objMap, objID)
		return fmt.Errorf("failed to insert object %d into quadtree at new position", objID)
	}

	return nil
}
