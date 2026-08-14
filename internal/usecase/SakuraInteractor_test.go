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

const sakuraMockOutput = `{"phase":0}`

func newSakuraMocks() (*interfaces.MockSakuraGame, *presenter.MockSakuraPresenter) {
	return new(interfaces.MockSakuraGame), new(presenter.MockSakuraPresenter)
}

// sakuraPassThroughPresenter は本物のドメインと組み合わせて使う素通しプレゼンター。
type sakuraPassThroughPresenter struct{ lastErr error }

func (p *sakuraPassThroughPresenter) Output(_ interfaces.SakuraGame, lastErr error) string {
	p.lastErr = lastErr
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}

func (p *sakuraPassThroughPresenter) ActionLogOutput(g interfaces.SakuraGame) string {
	return "log:" + string(rune('0'+len(g.GetActionLog())%10))
}

func (p *sakuraPassThroughPresenter) HintOutput(g interfaces.SakuraGame) string {
	return "hint:" + g.GetHint().Reason
}

// newSakuraRealInteractor は本物のドメインを載せたインタラクターを返す。
func newSakuraRealInteractor() (*usecase.SakuraInteractor, *domain.Sakura, *sakuraPassThroughPresenter) {
	g := domain.NewDefaultSakura()
	p := &sakuraPassThroughPresenter{}
	return usecase.NewSakuraInteractor(g, p), g, p
}

func TestNewSakuraInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockSakuraPresenter)
	assert.PanicsWithValue(t, "SakuraInteractor: sg must not be nil", func() {
		usecase.NewSakuraInteractor(nil, cp)
	})
	gm := new(interfaces.MockSakuraGame)
	assert.PanicsWithValue(t, "SakuraInteractor: cp must not be nil", func() {
		usecase.NewSakuraInteractor(gm, nil)
	})
}

