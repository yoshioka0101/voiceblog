package user

import (
	"context"
	"fmt"

	dbtx "github.com/yoshioka0101/voiceblog/backend/internal/db"
	"github.com/yoshioka0101/voiceblog/backend/internal/entity/repository"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

type UseCase struct {
	repo     repository.UserRepository
	txRunner dbtx.TxRunner
}

func NewUseCase(repo repository.UserRepository, txRunner dbtx.TxRunner) *UseCase {
	return &UseCase{repo: repo, txRunner: txRunner}
}

func (uc *UseCase) FindOrCreate(ctx context.Context, provider, subject, email, name string) (*entity.User, error) {
	return uc.repo.FindOrCreate(ctx, provider, subject, email, name)
}

// DeleteAccount はユーザーと関連データ（記事・文字起こし・プロンプト・外部連携トークン・投稿履歴）を
// 1 トランザクションで物理削除する。App Store のアカウント削除要件に対応する。
func (uc *UseCase) DeleteAccount(ctx context.Context, userID int64) error {
	if err := uc.txRunner.RunInTx(ctx, func(ctx context.Context) error {
		return uc.repo.DeleteWithRelatedData(ctx, userID)
	}); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}
