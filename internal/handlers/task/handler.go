package task

import (
	"fmt"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTasks(c *gin.Context) {
	type TaskResponseType struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		Status    string `json:"status"`
		TimeSpend uint   `json:"time_spend"`
		Streak    uint   `json:"streak"`
	}

	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	var tasks []models.Task

	if err := db.DB.Where("user_id = ?", userId).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	var data []TaskResponseType
	for _, task := range tasks {
		data = append(data, TaskResponseType{
			ID:        task.ID,
			Title:     task.Title,
			Status:    task.Status,
			TimeSpend: task.TimeSpend,
			Streak:    task.Streak,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func CreateTask(c *gin.Context) {
	var body struct {
		Title     string `json:"title"`
		Status    string `json:"status"`
		TimeSpend uint   `json:"time_spend"`
		Streak    uint   `json:"streak"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	task := models.Task{
		UserId:    userId,
		Title:     body.Title,
		Status:    body.Status,
		TimeSpend: body.TimeSpend,
		Streak:    body.Streak,
	}

	result := db.DB.Create(&task)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
		})
		return
	}

	fmt.Println("result", result)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task Created successfully.",
	})
}
