package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func spBuildScoredScopone(t *testing.T) *domain.Scopone {
	t.Helper()
	s := domain.NewDefaultScopone()
	s.Reset()
	return s
}

func TestScoponeWebPresenter_HintOutput(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	// HintOutput mirrors Output (the GUI computes its own hint client-side).
	var parsed map[string]any
	if err := json.Unmarshal([]byte(p.HintOutput(s)), &parsed); err != nil {
		t.Fatalf("hint output is not valid JSON: %v", err)
	}
	if _, ok := parsed["phase"]; !ok {
		t.Errorf("missing phase key in hint output")
	}
}

func TestScoponeWebPresenter_OutputJSON(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	out := p.Output(s, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, k := range []string{
		"players", "tableCards", "phase", "config", "teamScores",
		"roundNumber", "currentTurn", "dealerIdx", "winnerTeam", "isHumanTurn",
	} {
		if _, ok := parsed[k]; !ok {
			t.Errorf("missing key %q in output", k)
		}
	}
}

func TestScoponeWebPresenter_OutputError(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	out := p.Output(s, scoponeAssertErrWeb{})
	if !strings.Contains(out, "kaboom") {
		t.Errorf("expected error message in JSON, got: %s", out)
	}
}

// **Web は messageCode を受け取らないと生の識別子を出す。**`NewDomainErrorCode` で
// 作ったエラーは `Message` が空なので `Error()` がキーを返し、`GameMessageBox` は
// `messageCode` が空だと翻訳を通さない ── CUI を直しただけでは、同じ生文字列が
// 今度は Web に出る (#6457、レビュー指摘)。
func TestScoponeWebPresenter_ForwardsTheErrorMessageCode(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)

	err := domain.NewDomainErrorCode(domain.ErrInvalidCard, "scopone.errHandIndexOutOfRange",
		map[string]string{"idx": "7"})
	var got struct {
		Message       string            `json:"message"`
		MessageCode   string            `json:"messageCode"`
		MessageParams map[string]string `json:"messageParams"`
	}
	if uerr := json.Unmarshal([]byte(p.Output(s, err)), &got); uerr != nil {
		t.Fatalf("output not valid JSON: %v", uerr)
	}

	assert.Equal(t, "scopone.errHandIndexOutOfRange", got.MessageCode)
	assert.Equal(t, map[string]string{"idx": "7"}, got.MessageParams)

	// パラメータの無いコードでも同じ経路を通る。
	var plain struct {
		MessageCode string `json:"messageCode"`
	}
	plainOut := p.Output(s, domain.NewDomainErrorCode(domain.ErrInvalidPlay, "scopone.errCaptureRequired", nil))
	if uerr := json.Unmarshal([]byte(plainOut), &plain); uerr != nil {
		t.Fatalf("output not valid JSON: %v", uerr)
	}
	assert.Equal(t, "scopone.errCaptureRequired", plain.MessageCode)

	// **コードを持たないエラーは空のまま。**ここを緩めると「常に何か入れる」
	// 実装が通り、翻訳の無い文字列をキーとして送ってしまう。
	var generic struct {
		Message     string `json:"message"`
		MessageCode string `json:"messageCode"`
	}
	if uerr := json.Unmarshal([]byte(p.Output(s, scoponeAssertErrWeb{})), &generic); uerr != nil {
		t.Fatalf("output not valid JSON: %v", uerr)
	}
	assert.Equal(t, "kaboom", generic.Message)
	assert.Empty(t, generic.MessageCode)
}

type scoponeAssertErrWeb struct{}

func (scoponeAssertErrWeb) Error() string { return "kaboom" }

func TestScoponeWebPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spBuildScoredScopone(t)
	out := p.ActionLogOutput(s)
	var parsed any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Errorf("action log output not valid JSON: %v", err)
	}
}

// spPlayedOutScopone returns an all-CPU game driven to game end, exercising the
// roundEnd / gameEnd / score-detail / result-message render branches.
func spPlayedOutScopone(t *testing.T) *domain.Scopone {
	t.Helper()
	players := make([]*domain.ScopaPlayer, domain.ScoponePlayerCnt)
	for i := range players {
		players[i] = domain.NewScopaPlayer(false)
	}
	s := domain.NewScopone(domain.NewTrumpCardsScopa(), players, domain.DefaultScoponeConfig())
	s.Reset()
	for guard := 0; !s.GetGameEndFlag() && guard < 200000; guard++ {
		switch s.GetPhase() {
		case domain.ScoponePhasePlayerTurn:
			s.CpuPlay()
		case domain.ScoponePhaseRoundEnd:
			s.NextRound()
		}
	}
	return s
}

func TestScoponeWebPresenter_OutputGameEnd(t *testing.T) {
	p := &presenter.ScoponeWebPresenter{}
	s := spPlayedOutScopone(t)
	out := p.Output(s, nil)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["gameEndFlag"] != true {
		t.Errorf("expected gameEndFlag true at game end, got %v", parsed["gameEndFlag"])
	}
	if _, ok := parsed["lastRoundDetail"]; !ok {
		t.Errorf("expected lastRoundDetail key at game end")
	}
}
