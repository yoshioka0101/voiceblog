// Package validation は HTTP 層の OpenAPI リクエスト型を検証し、
// usecase の入力型へ変換する。レスポンス方向の presenter と対になる層で、
// ここを通ることで api パッケージの型がドメイン側へ漏れない。
package validation

import (
	"strconv"

	"github.com/yoshioka0101/voiceblog/backend/internal/apperr"
)

// IDParam はパスパラメータを正の int64 に変換する。
// ゼロ・負数・数値以外はこの層で弾く。
func IDParam(raw, name string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.BadRequestWithCode("invalid_id", "invalid "+name)
	}
	return id, nil
}
