package services


import (
    "mime/multipart"
    "os"
    "io"
    "github.com/google/uuid"
    "path/filepath"
    "strings"
    "net/http"
    "fmt"
    "time"
)




type FileUploadService struct {
    path string
    allowedMimTypes []string
    maxUploadSize int64
}




var cacheFileUploadService *FileUploadService


func NewFileUploadService(path string, maxUploadSize int64) *FileUploadService {
    if  cacheFileUploadService == nil {
        cacheFileUploadService = &FileUploadService{path: path, allowedMimTypes : []string{"image/png", "image/jpeg"}, maxUploadSize: maxUploadSize}
    }
    return cacheFileUploadService
}




func (i *FileUploadService) UploadFile(file multipart.File) (string, error) {
    fmt.Println(file)

    var (
        fileName string
        filePath string
    )
    // check for mim Type
    if i.ValidMimType(file) == false {
        return "", fmt.Errorf("Invalid mim type we only support %s ", strings.Join(i.allowedMimTypes, ",")) 

    }

    fileName =  fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Unix())
    filePath = filepath.Join(i.path, fileName)

    dst , err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("unexpected error please try again") 
    }

    _ , err = io.Copy(dst, file)

    if err != nil {

        return "", fmt.Errorf("unexpected error please try again") 
    }

    return fileName , nil
}


func (i *FileUploadService) ValidMimType(file multipart.File) bool {

    var (
        valid bool
    )

    valid = false
    buffer := make([]byte, 512)

    bytesRead , err := file.Read(buffer)
    if err != nil {
            fmt.Println("file Upload ", err)
            return false
    }

    contentType := http.DetectContentType(buffer[:bytesRead])

    for _, value := range i.allowedMimTypes {
        if value == contentType {
            valid = true
        }
    } 

    if _, err := file.Seek(0, io.SeekStart); err != nil {
        fmt.Println("file Upload ", err)
        return false 
	}


    return valid 
}

