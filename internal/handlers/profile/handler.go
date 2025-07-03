package profile

import (
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID         uint   `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Theme      string `json:"theme"`
	ActiveTask *uint  `json:"active_task"`
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
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func UpdateProfile(c *gin.Context) {
	type UpdateProfileInput struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
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
		user.ActiveTask = body.ActiveTask
	} else {
		user.ActiveTask = nil
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

	c.JSON(http.StatusOK, gin.H{"message": successMessage})
}

func UpdateTheme(c *gin.Context) {
	var body struct {
		Theme string `json:"theme"`
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

	if body.Theme != "" {
		user.Theme = body.Theme
	}

	if err := db.DB.Save(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theme!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme updated successfully."})
}
