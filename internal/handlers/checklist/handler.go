package checklist

import (
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateChecklist(c *gin.Context) {
	user, _ := c.Get("user")
	userId := user.(models.User).ID

	var body struct {
		Title  string `json:"title"`
		TaskId uint   `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	checklist := models.Checklist{
		UserId:    userId,
		Title:     body.Title,
		TaskId:    body.TaskId,
		Completed: false,
	}

	if db.DB.Create(&checklist).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checklist!"})
		return
	}

	var task models.Task
	db.DB.Where("id = ?", body.TaskId).First(&task)
	if err := utils.RecalculateProgress(body.TaskId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate goal progress!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Checklist created successfully.",
		"data":     checklist,
		"progress": task.Progress,
	})
}

func GetAllChecklists(c *gin.Context) {
	user, _ := c.Get("user")
	userId := user.(models.User).ID

	taskId := c.Query("task_id")

	var checklists []models.Checklist

	if db.DB.Where("user_id = ? AND task_id = ?", userId, taskId).Find(&checklists).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve checklists!"})
		return
	}

	response := make([]gin.H, 0, len(checklists))

	for _, checklist := range checklists {
		response = append(response, gin.H{
			"id":           checklist.ID,
			"created_at":   checklist.CreatedAt,
			"title":        checklist.Title,
			"task_id":      checklist.TaskId,
			"user_id":      checklist.UserId,
			"completed":    checklist.Completed,
			"completed_at": checklist.CompletedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func UpdateChecklist(c *gin.Context) {
	user, _ := c.Get("user")
	userId := user.(models.User).ID

	id := c.Param("id")

	var body struct {
		Title     string `json:"title"`
		Completed *bool  `json:"completed"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	var checklist models.Checklist

	if db.DB.Where("user_id = ? AND id = ?", userId, id).First(&checklist).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find checklist!"})
		return
	}

	if body.Title != "" {
		checklist.Title = body.Title
	}
	if body.Completed != nil {
		checklist.Completed = *body.Completed
	}

	if db.DB.Save(&checklist).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save checklist!"})
		return
	}

	var task models.Task
	if body.Completed != nil {
		db.DB.Where("id = ?", checklist.TaskId).First(&task)
		if err := utils.RecalculateProgress(task.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate goal progress!"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Updated successfully.",
		"data":     checklist,
		"progress": task.Progress,
	})
}

func DeleteChecklist(c *gin.Context) {
	user, _ := c.Get("user")
	userId := user.(models.User).ID

	id := c.Param("id")

	var checklist = models.Checklist{}

	if db.DB.Where("user_id = ? AND id = ?", userId, id).Find(&checklist).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find checklist!"})
		return
	}

	if db.DB.Delete(&checklist).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete checklist!"})
		return
	}

	var task models.Task
	db.DB.Where("id = ?", checklist.TaskId).First(&task)
	if err := utils.RecalculateProgress(task.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate goal progress!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Deleted successfully.",
		"progress": task.Progress,
	})
}

func SaveChecklists(c *gin.Context) {
	user, _ := c.Get("user")
	userId := user.(models.User).ID

	var body []struct {
		Title  string `json:"title"`
		TaskId uint   `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	var checklists []models.Checklist

	for _, checklist := range body {
		checklists = append(checklists, models.Checklist{
			Title:     checklist.Title,
			Completed: false,
			TaskId:    checklist.TaskId,
			UserId:    userId,
		})
	}

	if db.DB.Save(&checklists).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checklist!"})
		return
	}

	var task models.Task
	db.DB.Where("id = ?", body[0].TaskId).First(&task)
	if err := utils.RecalculateProgress(task.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recalculate goal progress!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Checklists added successfully.",
		"data":     checklists,
		"progress": task.Progress,
	})
}
