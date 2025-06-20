package subtasks

import (
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllSubtasks(c *gin.Context) {
	type TaskResponseType struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		Status    string `json:"status"`
		TimeSpend uint   `json:"time_spend"`
		Streak    uint   `json:"streak"`
		ParentId  *uint  `json:"parent_id"`
	}

	taskId := c.Param("id")

	var subtasks []models.Task

	if err := db.DB.Where("parent_id = ?", taskId).Find(&subtasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subtasks"})
		return
	}

	var data []TaskResponseType
	for _, task := range subtasks {
		data = append(data, TaskResponseType{
			ID:        task.ID,
			Title:     task.Title,
			Status:    task.Status,
			TimeSpend: task.TimeSpend,
			Streak:    task.Streak,
			ParentId:  task.ParentId,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
