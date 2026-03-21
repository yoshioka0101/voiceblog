package usecase

import (
	"context"

	promptdomain "github.com/yoshioka0101/voiceblog/backend/internal/feature/prompt/domain"
)

type UseCase struct {
	repo promptdomain.Repository
}

func New(repo promptdomain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) ListVisibleByUserID(ctx context.Context, userID int64) ([]*promptdomain.Prompt, error) {
	return uc.repo.ListVisibleByUserID(ctx, userID)
}

func (uc *UseCase) Create(ctx context.Context, params promptdomain.CreateParams) (*promptdomain.Prompt, error) {
	return uc.repo.Create(ctx, params)
}

func (uc *UseCase) Update(ctx context.Context, userID, promptID int64, params promptdomain.PatchParams) (*promptdomain.Prompt, error) {
	prompt, err := uc.repo.FindByID(ctx, promptID)
	if err != nil {
		return nil, err
	}
	if !prompt.OwnedBy(userID) {
		return nil, promptdomain.ErrForbidden
	}

	updateParams := promptdomain.UpdateParams{
		ID:       prompt.ID,
		Name:     prompt.Name,
		Body:     prompt.Body,
		IsActive: prompt.IsActive,
	}
	if params.Name != nil {
		updateParams.Name = *params.Name
	}
	if params.Body != nil {
		updateParams.Body = *params.Body
	}
	if params.IsActive != nil {
		updateParams.IsActive = *params.IsActive
	}

	return uc.repo.Update(ctx, updateParams)
}

func (uc *UseCase) Delete(ctx context.Context, userID, promptID int64) error {
	prompt, err := uc.repo.FindByID(ctx, promptID)
	if err != nil {
		return err
	}
	if !prompt.OwnedBy(userID) {
		return promptdomain.ErrForbidden
	}

	return uc.repo.Delete(ctx, promptID)
}
