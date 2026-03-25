package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"

	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/externaltoken"
	"github.com/yoshioka0101/voiceblog/backend/models"
)

type ExternalTokenRepository struct {
	db bob.DB
}

func NewExternalTokenRepository(db *sql.DB) *ExternalTokenRepository {
	return &ExternalTokenRepository{db: bob.NewDB(db)}
}

func (r *ExternalTokenRepository) Upsert(ctx context.Context, value *entity.ExternalToken) (*entity.ExternalToken, error) {
	now := time.Now()
	modelValue, err := models.ExternalTokens.Insert(
		&models.ExternalTokenSetter{
			UserID:         omit.From(value.UserID),
			Provider:       omit.From(value.Provider),
			EncryptedToken: omit.From(value.EncryptedToken),
			Nonce:          omit.From(value.Nonce),
			UpdatedAt:      omit.From(now),
		},
		im.OnConflict("user_id", "provider").DoUpdate(
			im.SetExcluded("encrypted_token", "nonce", "updated_at"),
		),
	).One(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("upsert external token: %w", err)
	}

	return toExternalTokenEntity(modelValue), nil
}

func (r *ExternalTokenRepository) FindByUserIDAndProvider(ctx context.Context, userID int64, provider string) (*entity.ExternalToken, error) {
	modelValue, err := models.ExternalTokens.Query(
		sm.Where(
			models.ExternalTokens.Columns.UserID.EQ(psql.Arg(userID)).And(
				models.ExternalTokens.Columns.Provider.EQ(psql.Arg(provider)),
			),
		),
	).One(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find external token: %w", err)
	}

	return toExternalTokenEntity(modelValue), nil
}

func (r *ExternalTokenRepository) DeleteByUserIDAndProvider(ctx context.Context, userID int64, provider string) error {
	_, err := models.ExternalTokens.Delete(
		dm.Where(
			models.ExternalTokens.Columns.UserID.EQ(psql.Arg(userID)).And(
				models.ExternalTokens.Columns.Provider.EQ(psql.Arg(provider)),
			),
		),
	).Exec(ctx, r.db)
	if err != nil {
		return fmt.Errorf("delete external token: %w", err)
	}

	return nil
}

func (r *ExternalTokenRepository) ListByUserID(ctx context.Context, userID int64) ([]*entity.ExternalToken, error) {
	modelValues, err := models.ExternalTokens.Query(
		sm.Where(models.ExternalTokens.Columns.UserID.EQ(psql.Arg(userID))),
		sm.OrderBy(models.ExternalTokens.Columns.Provider).Asc(),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("list external tokens: %w", err)
	}

	result := make([]*entity.ExternalToken, 0, len(modelValues))
	for _, v := range modelValues {
		result = append(result, toExternalTokenEntity(v))
	}

	return result, nil
}

func toExternalTokenEntity(v *models.ExternalToken) *entity.ExternalToken {
	if v == nil {
		return nil
	}

	return &entity.ExternalToken{
		ID:             v.ID,
		UserID:         v.UserID,
		Provider:       v.Provider,
		EncryptedToken: v.EncryptedToken,
		Nonce:          v.Nonce,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}
