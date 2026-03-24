package presenter

import (
	"github.com/yoshioka0101/voiceblog/backend/internal/api"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/article"
)

func Article(value *entity.Article) api.ArticleResponse {
	return api.ArticleResponse{
		Id:             value.ID,
		UserId:         value.UserID,
		PromptRunJobId: value.PromptRunJobID,
		Title:          value.Title,
		Content:        value.Content,
		DeletedAt:      value.DeletedAt,
		CreatedAt:      value.CreatedAt,
		UpdatedAt:      value.UpdatedAt,
	}
}

func Articles(values []*entity.Article) api.ArticlesResponse {
	responses := make(api.ArticlesResponse, 0, len(values))
	for _, value := range values {
		responses = append(responses, Article(value))
	}
	return responses
}
