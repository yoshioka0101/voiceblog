package presenter

import (
	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/user"
)

func User(value *entity.User) api.UserResponse {
	return api.UserResponse{
		Id:           value.ID,
		Email:        stringOrNil(value.Email),
		Name:         stringOrNil(value.Name),
		AuthProvider: value.AuthProvider,
	}
}

func stringOrNil(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
