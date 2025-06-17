package task

import (
	"fmt"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllTasks(c *gin.Context) {
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
		"data":    gin.H{"id": task.ID},
	})
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	var task models.Task
	if err := db.DB.Where("id = ? AND user_id = ?", id, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if err := db.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func GetTask(c *gin.Context) {
	id := c.Param("id")
	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	var task models.Task
	if err := db.DB.Where("id = ? AND user_id = ?", id, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"status":      task.Status,
			"time_spend":  task.TimeSpend,
			"streak":      task.Streak,
			"description": task.Description,
		},
	})
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	var body struct {
		Title       *string `json:"title"`
		Status      *string `json:"status"`
		TimeSpend   *uint   `json:"time_spend"`
		Streak      *uint   `json:"streak"`
		Description *string `json:"description"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	var task models.Task
	if err := db.DB.Where("id = ? AND user_id = ?", id, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if body.Title != nil {
		task.Title = *body.Title
	}
	if body.Status != nil {
		task.Status = *body.Status
	}
	if body.TimeSpend != nil {
		task.TimeSpend = *body.TimeSpend
	}
	if body.Streak != nil {
		task.Streak = *body.Streak
	}
	if body.Description != nil {
		task.Description = *body.Description
	}

	if err := db.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}
