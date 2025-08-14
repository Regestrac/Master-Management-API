package workspace

import (
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"master-management-api/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetWorkspaces(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var workspaces []models.Workspace

	// Join workspace_members and workspaces to find all workspaces for this user
	if err := db.DB.
		Table("workspaces").
		Select("workspaces.id, workspaces.name, workspaces.manager_id, workspaces.created_at").
		Joins("JOIN members wm ON wm.workspace_id = workspaces.id").
		Where("wm.user_id = ?", userId).
		Scan(&workspaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
	})

}

func CreateWorkspace(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var body struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	inviteCode := utils.GenerateInviteCode(4)

	workspace := models.Workspace{
		Name:       body.Name,
		ManagerId:  userId,
		InviteCode: inviteCode,
	}

	if err := db.DB.Create(&workspace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create a workspace. please try again later."})
		return
	}

	currentTime := time.Now()

	member := models.Member{
		WorkspaceId: workspace.ID,
		UserId:      userId,
		Role:        "manager",
		JoinedAt:    &currentTime,
	}

	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join workspace as manager! Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workspace created successfully.",
		"data":    workspace,
	})
}

func JoinWorkspace(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var body struct {
		InviteCode string `json:"invite_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body!"})
		return
	}

	var workspace models.Workspace
	if err := db.DB.Where("invite_code = ?", body.InviteCode).First(&workspace).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid invite code"})
		return
	}

	var existing models.Member
	if err := db.DB.Where("workspace_id = ? AND user_id = ?", workspace.ID, userId).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already a member"})
		return
	}

	currentTime := time.Now()

	member := models.Member{
		WorkspaceId: workspace.ID,
		UserId:      userId,
		Role:        "member",
		JoinedAt:    &currentTime,
	}

	if err := db.DB.Create(&member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join workspace! Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined successfully."})
}
