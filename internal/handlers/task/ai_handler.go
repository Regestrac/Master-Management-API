package task

import (
	"fmt"
	"master-management-api/pkg/ai"
	"net/http"
	"strings"

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
