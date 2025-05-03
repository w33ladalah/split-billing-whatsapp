package processor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/w33ladalah/split-billing-whatsapp/internal/models"
)

// ImageProcessor handles bill image processing via GPT-4o
// It sends the image and a system prompt to the OpenAI API and parses the response.
type ImageProcessor struct {
	openaiAPIKey string
	openaiModel  string
	gpt4oPrompt  string
}

func NewImageProcessor() *ImageProcessor {
	apiKey := os.Getenv("OPENAI_API_KEY")
	prompt := os.Getenv("GPT4O_BILL_PROMPT")
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	if prompt == "" {
		prompt = "You are a bill extraction assistant. Given a photo of a bill/receipt, extract all items with their names and prices, and the total. Return as JSON: {\"items\":[{\"name\":\"string\",\"amount\":float}],\"total\":float}."
	}
	return &ImageProcessor{
		openaiAPIKey: apiKey,
		openaiModel:  model,
		gpt4oPrompt:  prompt,
	}
}

// ProcessBillImage sends the image to GPT-4o and parses the response into a Bill
func (p *ImageProcessor) ProcessBillImage(imgData []byte) (*models.Bill, error) {
	jsonResp, err := p.callGPT4o(imgData)
	if err != nil {
		return nil, fmt.Errorf("GPT-4o error: %w", err)
	}
	if strings.TrimSpace(jsonResp) == "" {
		fmt.Println("[DEBUG] GPT-4o returned empty response body")
		return nil, errors.New("GPT-4o returned empty response")
	}
	return p.parseBillJSON(jsonResp)
}

// callGPT4o sends the image and prompt to the OpenAI API and returns the raw JSON string response
func (p *ImageProcessor) callGPT4o(imgData []byte) (string, error) {
	if p.openaiAPIKey == "" {
		return "", errors.New("OPENAI_API_KEY not set in env")
	}
	url := "https://api.openai.com/v1/chat/completions"

	imgBase64 := encodeToBase64(imgData)
	messages := []map[string]any{
		{"role": "system", "content": p.gpt4oPrompt},
		{
			"role": "user",
			"content": []any{
				map[string]any{"type": "text", "text": "[image attached]"},
				map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/jpeg;base64," + imgBase64}},
			},
		},
	}

	payload := map[string]any{
		"model":      p.openaiModel,
		"messages":   messages,
		"max_tokens": 512,
		"tools": []any{
			map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        "extract_receipt_info",
					"description": "Extract purchased items and total from receipt text",
					"parameters": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"items": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"name": map[string]any{
											"type": "string",
										},
										"quantity": map[string]any{
											"type": "number",
										},
										"price": map[string]any{
											"type": "number",
										},
									},
									"required": []string{"name", "price"},
								},
							},
							"total": map[string]any{
								"type": "number",
							},
						},
						"required": []string{"items", "total"},
					},
				},
			},
		},
		"tool_choice": map[string]any{
			"type": "function",
			"function": map[string]any{
				"name": "extract_receipt_info",
			},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	fmt.Println("[DEBUG] OpenAI API payload:")
	fmt.Println(string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Parse OpenAI response to extract the assistant's message content
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", errors.New("no choices in OpenAI response")
	}
	return result.Choices[0].Message.Content, nil
}

// parseBillJSON parses the JSON response into a Bill
func (p *ImageProcessor) parseBillJSON(jsonStr string) (*models.Bill, error) {
	if strings.TrimSpace(jsonStr) == "" {
		fmt.Println("[DEBUG] No content to parse from GPT-4o")
		return nil, errors.New("no content to parse from GPT-4o")
	}
	fmt.Println("[DEBUG] Raw GPT-4o output:")
	fmt.Println(jsonStr)
	var parsed struct {
		Items []struct {
			Name   string  `json:"name"`
			Amount float64 `json:"amount"`
		} `json:"items"`
		Total float64 `json:"total"`
	}
	dec := json.NewDecoder(strings.NewReader(jsonStr))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&parsed); err != nil {
		fmt.Printf("[DEBUG] Failed to parse JSON. Raw content: %s\n", jsonStr)
		return nil, fmt.Errorf("failed to parse %s JSON: %w", p.openaiModel, err)
	}
	bill := models.NewBill("Bill from image")
	for _, item := range parsed.Items {
		bill.AddItem(item.Name, fmt.Sprintf("%.2f", item.Amount))
	}
	bill.Total = parsed.Total
	return bill, nil
}

// Helper: encode image to base64
func encodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
