package usecase

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newColourWhistInteractorForTest() (*interfaces.MockColourWhistGame,
	*presenter.MockColourWhistPresenter, *ColourWhistInteractor,
) {
	mg := new(interfaces.MockColourWhistGame)
	mp := new(presenter.MockColourWhistPresenter)
	return mg, mp, NewColourWhistInteractor(mg, mp)
}

func TestNewColourWhistInteractor(t *testing.T) {
	_, _, ci := newColourWhistInteractorForTest()
	assert.NotNil(t, ci)
}

func TestNewColourWhistInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockColourWhistPresenter)
	assert.Panics(t, func() { NewColourWhistInteractor(nil, mp) })

	mg := new(interfaces.MockColourWhistGame)
	assert.Panics(t, func() { NewColourWhistInteractor(mg, nil) })
}

func TestColourWhistInteractor_Reset(t *testing.T) {
	mg, mp, ci := newColourWhistInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。** 引数はどれも int なので中身まで固定します。
func TestColourWhistInteractor_ActionsReachTheDomain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*interfaces.MockColourWhistGame)
		invoke func(*ColourWhistInteractor) string
		method string
		args   []any
	}{
		{
			name:   "Bid",
			setup:  func(m *interfaces.MockColourWhistGame) { m.On("Bid", domain.ColourWhistContractAlleen).Return(nil) },
			invoke: func(ci *ColourWhistInteractor) string { return ci.Bid(domain.ColourWhistContractAlleen) },
			method: "Bid", args: []any{domain.ColourWhistContractAlleen},
		},
		{
			// **パスも契約値のひとつ (0)。**
			name:   "Bid pass",
			setup:  func(m *interfaces.MockColourWhistGame) { m.On("Bid", domain.ColourWhistContractNone).Return(nil) },
			invoke: func(ci *ColourWhistInteractor) string { return ci.Bid(domain.ColourWhistContractNone) },
			method: "Bid", args: []any{domain.ColourWhistContractNone},
		},
		{
			name:   "Call",
			setup:  func(m *interfaces.MockColourWhistGame) { m.On("Call", domain.CardDesignDiamond).Return(nil) },
			invoke: func(ci *ColourWhistInteractor) string { return ci.Call(domain.CardDesignDiamond) },
			method: "Call", args: []any{domain.CardDesignDiamond},
		},
		{
			name:   "PlayCard",
			setup:  func(m *interfaces.MockColourWhistGame) { m.On("PlayCard", 6).Return(nil) },
			invoke: func(ci *ColourWhistInteractor) string { return ci.PlayCard(6) },
			method: "PlayCard", args: []any{6},
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockColourWhistGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *ColourWhistInteractor) string { return ci.NextRound() },
			method: "NextRound", args: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newColourWhistInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(ci))
			mg.AssertCalled(t, tt.method, tt.args...)
		})
	}
}

func TestColourWhistInteractor_ErrorsReachThePresenter(t *testing.T) {
	mg, mp, ci := newColourWhistInteractorForTest()
	bidErr := errors.New("競りで宣言できない契約です")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Bid", domain.ColourWhistContractTroel).Return(bidErr)
	mp.On("Output", mg, bidErr).Return("error output")

	// **Troel は競りで宣言できない。** ドメインのエラーがそのまま届きます。
	assert.Equal(t, "error output", ci.Bid(domain.ColourWhistContractTroel))
	mp.AssertCalled(t, "Output", mg, bidErr)
}

// **終局後の操作はドメインまで届かない。**
func TestColourWhistInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*ColourWhistInteractor) string
		method string
	}{
		{"Bid", func(ci *ColourWhistInteractor) string { return ci.Bid(1) }, "Bid"},
		{"Call", func(ci *ColourWhistInteractor) string { return ci.Call(1) }, "Call"},
		{"PlayCard", func(ci *ColourWhistInteractor) string { return ci.PlayCard(0) }, "PlayCard"},
		{"NextRound", func(ci *ColourWhistInteractor) string { return ci.NextRound() }, "NextRound"},
		{"GiveUp", func(ci *ColourWhistInteractor) string { return ci.GiveUp() }, "GiveUp"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newColourWhistInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(ci))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestColourWhistInteractor_GiveUp(t *testing.T) {
	mg, mp, ci := newColourWhistInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("GiveUp").Return()
	mp.On("Output", mg, nil).Return("gave up")

	assert.Equal(t, "gave up", ci.GiveUp())
	mg.AssertCalled(t, "GiveUp")
}

func TestColourWhistInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, ci := newColourWhistInteractorForTest()
	cfg := domain.DefaultColourWhistConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint output", ci.Hint())
	assert.Equal(t, "log output", ci.ActionLog())
}

// **保存して読み直しても同じ盤面になる。** Troel の状態も保たれます。
func TestRestoreColourWhistInteractor(t *testing.T) {
	mp := new(presenter.MockColourWhistPresenter)

	game := domain.NewDefaultColourWhist()
	game.Reset()
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ci, err := RestoreColourWhistInteractor(data, mp)
	require.NoError(t, err)
	require.NotNil(t, ci)
	assert.Equal(t, game.GetPhase(), ci.Game.GetPhase())
	assert.Equal(t, game.GetContract(), ci.Game.GetContract())
	assert.Equal(t, game.IsTroelForced(), ci.Game.IsTroelForced())

	snap, err := ci.Snapshot()
	require.NoError(t, err)
	again, err := RestoreColourWhistInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetTrumpSuit(), again.Game.GetTrumpSuit())
}

func TestRestoreColourWhistInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockColourWhistPresenter)

	_, err := RestoreColourWhistInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	_, err = RestoreColourWhistInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
