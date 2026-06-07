package dto




type JWT struct {
       Access_token  string `json:"access_token"`
}




func NewJwt(jwt string) JWT {

    return JWT{Access_token: jwt}
}
