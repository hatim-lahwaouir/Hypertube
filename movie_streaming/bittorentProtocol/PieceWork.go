package bittorentProtocol

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type PieceWork struct {
	Size            uint32
	Downloaded      atomic.Uint32
	Index           uint32
	NPiece          uint32
	Mu              sync.Mutex
	SizeOfAllPieces uint32
	BlockSize       uint32
	PiecesState     []PieceState
	Buffer          []byte
	sha1            [20]byte
	start 			time.Time
}

type PieceState uint8

const (
	BlockStatePending   = iota // Requested but not yet received
	BlockStateCompleted        // Received and written to storage
)

func NewPieceWork(index uint32, size uint32, allPiecesSize uint32, expectedSha1 [20]byte) *PieceWork {

	if size * (index + 1) > allPiecesSize{
		fmt.Println("piece size ",size, allPiecesSize - (size * index))
		size = allPiecesSize - (size * index)
		// os.Exit(1)
	}


	p := &PieceWork{
		Index:           index,
		Size:            size,
		NPiece:          (size + (16384 - 1)) / 16384,
		BlockSize:       16384,
		SizeOfAllPieces: allPiecesSize,
		sha1:            expectedSha1,
		start: time.Now(),
	}

	p.PiecesState = make([]PieceState, p.NPiece)

	// FIX: Must allocate the buffer to hold the actual file bytes!
	p.Buffer = make([]byte, size)

	return p
}

func (p *PieceWork) Done() bool {
	return p.Downloaded.Load() == p.Size
}

func (p *PieceWork) PrintState() {
	fmt.Println(p.Index, "piece took ", time.Since(p.start))
}

func (p *PieceWork) PieceDone(pieceSize uint32) {
	p.Downloaded.Add(pieceSize)
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
	p.Downloaded.Add(uint32(len(buf))) 
}


func (p *PieceWork) ValidateEntigrity() bool {
	hash := sha1.New()
	hash.Write(p.Buffer)
	hashedData := hash.Sum(nil)

	return bytes.Equal(hashedData, p.sha1[:])
}


func (p *PieceWork) GetDownloaded() uint32 {

	return p.Downloaded.Load()
}
