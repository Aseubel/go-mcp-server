package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ServiceKeyHeader authenticates the MCP gateway deployment.
	ServiceKeyHeader = "X-MCP-Service-Key"
	// DeveloperKeyHeader carries the end-user developer key to the Java boundary.
	DeveloperKeyHeader = "X-Developer-API-Key"
	// DeveloperKeyContextKey is the request context key used by extension tools.
	DeveloperKeyContextKey = "developerApiKey"
)

// RequireServiceKey authenticates the internal MCP endpoint with its
// deployment-level service key. It does not identify an end user.
func RequireServiceKey(expectedKey string) gin.HandlerFunc {
	configuredKey := strings.TrimSpace(expectedKey)

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		serviceKey := strings.TrimSpace(c.GetHeader(ServiceKeyHeader))
		if configuredKey == "" || serviceKey == "" || len(serviceKey) != len(configuredKey) ||
			subtle.ConstantTimeCompare([]byte(serviceKey), []byte(configuredKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or missing MCP service key"})
			return
		}

		c.Next()
	}
}

// RequireDeveloperKey authenticates the public MCP endpoint at the gateway
// boundary. Java performs the authoritative key lookup and scope check when a
// user-scoped tool is called.
func RequireDeveloperKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		developerKey := extractDeveloperKey(c)
		if developerKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing developer API key"})
			return
		}

		c.Set(DeveloperKeyContextKey, developerKey)
		c.Next()
	}
}

// Auth is retained as an internal compatibility alias for callers that used
// the previous service-key-only middleware.
func Auth(expectedKey string) gin.HandlerFunc {
	return RequireServiceKey(expectedKey)
}

func extractDeveloperKey(c *gin.Context) string {
	developerKey := strings.TrimSpace(c.GetHeader(DeveloperKeyHeader))
	if developerKey == "" {
		// Keep the existing header usable for user-scoped developer keys.
		developerKey = strings.TrimSpace(c.GetHeader("X-API-Key"))
	}
	if developerKey == "" {
		developerKey = bearerToken(c.GetHeader("Authorization"))
	}
	return developerKey
}

func bearerToken(value string) string {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
