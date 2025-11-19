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
	if goal.TargetType == nil {
		return 0
	}
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
	case "repetition", "sessions", "points", "percentage":
		totalTarget := 0.0
		if goal.TargetValue != nil {
			totalTarget = *goal.TargetValue
		}
		targetProgress := 0.0
		if goal.TargetProgress != nil {
			targetProgress = *goal.TargetProgress
		}
		if totalTarget == 0 {
			return 0
		}
		return (targetProgress / totalTarget) * 100
	default:
		return 0
	}
}

func RecalculateProgress(id uint) (float64, error) {
	var task models.Task
	if err := db.DB.First(&task, "id = ?", id).Error; err != nil {
		return 0, err
	}

	// 1️⃣ Checklist Progress
	var checklistTotal, checklistDone int64
	db.DB.Model(&models.Checklist{}).
		Where("task_id = ? AND deleted_at IS NULL", id).
		Count(&checklistTotal)

	db.DB.Model(&models.Checklist{}).
		Where("task_id = ? AND completed = true AND deleted_at IS NULL", id).
		Count(&checklistDone)

	checklistProgress := 0.0
	if checklistTotal > 0 {
		checklistProgress = (float64(checklistDone) / float64(checklistTotal)) * 100
	}

	// 2️⃣ Subtask Progress (only for parent goals)
	var subtaskTotal, subtaskDone int64
	db.DB.Model(&models.Task{}).
		Where("parent_id = ? AND deleted_at IS NULL", id).
		Count(&subtaskTotal)

	db.DB.Model(&models.Task{}).
		Where("parent_id = ? AND status = 'completed' AND deleted_at IS NULL", id).
		Count(&subtaskDone)

	subtaskProgress := 0.0
	if subtaskTotal > 0 {
		subtaskProgress = (float64(subtaskDone) / float64(subtaskTotal)) * 100
	}

	// 3️⃣ Activity (time/count) Progress
	activityProgress := 0.0
	if task.TargetType != nil {
		activityProgress = math.Min(CalculateActivityProgress(task), 100)
	}

	// 4️⃣ Dynamic weight distribution
	totalWeight := 0.0
	if checklistTotal > 0 {
		totalWeight += 1
	}
	if subtaskTotal > 0 {
		totalWeight += 1
	}
	if task.TargetType != nil {
		totalWeight += 1
	}

	finalProgress := 0.0
	if totalWeight > 0 {
		if checklistTotal > 0 {
			finalProgress += (checklistProgress / totalWeight)
		}
		if subtaskTotal > 0 {
			finalProgress += (subtaskProgress / totalWeight)
		}
		if task.TargetType != nil {
			finalProgress += (activityProgress / totalWeight)
		}
	}

	// 5️⃣ Save progress to DB
	if err := db.DB.Model(&task).Update("progress", finalProgress).Error; err != nil {
		return 0, err
	}

	return finalProgress, nil
}
