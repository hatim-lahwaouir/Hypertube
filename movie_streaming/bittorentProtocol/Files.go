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
	Offset int64
}

func Newfile(ParentPath string , filePath string, size uint32, PieceSize int64, offset int64) *File {
	return &File{FilePath: filepath.Join(ParentPath, filePath), Size: int64(size), created: false, PieceSize: PieceSize, Offset: offset}
}


func (f *File) NewFileUploads() *FileUploads {
	return &FileUploads{FilePath: f.FilePath, Size: f.Size, PieceSize: f.PieceSize}
}


func (f *File) Create() error {
	// make sure fiest that all sub directories of the file are downloaded
	dirPath := filepath.Dir(f.FilePath)
	fmt.Println(">>>>>>>>>>>>>>>>", dirPath, f.FilePath)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	fd, err := os.OpenFile(f.FilePath, os.O_RDWR|os.O_CREATE, 0644)
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

func (f *File) WriteData(offset int64,buf []byte) error {
	f.FD.WriteAt(buf, offset)
	return nil
}
