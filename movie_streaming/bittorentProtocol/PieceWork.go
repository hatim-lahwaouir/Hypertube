package bittorentProtocol

import (
	"bytes"
	"crypto/sha1"
	"sync"
)
type PieceWork struct {
	Size            uint32
	Downloaded      uint32
	Index           uint32
	NPiece          uint32
	Mu              sync.Mutex
	SizeOfAllPieces uint32
	BlockSize       uint32
	PiecesState   []PieceState
	Buffer        []byte
	sha1 [20]byte
}

type PieceState uint8


const (
    BlockStatePending = iota             // Requested but not yet received
    BlockStateCompleted           // Received and written to storage
)


func NewPieceWork(index uint32, size uint32, allPiecesSize uint32, sha1 [20]byte) *PieceWork {

	p := &PieceWork{Index: index, Size: size, NPiece: ((size + (16384 - 1)) / 16384), BlockSize: uint32(16384), 
		 SizeOfAllPieces: allPiecesSize, sha1: sha1}

	p.PiecesState = make([]PieceState, p.NPiece)
	
	
	return p
}

func (p *PieceWork) Done() bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return p.Downloaded == p.Size
}

func (p *PieceWork) PieceDone(pieceSize uint32) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.Downloaded += pieceSize
}


func (p *PieceWork) Pending(begin  uint32) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.PiecesState[begin / p.BlockSize] = BlockStatePending
}



func (p *PieceWork) IsThisDone(begin  uint32) bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return p.PiecesState[begin / p.BlockSize] == BlockStateCompleted
}


func (p *PieceWork) SetPiece(buf []byte, begin uint32) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	// completed piece
	p.PiecesState[begin / p.BlockSize] = BlockStateCompleted
	index := begin / p.BlockSize

	// validating entgrity of the piece
	hash := sha1.New()
    hash.Write(buf)
    hashedData := hash.Sum(nil)

	if bytes.Equal(hashedData, p.sha1[:]) {
		p.PiecesState[begin / p.BlockSize] = BlockStatePending
		return 
	} 
	copy(p.Buffer[(index):], buf)
}