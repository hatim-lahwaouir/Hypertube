package types

type CurrentState struct {
	Downloaded int64
	Left       int64
	InfoHash   [20]byte
	PeerId     [20]byte
	Port       uint16
}
