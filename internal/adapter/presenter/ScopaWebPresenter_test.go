package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func buildScoredScopa(t *testing.T) *domain.Scopa {
	t.Helper()
	s := domain.NewDefaultScopa()
	s.Reset()
	return s
}

func TestScopaWebPresenter_OutputJSON(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	out := p.Output(s, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, k := range []string{"players", "tableCards", "phase", "config"} {
		if _, ok := parsed[k]; !ok {
			t.Errorf("missing key %q in output", k)
		}
	}
}

func TestScopaWebPresenter_HintOutput(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	// HintOutput mirrors Output (the GUI computes its own hint client-side).
	out := p.HintOutput(s)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("hint output is not valid JSON: %v", err)
	}
	if _, ok := parsed["phase"]; !ok {
		t.Errorf("missing phase key in hint output")
	}
}

func TestScopaWebPresenter_OutputError(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	out := p.Output(s, scopaAssertErrWeb{})
	if !strings.Contains(out, "kaboom") {
		t.Errorf("expected error message in JSON, got: %s", out)
	}
}

// **Web は messageCode を受け取らないと生の識別子を出す。**`NewDomainErrorCode` で
// 作ったエラーは `Message` が空なので `Error()` がキーを返し、`GameMessageBox` は
// `messageCode` が空だと翻訳を通さない ── ドメインを直しただけでは、同じ生文字列が
// 今度は Web に出る (#6846、Scopone #6457 で実際に踏んだ形)。
func TestScopaWebPresenter_ForwardsTheErrorMessageCode(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)

	decode := func(t *testing.T, err error) (string, string, map[string]string) {
		t.Helper()
		var got struct {
			Message       string            `json:"message"`
			MessageCode   string            `json:"messageCode"`
			MessageParams map[string]string `json:"messageParams"`
		}
		if uerr := json.Unmarshal([]byte(p.Output(s, err)), &got); uerr != nil {
			t.Fatalf("output not valid JSON: %v", uerr)
		}
		return got.Message, got.MessageCode, got.MessageParams
	}

	_, code, params := decode(t, domain.NewDomainErrorCode(domain.ErrInvalidCard,
		"scopa.errHandIndexOutOfRange", map[string]string{"idx": "7"}))
	assert.Equal(t, "scopa.errHandIndexOutOfRange", code)
	assert.Equal(t, map[string]string{"idx": "7"}, params)

	// パラメータの無いコードでも同じ経路を通る。
	_, code, _ = decode(t, domain.NewDomainErrorCode(domain.ErrInvalidPlay, "scopa.errCaptureRequired", nil))
	assert.Equal(t, "scopa.errCaptureRequired", code)

	// **コードを持たないエラーは空のまま。**ここを緩めると「常に何か入れる」実装が
	// 通り、訳の無い文字列をキーとして送ってしまう。
	msg, code, _ := decode(t, scopaAssertErrWeb{})
	assert.Equal(t, "kaboom", msg)
	assert.Empty(t, code)
}

type scopaAssertErrWeb struct{}

func (scopaAssertErrWeb) Error() string { return "kaboom" }

func TestScopaWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScopaWebPresenter{}
	s := buildScoredScopa(t)
	out := p.ActionLogOutput(s)
	var parsed any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Errorf("action log output not valid JSON: %v", err)
	}
}
