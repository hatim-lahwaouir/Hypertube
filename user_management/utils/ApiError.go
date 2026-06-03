package utils


import (
    "net/http"
    "encoding/json"
    "fmt"
)



type ApiError struct {
    Status int
    Msg any
}

func (a ApiError) Error() string {

    return fmt.Sprintf("Api error status code %d", a.Status)

}

func NewApiError(status int, msg any) ApiError {

    return ApiError{Status: status, Msg: msg} 

}

type APIFunc func(http.ResponseWriter, *http.Request) error





func WriteResp(w http.ResponseWriter, status int, v any)  error {
    w.WriteHeader(status)
    w.Header().Set("content-Type", "application/json")
    return json.NewEncoder(w).Encode(v)

}

func MakeHandler(h APIFunc) func(http.ResponseWriter, *http.Request){


    return func(w http.ResponseWriter, r *http.Request) {
            if err := h(w,r); err != nil {
                if apiError, ok := err.(ApiError); ok {
                    WriteResp(w, apiError.Status, apiError)
                }else {
                    apiError := ApiError{Status: http.StatusInternalServerError, Msg:  "Internal Server Error"}
                    WriteResp(w, apiError.Status, apiError)
                }
            }
    }
}







