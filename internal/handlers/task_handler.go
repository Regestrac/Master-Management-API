package handlers

import "github.com/gin-gonic/gin"

func GetTasks(c *gin.Context) {
	// This function will handle the GET request to fetch tasks
	// For now, we will return a placeholder response
	c.JSON(200, gin.H{
		"message": "List of tasks",
	})
}
