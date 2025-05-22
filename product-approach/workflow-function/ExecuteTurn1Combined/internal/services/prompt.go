// internal/services/prompt.go
package prompt

import (
	"context"
	"fmt"

	"ExecuteTurn1Combined/internal/config"
	"ExecuteTurn1Combined/internal/errors"
	"ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/templateloader"
)

// PromptService defines prompt-generation operations.
type PromptService interface {
	// GenerateTurn1Prompt renders the Turn-1 template given the verification context and system prompt.
	GenerateTurn1Prompt(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, error)
}

type promptService struct {
	loader  *templateloader.Loader
	version string
}

// NewPromptService constructs a PromptService using the configured template loader.
func NewPromptService(cfg config.Config) (PromptService, error) {
	loader, err := templateloader.New(templateloader.Config{
		BasePath:     cfg.Prompts.TemplateBasePath,
		CacheEnabled: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init template loader: %w", err)
	}
	return &promptService{
		loader:  loader,
		version: cfg.Prompts.TemplateVersion,
	}, nil
}

// GenerateTurn1Prompt renders the "turn1" template with the given context.
func (p *promptService) GenerateTurn1Prompt(
	ctx context.Context,
	vCtx models.VerificationContext,
	systemPrompt string,
) (string, error) {
	// assemble template data
	data := struct {
		VerificationContext models.VerificationContext
		SystemPrompt        string
	}{
		VerificationContext: vCtx,
		SystemPrompt:        systemPrompt,
	}

	tmpl, err := p.loader.RenderTemplateWithVersion("turn1", p.version, data)
	if err != nil {
		return "", errors.WrapNonRetryable(
			fmt.Errorf("render template: %w", err),
			errors.StagePromptGeneration,
			"failed to render turn-1 prompt",
		)
	}
	return tmpl, nil
}
