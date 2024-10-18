package middleware

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT Middleware
func JWTAuthMiddleware(userRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove the "Bearer " prefix and extract the token
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if len(userRoles) > 0 {
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if len(userRoles) > 0 && !slices.Contains(userRoles, claims["scope"].(string)) {
					log.Printf("User does not have the required role: %v not in %v", claims["scope"].(string), userRoles)
					c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
					c.Abort()
					return
				}
			}
		}

		userId, err := GetUserIDFromJWT(*token)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store user_id in context
		c.Set("user_id", userId) // uint
		// If the token is valid, proceed to the next handler.
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			// response 200 if it's a preflight request
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}
