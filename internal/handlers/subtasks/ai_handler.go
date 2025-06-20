package subtasks

import (
	"fmt"
	"master-management-api/pkg/ai"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// This function will handle the generation of subtasks using AI.
// It will use the Gemini client to generate subtasks based on the main task description.
func GenerateSubTasks(c *gin.Context) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.BindJSON(&body); err != nil || strings.TrimSpace(body.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	prompt := fmt.Sprintf(`
		Generate clear subtasks for this goal/task: "%s".
		Use this description only if you need additional context: "%s".
		Make sure the subtasks are relevant to the task.
		Only generate a maximum of 3 tasks if a specific number is not provided.
		strictly follow the JSON output format as in example.

		For example:
		if the goal/task is "Full body workout", the subtasks could be like:
		[{"title": "Arm workouts", "description": "Exercises targeting arms."},
		{"title": "Abs workouts", "description": "Exercises to strengthen and grow abs."},
		{"title": "Cardio exercises", "description": "Exercises to improve cardiovascular fitness."},
		{"title": "Stretching exercises", "description": "Exercises to improve flexibility and reduce tension."}]
	`, body.Title, body.Description)

	subtasks, err := ai.Generate(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subtasks})
}
