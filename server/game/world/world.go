package world

type World struct {
	sizeX int
	sizeY int
}

func NewWorld(sizeX int, sizeY int) *World {
	return &World{
		sizeX: sizeX,
		sizeY: sizeY,
	}
}

func (w *World) GetSizeX() int {
	return w.sizeX
}

func (w *World) GetSizeY() int {
	return w.sizeY
}
