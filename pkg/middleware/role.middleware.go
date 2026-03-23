package middleware

import (
	"ai-symptom-checker/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware restricts access to specific roles
// Usage: router.Use(middleware.RoleMiddleware("admin"))
//
//	router.Use(middleware.RoleMiddleware("doctor", "admin"))
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			utils.Unauthorized(c, "No role found in token")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			utils.Unauthorized(c, "Invalid role format")
			c.Abort()
			return
		}

		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "You do not have permission to access this resource")
		c.Abort()
	}
}
