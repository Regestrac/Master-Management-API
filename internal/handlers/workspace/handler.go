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
		Select("workspaces.*").
		Joins("JOIN members wm ON wm.workspace_id = workspaces.id").
		Where("wm.user_id = ? AND wm.deleted_at IS NULL", userId).
		Scan(&workspaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
		return
	}

	response := []gin.H{}
	for _, workspace := range workspaces {
		response = append(response, gin.H{
			"id":         workspace.ID,
			"name":       workspace.Name,
			"manager_id": workspace.ManagerId,
			"created_at": workspace.CreatedAt,
			"type":       workspace.Type,
		})
	}

	c.JSON(http.StatusOK, gin.H{"workspaces": response})

}

func GetWorkspaceById(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	workspaceId := c.Param("workspaceId")

	var workspace models.Workspace
	if err := db.DB.Table("workspaces").
		Select("workspaces.*").
		Joins("JOIN members ON members.workspace_id = workspaces.id").
		Where("members.user_id = ? AND workspaces.id = ? AND members.deleted_at IS NULL", userId, workspaceId).
		First(&workspace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workspace details!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workspace})
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

	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join workspace! Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined successfully.",
		"workspace": gin.H{
			"id":   workspace.ID,
			"name": workspace.Name,
		},
	})
}

func LeaveWorkspace(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	workspaceId := c.Param("workspaceId")

	// Check if workspace exists and user is a member
	var member models.Member
	if err := db.DB.Where("workspace_id = ? AND user_id = ?", workspaceId, userId).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not a member of this workspace"})
		return
	}

	// Check if user is not the manager
	var workspace models.Workspace
	if err := db.DB.First(&workspace, workspaceId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}

	if workspace.ManagerId == userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Workspace manager cannot leave. Transfer ownership first"})
		return
	}

	// Delete the member record
	if err := db.DB.Delete(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave workspace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the workspace"})
}

func GetMembers(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized!"})
		return
	}

	workspaceId := c.Param("workspaceId")

	type MemberWithName struct {
		models.Member
		Name string `json:"name"`
	}

	var members []MemberWithName
	if err := db.DB.Table("members").
		Select("members.*, CONCAT(users.first_name, ' ', users.last_name) as name").
		Joins("JOIN users ON users.id = members.user_id").
		Where("members.workspace_id = ? AND members.deleted_at IS NULL", workspaceId).
		Scan(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve members list!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func GetWorkspaceTasks(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	workspaceId := c.Param("workspaceId")

	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var tasks []models.Task
	if err := db.DB.Where("workspace_id = ?", workspaceId).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
