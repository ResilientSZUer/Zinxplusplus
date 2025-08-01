package aoi

import "zinxplusplus/ziface"

type IAoiManager = ziface.IAoiManager

type IPoint interface {
	GetX() float32
	GetZ() float32
	GetID() uint64
}
