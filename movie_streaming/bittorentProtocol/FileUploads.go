package bittorentProtocol

import (
	"os"
)




type FileUploads struct {
	FilePath     string
	Size         int64
	PieceSize    int64
}





func GetCurrentPiece(piece uint32, begin uint32, length uint32,files []*FileUploads) ([]byte, error){
	// here you should check if we already has the piece asked 
	Offset := files[0].PieceSize * int64(piece) + int64(begin)

	current := int64(0)
	bytesRequested := int64(length)
	var buf  = make([]byte, length)
	bufPos := int64(0)

	for _, f := range(files){
		fileEnd := current + f.Size
		if  Offset >= current  && Offset < fileEnd{
			OffsetStartInFile  := Offset - current
			bytesAvailable := f.Size - OffsetStartInFile
	
			bytesToRead := bytesRequested
			
			if bytesAvailable < bytesRequested{
				bytesToRead = bytesAvailable		
			}

			if bytesToRead > 0 {
				file, err := os.Open(f.FilePath)
				if err != nil {
					return nil, err
				}
				_, err = file.ReadAt(buf[bufPos : bufPos+bytesToRead], OffsetStartInFile)


				file.Close()
				if err != nil {
					return nil, err
				}
			}

			Offset += bytesToRead
			bytesRequested -= bytesToRead
			bufPos += bytesToRead

			if bytesRequested == 0 {
				break
			}
		}
		current += f.Size
	}
	return buf, nil
}