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

const zwanzigerrufenMockOutput = `{"phase":0}`

func newZwanzigerrufenMocks() (*interfaces.MockZwanzigerrufenGame, *presenter.MockZwanzigerrufenPresenter) {
	return new(interfaces.MockZwanzigerrufenGame), new(presenter.MockZwanzigerrufenPresenter)
}

// zwanzigerrufenPassThrough は本物のドメインと組み合わせて使う素通しプレゼンター。
type zwanzigerrufenPassThrough struct{}

func (p *zwanzigerrufenPassThrough) Output(_ interfaces.ZwanzigerrufenGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}

func (p *zwanzigerrufenPassThrough) ActionLogOutput(g interfaces.ZwanzigerrufenGame) string {
	if len(g.GetActionLog()) == 0 {
		return "empty"
	}
	return "log"
}

func (p *zwanzigerrufenPassThrough) HintOutput(g interfaces.ZwanzigerrufenGame) string {
	if h := g.GetHint(); h != nil {
		return "hint:" + h.Reason
	}
	return "hint:none"
}

// newZwanzigerrufenReal は本物のドメインを載せたインタラクターを返す。
func newZwanzigerrufenReal(deals int) (*usecase.ZwanzigerrufenInteractor, *domain.Zwanzigerrufen) {
	players := make([]*domain.ZwanzigerrufenPlayer, domain.ZwanzigerrufenPlayerCnt)
	players[0] = domain.NewZwanzigerrufenPlayer(true)
	for i := 1; i < domain.ZwanzigerrufenPlayerCnt; i++ {
		players[i] = domain.NewZwanzigerrufenPlayer(false)
	}
	g := domain.NewZwanzigerrufen(players, domain.ZwanzigerrufenConfig{TargetDeals: deals})
	return usecase.NewZwanzigerrufenInteractor(g, &zwanzigerrufenPassThrough{}), g
}

func TestNewZwanzigerrufenInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockZwanzigerrufenPresenter)
	assert.PanicsWithValue(t, "ZwanzigerrufenInteractor: g must not be nil", func() {
		usecase.NewZwanzigerrufenInteractor(nil, tp)
	})
	gm := new(interfaces.MockZwanzigerrufenGame)
	assert.PanicsWithValue(t, "ZwanzigerrufenInteractor: tp must not be nil", func() {
		usecase.NewZwanzigerrufenInteractor(gm, nil)
	})
}

