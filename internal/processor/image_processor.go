package processor

import (
	"bytes"
	"context"
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
	openaiAPIKey  string
	gpt4oPrompt   string
}

func NewImageProcessor() *ImageProcessor {
	apiKey := os.Getenv("OPENAI_API_KEY")
	prompt := os.Getenv("GPT4O_BILL_PROMPT")
	if prompt == "" {
		prompt = "You are a bill extraction assistant. Given a photo of a bill/receipt, extract all items with their names and prices, and the total. Return as JSON: {\"items\":[{"name":"string","amount":float}],\"total\":float}."
	}
	return &ImageProcessor{
		openaiAPIKey: apiKey,
		gpt4oPrompt:  prompt,
	}
}

// ProcessBillImage sends the image to GPT-4o and parses the response into a Bill
func (p *ImageProcessor) ProcessBillImage(imgData []byte) (*models.Bill, error) {
	jsonResp, err := p.callGPT4o(imgData)
	if err != nil {
		return nil, fmt.Errorf("GPT-4o error: %w", err)
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
	messages := []map[string]string{
		{"role": "system", "content": p.gpt4oPrompt},
		{"role": "user", "content": "[image attached]"},
	}
	// OpenAI API expects the image as a separate part, but for this mockup, we'll assume you have a way to send it.
	// In a real implementation, use OpenAI's vision API endpoints.

	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": messages,
		"max_tokens": 512,
		"tools": []interface{}{},
		"attachments": []map[string]string{
			{"type": "image", "data": imgBase64, "mime": "image/jpeg"},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

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
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
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
		return nil, fmt.Errorf("failed to parse GPT-4o JSON: %w", err)
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
	return "" // TODO: Implement base64 encoding or use a real OpenAI vision endpoint
}
