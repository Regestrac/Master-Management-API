package settings

import (
	"database/sql"
	"fmt"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

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
		UserId            *uint   `json:"user_id"`
		DateFormat        *string `json:"date_format"`
		TimeFormat        *string `json:"time_format"`
		FirstDayOfWeek    *string `json:"first_day_of_week"`
		WorkWeek          *string `json:"work_week"`
		Theme             *string `json:"theme"`
		AccentColor       *string `json:"accent_color"`
		FocusDuration     *uint   `json:"focus_duration"`
		ShortBreak        *uint   `json:"short_break"`
		LongBreak         *uint   `json:"long_break"`
		AutoBreak         *bool   `json:"auto_break"`
		LongBreakAfter    *uint   `json:"long_break_after"`
		GoalDuration      *uint   `json:"goal_duration"`
		WeeklyTargetHours *uint   `json:"weekly_target_hours"`
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
	userId := userDataRaw.(models.User).ID

	if userId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var result struct {
		TotalBytes      int64 `json:"total_bytes"`
		TasksBytes      int64 `json:"tasks_bytes"`
		GoalsBytes      int64 `json:"goals_bytes"`
		WorkspacesBytes int64 `json:"workspaces_bytes"`
	}

	// Helper to run pg_column_size aggregation
	calcSize := func(table, condition string, args ...interface{}) int64 {
		var size sql.NullInt64
		query := fmt.Sprintf("SELECT COALESCE(SUM(pg_column_size(t)),0) FROM %s t WHERE %s", table, condition)
		if err := db.DB.Raw(query, args...).Scan(&size).Error; err != nil {
			return 0
		}
		if size.Valid {
			return size.Int64
		}
		return 0
	}

	// Calculate sizes
	notesSize := calcSize("notes", "t.user_id = ?", userId)
	checklistsSize := calcSize("checklists", "t.user_id = ?", userId)
	membersSize := calcSize("members", "t.user_id = ?", userId)
	userSize := calcSize("users", "t.id = ?", userId)
	SessionsSize := calcSize("task_sessions", "t.user_id = ?", userId)
	HistorySize := calcSize("task_histories", "t.user_id = ?", userId)
	SettingsSize := calcSize("user_settings", "t.user_id = ?", userId)

	result.TasksBytes = calcSize("tasks", "t.user_id = ? AND t.type = 'task'", userId)
	result.GoalsBytes = calcSize("tasks", "t.user_id = ? AND t.type = 'goal'", userId)
	result.WorkspacesBytes = calcSize("workspaces", "t.owner_id = ?", userId) + membersSize

	result.TotalBytes = result.TasksBytes +
		result.GoalsBytes +
		result.WorkspacesBytes +
		SessionsSize +
		HistorySize +
		SettingsSize +
		notesSize +
		checklistsSize +
		userSize

	c.JSON(http.StatusOK, result)
}
