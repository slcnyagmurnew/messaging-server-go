package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Index godoc
// @Summary      Show service status
// @Description  Returns a welcome message
// @Tags         root
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       / [get]
func Index(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "messaging-server", "status": "ok"})
}

// Health godoc
// @Summary      Health check
// @Description  Returns OK if the service is alive
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /health [get]
func Health(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"message": "healthy", "status": "ok"})
}
