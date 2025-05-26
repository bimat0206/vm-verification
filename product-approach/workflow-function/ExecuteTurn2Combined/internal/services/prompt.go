package services

// PromptService generates prompts for Turn 2

type PromptService interface {
    GenerateTurn2Prompt(systemPrompt string, imageData string, previousAnalysis interface{}) (string, error)
}

func NewPromptService(cfg interface{}, log interface{}) (PromptService, error) {
    return &noopPrompt{}, nil
}

type noopPrompt struct{}

func (n *noopPrompt) GenerateTurn2Prompt(systemPrompt string, imageData string, previousAnalysis interface{}) (string, error) {
    return systemPrompt, nil
}
