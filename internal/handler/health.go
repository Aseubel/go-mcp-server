package handler

import "github.com/gin-gonic/gin"

// Health handles the health check endpoint.
func Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
