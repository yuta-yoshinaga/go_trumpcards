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

const tapptarockMockOutput = `{"phase":0}`

func newTappTarockMocks() (*interfaces.MockTappTarockGame, *presenter.MockTappTarockPresenter) {
	return new(interfaces.MockTappTarockGame), new(presenter.MockTappTarockPresenter)
}

// tapptarockPassThrough は本物のドメインと組み合わせて使う素通しプレゼンター。
type tapptarockPassThrough struct{}

func (p *tapptarockPassThrough) Output(_ interfaces.TappTarockGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}

func (p *tapptarockPassThrough) ActionLogOutput(g interfaces.TappTarockGame) string {
	if len(g.GetActionLog()) == 0 {
		return "empty"
	}
	return "log"
}

func (p *tapptarockPassThrough) HintOutput(g interfaces.TappTarockGame) string {
	if h := g.GetHint(); h != nil {
		return "hint:" + h.Reason
	}
	return "hint:none"
}

// newTappTarockReal は本物のドメインを載せたインタラクターを返す。
func newTappTarockReal(deals int) (*usecase.TappTarockInteractor, *domain.TappTarock) {
	players := make([]*domain.TappTarockPlayer, domain.TappTarockPlayerCnt)
	players[0] = domain.NewTappTarockPlayer(true)
	for i := 1; i < domain.TappTarockPlayerCnt; i++ {
		players[i] = domain.NewTappTarockPlayer(false)
	}
	g := domain.NewTappTarock(players, domain.TappTarockConfig{TargetDeals: deals})
	return usecase.NewTappTarockInteractor(g, &tapptarockPassThrough{}), g
}

func TestNewTappTarockInteractor_NilGuards(t *testing.T) {
	tp := new(presenter.MockTappTarockPresenter)
	assert.PanicsWithValue(t, "TappTarockInteractor: g must not be nil", func() {
		usecase.NewTappTarockInteractor(nil, tp)
	})
	gm := new(interfaces.MockTappTarockGame)
	assert.PanicsWithValue(t, "TappTarockInteractor: tp must not be nil", func() {
		usecase.NewTappTarockInteractor(gm, nil)
	})
}

