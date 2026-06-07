package middleware 


import (
    "net/http"
    "github.com/hatim-lahwaouir/Hypertube/user_management/services"
    "context"
)








func Auth(next http.Handler) http.Handler {
    
    var authService = services.NewAuthService()



    return http.HandlerFunc(func  (w http.ResponseWriter, r *http.Request)  {
        token := r.Header.Get("Authorization")

        claims, err := authService.ValidateToken(token)

        if err != nil {

            w.Header().Set("Content-Type", "text/plain; charset=utf-8")
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Status Unauthorized"))
            return 
        }

        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
