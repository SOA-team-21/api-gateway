package middleware

import (
	"net/http"
	"strings"

	"api-gateway.com/utils"
)

func JwtMiddleware(next http.Handler, protectedPaths []*utils.Path) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		//TODO: Add protected routes
		// protected, _ := isProtectedPath(r.URL.Path, protectedPaths)
		// if !protected {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }
		//TODO: Call auth service to validate token
	})
}

func isProtectedPath(path string, protectedPaths []*utils.Path) (bool, string) {
	for _, p := range protectedPaths {
		if strings.Contains(path, p.Path) {
			return true, p.Role
		}
	}
	return false, ""
}
