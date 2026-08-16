//go:build test

package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// ドメインのエラーは文言ではなくキーで運ぶ。以前は errors.New で完成した文字列を
// 返しており、presenter がそれを素通しにするので、どのロケールでも書いたときの
// 言語のまま出ていた (MissMilligan は英語、他ゲームは日本語) — #5556。
func TestDomainErrorCode_CarriesAKeyNotAPhrase(t *testing.T) {
	err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "missmilligan.errNotDescendingRun", nil)

	if !errors.Is(err, domain.ErrInvalidPlay) {
		t.Error("errors.Is は sentinel を見えたままにすること")
	}
	if got := err.MessageCode(); got != "missmilligan.errNotDescendingRun" {
		t.Errorf("MessageCode() = %q", got)
	}
	// Error() は失っても困るので、キーそのものを返す。ログに出たときに何の
	// エラーか分かり、かつ「翻訳されていない」ことが一目で分かる。
	if err.Error() == "" {
		t.Error("Error() が空だと、コードを読まない経路で無言になる")
	}
}

func TestDomainErrorCode_CarriesParams(t *testing.T) {
	err := domain.NewDomainErrorCode(domain.ErrInvalidCard, "missmilligan.errBadColumn",
		map[string]string{"col": "9"})
	if got := err.MessageParams()["col"]; got != "9" {
		t.Errorf("params[col] = %q", got)
	}
}

// 文言を持つ従来の DomainError は、コードを持たないと明確に答えること。
// ここが空でないと presenter が「翻訳できる」と誤解して未定義キーを引く。
func TestDomainError_WithoutCodeReportsNone(t *testing.T) {
	err := domain.NewDomainError(domain.ErrInvalidPlay, "無効です")
	if got := err.MessageCode(); got != "" {
		t.Errorf("MessageCode() = %q, want empty", got)
	}
	// 文言形式は今までどおり文言を返す。ここが壊れると、変換していない
	// 全ゲームのエラー表示が一斉に消える。
	if got := err.Error(); got != "無効です" {
		t.Errorf("Error() = %q", got)
	}
}

// 素の error はコードを持たない。presenter はこの判定で分岐する。
func TestDomainErrorCodeOf_PlainErrorHasNoCode(t *testing.T) {
	code, params := domain.ErrorMessageCode(errors.New("plain"))
	if code != "" || params != nil {
		t.Errorf("plain error should carry no code, got %q %v", code, params)
	}
	code, _ = domain.ErrorMessageCode(nil)
	if code != "" {
		t.Errorf("nil error should carry no code, got %q", code)
	}
}

// 文言形式の DomainError は、キーを訊かれても「無い」と答えること。ここが
// コードを返すと presenter が文言を i18n キーとして引き、画面から消える。
func TestErrorMessageCode_PhraseFormHasNoCode(t *testing.T) {
	code, params := domain.ErrorMessageCode(domain.NewDomainError(domain.ErrInvalidPlay, "無効です"))
	if code != "" || params != nil {
		t.Errorf("phrase-form DomainError should carry no code, got %q %v", code, params)
	}
}

// 包まれていても取り出せること。usecase 層は err をそのまま渡すが、将来
// fmt.Errorf("%w") で包む経路が増えても presenter の分岐は動く必要がある。
func TestErrorMessageCode_SeesThroughWrapping(t *testing.T) {
	inner := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "missmilligan.errColumnEmpty", nil)
	code, _ := domain.ErrorMessageCode(fmt.Errorf("while moving: %w", inner))
	if code != "missmilligan.errColumnEmpty" {
		t.Errorf("wrapped code = %q", code)
	}
}

// 文言もキーも無い DomainError は空文字を返す。ここで panic すると、
// ゼロ値の DomainError を作った経路がログごと落ちる。
func TestDomainError_EmptyReturnsEmptyString(t *testing.T) {
	if got := (&domain.DomainError{Sentinel: domain.ErrInvalidPlay}).Error(); got != "" {
		t.Errorf("Error() = %q, want empty", got)
	}
}
