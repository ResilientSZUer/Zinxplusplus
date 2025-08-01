package ziface

type IAoiManager interface {
	GetSurroundingObjectIDs(x, z float32) (objectIDs []uint64)

	AddObjectToGridByPos(objID uint64, x, z float32) error

	RemoveObjectFromGridByPos(objID uint64, x, z float32) error

	UpdateObjectPos(objID uint64, oldX, oldZ, newX, newZ float32) error
}
