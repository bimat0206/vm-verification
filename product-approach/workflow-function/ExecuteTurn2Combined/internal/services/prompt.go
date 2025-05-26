package services

import (
	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/templateloader"
)

// PromptService renders the Turn 2 prompt using a template and previous analysis.
type PromptService interface {
	GenerateTurn2Prompt(systemPrompt string, ctx models.VerificationContext, previous interface{}) (string, error)
}

type promptService struct {
	loader templateloader.TemplateLoader
	cfg    config.Config
}

// NewPromptService creates a prompt service.
func NewPromptService(cfg *config.Config) (PromptService, error) {
	loader, err := templateloader.New(templateloader.Config{BasePath: cfg.Prompts.TemplateBasePath})
	if err != nil {
		return nil, err
	}
	return &promptService{loader: loader, cfg: *cfg}, nil
}

func (p *promptService) GenerateTurn2Prompt(systemPrompt string, ctx models.VerificationContext, previous interface{}) (string, error) {
	data := map[string]interface{}{
		"SystemPrompt":     systemPrompt,
		"Verification":     ctx,
		"PreviousAnalysis": previous,
	}
	return p.loader.RenderTemplateWithVersion("turn2", p.cfg.Prompts.TemplateVersion, data)
}
