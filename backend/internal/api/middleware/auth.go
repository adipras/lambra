package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/lambra/pkg/authctx"
)

// AuthContextKey identifies the authenticated user in gin context
const AuthContextKey = "auth_user"

// AuthUser is the authenticated user stored in request context
type AuthUser struct {
	UUID     string
	Username string
	Role     string
}

// CurrentUser returns the authenticated user from gin context, or nil
func CurrentUser(c *gin.Context) *AuthUser {
	if v, ok := c.Get(AuthContextKey); ok {
		if user, ok := v.(*AuthUser); ok {
			return user
		}
	}
	return nil
}

// ClaimsValidator parses a JWT string and returns its claims
type ClaimsValidator func(tokenString string) (jwt.MapClaims, error)

// Auth returns middleware that requires a valid Bearer JWT
func Auth(validate ClaimsValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Authorization header must be Bearer {token}"})
			return
		}

		claims, err := validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Invalid or expired token"})
			return
		}

		username, _ := claims["username"].(string)
		uuid, _ := claims["uuid"].(string)
		role, _ := claims["role"].(string)
		if username == "" || uuid == "" {
			c.AbortWithStatusJSON(401, gin.H{"success": false, "message": "Invalid token claims"})
			return
		}

		c.Set(AuthContextKey, &AuthUser{UUID: uuid, Username: username, Role: role})
		// Also expose on request context for services
		c.Request = c.Request.WithContext(authctx.WithUser(c.Request.Context(), username))
		c.Next()
	}
}
