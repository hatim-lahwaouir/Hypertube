package services


import (
    "mime/multipart"
    "os"
    "io"
    "github.com/google/uuid"
    "bufio"
    "path/filepath"
    "strings"
    "net/url"
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

    var (
        fileName string
        filePath string
    )
    // check for mim Type
    valid, src := i.ValidMimType(file)
    if  valid  == false{
        return "", fmt.Errorf("Invalid mim type we only support %s ", strings.Join(i.allowedMimTypes, ",")) 

    }

    fileName =  fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Unix())
    filePath = filepath.Join(i.path, fileName)

    dst , err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("unexpected error please try again") 
    }
    defer dst.Close()
    _ , err = io.Copy(dst, src)

    if err != nil {

        return "", fmt.Errorf("unexpected error please try again") 
    }

    return fileName , nil
}


func (i *FileUploadService) ValidMimType(file io.Reader) (bool, io.Reader) {

    var (
        valid bool
    )


    bufReader := bufio.NewReader(file)
    fmt.Println(">>>>>>>>>", bufReader.Buffered())
    valid = false
    buffer, err := bufReader.Peek(512)
    if err != nil {
            fmt.Println("here", err)
            return false, bufReader
    }

    contentType := http.DetectContentType(buffer)

    fmt.Println("here", contentType)
    for _, value := range i.allowedMimTypes {
        if value == contentType {

            valid = true
        }
    } 


    return valid, bufReader
}

var client *http.Client = &http.Client{} 

func (i *FileUploadService) DownloadAnImage(imgUrl string) (string, error) {
    
    var (
        fileName string
        filePath string

        checkBuf [1]byte
    )


    const maxLimit int64 = 5 * 1024 * 1024  // 5 MB


    if _, err := url.Parse(imgUrl); err != nil {
        
        return "", fmt.Errorf("Invalid img URl")
    }
    req, err := http.NewRequest("GET", imgUrl , nil)
    if err != nil {

        return "", fmt.Errorf("error doing the request to get the img") 
    }


    req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")



    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("img URl isn't working")
    }
    defer resp.Body.Close()
    

   
    imgStream := io.LimitReader(resp.Body, maxLimit)
    valid, src := i.ValidMimType(imgStream)
    if valid  == false {
       return "", fmt.Errorf("Invalid mim type we only support %s ", strings.Join(i.allowedMimTypes, ",")) 
    }


    

	n, err := resp.Body.Read(checkBuf[:])
    fmt.Println(n, err)
	if n > 0 {
        //err := os.Remove(filePath)
        if err != nil {
		    return "", fmt.Errorf("error processing the img")
        }

		return "", fmt.Errorf("image exceeds the maximum allowed size of 5MB")
	}

    fileName =  fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Unix())
    filePath = filepath.Join(i.path, fileName)

    dst , err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("error processing the img") 
    }
    defer dst.Close()
    _ , err = io.Copy(dst, src)
    if err != nil {
        fmt.Println(err)
    }
    return fileName , nil
}






