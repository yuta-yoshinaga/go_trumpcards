//go:build test

package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// 不正な操作のエラーがロケールに追随すること。以前はドメインが英語で組み立てた
// 文を presenter が素通ししており、日本語ロケールでも英語が出ていた (#5556)。
func TestCuiErrorBlock_TranslatesCodedErrors(t *testing.T) {
	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "missmilligan.errNothingWaived", nil)

	render := func(lang string) string {
		i18n.SetLang(lang)
		defer i18n.SetLang("ja")
		var b strings.Builder
		cuiErrorBlock(&b, err)
		return b.String()
	}

	ja := render("ja")
	assert.Contains(t, ja, "取り置いたカードがありません")
	// キーがそのまま出ていたら、それは翻訳を素通りした証拠。
	assert.NotContains(t, ja, "missmilligan.")

	en := render("en")
	assert.Contains(t, en, "Nothing is waived")
	assert.NotContains(t, en, "取り置")
}

// コードを持たないエラーは今までどおり文言をそのまま出す。ここを翻訳に回すと
// 未定義キーとして文言自体が消える。
func TestCuiErrorBlock_LeavesUncodedErrorsAlone(t *testing.T) {
	var b strings.Builder
	cuiErrorBlock(&b, errors.New("plain failure"))
	assert.Contains(t, b.String(), "plain failure")
}

func TestCuiErrorBlock_WritesNothingWithoutAnError(t *testing.T) {
	var b strings.Builder
	cuiErrorBlock(&b, nil)
	assert.Empty(t, b.String())
}

// params は並びを固定する。map の反復順は毎回変わるので、そのまま流すと
// 同じエラーが実行ごとに違う文字列になる。
func TestI18nPairs_IsOrderStable(t *testing.T) {
	params := map[string]string{"b": "2", "a": "1", "c": "3"}
	for i := 0; i < 50; i++ {
		assert.Equal(t, []string{"a", "1", "b", "2", "c", "3"}, i18nPairs(params))
	}
	assert.Nil(t, i18nPairs(nil))
	assert.Nil(t, i18nPairs(map[string]string{}))
}
