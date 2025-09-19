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

	searchKey := c.Query("searchKey")

	var workspaces []models.Workspace
	query := db.DB.
		Table("workspaces").
		Select("workspaces.*").
		Joins("JOIN members wm ON wm.workspace_id = workspaces.id").
		Where("wm.user_id = ? AND wm.deleted_at IS NULL", userId)

	if searchKey != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+searchKey+"%")
	}

	if err := query.Scan(&workspaces).Error; err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite code"})
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
	searchKey := c.Query("searchKey")

	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var tasks []models.Task
	query := db.DB.Where("workspace_id = ? AND type = 'task'", workspaceId)

	if searchKey != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+searchKey+"%")
	}

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func GetWorkspaceGoals(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	workspaceId := c.Param("workspaceId")
	searchKey := c.Query("searchKey")

	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var tasks []models.Task
	query := db.DB.Where("workspace_id = ? AND type = 'goal'", workspaceId)

	if searchKey != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+searchKey+"%")
	}

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve goals!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": tasks})
}

func UpdateMember(c *gin.Context) {
	// userData, _ := c.Get("user")
	// userId := userData.(models.User).ID

	memberId := c.Param("memberId")

	var body struct {
		Role         *string `json:"role"`
		ProfileColor *string `json:"profile_color"`
		AvatarUrl    *string `json:"avatar_url"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	var member models.Member
	if err := db.DB.Where("id = ?", memberId).First(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve member details!"})
		return
	}

	if body.Role != nil {
		member.Role = *body.Role
	}
	if body.AvatarUrl != nil {
		member.AvatarUrl = *body.AvatarUrl
	}
	if body.ProfileColor != nil {
		member.ProfileColor = *body.ProfileColor
	}

	if err := db.DB.Save(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully."})
}

func RemoveMember(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	workspaceId := c.Param("workspaceId")
	memberToRemoveId := c.Param("memberId")

	// Check if requester is manager/admin
	var requesterMember models.Member
	if err := db.DB.Where("workspace_id = ? AND user_id = ?", workspaceId, userId).First(&requesterMember).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not a member of this workspace"})
		return
	}

	if requesterMember.Role != "manager" && requesterMember.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only managers and admins can remove members"})
		return
	}

	// Get member to remove
	var memberToRemove models.Member
	if err := db.DB.Where("workspace_id = ? AND id = ?", workspaceId, memberToRemoveId).First(&memberToRemove).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	// Check if removing a manager
	if memberToRemove.Role == "manager" {
		// Count remaining managers
		var managerCount int64
		db.DB.Model(&models.Member{}).Where("workspace_id = ? AND role = ? AND deleted_at IS NULL", workspaceId, "manager").Count(&managerCount)

		if managerCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the only manager. Assign another manager first"})
			return
		}
	}

	// Soft delete the member
	if err := db.DB.Delete(&memberToRemove).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}
