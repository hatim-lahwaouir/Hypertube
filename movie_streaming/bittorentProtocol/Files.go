package bittorentProtocol

import (

"os"
	"path/filepath"
)




type File struct {
	FilePath string
	Size  int64
	BytesWritten int64
	FD    *os.File
	created bool
}


func Newfile(filePath string, size uint32) *File{
	return &File{FilePath: filePath, Size: int64(size), created: false}
}

func (f *File) Create(path string) error {
	fd, err := os.OpenFile(filepath.Join(path, f.FilePath), os.O_RDWR|os.O_CREATE, 0644)	
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


func (f * File) Done() bool {
	done:=  f.BytesWritten >= f.Size

	if done {
		f.FD.Close()
	}
	return done
}


func (f * File) IsCreated() bool {
	return f.created
}


func (f *File) WriteData(buf []byte) ([]byte, error) {
	toWrite:= int64(len(buf))
	var (
		overflow []byte
	)

	overflow = nil

	if len(buf) + int(f.BytesWritten) > int(f.Size){
		toWrite = f.Size - f.BytesWritten
		overflow = buf[toWrite:]
	} 

	n, err := f.FD.WriteAt(buf, f.BytesWritten)
	if err != nil {
		return overflow, err
	}
	
	f.BytesWritten += int64(n)


	return overflow, nil
}







