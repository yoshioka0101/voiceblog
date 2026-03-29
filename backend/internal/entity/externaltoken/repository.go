package entity

import "context"

type Repository interface {
	StoreExternalToken(ctx context.Context, value *ExternalToken) (*ExternalToken, error)
	FindByUserIDAndProvider(ctx context.Context, userID int64, provider string) (*ExternalToken, error)
	DeleteByUserIDAndProvider(ctx context.Context, userID int64, provider string) error
	ListByUserID(ctx context.Context, userID int64) ([]*ExternalToken, error)
}
