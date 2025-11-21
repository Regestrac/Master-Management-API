package workspace

import (
	"net/http"
	"time"

	"github.com/regestrac/master-management-api/internal/db"
	"github.com/regestrac/master-management-api/internal/models"
	"github.com/regestrac/master-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type ResponseType struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	UserId          uint       `json:"user_id"`
	TimeSpend       uint       `json:"time_spend"`
	Streak          uint       `json:"streak"`
	StartedAt       *time.Time `json:"started_at"`
	ParentId        *uint      `json:"parent_id"`
	LastAccessedAt  *time.Time `json:"last_accessed_at"`
	LastStartedAt   *time.Time `json:"last_started_at"`
	Priority        *string    `json:"priority"`
	Type            string     `json:"type"`
	DueDate         *time.Time `json:"due_date"`
	Category        *string    `json:"category"`
	Tags            *[]string  `json:"tags" gorm:"serializer:json"`
	Achievements    *[]string  `json:"achievements" gorm:"serializer:json"`
	WorkspaceId     *uint      `json:"workspace_id"`
	Assignees       *[]uint    `json:"assignees" gorm:"serializer:json"`
	CompletedAt     *time.Time `json:"completed_at"`
	Progress        *float64   `json:"progress"`
	TargetValue     *float64   `json:"target_value"`
	TargetType      *string    `json:"target_type"`
	TargetFrequency *string    `json:"target_frequency"`
	SubTaskCount    *int64     `json:"sub_task_count"`
}

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

	var response []ResponseType
	for _, task := range tasks {
		var subtaskCount int64

		if err := db.DB.Model(&models.Task{}).Where("parent_id = ?", task.ID).Count(&subtaskCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subtask count!"})
			return
		}

		response = append(response, ResponseType{
			ID:              task.ID,
			Title:           task.Title,
			Description:     task.Description,
			Status:          task.Status,
			UserId:          task.UserId,
			TimeSpend:       task.TimeSpend,
			Streak:          task.Streak,
			StartedAt:       task.StartedAt,
			ParentId:        task.ParentId,
			LastAccessedAt:  task.LastAccessedAt,
			LastStartedAt:   task.LastStartedAt,
			Priority:        task.Priority,
			Type:            task.Type,
			DueDate:         task.DueDate,
			Category:        task.Category,
			Tags:            task.Tags,
			Achievements:    task.Achievements,
			WorkspaceId:     task.WorkspaceId,
			Assignees:       task.Assignees,
			CompletedAt:     task.CompletedAt,
			Progress:        task.Progress,
			TargetValue:     task.TargetValue,
			TargetType:      task.TargetType,
			TargetFrequency: task.TargetFrequency,
			SubTaskCount:    &subtaskCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": response})
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

	var goals []models.Task
	query := db.DB.Where("workspace_id = ? AND type = 'goal'", workspaceId)

	if searchKey != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+searchKey+"%")
	}

	if err := query.Find(&goals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve goals!"})
		return
	}

	var response []ResponseType
	for _, goal := range goals {
		var subGoalsCount int64

		if err := db.DB.Model(&models.Task{}).Where("parent_id = ?", goal.ID).Count(&subGoalsCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subtask count!"})
			return
		}

		response = append(response, ResponseType{
			ID:              goal.ID,
			Title:           goal.Title,
			Description:     goal.Description,
			Status:          goal.Status,
			UserId:          goal.UserId,
			TimeSpend:       goal.TimeSpend,
			Streak:          goal.Streak,
			StartedAt:       goal.StartedAt,
			ParentId:        goal.ParentId,
			LastAccessedAt:  goal.LastAccessedAt,
			LastStartedAt:   goal.LastStartedAt,
			Priority:        goal.Priority,
			Type:            goal.Type,
			DueDate:         goal.DueDate,
			Category:        goal.Category,
			Tags:            goal.Tags,
			Achievements:    goal.Achievements,
			WorkspaceId:     goal.WorkspaceId,
			Assignees:       goal.Assignees,
			CompletedAt:     goal.CompletedAt,
			Progress:        goal.Progress,
			TargetValue:     goal.TargetValue,
			TargetType:      goal.TargetType,
			TargetFrequency: goal.TargetFrequency,
			SubTaskCount:    &subGoalsCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"goals": response})
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
