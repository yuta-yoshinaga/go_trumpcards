//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const trogguMockOutput = `{"phase":0}`

// trogguPassThrough は本物のドメインと組み合わせて使う素通しプレゼンター。
type trogguPassThrough struct{}

func (p *trogguPassThrough) Output(_ interfaces.TrogguGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}

func (p *trogguPassThrough) ActionLogOutput(g interfaces.TrogguGame) string {
	if len(g.GetActionLog()) == 0 {
		return "empty"
	}
	return "log"
}

func (p *trogguPassThrough) HintOutput(g interfaces.TrogguGame) string {
	if h := g.GetHint(); h != nil {
		return "hint:" + h.Reason
	}
	return "hint:none"
}

// newTrogguReal は本物のドメインを載せたインタラクターを返す。
func newTrogguReal(deals int) (*usecase.TrogguInteractor, *domain.Troggu) {
	players := make([]*domain.TrogguPlayer, domain.TrogguPlayerCnt)
	players[0] = domain.NewTrogguPlayer(true)
	for i := 1; i < domain.TrogguPlayerCnt; i++ {
		players[i] = domain.NewTrogguPlayer(false)
	}
	g := domain.NewTroggu(players, domain.TrogguConfig{TargetDeals: deals})
	return usecase.NewTrogguInteractor(g, &trogguPassThrough{}), g
}

func TestNewTrogguInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockTrogguPresenter)
	assert.PanicsWithValue(t, "TrogguInteractor: g must not be nil", func() {
		usecase.NewTrogguInteractor(nil, tp)
	})
	gm := new(interfaces.MockTrogguGame)
	assert.PanicsWithValue(t, "TrogguInteractor: tp must not be nil", func() {
		usecase.NewTrogguInteractor(gm, nil)
	})
}

func TestTrogguInteractor_Bid_Error(t *testing.T) {
	gm := new(interfaces.MockTrogguGame)
	tp := new(presenter.MockTrogguPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(trogguMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerBid", domain.TrogguBidSolo).Return(assert.AnError)

	ti := usecase.NewTrogguInteractor(gm, tp)
	assert.Equal(t, trogguMockOutput, ti.Bid(domain.TrogguBidSolo))
	gm.AssertNotCalled(t, "CpuBid")
}

func TestTrogguInteractor_BlockedAfterGameEnd(t *testing.T) {
	gm := new(interfaces.MockTrogguGame)
	tp := new(presenter.MockTrogguPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(trogguMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	ti := usecase.NewTrogguInteractor(gm, tp)
	assert.Equal(t, trogguMockOutput, ti.Bid(domain.TrogguBidSolo))
	assert.Equal(t, trogguMockOutput, ti.Pass())
	assert.Equal(t, trogguMockOutput, ti.Play(0))
	gm.AssertNotCalled(t, "PlayerBid", mock.Anything)
	gm.AssertNotCalled(t, "PlayerPass")
	gm.AssertNotCalled(t, "PlayerPlayCard", mock.Anything)
}

func TestTrogguInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gm := new(interfaces.MockTrogguGame)
	tp := new(presenter.MockTrogguPresenter)
	tp.On("Output", mock.Anything, mock.Anything).Return(trogguMockOutput)
	ti := usecase.NewTrogguInteractor(gm, tp)

	assert.Equal(t, trogguMockOutput, ti.ResetWithConfig(domain.TrogguConfig{TargetDeals: 0}))
	gm.AssertNotCalled(t, "Reset")
	gm.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestTrogguInteractor_HintAndLog(t *testing.T) {
	gm := new(interfaces.MockTrogguGame)
	tp := new(presenter.MockTrogguPresenter)
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")
	ti := usecase.NewTrogguInteractor(gm, tp)
	assert.Equal(t, "hint", ti.Hint())
	assert.Equal(t, "log", ti.ActionLog())
}

// **Reset は人間の入力が要る場面まで進める。**
func TestTrogguInteractor_Reset_StopsWhereTheHumanMustAct(t *testing.T) {
	ti, g := newTrogguReal(1)
	require.Equal(t, "ok", ti.Reset())
	assert.True(t,
		g.IsHumanTurn() || g.GetPhase() == domain.TrogguPhaseTrickEnd ||
			g.GetPhase() == domain.TrogguPhaseRoundEnd || g.GetGameEndFlag(),
		"人間の入力を待つ場面で止まっていない (phase=%d)", g.GetPhase())
	assert.Equal(t, domain.TrogguConfig{TargetDeals: 1}, ti.GetConfig())
}

// マッチを最後まで進められる。
func TestTrogguInteractor_PlaysThroughToTheEnd(t *testing.T) {
	ti, g := newTrogguReal(1)
	require.Equal(t, "ok", ti.Reset())

	for range 3000 {
		if g.GetGameEndFlag() {
			break
		}
		switch g.GetPhase() {
		case domain.TrogguPhaseTrickEnd:
			ti.NextTrick()
		case domain.TrogguPhaseRoundEnd:
			ti.NextRound()
		case domain.TrogguPhaseBid:
			ti.Pass()
		case domain.TrogguPhasePlay:
			h := g.GetHint()
			require.NotNil(t, h)
			require.NotNil(t, h.CardIndex)
			ti.Play(*h.CardIndex)
		default:
			t.Fatalf("想定外のフェーズ %d", g.GetPhase())
		}
	}
	require.True(t, g.GetGameEndFlag(), "終局しなかった")
	// 終局後の入力は盤面を動かさない。
	assert.Equal(t, "ok", ti.Play(0))
	assert.Equal(t, "ok", ti.NextRound())
	assert.True(t, g.GetGameEndFlag())
}

// **トリック終了では止める。**
func TestTrogguInteractor_StopsAtTrickEnd(t *testing.T) {
	ti, g := newTrogguReal(1)
	require.Equal(t, "ok", ti.Reset())
	for range 50 {
		if g.GetPhase() != domain.TrogguPhaseBid {
			break
		}
		ti.Pass()
	}
	if g.GetPhase() != domain.TrogguPhasePlay {
		t.Skip("この配りではプレイフェーズに届かなかった")
	}

	h := g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	require.Equal(t, "ok", ti.Play(*h.CardIndex))
	if g.GetPhase() != domain.TrogguPhaseTrickEnd {
		return
	}
	assert.Len(t, g.GetLastTrickCards(), domain.TrogguPlayerCnt)
	before := g.GetTrickNumber()
	assert.Equal(t, "ok", ti.NextTrick())
	assert.Greater(t, g.GetTrickNumber(), before, "次のトリックへ進んでいない")
}

func TestTrogguInteractor_SnapshotRoundTrip(t *testing.T) {
	ti, g := newTrogguReal(2)
	require.Equal(t, "ok", ti.Reset())
	for range 20 {
		if g.GetPhase() != domain.TrogguPhaseBid {
			break
		}
		ti.Pass()
	}

	data, err := ti.Snapshot()
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(data, &probe))

	restored, err := usecase.RestoreTrogguInteractor(data, &trogguPassThrough{})
	require.NoError(t, err)
	assert.Equal(t, ti.GetConfig(), restored.GetConfig())
	assert.Equal(t, "ok", restored.NextTrick())
}

func TestTrogguInteractor_RestoreRejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreTrogguInteractor([]byte(`{`), &trogguPassThrough{})
	assert.Error(t, err)
}
