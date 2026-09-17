package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"tensechassignment/internal/auth"
	"tensechassignment/internal/utils"
)

const (
	CtxUserID   = "user_id"
	CtxTenantID = "tenant_id"
	CtxRoles    = "roles"
)

func Auth(validator *auth.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.Error(
				c,
				http.StatusUnauthorized,
				"missing or invalid authorization header",
			)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := validator.ValidateToken(token)
		if err != nil {
			utils.Error(
				c,
				http.StatusUnauthorized,
				"invalid or expired token",
			)
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxTenantID, claims.TenantID)
		c.Set(CtxRoles, claims.Roles)

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get(CtxRoles)
		if !exists {
			utils.Error(
				c,
				http.StatusForbidden,
				"roles not found",
			)
			c.Abort()
			return
		}

		roles, ok := value.([]string)
		if !ok {
			utils.Error(
				c,
				http.StatusForbidden,
				"invalid roles",
			)
			c.Abort()
			return
		}

		for _, role := range roles {
			if role == "admin" {
				c.Next()
				return
			}
		}

		utils.Error(
			c,
			http.StatusForbidden,
			"admin role required",
		)
		c.Abort()
	}
}

func TenantID(c *gin.Context) string {
	value, exists := c.Get(CtxTenantID)
	if !exists {
		return ""
	}

	tenantID, ok := value.(string)
	if !ok {
		return ""
	}

	return tenantID
}