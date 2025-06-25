package history

import (
	"log"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTaskHistory(c *gin.Context) {
	taskId := c.Param("id")

	var history []models.TaskHistory

	if err := db.DB.Where("task_id = ?", taskId).Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    history,
	})
}

func AddToHistory(c *gin.Context) {
	var body struct {
		Action string `json:"action"`
		Before string `json:"before"`
		After  string `json:"after"`
		TaskId uint   `json:"task_id"`
		UserId uint   `json:"user_id"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	history := models.TaskHistory{
		Action: body.Action,
		Before: body.Before,
		After:  body.After,
		TaskId: body.TaskId,
		UserId: body.UserId,
	}

	if err := db.DB.Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data":    history,
	})
}

func LogHistory(action string, before string, after string, taskId uint, userId uint) {
	history := models.TaskHistory{
		Action: action,
		Before: before,
		After:  after,
		TaskId: taskId,
		UserId: userId,
	}

	if err := db.DB.Create(&history).Error; err != nil {
		log.Printf("Failed to log history: %v", err)
		return
	}

}
