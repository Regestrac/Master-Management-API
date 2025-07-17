package task

import (
	"fmt"
	"master-management-api/internal/db"
	"master-management-api/internal/handlers/history"
	"master-management-api/internal/models"
	"master-management-api/internal/utils"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskResponseType struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	TimeSpend uint       `json:"time_spend"`
	Streak    uint       `json:"streak"`
	Type      string     `json:"type"`
	Priority  *string    `json:"priority"`
	DueDate   *time.Time `json:"due_date"`
	Category  string     `json:"category"`
}

func UpdateStreak(task *models.Task, saveStartTime bool) uint {
	currentTime := time.Now()

	if task.LastStartedAt != nil {
		yesterday := currentTime.AddDate(0, 0, -1).Truncate(24 * time.Hour)
		lastCompleted := task.LastStartedAt.Truncate(24 * time.Hour)
		if lastCompleted.Equal(yesterday) {
			if saveStartTime {
				task.Streak += 1
			}
		} else if lastCompleted.Before(yesterday) {
			task.Streak = 0
			if saveStartTime {
				task.Streak = 1
			}
		}
	}

	if saveStartTime {
		if task.Streak < 1 {
			task.Streak = 1
		}
		task.LastStartedAt = &currentTime
	}

	db.DB.Save(task)

	return task.Streak
}

