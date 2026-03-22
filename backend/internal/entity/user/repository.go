package user

import "context"

type Repository interface {
	FindOrCreate(ctx context.Context, provider, subject, email, name string) (*User, error)
}
