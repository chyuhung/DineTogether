package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func success(c *gin.Context, message string, data ...gin.H) {
	resp := gin.H{"message": message, "success": true}
	if len(data) > 0 {
		for k, v := range data[0] {
			resp[k] = v
		}
	}
	c.JSON(http.StatusOK, resp)
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message, "success": false})
}

func notFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message, "success": false})
}

func serverError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message, "success": false})
}

func unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": message, "success": false})
}

func forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{"error": message, "success": false})
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