func GetAllTasks(c *gin.Context) {
	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	// Get filters and sort
	status := c.QueryArray("status")
	priority := c.QueryArray("priority")
	sortBy := c.DefaultQuery("sortBy", "created_at")
	order := c.DefaultQuery("order", "asc")
	searchKey := c.Query("searchKey")

	var tasks []models.Task

	// Start query
	query := db.DB.Where("user_id = ? AND parent_id IS NULL", userId)

	// Apply filtering
	if len(status) > 0 && !utils.Contains(status, "all") {
		query = query.Where("status IN ?", status)
	}
	if len(priority) > 0 && !utils.Contains(priority, "all") {
		query = query.Where("priority IN ?", priority)
	}

	if searchKey != "" {
		likeQuery := "%" + searchKey + "%"
		query = query.Where(
			db.DB.Where("LOWER(title) LIKE LOWER(?)", likeQuery), // Add if want to include description in search -> .Or("LOWER(description) LIKE LOWER(?)", likeQuery)
		)
	}

	// Sanitize sorting
	validSorts := map[string]bool{
		"priority":   true,
		"status":     true,
		"due_date":   true,
		"created_at": true,
	}
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Apply sorting
	switch sortBy {
	case "priority":
		query = query.Order(
			fmt.Sprintf(
				`CASE
					WHEN priority = 'high' THEN 1
					WHEN priority = 'normal' THEN 2
					WHEN priority = 'low' THEN 3
					ELSE 4
				END %s`, order,
			),
		)
	case "status":
		query = query.Order(
			fmt.Sprintf(
				`CASE
					WHEN status = 'todo' THEN 1
					WHEN status = 'inprogress' THEN 2
					WHEN status = 'pending' THEN 3
					WHEN status = 'paused' THEN 4
					WHEN status = 'complete' THEN 5
					ELSE 6
				END %s`, order,
			),
		)
	default:
		query = query.Order(fmt.Sprintf("%s %s", sortBy, order))
	}

	// Run query
	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	if searchKey != "" {
		searchKeyLower := strings.ToLower(searchKey)
		sort.SliceStable(tasks, func(i, j int) bool {
			a := strings.ToLower(tasks[i].Title)
			b := strings.ToLower(tasks[j].Title)
			aScore := utils.MatchScore(a, searchKeyLower, strings.ToLower(tasks[i].Description), false) // use true to include description in search
			bScore := utils.MatchScore(b, searchKeyLower, strings.ToLower(tasks[j].Description), false)
			return aScore > bScore // higher score first
		})
	}

	// Map to response
	data := make([]TaskResponseType, 0, len(tasks))
	for _, task := range tasks {
		streak := UpdateStreak(&task, false)
		data = append(data, TaskResponseType{
			ID:        task.ID,
			Title:     task.Title,
			Status:    task.Status,
			TimeSpend: task.TimeSpend,
			Streak:    streak,
			Type:      task.Type,
			Priority:  task.Priority,
			DueDate:   task.DueDate,
			Category:  task.Category,
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
		Type      string `json:"type"`
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
		Type:      body.Type,
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
		"data": TaskResponseType{
			ID:        task.ID,
			Title:     task.Title,
			Status:    task.Status,
			TimeSpend: task.TimeSpend,
			Streak:    task.Streak,
			Type:      task.Type,
			Priority:  task.Priority,
			DueDate:   task.DueDate,
			Category:  task.Category,
		},
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

	// Fetch notes related to this task
	var notes []models.Note
	if err := db.DB.Where("task_id = ? AND user_id = ?", task.ID, userId).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
		return
	}

	// Fetch checklist related to this task
	var checklists []models.Checklist
	if err := db.DB.Where("task_id = ? AND user_id = ?", task.ID, userId).Find(&checklists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve checklist"})
		return
	}

	currentTime := time.Now()
	task.LastAccessedAt = &currentTime
	db.DB.Save(&task) // Update last accessed time

	UpdateStreak(&task, false)

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
			"type":        task.Type,
			"priority":    task.Priority,
			"created_at":  task.CreatedAt,
			"notes":       notes,
			"checklists":  checklists,
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
		Priority    *string `json:"priority"`
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
	if body.Priority != nil {
		if task.Priority != nil && *task.Priority != "" {
			history.LogHistory("priority_change", *task.Priority, *body.Priority, task.ID, userId)
		} else {
			history.LogHistory("priority_change", "", *body.Priority, task.ID, userId)
		}
		task.Priority = body.Priority
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

			UpdateStreak(&task, true)
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
		Streak         uint       `json:"streak"`
		Priority       *string    `json:"priority"`
		Type           string     `json:"type"`
	}

	var data []RecentTaskResponse
	for _, task := range tasks {
		streak := UpdateStreak(&task, false)
		data = append(data, RecentTaskResponse{
			ID:             task.ID,
			Title:          task.Title,
			Status:         task.Status,
			TimeSpend:      task.TimeSpend,
			LastAccessedAt: task.LastAccessedAt,
			Streak:         streak,
			Priority:       task.Priority,
			Type:           task.Type,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetTaskStats(c *gin.Context) {
	userRawData, _ := c.Get("user")
	userId := userRawData.(models.User).ID

	type TaskStats struct {
		Total      int64 `json:"total"`
		ToDo       int64 `json:"todo"`
		InProgress int64 `json:"in_progress"`
		Pending    int64 `json:"pending"`
		Paused     int64 `json:"paused"`
		Completed  int64 `json:"completed"`
	}

	var stats TaskStats

	db.DB.Model(&models.Task{}).Where("user_id = ? AND parent_id IS NULL", userId).Count(&stats.Total)
	db.DB.Model(&models.Task{}).Where("user_id = ? AND parent_id IS NULL AND status = ?", userId, "todo").Count(&stats.ToDo)
	db.DB.Model(&models.Task{}).Where("user_id = ? AND parent_id IS NULL AND status = ?", userId, "inprogress").Count(&stats.InProgress)
	db.DB.Model(&models.Task{}).Where("user_id = ? AND parent_id IS NULL AND status = ?", userId, "pending").Count(&stats.Pending)
	db.DB.Model(&models.Task{}).Where("user_id = ? AND parent_id IS NULL AND status = ?", userId, "paused").Count(&stats.Paused)
	db.DB.Model(&models.Task{}).Where("user_id = ? AND parent_id IS NULL AND status = ?", userId, "completed").Count(&stats.Completed)

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
