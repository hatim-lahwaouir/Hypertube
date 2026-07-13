package bittorentProtocol 


import (
    "errors"
    "encoding/binary"
    "io"
)


type messageID uint8
const MaxMessageLength = 131081


const (
    MsgChoke         messageID = 0
    MsgUnchoke       messageID = 1
    MsgInterested    messageID = 2
    MsgNotInterested messageID = 3
    MsgHave          messageID = 4
    MsgBitfield      messageID = 5
    MsgRequest       messageID = 6
    MsgPiece         messageID = 7
    MsgCancel        messageID = 8
)




type Msg struct {
    ID messageID
    Payload []byte
}






func (m *Msg) Serialize() []byte {
    var (
        length uint32
    )


    length = uint32(1 + len(m.Payload))

    buf := make([]byte, 4 + length)
    binary.BigEndian.PutUint32(buf[0:4], length)
    buf[4] = byte(m.ID)
    copy(buf[5:], m.Payload)

    return buf
}

func (m *Msg) ParseHave() (uint32, bool) {
    if len(m.Payload) != 4 {
        return 0, false
    }

    pieceIndex := binary.BigEndian.Uint32(m.Payload)
    return pieceIndex,true 
}



func (m *Msg) ParsePiece() (uint32, uint32, []byte, bool) {

    if (len(m.Payload) < 8){
        return 0,0,nil, false
    }

    pieceIndex := binary.BigEndian.Uint32(m.Payload[0:4])
    begin := binary.BigEndian.Uint32(m.Payload[4:8])
    buf := m.Payload[8:]
    
    
    return pieceIndex, begin, buf, true
}





func NewMessage(r io.Reader) (*Msg, error) {
    buf := make([]byte, 4)

    if n, err := io.ReadFull(r, buf); n != 4 || err != nil {
            return nil, err
    }
    length := uint32(binary.BigEndian.Uint32(buf))
    if length == 0{
        return nil, nil
    }

    if length >= MaxMessageLength{
        return nil, errors.New("to much data was sent") 
    }

    messageBuf := make([]byte, length)
    _, err := io.ReadFull(r, messageBuf)
    if err != nil {
        return nil, err
    }

    m := Msg{
        ID:      messageID(messageBuf[0]),
        Payload: messageBuf[1:],
    }

    return &m,nil
}



