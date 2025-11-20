package profile

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Regestrac/master-management-api/internal/db"
	"github.com/Regestrac/master-management-api/internal/handlers/history"
	"github.com/Regestrac/master-management-api/internal/handlers/task"
	"github.com/Regestrac/master-management-api/internal/models"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email"`
	Theme      string  `json:"theme"`
	ActiveTask *uint   `json:"active_task"`
	AvatarUrl  *string `json:"avatar_url"`
	TimeZone   *string `json:"time_zone"`
	Language   string  `json:"language"`
	Bio        string  `json:"bio"`
	Company    *string `json:"company"`
	JobTitle   *string `json:"job_title"`
}

func StartTaskSession(userID uint, taskID uint) error {
	var activeSession models.TaskSession
	if err := db.DB.Where("end_time IS NULL").First(&activeSession).Error; err == nil {
		now := time.Now()
		activeSession.EndTime = &now
		activeSession.Duration = int64(now.Sub(activeSession.StartTime).Seconds())
		if err := db.DB.Save(&activeSession).Error; err != nil {
			return err
		}
	}

	session := models.TaskSession{
		TaskID:    taskID,
		UserID:    userID,
		StartTime: time.Now(),
	}
	return db.DB.Create(&session).Error
}

func StopTaskSession() error {
	var session models.TaskSession
	if err := db.DB.Where("end_time IS NULL").First(&session).Error; err != nil {
		return err
	}

	now := time.Now()
	session.EndTime = &now
	session.Duration = int64(now.Sub(session.StartTime).Seconds())

	// update task.total_time_spend also
	// db.DB.Model(&models.Task{}).
	// 	Where("id = ?", session.TaskID).
	// 	Update("time_spend", gorm.Expr("time_spend + ?", session.Duration))

	return db.DB.Save(&session).Error
}

func GetProfile(c *gin.Context) {
	userDataRaw, exists := c.Get("user")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userData, ok := userDataRaw.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
		return
	}

	data := UserResponse{
		ID:         userData.ID,
		FirstName:  userData.FirstName,
		LastName:   userData.LastName,
		Email:      userData.Email,
		Theme:      userData.Theme,
		ActiveTask: userData.ActiveTask,
		AvatarUrl:  userData.AvatarUrl,
		TimeZone:   userData.TimeZone,
		Company:    userData.Company,
		Language:   userData.Language,
		Bio:        userData.Bio,
		JobTitle:   userData.JobTitle,
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func UpdateProfile(c *gin.Context) {
	type UpdateProfileInput struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		TimeZone  string `json:"time_zone"`
		Language  string `json:"language"`
		Bio       string `json:"bio"`
		Company   string `json:"company"`
		JobTitle  string `json:"job_title"`
	}

	userDataRaw, exists := c.Get("user")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userData, ok := userDataRaw.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
		return
	}

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	updates := map[string]interface{}{}
	if input.FirstName != "" {
		updates["first_name"] = input.FirstName
	}
	if input.LastName != "" {
		updates["last_name"] = input.LastName
	}
	if input.Email != "" {
		updates["email"] = input.Email
	}
	if input.TimeZone != "" {
		updates["time_zone"] = input.TimeZone
	}
	if input.Company != "" {
		updates["company"] = input.Company
	}
	if input.Bio != "" {
		updates["bio"] = input.Bio
	}
	if input.JobTitle != "" {
		updates["job_title"] = input.JobTitle
	}
	if input.Language != "" {
		updates["language"] = input.Language
	}

	if err := db.DB.Model(&userData).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	data := UserResponse{
		ID:         userData.ID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
		Theme:      userData.Theme,
		ActiveTask: userData.ActiveTask,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    data,
	})
}

func updateStartedAt(taskId uint, userId uint, c *gin.Context) {
	// Updates the log and started
	currentTime := time.Now()
	var currTask models.Task
	if err := db.DB.Where("id = ? AND user_id = ?", taskId, userId).First(&currTask).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found!"})
		return
	}
	if currTask.StartedAt == nil {
		currTask.StartedAt = &currentTime
		history.LogHistory("started", "", "", currTask.ID, userId)
		task.UpdateStreak(&currTask, true)
	} else {
		lastSessionTime := uint(currentTime.Sub(*currTask.StartedAt).Seconds())
		totalTimeSpend := currTask.TimeSpend + lastSessionTime

		history.LogHistory("stopped", strconv.FormatUint(uint64(currTask.TimeSpend), 10), strconv.FormatUint(uint64(totalTimeSpend), 10), currTask.ID, userId)

		currTask.StartedAt = nil
		currTask.TimeSpend = totalTimeSpend
		task.UpdateStreak(&currTask, false)
	}
}

