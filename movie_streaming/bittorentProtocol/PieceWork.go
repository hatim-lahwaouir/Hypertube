package bittorentProtocol 




type PieceWork struct {
	Size uint32
	Downloaded uint32
	Index uint32
	NPiece uint32
}

func NewPieceWork(index uint32,  size uint32) *PieceWork{
	return &PieceWork{Index:index, Size: size, NPiece: (size + (16384 - 1) / 16384 )}
}
