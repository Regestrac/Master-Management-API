package note

import (
	"net/http"

	"github.com/Regestrac/Master-Management-API/internal/db"
	"github.com/Regestrac/Master-Management-API/internal/models"

	"github.com/gin-gonic/gin"
)

func AddNote(c *gin.Context) {
	var body struct {
		Content     string `json:"content"`
		X           int    `json:"x"`
		Y           int    `json:"y"`
		Width       uint   `json:"width"`
		Height      uint   `json:"height"`
		TextColor   string `json:"text_color"`
		BgColor     string `json:"bg_color"`
		BorderColor string `json:"border_color"`
		TaskId      uint   `json:"task_id"`
		Variant     string `json:"variant"`
	}

	userDataRaw, _ := c.Get("user")
	user, ok := userDataRaw.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	note := models.Note{
		Content:     body.Content,
		TextColor:   body.TextColor,
		BgColor:     body.BgColor,
		BorderColor: body.BorderColor,
		X:           body.X,
		Y:           body.Y,
		Height:      body.Height,
		Width:       body.Width,
		Variant:     body.Variant,
		TaskId:      body.TaskId,
		UserId:      user.ID,
	}

	if db.DB.Create(&note).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add note!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note added successfully!", "data": note})
}

func UpdateNote(c *gin.Context) {
	var body struct {
		Content     string `json:"content"`
		X           int    `json:"x"`
		Y           int    `json:"y"`
		Width       uint   `json:"width"`
		Height      uint   `json:"height"`
		TextColor   string `json:"text_color"`
		BgColor     string `json:"bg_color"`
		BorderColor string `json:"border_color"`
		TaskId      uint   `json:"task_id"`
		UserId      uint   `json:"user_id"`
		Variant     string `json:"variant"`
		ID          uint   `json:"id" gorm:"primaryKey"`
	}

	noteId := c.Param("noteId")

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	note := models.Note{}
	if db.DB.Where("id = ?", noteId).First(&note).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Could not find note!"})
		return
	}

	if body.Content != "" {
		note.Content = body.Content
	}
	if body.TextColor != "" {
		note.TextColor = body.TextColor
	}
	if body.BgColor != "" {
		note.BgColor = body.BgColor
	}
	if body.BorderColor != "" {
		note.BorderColor = body.BorderColor
	}
	if body.X != 0 {
		note.X = body.X
	}
	if body.Y != 0 {
		note.Y = body.Y
	}
	if body.Width != 0 {
		note.Width = body.Width
	}
	if body.Height != 0 {
		note.Height = body.Height
	}
	if body.TaskId != 0 {
		note.TaskId = body.TaskId
	}
	if body.UserId != 0 {
		note.UserId = body.UserId
	}
	if body.Variant != "" {
		note.Variant = body.Variant
	}

	if db.DB.Save(&note).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note!"})
		return
	}

	noteResponse := models.Note{
		Content:     note.Content,
		TaskId:      note.TaskId,
		UserId:      note.UserId,
		TextColor:   note.TextColor,
		BgColor:     note.BgColor,
		BorderColor: note.BorderColor,
		X:           note.X,
		Y:           note.Y,
		Height:      note.Height,
		Width:       note.Width,
		Variant:     note.Variant,
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully.", "data": noteResponse})
}

func DeleteNote(c *gin.Context) {
	noteId := c.Param("noteId")

	note := models.Note{}
	if db.DB.Where("id = ?", noteId).First(&note).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Could not find note!"})
		return
	}

	if db.DB.Delete(&note).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully."})
}

func GetAllNotes(c *gin.Context) {
	taskId := c.Query("task_id")

	userRawData, _ := c.Get("user")
	userId := userRawData.(models.User).ID

	var notes []models.Note

	if db.DB.Where("user_id = ? AND task_id = ?", userId, taskId).Find(&notes).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": notes})
}
