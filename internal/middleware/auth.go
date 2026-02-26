package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth returns a middleware that checks for API Key.
// It extracts the API Key from:
// 1. Query parameter: "api_key=<key>"
// 2. Authorization header: "Bearer <key>" or just "<key>"
// 3. X-API-Key header: "<key>"
//
// The extracted key is stored in the context as "apiKey".
//
// If expectedKey is provided (non-empty), it validates the extracted key against it.
// If expectedKey is empty, it allows any key (or no key) and just stores it if present.
func Auth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var clientKey string

		// 1. Check Query Parameter
		clientKey = c.Query("api_key")

		// 2. Check Authorization Header
		if clientKey == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					clientKey = parts[1]
				} else {
					// Some clients might send the key directly in Authorization header
					clientKey = authHeader
				}
			}
		}

		// 3. Check X-API-Key Header
		if clientKey == "" {
			clientKey = c.GetHeader("X-API-Key")
		}

		// Store the key in context for downstream handlers/tools
		if clientKey != "" {
			c.Set("apiKey", clientKey)
		}

		// Validation logic
		if expectedKey != "" {
			// If server is configured with a key, enforce it
			if clientKey == "" || clientKey != expectedKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or missing API Key"})
				return
			}
		}

		// If expectedKey is empty, we allow the request to proceed.
		// The clientKey (if any) is available in the context for tools to use (e.g. for forwarding to backend).
		c.Next()
	}
}
