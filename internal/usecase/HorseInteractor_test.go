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

const horseMockOutput = `{"phase":0}`

// horsePassThrough は本物のドメインと組み合わせて使う素通しプレゼンター。
type horsePassThrough struct{}

func (p *horsePassThrough) Output(_ interfaces.HorseGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}

func (p *horsePassThrough) ActionLogOutput(g interfaces.HorseGame) string {
	if len(g.GetActionLog()) == 0 {
		return "empty"
	}
	return "log"
}

func (p *horsePassThrough) HintOutput(g interfaces.HorseGame) string {
	return "hint:" + g.GetDisciplineLetter()
}

// newHorseReal は本物のドメインを載せたインタラクターを返す。
func newHorseReal() (*usecase.HorseInteractor, *domain.Horse) {
	g := domain.NewDefaultHorse()
	return usecase.NewHorseInteractor(g, &horsePassThrough{}), g
}

// horseFoldOut は人間がフォールドしてハンドを閉じる。
func horseFoldOut(t *testing.T, hi *usecase.HorseInteractor, g *domain.Horse) {
	t.Helper()
	for range 40 {
		if g.GetGameEndFlag() || g.GetPhase() != domain.HorsePhaseHand {
			return
		}
		if out := hi.Action(domain.HoldemActionFold, 0, 0); out != "ok" {
			return
		}
	}
	t.Fatal("ハンドが閉じなかった")
}

func TestNewHorseInteractor_NilGuards(t *testing.T) {
	hp := new(presenter.MockHorsePresenter)
	assert.PanicsWithValue(t, "HorseInteractor: g must not be nil", func() {
		usecase.NewHorseInteractor(nil, hp)
	})
	gm := new(interfaces.MockHorseGame)
	assert.PanicsWithValue(t, "HorseInteractor: hp must not be nil", func() {
		usecase.NewHorseInteractor(gm, nil)
	})
}

func TestHorseInteractor_Action_Error(t *testing.T) {
	gm := new(interfaces.MockHorseGame)
	hp := new(presenter.MockHorsePresenter)
	hp.On("Output", mock.Anything, mock.Anything).Return(horseMockOutput)
	gm.On("GetGameEndFlag").Return(false)
	gm.On("PlayerAction", 0, 0, 0).Return(assert.AnError)

	hi := usecase.NewHorseInteractor(gm, hp)
	assert.Equal(t, horseMockOutput, hi.Action(0, 0, 0))
}

func TestHorseInteractor_BlockedAfterGameEnd(t *testing.T) {
	gm := new(interfaces.MockHorseGame)
	hp := new(presenter.MockHorsePresenter)
	hp.On("Output", mock.Anything, mock.Anything).Return(horseMockOutput)
	gm.On("GetGameEndFlag").Return(true)

	hi := usecase.NewHorseInteractor(gm, hp)
	assert.Equal(t, horseMockOutput, hi.Action(0, 0, 0))
	assert.Equal(t, horseMockOutput, hi.NextHand())
	gm.AssertNotCalled(t, "PlayerAction", mock.Anything, mock.Anything, mock.Anything)
	gm.AssertNotCalled(t, "NextHand")
}

func TestHorseInteractor_ResetWithConfig_Invalid(t *testing.T) {
	gm := new(interfaces.MockHorseGame)
	hp := new(presenter.MockHorsePresenter)
	hp.On("Output", mock.Anything, mock.Anything).Return(horseMockOutput)
	hi := usecase.NewHorseInteractor(gm, hp)

	// **3 席は種目が受け付けない卓サイズ。**
	assert.Equal(t, horseMockOutput, hi.ResetWithConfig(domain.HorseConfig{
		Seats: 3, InitialChips: domain.HorseDefaultChips, HandsPerDiscipline: 1,
	}))
	gm.AssertNotCalled(t, "Reset")
	gm.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestHorseInteractor_HintAndLog(t *testing.T) {
	gm := new(interfaces.MockHorseGame)
	hp := new(presenter.MockHorsePresenter)
	hp.On("HintOutput", mock.Anything).Return("hint")
	hp.On("ActionLogOutput", mock.Anything).Return("log")
	hi := usecase.NewHorseInteractor(gm, hp)
	assert.Equal(t, "hint", hi.Hint())
	assert.Equal(t, "log", hi.ActionLog())
}

// **Reset は人間が打てる場面から始まる。** ハンドが始まっていないと、画面には
// 押せるものが無いのに人間の番として出る。
func TestHorseInteractor_Reset_StartsAHand(t *testing.T) {
	hi, g := newHorseReal()
	require.Equal(t, "ok", hi.Reset())
	assert.Equal(t, domain.HorsePhaseHand, g.GetPhase())
	assert.Equal(t, 1, g.GetHandNumber())
	assert.Equal(t, domain.HorseHoldem, g.GetDiscipline(), "H から始まっていない")
}

// **チップは種目をまたいで持ち回る。** インタラクター越しでも同じ。
func TestHorseInteractor_CarriesChipsAcrossHands(t *testing.T) {
	hi, g := newHorseReal()
	require.Equal(t, "ok", hi.Reset())

	total := 0
	for i := range g.GetSeatCount() {
		total += g.GetSeatChips(i)
	}
	for range 6 {
		if g.GetGameEndFlag() {
			break
		}
		horseFoldOut(t, hi, g)
		if g.GetGameEndFlag() {
			break
		}
		require.Equal(t, "ok", hi.NextHand())
		after := 0
		for i := range g.GetSeatCount() {
			after += g.GetSeatChips(i)
		}
		assert.Equal(t, total, after, "ハンドをまたいで総量が変わった")
	}
}

func TestHorseInteractor_SnapshotRoundTrip(t *testing.T) {
	hi, g := newHorseReal()
	require.Equal(t, "ok", hi.Reset())
	require.Equal(t, "ok", hi.Action(domain.HoldemActionFold, 0, 0))

	data, err := hi.Snapshot()
	require.NoError(t, err)
	var probe map[string]any
	require.NoError(t, json.Unmarshal(data, &probe))

	restored, err := usecase.RestoreHorseInteractor(data, &horsePassThrough{})
	require.NoError(t, err)
	assert.Equal(t, hi.GetConfig(), restored.GetConfig())
	// **復元した卓で打ち続けられる。**
	assert.Contains(t, []string{"ok", "err:horse: not allowed in this phase"},
		restored.Action(domain.HoldemActionFold, 0, 0))
	_ = g
}

func TestHorseInteractor_RestoreRejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreHorseInteractor([]byte(`{`), &horsePassThrough{})
	assert.Error(t, err)
}
