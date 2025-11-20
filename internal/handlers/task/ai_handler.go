package task

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Regestrac/Master-Management-API/pkg/ai"

	"github.com/gin-gonic/gin"
)

func GenerateDescription(c *gin.Context) {
	var body struct {
		Topic string `json:"topic"`
	}

	if err := c.BindJSON(&body); err != nil || strings.TrimSpace(body.Topic) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic is required"})
		return
	}

	prompt := fmt.Sprintf("Generate a helpful, concise description for the task: %s", body.Topic)

	description, err := ai.Generate(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"description": description})
}

func GenerateTags(c *gin.Context) {
	var body struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Checklists   []string `json:"checklist"`
		ExistingTags []string `json:"existing_tags"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	checklists := strings.Join(body.Checklists, ", ")
	existingTags := strings.Join(body.ExistingTags, ", ")

	prompt := fmt.Sprintf(`
		You need to generate some clear and useful tags for the task/goal the user is doing.
		If available and needed, use the description for additional context.
		If available and needed, use the checklists of this task/goal for additional context.
		Only generate about 3-7 tags if a specific number is not provided.
		The title is: %s.
		The description is %s.
		The checklists are: [%s].
		The existing tags are: [%s]. Do not create or return existing tags.
		Provide the response in the following format: ["tag1", "tag2", "tag3"];
		Return the output as a JSON object, without any other text and without the backticks.
		Do not include any other text or comments.
	`, body.Title, body.Description, checklists, existingTags)

	fmt.Println(prompt)

	tags, err := ai.Generate(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
