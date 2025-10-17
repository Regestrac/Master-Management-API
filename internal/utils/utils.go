package utils

import (
	"crypto/rand"
	"encoding/base64"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"math"
	"strings"
)

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func MatchScore(title, searchKey, desc string, useDesc bool) int {
	if title == searchKey {
		return 100
	}
	if strings.HasPrefix(title, searchKey) {
		return 90
	}
	if strings.Contains(title, searchKey) {
		return 70
	}
	if useDesc && strings.Contains(desc, searchKey) {
		return 50
	}
	return 0
}

func GenerateInviteCode(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // handle properly in production
	}
	code := base64.URLEncoding.EncodeToString(b)
	return strings.ToUpper(code[:length])
}

func CalculateActivityProgress(goal models.Task) float64 {
	switch *goal.TargetType {
	case "hours":
		totalTarget := *goal.TargetValue * 60 * 60
		return (float64(goal.TimeSpend) / totalTarget) * 100
	case "days":
		totalTarget := *goal.TargetValue * 24 * 60 * 60
		return (float64(goal.TimeSpend) / totalTarget) * 100
	case "weeks":
		totalTarget := *goal.TargetValue * 7 * 24 * 60 * 60
		return (float64(goal.TimeSpend) / totalTarget) * 100
	case "months":
		// Approximating a month as 30 days
		totalTarget := *goal.TargetValue * 30 * 24 * 60 * 60
		return (float64(goal.TimeSpend) / totalTarget) * 100
	}
	return 0
}

func RecalculateGoalProgress(goalID uint) error {
	var goal models.Task
	if err := db.DB.First(&goal, "id = ? AND type = ?", goalID, "goal").Error; err != nil {
		return err
	}

	// Calculate checklist progress
	var checklistTotal, checklistDone int64
	db.DB.Model(&models.Checklist{}).Where("task_id = ? AND deleted_at IS NULL", goalID).Count(&checklistTotal)
	db.DB.Model(&models.Checklist{}).Where("task_id = ? AND completed = true AND deleted_at IS NULL", goalID).Count(&checklistDone)

	checklistProgress := 0.0
	if checklistTotal > 0 {
		checklistProgress = (float64(checklistDone) / float64(checklistTotal)) * 100
	}

	// Calculate time/count progress (placeholder)
	activityProgress := math.Min(CalculateActivityProgress(goal), 100)

	// Weightage
	finalProgress := 0.0
	if goal.ParentId != nil {
		finalProgress = (checklistProgress * 0.5) + (activityProgress * 0.5)
	} else {
		// Calculate subtask progress
		var subtaskTotal, subtaskDone int64
		db.DB.Model(&models.Task{}).Where("parent_id = ? AND deleted_at IS NULL", goalID).Count(&subtaskTotal)
		db.DB.Model(&models.Task{}).Where("parent_id = ? AND status = 'completed' AND deleted_at IS NULL", goalID).Count(&subtaskDone)

		subtaskProgress := 0.0
		if subtaskTotal > 0 {
			subtaskProgress = (float64(subtaskDone) / float64(subtaskTotal)) * 100
		}

		finalProgress = (checklistProgress * 0.25) + (subtaskProgress * 0.25) + (activityProgress * 0.5)
	}

	// Save to DB
	return db.DB.Model(&goal).Update("progress", finalProgress).Error
}
