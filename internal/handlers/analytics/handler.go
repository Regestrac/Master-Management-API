package analytics

import (
	"net/http"
	"time"

	"github.com/Regestrac/master-management-api/internal/db"
	"github.com/Regestrac/master-management-api/internal/models"

	"github.com/gin-gonic/gin"
)

func GetQuickMetrics(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if endDate != "" {
		t, err := time.Parse("02-01-2006", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endDate = t.AddDate(0, 0, 1).Format("2006-01-02")
	}
	prevStartDate := c.Query("prevStartDate")
	prevEndDate := c.Query("prevEndDate")

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var totalFocusTime int64
	var focusTimeChange *int64 = nil
	var tasksCompleted int64
	var tasksCompletedChange *int64 = nil

	if startDate != "" && endDate != "" {
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
	} else {
		if err := db.DB.Model(&models.TaskSession{}).
			Where("user_id = ?", userId).
			Select("COALESCE(SUM(duration),0)").
			Scan(&totalFocusTime).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get focus time!"})
			return
		}
		if err := db.DB.Model(&models.Task{}).
			Select("COUNT(*)").
			Where("user_id = ? AND completed_at IS NOT NULL", userId).
			Scan(&tasksCompleted).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completed tasks!"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"focus_time":             totalFocusTime,
		"focus_time_change":      focusTimeChange,
		"tasks_completed":        tasksCompleted,
		"tasks_completed_change": tasksCompletedChange,
	})
}

func GetProductivityTrendData(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if endDate != "" {
		t, err := time.Parse("02-01-2006", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endDate = t.AddDate(0, 0, 1).Format("2006-01-02")
	}

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	type DailySession struct {
		Date     string `json:"date"`
		Duration int64  `json:"duration"`
	}

	var sessions []DailySession

	if startDate != "" && endDate != "" {
		if err := db.DB.Model(&models.TaskSession{}).
			Select("DATE(start_time) as date, COALESCE(SUM(duration), 0) as duration").
			Where("user_id = ? AND start_time BETWEEN ? AND ?", userId, startDate, endDate).
			Group("DATE(start_time)").
			Order("date").
			Scan(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task sessions!"})
			return
		}
	} else {
		if err := db.DB.Model(&models.TaskSession{}).
			Select("DATE(start_time) as date, COALESCE(SUM(duration), 0) as duration").
			Where("user_id = ?", userId).
			Group("DATE(start_time)").
			Order("date").
			Scan(&sessions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task sessions!"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func GetTaskDistributionData(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if endDate != "" {
		t, err := time.Parse("02-01-2006", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endDate = t.AddDate(0, 0, 1).Format("2006-01-02")
	}

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	type TasksDistribution struct {
		Category string `json:"category"`
		Count    string `json:"count"`
	}

	var tasksCount []TasksDistribution
	query := db.DB.Model(&models.Task{}).Select("COUNT(*) as count, category").Where("user_id = ? AND parent_id IS NULL", userId)

	if startDate != "" && endDate != "" {
		query = query.Where("completed_at BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Group("category").
		Scan(&tasksCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task distribution data!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tasksCount})
}

func getGoalDuration(goalId uint, startDate string, endDate string) (int64, error) {
	var duration int64
	query := db.DB.Model(&models.TaskSession{}).Where("task_id = ?", goalId)

	if startDate != "" && endDate != "" {
		query.Where("start_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := query.Select("COALESCE(SUM(duration),0)").Scan(&duration).Error; err != nil {
		return 0, err
	}

	return duration, nil
}

func GetGoalProgressInsights(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if endDate != "" {
		t, err := time.Parse("02-01-2006", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endDate = t.AddDate(0, 0, 1).Format("2006-01-02")
	}

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	type GoalsResponseType struct {
		ID       uint   `json:"id"`
		Title    string `json:"title"`
		DueDate  string `json:"due_date"`
		Progress string `json:"progress"`
		Duration int64  `json:"duration"`
	}

	var goals []GoalsResponseType
	if err := db.DB.Model(&models.Task{}).
		Select("id, title, due_date, progress, 0 as duration").
		Where("user_id = ? AND status = 'inprogress' AND type = 'goal' AND parent_id IS NULL", userId).
		Scan(&goals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve goal progress insights!"})
		return
	}

	for i := range goals {
		duration, err := getGoalDuration(goals[i].ID, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get goal duration!"})
			return
		}
		goals[i].Duration = duration
	}

	c.JSON(http.StatusOK, gin.H{"data": goals})
}

func GetTimelyInsights(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if endDate != "" {
		t, err := time.Parse("02-01-2006", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endDate = t.AddDate(0, 0, 1).Format("2006-01-02")
	}

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	// Get duration by hour of day (0-23)
	var hourlyData []struct {
		Hour     int   `json:"hour"`
		Duration int64 `json:"duration"`
	}
	hourQuery := db.DB.Model(&models.TaskSession{}).
		Select("EXTRACT(HOUR FROM start_time) as hour, COALESCE(SUM(duration), 0) as duration").
		Where("user_id = ?", userId)
	if startDate != "" && endDate != "" {
		hourQuery = hourQuery.Where("start_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := hourQuery.Group("hour").Order("hour").Scan(&hourlyData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hourly data"})
		return
	}

	// Get duration by day of week (0-6, where 0 is Sunday)
	var dailyData []struct {
		Day      int   `json:"day"`
		Duration int64 `json:"duration"`
	}
	dayQuery := db.DB.Model(&models.TaskSession{}).
		Select("EXTRACT(DOW FROM start_time) as day, COALESCE(SUM(duration), 0) as duration").
		Where("user_id = ?", userId)
	if startDate != "" && endDate != "" {
		dayQuery = dayQuery.Where("start_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := dayQuery.Group("day").Order("day").Scan(&dailyData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch daily data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hourly_distribution": hourlyData,
		"daily_distribution":  dailyData,
	})
}

func GetFocusSessions(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if endDate != "" {
		t, err := time.Parse("02-01-2006", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endDate = t.AddDate(0, 0, 1).Format("2006-01-02")
	}

	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var data struct {
		TotalSessions int64 `json:"total_sessions"`
		Duration      int64 `json:"duration"`
	}

	query := db.DB.Model(&models.TaskSession{}).Where("user_id = ?", userId)
	if startDate != "" && endDate != "" {
		query.Where("start_time BETWEEN ? AND ?", startDate, endDate)
	}
	if err := query.Select("COUNT(*) as total_sessions, COALESCE(SUM(duration), 0) as duration").Scan(&data).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve focus session data!"})
		return
	}

	// Parse and normalize dates
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("02-01-2006", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate format (expected DD-MM-YYYY)"})
			return
		}
	}

	if endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate format (expected DD-MM-YYYY)"})
			return
		}
		end = t.AddDate(0, 0, 1) // include full end date
	}

	type Result struct {
		Category string `json:"category"`
		Duration int64  `json:"duration"`
	}

	var results []Result

	query2 := db.DB.Table("task_sessions AS ts").
		Select(`
			COALESCE(t.category, 'Uncategorised') AS category,
			SUM(ts.duration) AS duration
		`).
		Joins("LEFT JOIN tasks t ON t.id = ts.task_id")

	if !start.IsZero() && !end.IsZero() {
		query2 = query2.Where("ts.created_at BETWEEN ? AND ?", start, end)
	} else if !start.IsZero() {
		query2 = query2.Where("ts.created_at >= ?", start)
	} else if !end.IsZero() {
		query2 = query2.Where("ts.created_at < ?", end)
	}

	if err := query2.Group("category").Order("category ASC").Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "sessions": results})
}
