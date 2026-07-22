package bittorentProtocol

import (
	"fmt"
	"os"
	"path/filepath"
)

type File struct {
	FilePath     string
	Size         int64
	BytesWritten int64
	FD           *os.File
	created      bool
	PieceSize    int64
}

func Newfile(filePath string, size uint32, PieceSize int64) *File {
	return &File{FilePath: filePath, Size: int64(size), created: false, PieceSize: PieceSize}
}

func (f *File) Create(path string) error {
	// make sure fiest that all sub directories of the file are downloaded
	filePath := filepath.Join(path, f.FilePath)
	dirPath := filepath.Dir(filePath)
	fmt.Println(">>>>>>>>>>>>>>>>", dirPath)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	if err = fd.Truncate(f.Size); err != nil {
		fd.Close()
		return err
	}

	f.FD = fd
	f.created = true
	return nil
}

func (f *File) Done() bool {
	done := f.BytesWritten >= f.Size

	if done {
		f.FD.Close()
	}
	return done
}

func (f *File) IsCreated() bool {
	return f.created
}

func (f *File) WriteData(pieceIndex int, buf []byte) ([]byte, error) {


	offset := int64(pieceIndex) * (f.PieceSize)

	var overflow []byte
	toWrite := int64(len(buf))

	if offset+toWrite > int64(f.Size) {
		toWrite = int64(f.Size) - offset
		overflow = buf[toWrite:]
	}

	_, err := f.FD.WriteAt(buf[:toWrite], offset)
	if err != nil {
		return overflow, err
	}

	f.BytesWritten += toWrite

	return overflow, nil
}
