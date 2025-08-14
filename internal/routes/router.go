package routes

import (
	"fmt"
	"master-management-api/internal/handlers/auth"
	"master-management-api/internal/handlers/checklist"
	"master-management-api/internal/handlers/history"
	"master-management-api/internal/handlers/note"
	"master-management-api/internal/handlers/profile"
	"master-management-api/internal/handlers/subtasks"
	"master-management-api/internal/handlers/task"
	"master-management-api/internal/middleware"
	"net/http"
	"os"

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
	router.PATCH("/update-active-task", profile.UpdateActiveTask)
	router.PATCH("/update-theme", profile.UpdateTheme)

	router.POST("/task", task.CreateTask)
	router.GET("/tasks", task.GetAllTasks)
	router.DELETE("/tasks/:id", task.DeleteTask)
	router.GET("/tasks/:id", task.GetTask)
	router.PATCH("/tasks/:id", task.UpdateTask)
	router.GET("/recent-tasks", task.GetRecentTasks)
	router.GET("/tasks/stats", task.GetTaskStats)

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

	router.Run(os.Getenv("PORT"))
	fmt.Println("Listening to port" + os.Getenv("PORT"))
}
