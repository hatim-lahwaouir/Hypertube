package bittorentProtocol

import (
	"bytes"
	"crypto/sha1"
	"fmt"
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
	PiecesState     []PieceState
	Buffer          []byte
	sha1            [20]byte
}

type PieceState uint8

const (
	BlockStatePending   = iota // Requested but not yet received
	BlockStateCompleted        // Received and written to storage
)

func NewPieceWork(index uint32, size uint32, allPiecesSize uint32, expectedSha1 [20]byte) *PieceWork {

	p := &PieceWork{
		Index:           index,
		Size:            size,
		NPiece:          (size + (16384 - 1)) / 16384,
		BlockSize:       16384,
		SizeOfAllPieces: allPiecesSize,
		sha1:            expectedSha1,
	}

	p.PiecesState = make([]PieceState, p.NPiece)

	// FIX: Must allocate the buffer to hold the actual file bytes!
	p.Buffer = make([]byte, size)

	return p
}

func (p *PieceWork) Done() bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return p.Downloaded == p.Size
}

func (p *PieceWork) PrintState() {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	fmt.Printf("piece %d [%.2f%%/100%%]\n", p.Index, (float64(p.Downloaded)*100)/float64(p.Size))
}

func (p *PieceWork) PieceDone(pieceSize uint32) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.Downloaded += pieceSize
}

func (p *PieceWork) Pending(begin uint32) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.PiecesState[begin/p.BlockSize] = BlockStatePending
}

func (p *PieceWork) IsThisDone(begin uint32) bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return p.PiecesState[begin/p.BlockSize] == BlockStateCompleted
}

func (p *PieceWork) SetPiece(indexOfPice uint32, buf []byte, begin uint32) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if p.Index != indexOfPice {
		return
	}

	stateIndex := begin / p.BlockSize

	// Prevent double-counting downloaded bytes if a peer sends us a duplicate block
	if p.PiecesState[stateIndex] == BlockStateCompleted {
		return
	}

	// FIX: Copy using the byte offset (begin), NOT the chunk index
	// We slice from begin to (begin + length of the buffer)
	copy(p.Buffer[begin:begin+uint32(len(buf))], buf)

	p.PiecesState[stateIndex] = BlockStateCompleted
	p.Downloaded += uint32(len(buf))
}


func (p *PieceWork) ValidateEntigrity() bool {
	hash := sha1.New()
	hash.Write(p.Buffer)
	hashedData := hash.Sum(nil)

	return bytes.Equal(hashedData, p.sha1[:])
}


func (p *PieceWork) GetDownloaded() uint32 {
	return p.Downloaded
}