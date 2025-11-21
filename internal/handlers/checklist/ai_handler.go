package checklist

import (
	"fmt"
	"net/http"

	"github.com/regestrac/master-management-api/pkg/ai"

	"github.com/gin-gonic/gin"
)

func GenerateChecklist(c *gin.Context) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	prompt := fmt.Sprintf(`
		You are an AI assistant in the task/goal management app called 'Master Management'. Your name is Tessa.
		You need to generate some clear checklist for a %s the user is need to do.
		If available and needed, use the description for additional context.
		Only generate about 3-5 checklist if a specific number is not provided.
		Make sure the checklists are relevant to the %s.
		The title is: %s.
		The description is: %s.
		Provide the response in the following format: type ChecklistType = { id: number; title: string; }[];
		Return the output as a JSON object, without any other text and without the backticks.
		Do not include any other text or comments.
	`, body.Type, body.Type, body.Title, body.Description)

	response, err := ai.Generate(prompt)
	fmt.Println("err: ", err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate checklists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}
