package middleware

import (
	"ai-symptom-checker/config"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures allowed origins, methods, and headers
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {

			if strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") {
				return true
			}

			if config.App.FrontendURL != "" && origin == config.App.FrontendURL {
				return true
			}

			if strings.HasSuffix(origin, ".vercel.app") {
				return true
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "Cookie"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}