func TestZwanzigerrufenInteractor_Bid_Error(t *testing.T) {
	gm, tp := newZwanzigerrufenMocks()
	tp.On("Output", mock.Anything, mock.Anything).Return(zwanzigerrufenMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerBid", domain.ZwanzigerrufenBidRufer).Return(assert.AnError)

	zi := usecase.NewZwanzigerrufenInteractor(gm, tp)
	assert.Equal(t, zwanzigerrufenMockOutput, zi.Bid(domain.ZwanzigerrufenBidRufer))
	gm.AssertNotCalled(t, "CpuBid")
}

func TestZwanzigerrufenInteractor_BlockedAfterGameEnd(t *testing.T) {
	gm, tp := newZwanzigerrufenMocks()
	tp.On("Output", mock.Anything, mock.Anything).Return(zwanzigerrufenMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	zi := usecase.NewZwanzigerrufenInteractor(gm, tp)
	assert.Equal(t, zwanzigerrufenMockOutput, zi.Bid(domain.ZwanzigerrufenBidRufer))
	assert.Equal(t, zwanzigerrufenMockOutput, zi.Pass())
	assert.Equal(t, zwanzigerrufenMockOutput, zi.Discard([]int{0}))
	assert.Equal(t, zwanzigerrufenMockOutput, zi.Play(0))
	gm.AssertNotCalled(t, "PlayerBid", mock.Anything)
	gm.AssertNotCalled(t, "PlayerPass")
	gm.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
	gm.AssertNotCalled(t, "PlayerPlayCard", mock.Anything)
}

func TestZwanzigerrufenInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gm, tp := newZwanzigerrufenMocks()
	tp.On("Output", mock.Anything, mock.Anything).Return(zwanzigerrufenMockOutput)
	zi := usecase.NewZwanzigerrufenInteractor(gm, tp)

	assert.Equal(t, zwanzigerrufenMockOutput,
		zi.ResetWithConfig(domain.ZwanzigerrufenConfig{TargetDeals: 0}))
	gm.AssertNotCalled(t, "Reset")
	gm.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestZwanzigerrufenInteractor_HintAndLog(t *testing.T) {
	gm, tp := newZwanzigerrufenMocks()
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")
	zi := usecase.NewZwanzigerrufenInteractor(gm, tp)
	assert.Equal(t, "hint", zi.Hint())
	assert.Equal(t, "log", zi.ActionLog())
}

// **Reset は人間の入力が要る場面まで進める。** CPU の入札で止まると、画面には
// 押せるものが無いのに人間の番として表示される。
func TestZwanzigerrufenInteractor_Reset_StopsWhereTheHumanMustAct(t *testing.T) {
	zi, g := newZwanzigerrufenReal(1)
	require.Equal(t, "ok", zi.Reset())

	assert.True(t, g.IsHumanTurn() || g.GetPhase() == domain.ZwanzigerrufenPhaseTrickEnd,
		"人間の入力を待つ場面で止まっていない (phase=%d)", g.GetPhase())
	assert.Equal(t, domain.ZwanzigerrufenConfig{TargetDeals: 1}, zi.GetConfig())
}

// **トリック終了では止める。** 出揃った 4 枚を見せずに次へ進めない。
func TestZwanzigerrufenInteractor_StopsAtTrickEnd(t *testing.T) {
	zi, g := newZwanzigerrufenReal(1)
	require.Equal(t, "ok", zi.Reset())
	zwanzigerrufenDriveToPlay(t, zi, g)
	if g.GetPhase() != domain.ZwanzigerrufenPhasePlay {
		t.Skip("この配りではプレイフェーズに届かなかった")
	}

	// 人間が出したあと、CPU が打ち切ってトリックが揃うところまで進む。
	h := g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	require.Equal(t, "ok", zi.Play(*h.CardIndex))

	if g.GetPhase() == domain.ZwanzigerrufenPhaseTrickEnd {
		assert.Len(t, g.GetLastTrickCards(), domain.ZwanzigerrufenPlayerCnt)
		before := g.GetTrickNumber()
		assert.Equal(t, "ok", zi.NextTrick())
		assert.Greater(t, g.GetTrickNumber(), before, "次のトリックへ進んでいない")
	}
}

// マッチを最後まで進められる。
func TestZwanzigerrufenInteractor_PlaysThroughToTheEnd(t *testing.T) {
	zi, g := newZwanzigerrufenReal(1)
	require.Equal(t, "ok", zi.Reset())

	for range 3000 {
		if g.GetGameEndFlag() {
			break
		}
		switch g.GetPhase() {
		case domain.ZwanzigerrufenPhaseTrickEnd:
			zi.NextTrick()
		case domain.ZwanzigerrufenPhaseRoundEnd:
			zi.NextRound()
		case domain.ZwanzigerrufenPhaseBid:
			zi.Pass()
		case domain.ZwanzigerrufenPhaseTalon:
			h := g.GetHint()
			require.NotNil(t, h)
			zi.Discard(h.DiscardIndices)
		case domain.ZwanzigerrufenPhasePlay:
			h := g.GetHint()
			require.NotNil(t, h)
			require.NotNil(t, h.CardIndex)
			zi.Play(*h.CardIndex)
		default:
			t.Fatalf("想定外のフェーズ %d", g.GetPhase())
		}
	}
	require.True(t, g.GetGameEndFlag(), "終局しなかった")
	assert.NotNil(t, g.GetBreakdown())
	// 終局後の入力は盤面を動かさない。
	assert.Equal(t, "ok", zi.Play(0))
	assert.Equal(t, "ok", zi.NextRound())
	assert.True(t, g.GetGameEndFlag())
}

// zwanzigerrufenDriveToPlay は人間がプレイフェーズに立つまで進める。
func zwanzigerrufenDriveToPlay(t *testing.T, zi *usecase.ZwanzigerrufenInteractor, g *domain.Zwanzigerrufen) {
	t.Helper()
	for range 200 {
		if g.GetGameEndFlag() || g.GetPhase() == domain.ZwanzigerrufenPhasePlay {
			return
		}
		switch g.GetPhase() {
		case domain.ZwanzigerrufenPhaseBid:
			zi.Pass()
		case domain.ZwanzigerrufenPhaseTalon:
			h := g.GetHint()
			require.NotNil(t, h)
			zi.Discard(h.DiscardIndices)
		case domain.ZwanzigerrufenPhaseTrickEnd:
			zi.NextTrick()
		default:
			return
		}
	}
}

func TestZwanzigerrufenInteractor_SnapshotRoundTrip(t *testing.T) {
	zi, g := newZwanzigerrufenReal(2)
	require.Equal(t, "ok", zi.Reset())
	zwanzigerrufenDriveToPlay(t, zi, g)

	data, err := zi.Snapshot()
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(data, &probe))

	restored, err := usecase.RestoreZwanzigerrufenInteractor(data, &zwanzigerrufenPassThrough{})
	require.NoError(t, err)
	assert.Equal(t, zi.GetConfig(), restored.GetConfig())
	assert.Equal(t, "ok", restored.NextTrick())
}

func TestZwanzigerrufenInteractor_RestoreRejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreZwanzigerrufenInteractor([]byte(`{`), &zwanzigerrufenPassThrough{})
	assert.Error(t, err)
}
