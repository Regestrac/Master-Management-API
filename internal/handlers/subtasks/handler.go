package subtasks

import (
	"master-management-api/internal/db"
	"master-management-api/internal/handlers/history"
	"master-management-api/internal/models"
	"master-management-api/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllSubtasks(c *gin.Context) {
	type TaskResponseType struct {
		ID                 uint       `json:"id" gorm:"primaryKey"`
		Title              string     `json:"title"`
		Status             string     `json:"status"`
		TimeSpend          uint       `json:"time_spend"`
		Streak             uint       `json:"streak"`
		ParentId           *uint      `json:"parent_id"`
		Progress           *float64   `json:"progress"`
		ChecklistCompleted *int64     `json:"checklist_completed"`
		ChecklistTotal     *int64     `json:"checklist_total"`
		DueDate            *time.Time `json:"due_date"`
	}

	taskId := c.Param("id")

	var subtasks []models.Task

	if err := db.DB.Where("parent_id = ?", taskId).Order("title ASC").Find(&subtasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subtasks"})
		return
	}

	var data []TaskResponseType
	for _, task := range subtasks {
		var result struct {
			TotalCount     int64
			CompletedCount int64
		}

		// Get checklist counts
		// if err := db.DB.Model(&models.Checklist{}).
		// 	Where("task_id = ?", task.ID).
		// 	Count(&totalCount).Error; err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count checklists"})
		// 	return
		// }

		// if err := db.DB.Model(&models.Checklist{}).
		// 	Where("task_id = ? AND completed = ?", task.ID, true).
		// 	Count(&completedCount).Error; err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count completed checklists"})
		// 	return
		// }
		if err := db.DB.Model(&models.Checklist{}).
			Select("COUNT(*) AS total_count, SUM(CASE WHEN completed = ? THEN 1 ELSE 0 END) AS completed_count", true).
			Where("task_id = ?", task.ID).
			Scan(&result).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count checklists"})
			return
		}

		data = append(data, TaskResponseType{
			ID:                 task.ID,
			Title:              task.Title,
			Status:             task.Status,
			TimeSpend:          task.TimeSpend,
			Streak:             task.Streak,
			ParentId:           task.ParentId,
			Progress:           task.Progress,
			ChecklistCompleted: &result.CompletedCount,
			ChecklistTotal:     &result.TotalCount,
			DueDate:            task.DueDate,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func SaveSubtasks(c *gin.Context) {
	var body []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		TimeSpend   uint   `json:"time_spend"`
		Streak      uint   `json:"streak"`
		ParentId    *uint  `json:"parent_id"`
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	parentId := uint(id)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	var subtasks []models.Task
	for _, item := range body {
		subtasks = append(subtasks, models.Task{
			Title:       item.Title,
			Description: item.Description,
			Status:      item.Status,
			TimeSpend:   item.TimeSpend,
			Streak:      item.Streak,
			UserId:      userId,
			ParentId:    &parentId,
		})
	}

	if err := db.DB.Create(&subtasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subtasks"})
		return
	}

	for _, task := range subtasks {
		history.LogHistory("created", "", task.Title, task.ID, userId)
	}

	var task models.Task
	var taskProgress float64
	if body[0].ParentId != nil {
		db.DB.Where("id = ?", body[0].ParentId).First(&task)
		progress, err := utils.RecalculateProgress(task.ID)
		taskProgress = progress
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate progress!"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     subtasks,
		"progress": taskProgress,
	})
}
