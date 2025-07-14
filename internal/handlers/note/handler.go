package note

import (
	"fmt"
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddNote(c *gin.Context) {
	var body struct {
		Text      string `json:"text"`
		TaskId    uint   `json:"task_id"`
		TextColor string `json:"text_color"`
		BgColor   string `json:"bg_color"`
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
		Text:      body.Text,
		TaskId:    body.TaskId,
		UserId:    user.ID,
		TextColor: body.TextColor,
		BgColor:   body.BgColor,
	}

	if db.DB.Create(&note).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add note!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note added successfully!", "data": note})
}

func UpdateNote(c *gin.Context) {
	var body struct {
		Text      string `json:"text"`
		TextColor string `json:"text_color"`
		BgColor   string `json:"bg_color"`
		TaskId    uint   `json:"task_id"`
		ID        uint   `json:"id"`
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

	fmt.Println("first note: ", note)
	fmt.Println("body: ", body)

	if body.Text != "" {
		note.Text = body.Text
	}
	if body.TextColor != "" {
		note.TextColor = body.TextColor
	}
	if body.BgColor != "" {
		note.BgColor = body.BgColor
	}

	fmt.Println("note before save: ", note)

	if db.DB.Save(&note).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note!"})
		return
	}

	noteResponse := models.Note{
		Text:      note.Text,
		TaskId:    note.TaskId,
		UserId:    note.UserId,
		TextColor: note.TextColor,
		BgColor:   note.BgColor,
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
