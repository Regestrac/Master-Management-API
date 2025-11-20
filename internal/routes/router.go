package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Regestrac/Master-Management-API/internal/handlers/analytics"
	"github.com/Regestrac/Master-Management-API/internal/handlers/auth"
	"github.com/Regestrac/Master-Management-API/internal/handlers/checklist"
	"github.com/Regestrac/Master-Management-API/internal/handlers/history"
	"github.com/Regestrac/Master-Management-API/internal/handlers/note"
	"github.com/Regestrac/Master-Management-API/internal/handlers/profile"
	"github.com/Regestrac/Master-Management-API/internal/handlers/settings"
	"github.com/Regestrac/Master-Management-API/internal/handlers/subtasks"
	"github.com/Regestrac/Master-Management-API/internal/handlers/task"
	"github.com/Regestrac/Master-Management-API/internal/handlers/workspace"
	"github.com/Regestrac/Master-Management-API/internal/middleware"

	"github.com/gin-gonic/gin"
)

func handleNoRoute(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "Invalid API route or endpoint"})
}

func SetupRouter() {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.NoRoute(handleNoRoute)

	router.POST("/signup", auth.SignUp)
	router.POST("/login", auth.Login)

	router.Use(middleware.RequireAuth)

	router.GET("/validate", auth.Validate)
	router.POST("/logout", auth.Logout)

	router.GET("/profile", profile.GetProfile)
	router.PUT("/profile", profile.UpdateProfile)
	router.GET("/profile/monthly-stats", profile.GetMonthlyStats)
	router.PATCH("/update-active-task", profile.UpdateActiveTask)
	router.GET("/dashboard/quick-stats", profile.GetQuickStats)

	router.POST("/task", task.CreateTask)
	router.GET("/tasks", task.GetAllTasks)
	router.DELETE("/tasks/:id", task.DeleteTask)
	router.GET("/tasks/:id", task.GetTask)
	router.PATCH("/tasks/:id", task.UpdateTask)
	router.GET("/recent-tasks", task.GetRecentTasks)
	router.GET("/tasks/stats", task.GetTaskStats)
	router.GET("/tasks/categories", task.GetCategories)

	router.GET("/tasks/:id/history", history.GetTaskHistory)
	router.POST("/tasks/:id/history", history.AddToHistory)
	router.POST("/task/:id/generate-description", task.GenerateDescription)

	router.GET("/tasks/:id/subtasks", subtasks.GetAllSubtasks)
	router.POST("/tasks/:id/generate-subtasks", subtasks.GenerateSubTasks)
	router.POST("/tasks/:id/subtasks", subtasks.SaveSubtasks)

	router.POST("/tasks/:id/generate-tags", task.GenerateTags)

	router.GET("/goals/stats", task.GetGoalStats)
	router.GET("/goals/active", task.GetActiveGoals)

	router.POST("/note", note.AddNote)
	router.GET("/notes", note.GetAllNotes)
	router.PATCH("/notes/:noteId", note.UpdateNote)
	router.DELETE("/notes/:noteId", note.DeleteNote)

	router.POST("/checklist", checklist.CreateChecklist)
	router.GET("/checklists", checklist.GetAllChecklists)
	router.PATCH("/checklists/:id", checklist.UpdateChecklist)
	router.DELETE("/checklists/:id", checklist.DeleteChecklist)
	router.POST("/checklists/:id/generate-checklists", checklist.GenerateChecklist)
	router.POST("/checklists", checklist.SaveChecklists)

	router.POST("/workspace", workspace.CreateWorkspace)
	router.GET("/workspaces", workspace.GetWorkspaces)
	router.GET("/workspaces/:workspaceId", workspace.GetWorkspaceById)
	router.GET("/workspaces/:workspaceId/tasks", workspace.GetWorkspaceTasks)
	router.GET("/workspaces/:workspaceId/goals", workspace.GetWorkspaceGoals)
	router.POST("/workspaces/:workspaceId/leave", workspace.LeaveWorkspace)

	router.POST("/workspace/join", workspace.JoinWorkspace)
	router.GET("/workspaces/:workspaceId/members", workspace.GetMembers)
	router.DELETE("/workspaces/:workspaceId/members/:memberId", workspace.RemoveMember)
	router.PATCH("/workspaces/:workspaceId/members/:memberId", workspace.UpdateMember)

	router.GET("/settings", settings.GetUserSettings)
	router.PATCH("/settings", settings.UpdateUserSettings)
	router.PUT("/settings/reset", settings.ResetSettings)
	router.GET("/settings/storage", settings.GetUserStorageUsage)
	router.PATCH("/update-theme", settings.UpdateTheme)

	router.GET("/analytics/quick-metrics", analytics.GetQuickMetrics)
	router.GET("/analytics/productivity-chart", analytics.GetProductivityTrendData)
	router.GET("/analytics/task-distribution", analytics.GetTaskDistributionData)
	router.GET("/analytics/goal-progress", analytics.GetGoalProgressInsights)
	router.GET("/analytics/timely-insights", analytics.GetTimelyInsights)
	router.GET("/analytics/focus-sessions", analytics.GetFocusSessions)

	router.Run(os.Getenv("PORT"))
	fmt.Println("Listening to port" + os.Getenv("PORT"))
}