func UpdateActiveTask(c *gin.Context) {
	var body struct {
		ActiveTask *uint `json:"active_task"`
	}

	userDataRaw, _ := c.Get("user")
	user, ok := userDataRaw.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if body.ActiveTask != nil {
		updateStartedAt(*body.ActiveTask, user.ID, c)
		user.ActiveTask = body.ActiveTask
		if err := StartTaskSession(user.ID, *body.ActiveTask); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log start session! " + err.Error()})
			return
		}
	} else {
		updateStartedAt(*user.ActiveTask, user.ID, c)
		user.ActiveTask = nil
		if err := StopTaskSession(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log stop session! " + err.Error()})
			return
		}
	}

	if err := db.DB.Save(user).Error; err != nil {
		var errorMessage string
		if body.ActiveTask != nil {
			errorMessage = "Failed to start task!"
		} else {
			errorMessage = "Failed to stop task!"
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage})
		return
	}

	var successMessage string
	if body.ActiveTask != nil {
		successMessage = "Successfully started the task"
	} else {
		successMessage = "Successfully stopped the task"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     successMessage,
		"active_task": user.ActiveTask,
	})
}

func GetQuickStats(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var stats struct {
		TotalTasks     int64 `json:"total_tasks"`
		TotalTimeSpend int64 `json:"total_time_spend"`
		CompletedToday int64 `json:"completed_today"`
		CurrentStreak  int64 `json:"current_streak"`
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	if err := db.DB.Model(&models.Task{}).
		Select(`
			COUNT(CASE WHEN type = 'task' AND parent_id IS NULL AND workspace_id IS NULL THEN 1 END) as total_tasks,
			COALESCE(SUM(time_spend), 0) as total_time_spend,
			COALESCE(MAX(streak), 0) as current_streak,
			COUNT(CASE WHEN status = 'completed' AND updated_at >= ? AND updated_at < ? THEN 1 END) as completed_today
		`, today, tomorrow).
		Where("user_id = ?", userId).
		Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func GetMonthlyStats(c *gin.Context) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	today := now.UTC().AddDate(0, 0, 1).Truncate(24 * time.Hour)

	prevMonthStart := startOfMonth.AddDate(0, -1, 0)
	prevMonthEnd := startOfMonth.Truncate(24 * time.Hour)

	var stats struct {
		TasksCompleted     int64   `json:"tasks_completed"`
		TaskDifference     int64   `json:"task_difference"`
		GoalsAchieved      int64   `json:"goals_achieved"`
		GoalDifference     int64   `json:"goal_difference"`
		FocusDuration      int64   `json:"focus_duration"`
		DurationDifference int64   `json:"duration_difference"`
		ProductivityScore  float64 `json:"productivity_score"`
		ScoreDifference    float64 `json:"score_difference"`
	}

	var prevStats struct {
		TasksCompleted    int64   `json:"tasks_completed"`
		GoalsAchieved     int64   `json:"goals_achieved"`
		FocusDuration     int64   `json:"focus_duration"`
		ProductivityScore float64 `json:"productivity_score"`
	}

	if err := db.DB.Model(models.Task{}).Select(`
		COUNT(CASE WHEN type = 'task' THEN 1 END) as tasks_completed,
		COUNT(CASE WHEN type = 'goal' THEN 1 END) as goals_achieved
	`).Where("status = 'completed' AND completed_at BETWEEN ? AND ?", startOfMonth, today).
		Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve completed tasks and goals count!"})
		return
	}
	if err := db.DB.Model(models.Task{}).Select(`
		COUNT(CASE WHEN type = 'task' THEN 1 END) as tasks_completed,
		COUNT(CASE WHEN type = 'goal' THEN 1 END) as goals_achieved
	`).Where("status = 'completed' AND completed_at BETWEEN ? AND ?", prevMonthStart, prevMonthEnd).
		Scan(&prevStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve completed tasks and goals count!"})
		return
	}
	stats.TaskDifference = stats.TasksCompleted - prevStats.TasksCompleted
	stats.GoalDifference = stats.GoalsAchieved - prevStats.GoalsAchieved

	if err := db.DB.Model(models.TaskSession{}).
		Select("COALESCE(SUM(duration), 0)").
		Where("start_time BETWEEN ? AND ?", startOfMonth, today).
		Scan(&stats.FocusDuration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve focus duration!"})
		return
	}
	if err := db.DB.Model(models.TaskSession{}).
		Select("COALESCE(SUM(duration), 0)").
		Where("start_time BETWEEN ? AND ?", prevMonthStart, prevMonthEnd).
		Scan(&prevStats.FocusDuration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve focus duration!"})
		return
	}
	stats.DurationDifference = stats.FocusDuration - prevStats.FocusDuration

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
