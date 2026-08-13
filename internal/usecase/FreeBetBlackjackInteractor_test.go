package usecase

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newFreeBetInteractorForTest() (*interfaces.MockFreeBetBlackjackGame,
	*presenter.MockFreeBetBlackjackPresenter, *FreeBetBlackjackInteractor,
) {
	mg := new(interfaces.MockFreeBetBlackjackGame)
	mp := new(presenter.MockFreeBetBlackjackPresenter)
	return mg, mp, NewFreeBetBlackjackInteractor(mg, mp)
}

func TestNewFreeBetBlackjackInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockFreeBetBlackjackPresenter)
	assert.Panics(t, func() { NewFreeBetBlackjackInteractor(nil, mp) })

	mg := new(interfaces.MockFreeBetBlackjackGame)
	assert.Panics(t, func() { NewFreeBetBlackjackInteractor(mg, nil) })
}

func TestFreeBetInteractor_Reset(t *testing.T) {
	mg, mp, ci := newFreeBetInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。**
func TestFreeBetInteractor_ActionsReachTheDomain(t *testing.T) {
	for _, tt := range []struct {
		name   string
		setup  func(*interfaces.MockFreeBetBlackjackGame)
		invoke func(*FreeBetBlackjackInteractor) string
		method string
		args   []any
	}{
		{
			name:   "PlaceBet",
			setup:  func(m *interfaces.MockFreeBetBlackjackGame) { m.On("PlaceBet", 50).Return(nil) },
			invoke: func(ci *FreeBetBlackjackInteractor) string { return ci.PlaceBet(50) },
			method: "PlaceBet", args: []any{50},
		},
		{
			name:   "Hit",
			setup:  func(m *interfaces.MockFreeBetBlackjackGame) { m.On("Hit").Return(nil) },
			invoke: func(ci *FreeBetBlackjackInteractor) string { return ci.Hit() },
			method: "Hit", args: nil,
		},
		{
			name:   "Stand",
			setup:  func(m *interfaces.MockFreeBetBlackjackGame) { m.On("Stand").Return(nil) },
			invoke: func(ci *FreeBetBlackjackInteractor) string { return ci.Stand() },
			method: "Stand", args: nil,
		},
		{
			name:   "FreeDouble",
			setup:  func(m *interfaces.MockFreeBetBlackjackGame) { m.On("FreeDouble").Return(nil) },
			invoke: func(ci *FreeBetBlackjackInteractor) string { return ci.FreeDouble() },
			method: "FreeDouble", args: nil,
		},
		{
			name:   "FreeSplit",
			setup:  func(m *interfaces.MockFreeBetBlackjackGame) { m.On("FreeSplit").Return(nil) },
			invoke: func(ci *FreeBetBlackjackInteractor) string { return ci.FreeSplit() },
			method: "FreeSplit", args: nil,
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockFreeBetBlackjackGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *FreeBetBlackjackInteractor) string { return ci.NextRound() },
			method: "NextRound", args: nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newFreeBetInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(ci))
			mg.AssertCalled(t, tt.method, tt.args...)
		})
	}
}

// **無料ダブル / 無料スプリットの可否はドメインだけが持つ。**
//
// ハードの 9-11 か、10 点札のペアでないか、といった条件をインタラクタが判定し直すと
// 規則が 2 か所に増える。エラーは握りつぶさずそのまま届ける。
func TestFreeBetInteractor_RuleChecksStayInTheDomain(t *testing.T) {
	mg, mp, ci := newFreeBetInteractorForTest()
	dblErr := errors.New("この手札は無料ダブルできません")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("FreeDouble").Return(dblErr)
	mp.On("Output", mg, dblErr).Return("error output")

	assert.Equal(t, "error output", ci.FreeDouble())
	mp.AssertCalled(t, "Output", mg, dblErr)
	mg.AssertNotCalled(t, "CanFreeDouble")
	mg.AssertNotCalled(t, "CanFreeSplit")
}

// **終局後の操作はドメインまで届かない。**
func TestFreeBetInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*FreeBetBlackjackInteractor) string
		method string
	}{
		{"PlaceBet", func(ci *FreeBetBlackjackInteractor) string { return ci.PlaceBet(50) }, "PlaceBet"},
		{"Hit", func(ci *FreeBetBlackjackInteractor) string { return ci.Hit() }, "Hit"},
		{"Stand", func(ci *FreeBetBlackjackInteractor) string { return ci.Stand() }, "Stand"},
		{"FreeDouble", func(ci *FreeBetBlackjackInteractor) string { return ci.FreeDouble() }, "FreeDouble"},
		{"FreeSplit", func(ci *FreeBetBlackjackInteractor) string { return ci.FreeSplit() }, "FreeSplit"},
		{"NextRound", func(ci *FreeBetBlackjackInteractor) string { return ci.NextRound() }, "NextRound"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newFreeBetInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(ci))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestFreeBetInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, ci := newFreeBetInteractorForTest()
	cfg := domain.DefaultFreeBetBlackjackConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint output", ci.Hint())
	assert.Equal(t, "log output", ci.ActionLog())
}

func TestFreeBetInteractor_ResetWithConfig(t *testing.T) {
	t.Run("正しい設定は通る", func(t *testing.T) {
		mg, mp, ci := newFreeBetInteractorForTest()
		cfg := domain.FreeBetBlackjackConfig{InitialChips: 2000, DefaultAnte: 20}
		mg.On("SetConfig", cfg).Return()
		mg.On("Reset").Return()
		mp.On("Output", mg, nil).Return("reset output")

		assert.Equal(t, "reset output", ci.ResetWithConfig(cfg))
		mg.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("範囲外の設定は弾かれる", func(t *testing.T) {
		mg, mp, ci := newFreeBetInteractorForTest()
		bad := domain.FreeBetBlackjackConfig{
			InitialChips: domain.FreeBetChipsMax + 1, DefaultAnte: 50,
		}
		mp.On("Output", mg, mock.Anything).Return("bad config")

		assert.NotEmpty(t, ci.ResetWithConfig(bad))
		mg.AssertNotCalled(t, "SetConfig")
		mg.AssertNotCalled(t, "Reset")
	})
}

// **保存して読み直しても同じ盤面になる。** ハウス出資も保たれる。
func TestRestoreFreeBetBlackjackInteractor(t *testing.T) {
	mp := new(presenter.MockFreeBetBlackjackPresenter)

	game := domain.NewDefaultFreeBetBlackjack()
	game.Reset()
	require.NoError(t, game.PlaceBet(50))
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ci, err := RestoreFreeBetBlackjackInteractor(data, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetPhase(), ci.Game.GetPhase())
	assert.Equal(t, game.GetAnteBet(), ci.Game.GetAnteBet())
	assert.Equal(t, game.GetChips(), ci.Game.GetChips())
	assert.Equal(t, game.GetFreeBets(), ci.Game.GetFreeBets())

	snap, err := ci.Snapshot()
	require.NoError(t, err)
	again, err := RestoreFreeBetBlackjackInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetAnteBet(), again.Game.GetAnteBet())
}

func TestRestoreFreeBetBlackjackInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockFreeBetBlackjackPresenter)

	_, err := RestoreFreeBetBlackjackInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	_, err = RestoreFreeBetBlackjackInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
