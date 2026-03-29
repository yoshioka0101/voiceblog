package usecase

import (
	"context"
	"strings"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	dbtx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/prompt"
)

var (
	ErrNotFound  = apperr.NotFound("prompt")
	ErrForbidden = apperr.ErrForbidden
)

type CreatePromptInput struct {
	UserID   int64
	Name     string
	Body     string
	IsActive *bool
}

type UpdatePromptInput struct {
	UserID   int64
	PromptID int64
	Name     *string
	Body     *string
	IsActive *bool
}

type UseCase struct {
	repo     entity.Repository
	txRunner dbtx.TxRunner
}

func NewUseCase(repo entity.Repository, txRunners ...dbtx.TxRunner) *UseCase {
	var txRunner dbtx.TxRunner
	if len(txRunners) > 0 {
		txRunner = txRunners[0]
	}

	return &UseCase{
		repo:     repo,
		txRunner: txRunner,
	}
}

func (uc *UseCase) ListVisiblePromptsByUserID(ctx context.Context, userID int64) ([]*entity.Prompt, error) {
	return uc.repo.ListVisiblePromptsByUserID(ctx, userID)
}

func (uc *UseCase) CreatePrompt(ctx context.Context, input CreatePromptInput) (*entity.Prompt, error) {
	name := strings.TrimSpace(input.Name)
	body := strings.TrimSpace(input.Body)
	if name == "" || body == "" {
		return nil, apperr.BadRequest("name and body are required")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	userID := input.UserID
	return uc.repo.CreatePrompt(ctx, &entity.Prompt{
		UserID:   &userID,
		Name:     name,
		Body:     body,
		IsActive: isActive,
	})
}

func (uc *UseCase) UpdatePrompt(ctx context.Context, input UpdatePromptInput) (*entity.Prompt, error) {
	if input.Name == nil && input.Body == nil && input.IsActive == nil {
		return nil, apperr.BadRequest("at least one field is required")
	}

	if uc.txRunner == nil {
		return uc.updatePrompt(ctx, input)
	}

	var prompt *entity.Prompt
	err := uc.txRunner.RunInTx(ctx, func(txCtx context.Context) error {
		var err error
		prompt, err = uc.updatePrompt(txCtx, input)
		return err
	})
	if err != nil {
		return nil, err
	}

	return prompt, nil
}

func (uc *UseCase) updatePrompt(ctx context.Context, input UpdatePromptInput) (*entity.Prompt, error) {
	prompt, err := uc.repo.FindPromptByID(ctx, input.PromptID)
	if err != nil {
		return nil, err
	}
	if !ownedByUser(prompt, input.UserID) {
		return nil, ErrForbidden
	}

	if input.Name != nil {
		trimmedName := strings.TrimSpace(*input.Name)
		if trimmedName == "" {
			return nil, apperr.BadRequest("name must not be blank")
		}
		prompt.Name = trimmedName
	}

	if input.Body != nil {
		trimmedBody := strings.TrimSpace(*input.Body)
		if trimmedBody == "" {
			return nil, apperr.BadRequest("body must not be blank")
		}
		prompt.Body = trimmedBody
	}

	if input.IsActive != nil {
		prompt.IsActive = *input.IsActive
	}

	return uc.repo.UpdatePrompt(ctx, prompt)
}

func (uc *UseCase) DeletePrompt(ctx context.Context, userID, promptID int64) error {
	if uc.txRunner == nil {
		return uc.deletePrompt(ctx, userID, promptID)
	}

	return uc.txRunner.RunInTx(ctx, func(txCtx context.Context) error {
		return uc.deletePrompt(txCtx, userID, promptID)
	})
}

func (uc *UseCase) deletePrompt(ctx context.Context, userID, promptID int64) error {
	prompt, err := uc.repo.FindPromptByID(ctx, promptID)
	if err != nil {
		return err
	}
	if !ownedByUser(prompt, userID) {
		return ErrForbidden
	}

	return uc.repo.DeletePrompt(ctx, promptID)
}

func ownedByUser(prompt *entity.Prompt, userID int64) bool {
	return prompt.UserID != nil && *prompt.UserID == userID
}
