package analytics

import (
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetQuickMetrics(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	prevStartDate := c.Query("prevStartDate")
	prevEndDate := c.Query("prevEndDate")

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var totalFocusTime int64
	var focusTimeChange *int64 = nil
	var tasksCompleted int64
	var tasksCompletedChange *int64 = nil

	if err := db.DB.Model(&models.TaskSession{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userId, startDate, endDate).
		Select("COALESCE(SUM(duration),0)").
		Scan(&totalFocusTime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get focus time!"})
		return
	}
	if prevStartDate != "" && prevEndDate != "" {
		var prevFocusTime int64
		if err := db.DB.Model(&models.TaskSession{}).
			Where("user_id = ? AND created_at BETWEEN ? AND ?", userId, prevStartDate, prevEndDate).
			Select("COALESCE(SUM(duration),0)").
			Scan(&prevFocusTime).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get focus time!"})
			return
		}
		change := totalFocusTime - prevFocusTime
		focusTimeChange = &change
	}

	if err := db.DB.Model(&models.Task{}).
		Select("COUNT(*)").
		Where("user_id = ? AND completed_at BETWEEN ? AND ?", userId, startDate, endDate).
		Scan(&tasksCompleted).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completed tasks!"})
		return
	}
	if prevStartDate != "" && prevEndDate != "" {
		var prevTasksCompleted int64
		if err := db.DB.Model(&models.Task{}).
			Select("COUNT(*)").
			Where("user_id = ? AND completed_at BETWEEN ? AND ?", userId, prevStartDate, prevEndDate).
			Scan(&prevTasksCompleted).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completed tasks!"})
			return
		}
		change := tasksCompleted - prevTasksCompleted
		tasksCompletedChange = &change
	}

	c.JSON(http.StatusOK, gin.H{
		"focus_time":             totalFocusTime,
		"focus_time_change":      focusTimeChange,
		"tasks_completed":        tasksCompleted,
		"tasks_completed_change": tasksCompletedChange,
	})
}
