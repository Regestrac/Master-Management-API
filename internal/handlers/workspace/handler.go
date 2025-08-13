package workspace

import (
	"master-management-api/internal/db"
	"master-management-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetWorkspaces(c *gin.Context) {
	userData, _ := c.Get("user")
	userId := userData.(models.User).ID

	var workspaces []models.Workspace

	// Join workspace_members and workspaces to find all workspaces for this user
	if err := db.DB.
		Table("workspaces").
		Select("workspaces.id, workspaces.name, workspaces.owner_id, workspaces.created_at").
		Joins("JOIN workspace_members wm ON wm.workspace_id = workspaces.id").
		Where("wm.user_id = ?", userId).
		Scan(&workspaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
	})

}
