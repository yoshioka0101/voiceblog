package usecase

import (
	"context"
	"fmt"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
	entity "github.com/yoshioka0101/voiceblog/backend/internal/entity/externaltoken"
	"github.com/yoshioka0101/voiceblog/backend/internal/infra/crypto"
)

var supportedProviders = map[string]bool{
	"qiita":  true,
	"hatena": true,
}

// Verifier checks whether a token is valid for a given provider.
type Verifier interface {
	Verify(ctx context.Context, token string) error
}

type IntegrationStatus struct {
	Provider  string
	Connected bool
}

type UseCase struct {
	repo      entity.Repository
	encryptor *crypto.AESEncryptor
	verifiers map[string]Verifier
}

func NewUseCase(repo entity.Repository, encryptor *crypto.AESEncryptor, verifiers map[string]Verifier) *UseCase {
	return &UseCase{
		repo:      repo,
		encryptor: encryptor,
		verifiers: verifiers,
	}
}

func (uc *UseCase) StoreToken(ctx context.Context, userID int64, provider, plaintext string) error {
	if !supportedProviders[provider] {
		return apperr.BadRequest(fmt.Sprintf("unsupported provider: %s", provider))
	}
	if plaintext == "" {
		return apperr.BadRequest("token must not be empty")
	}

	// Verify token before storing
	if v, ok := uc.verifiers[provider]; ok {
		if err := v.Verify(ctx, plaintext); err != nil {
			return apperr.BadRequest(fmt.Sprintf("token verification failed: %s", err.Error()))
		}
	}

	ciphertext, nonce, err := uc.encryptor.Encrypt([]byte(plaintext))
	if err != nil {
		return apperr.InternalError("failed to encrypt token")
	}

	_, err = uc.repo.Upsert(ctx, &entity.ExternalToken{
		UserID:         userID,
		Provider:       provider,
		EncryptedToken: ciphertext,
		Nonce:          nonce,
	})
	if err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	return nil
}

func (uc *UseCase) DeleteToken(ctx context.Context, userID int64, provider string) error {
	if !supportedProviders[provider] {
		return apperr.BadRequest(fmt.Sprintf("unsupported provider: %s", provider))
	}

	return uc.repo.DeleteByUserIDAndProvider(ctx, userID, provider)
}

func (uc *UseCase) GetIntegrations(ctx context.Context, userID int64) ([]IntegrationStatus, error) {
	tokens, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}

	connected := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		connected[t.Provider] = true
	}

	result := make([]IntegrationStatus, 0, len(supportedProviders))
	for provider := range supportedProviders {
		result = append(result, IntegrationStatus{
			Provider:  provider,
			Connected: connected[provider],
		})
	}

	return result, nil
}

func (uc *UseCase) DecryptToken(ctx context.Context, userID int64, provider string) (string, error) {
	token, err := uc.repo.FindByUserIDAndProvider(ctx, userID, provider)
	if err != nil {
		return "", fmt.Errorf("find token: %w", err)
	}
	if token == nil {
		return "", apperr.BadRequest(fmt.Sprintf("%s is not connected", provider))
	}

	plaintext, err := uc.encryptor.Decrypt(token.EncryptedToken, token.Nonce)
	if err != nil {
		return "", apperr.InternalError("failed to decrypt token")
	}

	return string(plaintext), nil
}
