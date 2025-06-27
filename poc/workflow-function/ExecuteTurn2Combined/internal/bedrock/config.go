package bedrock

import "time"

type Config struct {
	ModelID          string
	AnthropicVersion string
	MaxTokens        int
	Temperature      float64
	TopP             float64
	ThinkingType     string
	ThinkingBudget   int
	Timeout          time.Duration
	Region           string
}
