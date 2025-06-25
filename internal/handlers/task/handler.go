package task

import (
	"master-management-api/internal/db"
	"master-management-api/internal/handlers/history"
	"master-management-api/internal/models"
	"net/http"
	"strconv"
	"time"

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

	if err := db.DB.Where("user_id = ? AND parent_id IS NULL", userId).Find(&tasks).Error; err != nil {
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
		ParentId  *uint  `json:"parent_id"` // Optional parent ID for subtasks
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
		ParentId:  body.ParentId, // Set parent ID if provided
	}

	result := db.DB.Create(&task)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
		})
		return
	}

	if body.ParentId != nil {
		history.LogHistory("subtask", "", body.Title, *body.ParentId, userId)
	}

	history.LogHistory("created", "", body.Title, task.ID, userId)

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

	currentTime := time.Now()
	task.LastAccessedAt = &currentTime
	db.DB.Save(&task) // Update last accessed time

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"status":      task.Status,
			"time_spend":  task.TimeSpend,
			"streak":      task.Streak,
			"description": task.Description,
			"started_at":  task.StartedAt,
			"parent_id":   task.ParentId,
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
		StartedAt   *string `json:"started_at"` // Accept time as string or empty
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

	// Update fields
	if body.Title != nil {
		history.LogHistory("title_update", task.Title, *body.Title, task.ID, userId)
		task.Title = *body.Title
	}
	if body.Status != nil {
		history.LogHistory("status_update", task.Status, *body.Status, task.ID, userId)
		task.Status = *body.Status
	}
	if body.TimeSpend != nil {
		history.LogHistory("stopped", strconv.FormatUint(uint64(task.TimeSpend), 10), strconv.FormatUint(uint64(*body.TimeSpend), 10), task.ID, userId)
		task.TimeSpend = *body.TimeSpend
	}
	if body.Streak != nil {
		task.Streak = *body.Streak
	}
	if body.Description != nil {
		history.LogHistory("description_update", task.Description, *body.Description, task.ID, userId)
		task.Description = *body.Description
	}

	// Handle StartedAt
	if body.StartedAt != nil {
		if *body.StartedAt == "" {
			task.StartedAt = nil
		} else {
			parsedTime, err := time.Parse(time.RFC3339, *body.StartedAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid started_at format. Use ISO 8601"})
				return
			}
			history.LogHistory("started", "", "", task.ID, userId)
			task.StartedAt = &parsedTime

			currentTime := time.Now()
			task.LastStartedAt = &currentTime

			if task.LastStartedAt != nil {
				yesterday := currentTime.AddDate(0, 0, -1).Truncate(24 * time.Hour)
				lastCompleted := task.LastStartedAt.Truncate(24 * time.Hour)
				if lastCompleted.Equal(yesterday) {
					task.Streak += 1
				} else if lastCompleted.Before(yesterday) {
					task.Streak = 1 // reset
				} else {
					if task.Streak == 0 {
						task.Streak = 1
					}
				}
			} else {
				task.Streak = 1
			}

			db.DB.Save(&task)
		}
	}

	if err := db.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func GetRecentTasks(c *gin.Context) {
	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	var tasks []models.Task
	if err := db.DB.Where("user_id = ? AND parent_id IS NULL AND last_accessed_at IS NOT NULL", userId).Order("last_accessed_at DESC").Limit(5).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recent tasks"})
		return
	}

	type RecentTaskResponse struct {
		ID             uint       `json:"id"`
		Title          string     `json:"title"`
		Status         string     `json:"status"`
		TimeSpend      uint       `json:"time_spend"`
		LastAccessedAt *time.Time `json:"last_accessed_at"`
	}

	var data []RecentTaskResponse
	for _, task := range tasks {
		data = append(data, RecentTaskResponse{
			ID:             task.ID,
			Title:          task.Title,
			Status:         task.Status,
			TimeSpend:      task.TimeSpend,
			LastAccessedAt: task.LastAccessedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