func TestTappTarockInteractor_Bid_Error(t *testing.T) {
	gm, tp := newTappTarockMocks()
	tp.On("Output", mock.Anything, mock.Anything).Return(tapptarockMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerBid", domain.TappTarockBidDreier).Return(assert.AnError)

	zi := usecase.NewTappTarockInteractor(gm, tp)
	assert.Equal(t, tapptarockMockOutput, zi.Bid(domain.TappTarockBidDreier))
	gm.AssertNotCalled(t, "CpuBid")
}

func TestTappTarockInteractor_BlockedAfterGameEnd(t *testing.T) {
	gm, tp := newTappTarockMocks()
	tp.On("Output", mock.Anything, mock.Anything).Return(tapptarockMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	zi := usecase.NewTappTarockInteractor(gm, tp)
	assert.Equal(t, tapptarockMockOutput, zi.Bid(domain.TappTarockBidDreier))
	assert.Equal(t, tapptarockMockOutput, zi.Pass())
	assert.Equal(t, tapptarockMockOutput, zi.Discard([]int{0}))
	assert.Equal(t, tapptarockMockOutput, zi.Play(0))
	gm.AssertNotCalled(t, "PlayerBid", mock.Anything)
	gm.AssertNotCalled(t, "PlayerPass")
	gm.AssertNotCalled(t, "PlayerDiscard", mock.Anything)
	gm.AssertNotCalled(t, "PlayerPlayCard", mock.Anything)
}

func TestTappTarockInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gm, tp := newTappTarockMocks()
	tp.On("Output", mock.Anything, mock.Anything).Return(tapptarockMockOutput)
	zi := usecase.NewTappTarockInteractor(gm, tp)

	assert.Equal(t, tapptarockMockOutput,
		zi.ResetWithConfig(domain.TappTarockConfig{TargetDeals: 0}))
	gm.AssertNotCalled(t, "Reset")
	gm.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestTappTarockInteractor_HintAndLog(t *testing.T) {
	gm, tp := newTappTarockMocks()
	tp.On("HintOutput", mock.Anything).Return("hint")
	tp.On("ActionLogOutput", mock.Anything).Return("log")
	zi := usecase.NewTappTarockInteractor(gm, tp)
	assert.Equal(t, "hint", zi.Hint())
	assert.Equal(t, "log", zi.ActionLog())
}

// **Reset は人間の入力が要る場面まで進める。** CPU の入札で止まると、画面には
// 押せるものが無いのに人間の番として表示される。
func TestTappTarockInteractor_Reset_StopsWhereTheHumanMustAct(t *testing.T) {
	zi, g := newTappTarockReal(1)
	require.Equal(t, "ok", zi.Reset())

	assert.True(t, g.IsHumanTurn() || g.GetPhase() == domain.TappTarockPhaseTrickEnd,
		"人間の入力を待つ場面で止まっていない (phase=%d)", g.GetPhase())
	assert.Equal(t, domain.TappTarockConfig{TargetDeals: 1}, zi.GetConfig())
}

// **トリック終了では止める。** 出揃った 4 枚を見せずに次へ進めない。
func TestTappTarockInteractor_StopsAtTrickEnd(t *testing.T) {
	zi, g := newTappTarockReal(1)
	require.Equal(t, "ok", zi.Reset())
	tapptarockDriveToPlay(t, zi, g)
	if g.GetPhase() != domain.TappTarockPhasePlay {
		t.Skip("この配りではプレイフェーズに届かなかった")
	}

	// 人間が出したあと、CPU が打ち切ってトリックが揃うところまで進む。
	h := g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	require.Equal(t, "ok", zi.Play(*h.CardIndex))

	if g.GetPhase() == domain.TappTarockPhaseTrickEnd {
		assert.Len(t, g.GetLastTrickCards(), domain.TappTarockPlayerCnt)
		before := g.GetTrickNumber()
		assert.Equal(t, "ok", zi.NextTrick())
		assert.Greater(t, g.GetTrickNumber(), before, "次のトリックへ進んでいない")
	}
}

// マッチを最後まで進められる。
func TestTappTarockInteractor_PlaysThroughToTheEnd(t *testing.T) {
	zi, g := newTappTarockReal(1)
	require.Equal(t, "ok", zi.Reset())

	for range 3000 {
		if g.GetGameEndFlag() {
			break
		}
		switch g.GetPhase() {
		case domain.TappTarockPhaseTrickEnd:
			zi.NextTrick()
		case domain.TappTarockPhaseRoundEnd:
			zi.NextRound()
		case domain.TappTarockPhaseBid:
			zi.Pass()
		case domain.TappTarockPhaseTalon:
			h := g.GetHint()
			require.NotNil(t, h)
			zi.Discard(h.DiscardIndices)
		case domain.TappTarockPhasePlay:
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

// tapptarockDriveToPlay は人間がプレイフェーズに立つまで進める。
func tapptarockDriveToPlay(t *testing.T, zi *usecase.TappTarockInteractor, g *domain.TappTarock) {
	t.Helper()
	for range 200 {
		if g.GetGameEndFlag() || g.GetPhase() == domain.TappTarockPhasePlay {
			return
		}
		switch g.GetPhase() {
		case domain.TappTarockPhaseBid:
			zi.Pass()
		case domain.TappTarockPhaseTalon:
			h := g.GetHint()
			require.NotNil(t, h)
			zi.Discard(h.DiscardIndices)
		case domain.TappTarockPhaseTrickEnd:
			zi.NextTrick()
		default:
			return
		}
	}
}

func TestTappTarockInteractor_SnapshotRoundTrip(t *testing.T) {
	zi, g := newTappTarockReal(2)
	require.Equal(t, "ok", zi.Reset())
	tapptarockDriveToPlay(t, zi, g)

	data, err := zi.Snapshot()
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(data, &probe))

	restored, err := usecase.RestoreTappTarockInteractor(data, &tapptarockPassThrough{})
	require.NoError(t, err)
	assert.Equal(t, zi.GetConfig(), restored.GetConfig())
	assert.Equal(t, "ok", restored.NextTrick())
}

func TestTappTarockInteractor_RestoreRejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreTappTarockInteractor([]byte(`{`), &tapptarockPassThrough{})
	assert.Error(t, err)
}