func TestSakuraInteractor_Play_Error(t *testing.T) {
	gm, cp := newSakuraMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(sakuraMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(true)
	gm.On("PlayerPlay", 0, -1).Return(assert.AnError)

	si := usecase.NewSakuraInteractor(gm, cp)
	assert.Equal(t, sakuraMockOutput, si.Play(0, -1))
	gm.AssertNotCalled(t, "CpuPlay")
}

func TestSakuraInteractor_Play_NotPlayable(t *testing.T) {
	gm, cp := newSakuraMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(sakuraMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("IsHumanTurn").Return(false)

	si := usecase.NewSakuraInteractor(gm, cp)
	assert.Equal(t, sakuraMockOutput, si.Play(0, -1))
	gm.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestSakuraInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gm, cp := newSakuraMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(sakuraMockOutput)
	si := usecase.NewSakuraInteractor(gm, cp)

	assert.Equal(t, sakuraMockOutput, si.ResetWithConfig(domain.SakuraConfig{Seats: 9, Rounds: 3}))
	gm.AssertNotCalled(t, "Reset")
	gm.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestSakuraInteractor_ResetWithConfig_AppliesSeats(t *testing.T) {
	si, g, _ := newSakuraRealInteractor()
	require.Equal(t, "ok", si.ResetWithConfig(domain.SakuraConfig{Seats: 4, Rounds: 2}))

	assert.Equal(t, 4, g.GetPlayerCnt(), "席数が反映されていない")
	assert.Equal(t, domain.SakuraConfig{Seats: 4, Rounds: 2}, si.GetConfig())
	// 席数を替えた局面も保存・復元できる。
	data, err := si.Snapshot()
	require.NoError(t, err)
	_, err = usecase.RestoreSakuraInteractor(data, &sakuraPassThroughPresenter{})
	assert.NoError(t, err)
}

func TestSakuraInteractor_HintAndLog(t *testing.T) {
	gm, cp := newSakuraMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")
	si := usecase.NewSakuraInteractor(gm, cp)
	assert.Equal(t, "hint", si.Hint())
	assert.Equal(t, "log", si.ActionLog())
}

// **Reset は人間の手番まで進める。** CPU が親の局面で止まると、画面には打てる手が
// 無いのに人間の入力を待つ盤面が出る。
func TestSakuraInteractor_Reset_StopsOnTheHumanTurn(t *testing.T) {
	si, g, _ := newSakuraRealInteractor()
	require.Equal(t, "ok", si.Reset())
	assert.True(t, g.IsHumanTurn(), "人間の手番で止まっていない")
	assert.Equal(t, domain.SakuraPhasePlay, g.GetPhase())
}

// 人間が 1 手打つと、次に人間の番が回るまで CPU が打ち切る。
func TestSakuraInteractor_Play_RunsTheCpusUntilTheHumanIsBack(t *testing.T) {
	si, g, _ := newSakuraRealInteractor()
	require.Equal(t, "ok", si.Reset())
	seat := g.HumanSeat()
	handBefore := g.GetPlayer(seat).GetCardsSize()
	cpuBefore := g.GetPlayer((seat + 1) % g.GetPlayerCnt()).GetCardsSize()

	h := g.GetHint()
	require.Equal(t, "ok", si.Play(h.CardIndex, h.FieldIndex))

	assert.Equal(t, handBefore-1, g.GetPlayer(seat).GetCardsSize())
	assert.Less(t, g.GetPlayer((seat+1)%g.GetPlayerCnt()).GetCardsSize(), cpuBefore,
		"CPU が打っていない")
	assert.True(t, g.IsHumanTurn() || g.GetPhase() != domain.SakuraPhasePlay)
}

// ラウンド終了で止まり、NextRound で配り直す。
func TestSakuraInteractor_NextRound_DealsAgain(t *testing.T) {
	si, g, _ := newSakuraRealInteractor()
	require.Equal(t, "ok", si.Reset())
	for range 200 {
		if g.GetPhase() != domain.SakuraPhasePlay {
			break
		}
		h := g.GetHint()
		require.Equal(t, "ok", si.Play(h.CardIndex, h.FieldIndex))
	}
	require.Equal(t, domain.SakuraPhaseRoundEnd, g.GetPhase(), "ラウンドが終わっていない")
	require.NotNil(t, g.GetLastResult())

	assert.Equal(t, "ok", si.NextRound())
	assert.Equal(t, 2, g.GetRound())
	assert.Equal(t, domain.SakuraHandSize, g.GetPlayer(0).GetCardsSize())
}

// 終局まで進めても入力を受け付け続ける (打ち止めで固まらない)。
func TestSakuraInteractor_PlaysThroughToTheEnd(t *testing.T) {
	si, g, _ := newSakuraRealInteractor()
	require.Equal(t, "ok", si.Reset())
	for range 2000 {
		if g.GetGameEndFlag() {
			break
		}
		if g.GetPhase() == domain.SakuraPhaseRoundEnd {
			si.NextRound()
			continue
		}
		h := g.GetHint()
		si.Play(h.CardIndex, h.FieldIndex)
	}
	require.True(t, g.GetGameEndFlag(), "終局していない")
	// 終局後の入力は盤面を動かさない。
	assert.Equal(t, "ok", si.Play(0, -1))
	assert.Equal(t, "ok", si.NextRound())
	assert.True(t, g.GetGameEndFlag())
}

func TestSakuraInteractor_SnapshotRoundTrip(t *testing.T) {
	si, g, _ := newSakuraRealInteractor()
	require.Equal(t, "ok", si.Reset())
	h := g.GetHint()
	require.Equal(t, "ok", si.Play(h.CardIndex, h.FieldIndex))

	data, err := si.Snapshot()
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(data, &probe))

	restored, err := usecase.RestoreSakuraInteractor(data, &sakuraPassThroughPresenter{})
	require.NoError(t, err)
	assert.Equal(t, si.GetConfig(), restored.GetConfig())
	// 復元した局面から続けられる。
	assert.Equal(t, "ok", restored.Reset())
}

func TestSakuraInteractor_RestoreRejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreSakuraInteractor([]byte(`{`), &sakuraPassThroughPresenter{})
	assert.Error(t, err)
}
