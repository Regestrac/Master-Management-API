package ai

import (
	"context"
	"os"

	"google.golang.org/genai"
)

var (
	client *genai.Client
	model  string = "gemini-2.0-flash"
)

// Init initializes the Gemini client. Call this once in your main.go or similar.
func Init() error {
	var err error
	ctx := context.Background()

	client, err = genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})

	return err
}

// Generate response based on given prompt.
func Generate(prompt string) (string, error) {
	ctx := context.Background()

	// config := &genai.GenerateContentConfig{
	// ThinkingConfig: &genai.ThinkingConfig{
	// 	ThinkingBudget: int32(0), // Disables thinking
	// },
	// SystemInstruction: genai.NewContentFromText("Answer in less than 500 characters.", genai.RoleUser),
	// }

	result, err := client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}
