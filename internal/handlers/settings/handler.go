package settings

import (
	"database/sql"
	"net/http"

	"github.com/Regestrac/Master-Management-API/internal/db"
	"github.com/Regestrac/Master-Management-API/internal/models"

	"github.com/gin-gonic/gin"
)

func CreateUserSettings(userId uint) error {
	settings := models.UserSettings{
		UserId:            userId,
		DateFormat:        "MM/DD/YYYY",
		TimeFormat:        "12",
		FirstDayOfWeek:    "sunday",
		WorkWeek:          "5",
		Theme:             "dark",
		AccentColor:       "Purple",
		FocusDuration:     25,
		ShortBreak:        5,
		LongBreak:         20,
		AutoBreak:         true,
		LongBreakAfter:    4,
		GoalDuration:      30,
		WeeklyTargetHours: 5,
	}

	if err := db.DB.Create(&settings).Error; err != nil {
		return err
	}

	return nil
}

func GetUserSettings(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var settings models.UserSettings
	if err := db.DB.Where("user_id = ?", userId).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user settings!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func UpdateUserSettings(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var body struct {
		UserId                *uint   `json:"user_id"`
		DateFormat            *string `json:"date_format"`
		TimeFormat            *string `json:"time_format"`
		FirstDayOfWeek        *string `json:"first_day_of_week"`
		WorkWeek              *string `json:"work_week"`
		Theme                 *string `json:"theme"`
		AccentColor           *string `json:"accent_color"`
		FocusDuration         *uint   `json:"focus_duration"`
		ShortBreak            *uint   `json:"short_break"`
		LongBreak             *uint   `json:"long_break"`
		AutoBreak             *bool   `json:"auto_break"`
		LongBreakAfter        *uint   `json:"long_break_after"`
		GoalDuration          *uint   `json:"goal_duration"`
		WeeklyTargetHours     *uint   `json:"weekly_target_hours"`
		TaskReminder          *bool   `json:"task_reminder"`
		GoalProgress          *bool   `json:"goal_progress"`
		SessionBreaks         *bool   `json:"session_breaks"`
		DailySummary          *bool   `json:"daily_summary"`
		Milestone             *bool   `json:"milestone"`
		NewFeature            *bool   `json:"new_feature"`
		CloudSync             *bool   `json:"cloud_sync"`
		KeepCompletedFor      *string `json:"keep_completed_for"`
		AnalyticDataRetention *string `json:"analytic_data_retention"`
		AutoDeleteOldData     *bool   `json:"auto_delete_old_data"`
		DebugMode             *bool   `json:"debug_mode"`
		BetaFeatures          *bool   `json:"beta_features"`
		Telemetry             *bool   `json:"telemetry"`
		AIAssistant           *bool   `json:"ai_assistant"`
		AdvancedAnalytics     *bool   `json:"advanced_analytics"`
		TeamCollaboration     *bool   `json:"team_collaboration"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		return
	}

	var settings models.UserSettings
	if err := db.DB.Where("user_id = ?", userId).First(&settings).Error; err != nil {
		if err := CreateUserSettings(userId); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user settings!"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user settings!"})
		return
	}

	if body.DateFormat != nil {
		settings.DateFormat = *body.DateFormat
	}
	if body.TimeFormat != nil {
		settings.TimeFormat = *body.TimeFormat
	}
	if body.FirstDayOfWeek != nil {
		settings.FirstDayOfWeek = *body.FirstDayOfWeek
	}
	if body.WorkWeek != nil {
		settings.WorkWeek = *body.WorkWeek
	}
	if body.Theme != nil {
		settings.Theme = *body.Theme
	}
	if body.AccentColor != nil {
		settings.AccentColor = *body.AccentColor
	}
	if body.FocusDuration != nil {
		settings.FocusDuration = *body.FocusDuration
	}
	if body.ShortBreak != nil {
		settings.ShortBreak = *body.ShortBreak
	}
	if body.LongBreak != nil {
		settings.LongBreak = *body.LongBreak
	}
	if body.AutoBreak != nil {
		settings.AutoBreak = *body.AutoBreak
	}
	if body.LongBreakAfter != nil {
		settings.LongBreakAfter = *body.LongBreakAfter
	}
	if body.GoalDuration != nil {
		settings.GoalDuration = *body.GoalDuration
	}
	if body.WeeklyTargetHours != nil {
		settings.WeeklyTargetHours = *body.WeeklyTargetHours
	}
	if body.TaskReminder != nil {
		settings.TaskReminder = *body.TaskReminder
	}
	if body.GoalProgress != nil {
		settings.GoalProgress = *body.GoalProgress
	}
	if body.SessionBreaks != nil {
		settings.SessionBreaks = *body.SessionBreaks
	}
	if body.DailySummary != nil {
		settings.DailySummary = *body.DailySummary
	}
	if body.Milestone != nil {
		settings.Milestone = *body.Milestone
	}
	if body.NewFeature != nil {
		settings.NewFeature = *body.NewFeature
	}
	if body.CloudSync != nil {
		settings.CloudSync = *body.CloudSync
	}
	if body.KeepCompletedFor != nil {
		settings.KeepCompletedFor = *body.KeepCompletedFor
	}
	if body.AnalyticDataRetention != nil {
		settings.AnalyticDataRetention = *body.AnalyticDataRetention
	}
	if body.AutoDeleteOldData != nil {
		settings.AutoDeleteOldData = *body.AutoDeleteOldData
	}
	if body.DebugMode != nil {
		settings.DebugMode = *body.DebugMode
	}
	if body.BetaFeatures != nil {
		settings.BetaFeatures = *body.BetaFeatures
	}
	if body.Telemetry != nil {
		settings.Telemetry = *body.Telemetry
	}
	if body.AIAssistant != nil {
		settings.AIAssistant = *body.AIAssistant
	}
	if body.AdvancedAnalytics != nil {
		settings.AdvancedAnalytics = *body.AdvancedAnalytics
	}
	if body.TeamCollaboration != nil {
		settings.TeamCollaboration = *body.TeamCollaboration
	}

	body.UserId = &userId

	if err := db.DB.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user settings!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func ResetSettings(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var settings models.UserSettings
	if err := db.DB.Where("user_id = ?", userId).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user settings!"})
		return
	}

	settings.DateFormat = "DD/MM/YYYY"
	settings.TimeFormat = "12"
	settings.FirstDayOfWeek = "sunday"
	settings.WorkWeek = "5"
	settings.Theme = "dark"
	settings.AccentColor = "Purple"
	settings.FocusDuration = 25
	settings.ShortBreak = 5
	settings.LongBreak = 20
	settings.AutoBreak = false
	settings.LongBreakAfter = 4
	settings.GoalDuration = 30
	settings.WeeklyTargetHours = 5

	if err := db.DB.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user settings!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func UpdateTheme(c *gin.Context) {
	var body struct {
		Theme string `json:"theme"`
	}

	userDataRaw, _ := c.Get("user")
	userId := userDataRaw.(models.User).ID

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var settings models.UserSettings
	if err := db.DB.Where("user_id = ?", userId).First(&settings).Error; err != nil {
		if err := CreateUserSettings(userId); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user settings!"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user settings!"})
		return
	}

	if body.Theme != "" {
		settings.Theme = body.Theme
	}

	if err := db.DB.Save(settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theme!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Theme updated successfully.",
		"theme":   settings.Theme,
	})
}

func GetUserStorageUsage(c *gin.Context) {
	userDataRaw, _ := c.Get("user")
	userID := userDataRaw.(models.User).ID

	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var result struct {
		TotalBytes      int64 `json:"total_bytes"`
		PersonalTasks   int64 `json:"tasks_bytes"`
		PersonalGoals   int64 `json:"goals_bytes"`
		WorkspacesBytes int64 `json:"workspaces_bytes"`
	}

	// helper
	calcSize := func(query string, args ...interface{}) int64 {
		var size sql.NullInt64
		if err := db.DB.Raw(query, args...).Scan(&size).Error; err != nil {
			return 0
		}
		if size.Valid {
			return size.Int64
		}
		return 0
	}

	// 1. Personal tasks/goals (not inside a workspace)
	result.PersonalTasks = calcSize(`
    SELECT 
			(SELECT COALESCE(SUM(pg_column_size(t)),0) FROM tasks t WHERE t.user_id = ? AND t.type = 'task' AND t.workspace_id IS NULL) +
			(SELECT COALESCE(SUM(pg_column_size(n)),0) FROM notes n JOIN tasks t ON n.task_id = t.id WHERE t.user_id = ? AND t.type = 'task' AND t.workspace_id IS NULL) +
			(SELECT COALESCE(SUM(pg_column_size(c)),0) FROM checklists c JOIN tasks t ON c.task_id = t.id WHERE t.user_id = ? AND t.type = 'task' AND t.workspace_id IS NULL)
	`, userID, userID, userID)

	result.PersonalGoals = calcSize(`
    SELECT 
			(SELECT COALESCE(SUM(pg_column_size(t)),0) FROM tasks t WHERE t.user_id = ? AND t.type = 'goal' AND t.workspace_id IS NULL) +
			(SELECT COALESCE(SUM(pg_column_size(n)),0) FROM notes n JOIN tasks t ON n.task_id = t.id WHERE t.user_id = ? AND t.type = 'goal' AND t.workspace_id IS NULL) +
			(SELECT COALESCE(SUM(pg_column_size(c)),0) FROM checklists c JOIN tasks t ON c.task_id = t.id WHERE t.user_id = ? AND t.type = 'goal' AND t.workspace_id IS NULL)
	`, userID, userID, userID)

	// 2. Workspace-related data
	result.WorkspacesBytes = calcSize(`
    WITH ws AS (
        SELECT id FROM workspaces WHERE owner_id = ?
        UNION
        SELECT workspace_id FROM members WHERE user_id = ?
    )
    SELECT 
        COALESCE(SUM(pg_column_size(w)),0) + 
        (SELECT COALESCE(SUM(pg_column_size(t)),0) FROM tasks t WHERE t.workspace_id IN (SELECT id FROM ws)) +
        (SELECT COALESCE(SUM(pg_column_size(n)),0) FROM notes n JOIN tasks t ON n.task_id = t.id WHERE t.workspace_id IN (SELECT id FROM ws)) +
        (SELECT COALESCE(SUM(pg_column_size(c)),0) FROM checklists c JOIN tasks t ON c.task_id = t.id WHERE t.workspace_id IN (SELECT id FROM ws)) +
        (SELECT COALESCE(SUM(pg_column_size(m)),0) FROM members m WHERE m.workspace_id IN (SELECT id FROM ws))
    FROM workspaces w
    WHERE w.id IN (SELECT id FROM ws)
	`, userID, userID)

	// 3. Sessions, history, settings
	SessionsBytes := calcSize(`SELECT COALESCE(SUM(pg_column_size(s)),0) FROM task_sessions s WHERE s.user_id = ?`, userID)
	HistoryBytes := calcSize(`SELECT COALESCE(SUM(pg_column_size(h)),0) FROM task_histories h WHERE h.user_id = ?`, userID)
	SettingsBytes := calcSize(`SELECT COALESCE(SUM(pg_column_size(us)),0) FROM user_settings us WHERE us.user_id = ?`, userID)

	// 4. Total
	result.TotalBytes = result.PersonalTasks + result.PersonalGoals + result.WorkspacesBytes +
		SessionsBytes + HistoryBytes + SettingsBytes

	c.JSON(http.StatusOK, result)
}
