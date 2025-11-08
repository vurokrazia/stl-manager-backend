package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Classifier interface {
	Classify(ctx context.Context, fileName string, allowedCategories []string) ([]string, error)
	IsEnabled() bool
}

type OpenAIClassifier struct {
	client  *openai.Client
	model   string
	enabled bool
}

func NewOpenAIClassifier(apiKey string) *OpenAIClassifier {
	// Only enable if API key is provided and not empty
	enabled := apiKey != "" && apiKey != "YOUR_API_KEY_HERE"
	var client *openai.Client
	if enabled {
		client = openai.NewClient(apiKey)
	}
	return &OpenAIClassifier{
		client:  client,
		model:   openai.GPT4oMini, // Fast and cheap
		enabled: enabled,
	}
}

func (c *OpenAIClassifier) IsEnabled() bool {
	return c.enabled
}

func (c *OpenAIClassifier) Classify(ctx context.Context, fileName string, allowedCategories []string) ([]string, error) {
	// If not enabled, return empty (no classification)
	if !c.enabled || c.client == nil {
		return []string{}, nil
	}

	systemPrompt := `You are a classifier. You receive a filename and a catalog of categories.
Return ONLY a JSON array of strings from the catalog. No extra text.`

	categoriesJSON, _ := json.Marshal(allowedCategories)
	userPrompt := fmt.Sprintf(`file_name: "%s"
allowed_categories: %s

Instructions:
- Choose 0-3 categories from the catalog that describe the file by its NAME.
- If it doesn't fit, return [].
- Respond ONLY with JSON: ["cat1","cat2"]

Examples:
benchy_calibration.stl          -> ["calibration"]
iphone_magsafe_mount_v2.zip     -> ["phone_accessory","mount"]
orc_mini_pack.rar               -> ["miniature","figurine"]
nozzle_holder_4010_fan.stl      -> ["tool_holder","printer_upgrade"]
gt2_pulley_adapter_v3.stl       -> ["adapter","mechanical_part"]
porsche_911_body.zip            -> ["vehicle","rc_part"]
gojo_figure.stl                 -> ["figurine","anime"]
`, fileName, string(categoriesJSON))

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   100,
	})

	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return []string{}, nil
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Parse JSON response
	var categories []string
	if err := json.Unmarshal([]byte(content), &categories); err != nil {
		// If parsing fails, try to extract JSON from text
		start := strings.Index(content, "[")
		end := strings.LastIndex(content, "]")
		if start >= 0 && end > start {
			jsonStr := content[start : end+1]
			if err := json.Unmarshal([]byte(jsonStr), &categories); err != nil {
				return []string{}, nil
			}
		} else {
			return []string{}, nil
		}
	}

	// Validate categories are in allowed list
	validCategories := make([]string, 0, len(categories))
	allowedMap := make(map[string]bool)
	for _, cat := range allowedCategories {
		allowedMap[strings.ToLower(cat)] = true
	}

	for _, cat := range categories {
		if allowedMap[strings.ToLower(cat)] {
			validCategories = append(validCategories, strings.ToLower(cat))
		}
	}

	return validCategories, nil
}
