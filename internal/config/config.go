package config

import "os"

var (
	OpenaiAPIKey  string
	OpenaiModel   string
	GptBillPrompt string
)

func GetOpenaiAPIKey() string {
	if OpenaiAPIKey != "" {
		return OpenaiAPIKey
	}
	return os.Getenv("OPENAI_API_KEY")
}

func GetOpenaiModel() string {
	if OpenaiModel != "" {
		return OpenaiModel
	}
	return os.Getenv("OPENAI_MODEL")
}

func GetGptBillPrompt() string {
	if GptBillPrompt != "" {
		return GptBillPrompt
	}
	return os.Getenv("GPT4O_BILL_PROMPT")
}
