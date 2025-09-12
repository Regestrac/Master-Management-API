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
	ID        uint       `json:"id" gorm:"primaryKey"`
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
	taskType := c.Query("type")

	var tasks []models.Task

	// Start query
	query := db.DB.Where("user_id = ? AND parent_id IS NULL AND workspace_id IS NULL", userId)

	// Apply filtering
	if len(status) > 0 && !utils.Contains(status, "all") {
		query = query.Where("status IN ?", status)
	}
	if len(priority) > 0 && !utils.Contains(priority, "all") {
		query = query.Where("priority IN ?", priority)
	}
	if taskType == "task" || taskType == "goal" {
		query = query.Where("type = ?", taskType)
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
		Title       string `json:"title"`
		Status      string `json:"status"`
		TimeSpend   uint   `json:"time_spend"`
		Streak      uint   `json:"streak"`
		ParentId    *uint  `json:"parent_id"` // Optional parent ID for subtasks
		Type        string `json:"type"`
		WorkspaceId *uint  `json:"workspace_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body!"})
		return
	}

	if body.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required!"})
		return
	}

	if body.Type != "task" && body.Type != "goal" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be either 'task' or 'goal'!"})
		return
	}

	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	task := models.Task{
		UserId:      userId,
		Title:       body.Title,
		Status:      body.Status,
		TimeSpend:   body.TimeSpend,
		Streak:      body.Streak,
		ParentId:    body.ParentId, // Set parent ID if provided
		Type:        body.Type,
		WorkspaceId: body.WorkspaceId,
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
		if task.ParentId != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sub task"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	if task.ParentId != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Sub Task deleted successfully"})
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
			"tags":        task.Tags,
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
		Title       *string   `json:"title"`
		Status      *string   `json:"status"`
		TimeSpend   *uint     `json:"time_spend"`
		Streak      *uint     `json:"streak"`
		Description *string   `json:"description"`
		StartedAt   *string   `json:"started_at"` // Accept time as string or empty
		Priority    *string   `json:"priority"`
		Tags        *[]string `json:"tags"`
		Assignees   *[]uint   `json:"assignees"`
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
	if body.Tags != nil {
		task.Tags = body.Tags
	}
	if body.Assignees != nil {
		task.Assignees = body.Assignees
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
	if err := db.DB.Where("user_id = ? AND parent_id IS NULL AND last_accessed_at IS NOT NULL AND workspace_id IS NULL AND type = 'task'", userId).Order("last_accessed_at DESC").Limit(5).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recent tasks"})
		return
	}

	type RecentTaskResponse struct {
		ID             uint       `json:"id" gorm:"primaryKey"`
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
		Completed  int64 `json:"completed"`
		Todo       int64 `json:"todo"`
		InProgress int64 `json:"in_progress"`
		Pending    int64 `json:"pending"`
		Paused     int64 `json:"paused"`
		OverDue    int64 `json:"overdue"`
	}

	var taskStats TaskStats

	commonQuery := "type = 'task' AND parent_id IS NULL AND deleted_at IS NULL"

	err := db.DB.Model(&models.Task{}).Raw(fmt.Sprintf(`
		SELECT
			COUNT (*) FILTER (WHERE %s) AS total,
			COUNT (*) FILTER (WHERE status = 'completed' AND %s) as completed,
			COUNT (*) FILTER (WHERE status = 'todo' AND %s) as todo,
			COUNT (*) FILTER (WHERE status = 'inprogress' AND %s) as in_progress,
			COUNT (*) FILTER (WHERE status = 'pending' AND %s) as pending,
			COUNT (*) FILTER (WHERE status = 'paused' AND %s) as paused,
			COUNT (*) FILTER (WHERE due_date IS NOT NULL AND due_date < NOW() AND status != 'completed' AND %s) AS overdue
		FROM tasks
		WHERE user_id = ?
	`, commonQuery, commonQuery, commonQuery, commonQuery, commonQuery, commonQuery, commonQuery), userId).Scan(&taskStats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task stats!"})
		return
	}

	data := []map[string]interface{}{
		{"status": "total", "count": taskStats.Total},
		{"status": "completed", "count": taskStats.Completed},
		{"status": "todo", "count": taskStats.Todo},
		{"status": "in_progress", "count": taskStats.InProgress},
		{"status": "pending", "count": taskStats.Pending},
		{"status": "paused", "count": taskStats.Paused},
		{"status": "overdue", "count": taskStats.OverDue},
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetGoalStats(c *gin.Context) {
	userRawData, _ := c.Get("user")
	userId := userRawData.(models.User).ID

	type GoalStats struct {
		Total           uint  `json:"total"`
		Active          uint  `json:"active"`
		Completed       uint  `json:"completed"`
		Paused          uint  `json:"paused"`
		HighPriority    uint  `json:"high_priority"`
		AverageProgress *uint `json:"average_progress"`
	}

	var goalStats GoalStats

	commonQuery := "type = 'goal' AND parent_id IS NULL AND deleted_at IS NULL"

	err := db.DB.Model(&models.Task{}).Raw(fmt.Sprintf(`
		SELECT
			COUNT (*) FILTER (WHERE %s) AS total,
			COUNT (*) FILTER (WHERE status = 'inprogress' AND %s) as active,
			COUNT (*) FILTER (WHERE status = 'completed' AND %s) as completed,
			COUNT (*) FILTER (WHERE status = 'paused' AND %s) as paused,
			COUNT (*) FILTER (WHERE priority = 'high' AND %s) as high_priority
		FROM tasks
		WHERE user_id = ?
	`, commonQuery, commonQuery, commonQuery, commonQuery, commonQuery), userId).Scan(&goalStats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve goal stats!"})
		return
	}

	var averageProgress float64 = 0
	if goalStats.Total > 0 {
		averageProgress = (float64(goalStats.Completed) / float64(goalStats.Total)) * 100
	}

	data := []map[string]interface{}{
		{"status": "total", "count": goalStats.Total},
		{"status": "active", "count": goalStats.Active},
		{"status": "completed", "count": goalStats.Completed},
		{"status": "paused", "count": goalStats.Paused},
		{"status": "high_priority", "count": goalStats.HighPriority},
		{"status": "average_progress", "count": averageProgress},
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func GetActiveGoals(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var goals []models.Task
	if err := db.DB.Where("user_id = ? AND parent_id IS NULL AND status = 'inprogress' AND workspace_id IS NULL AND type = 'goal'", userId).Find(&goals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active goals"})
		return
	}

	type ActiveGoalsResponse struct {
		ID             uint       `json:"id" gorm:"primaryKey"`
		Title          string     `json:"title"`
		Status         string     `json:"status"`
		TimeSpend      uint       `json:"time_spend"`
		Type           string     `json:"type"`
		Streak         uint       `json:"streak"`
		LastAccessedAt *time.Time `json:"last_accessed_at"`
	}

	var data []ActiveGoalsResponse
	for _, goal := range goals {
		streak := UpdateStreak(&goal, false)
		data = append(data, ActiveGoalsResponse{
			ID:             goal.ID,
			Streak:         streak,
			Type:           goal.Type,
			Title:          goal.Title,
			Status:         goal.Status,
			TimeSpend:      goal.TimeSpend,
			LastAccessedAt: goal.LastAccessedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
