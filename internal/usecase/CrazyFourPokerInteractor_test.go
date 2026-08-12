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

func newCrazyFourPokerInteractorForTest() (*interfaces.MockCrazyFourPokerGame,
	*presenter.MockCrazyFourPokerPresenter, *CrazyFourPokerInteractor,
) {
	mg := new(interfaces.MockCrazyFourPokerGame)
	mp := new(presenter.MockCrazyFourPokerPresenter)
	return mg, mp, NewCrazyFourPokerInteractor(mg, mp)
}

func TestNewCrazyFourPokerInteractor(t *testing.T) {
	_, _, ci := newCrazyFourPokerInteractorForTest()
	assert.NotNil(t, ci)
}

func TestNewCrazyFourPokerInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockCrazyFourPokerPresenter)
	assert.Panics(t, func() { NewCrazyFourPokerInteractor(nil, mp) })

	mg := new(interfaces.MockCrazyFourPokerGame)
	assert.Panics(t, func() { NewCrazyFourPokerInteractor(mg, nil) })
}

func TestCrazyFourPokerInteractor_Reset(t *testing.T) {
	mg, mp, ci := newCrazyFourPokerInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。**
//
// アンティも Queens Up も倍率もすべて int なので、順番を取り違えても型では
// 気付けない。**引数の中身まで**固定する。
func TestCrazyFourPokerInteractor_ActionsReachTheDomain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*interfaces.MockCrazyFourPokerGame)
		invoke func(*CrazyFourPokerInteractor) string
		method string
		args   []any
	}{
		{
			name:   "PlaceBet",
			setup:  func(m *interfaces.MockCrazyFourPokerGame) { m.On("PlaceBet", 50, 20).Return(nil) },
			invoke: func(ci *CrazyFourPokerInteractor) string { return ci.PlaceBet(50, 20) },
			method: "PlaceBet", args: []any{50, 20},
		},
		{
			// **Queens Up 0 は「置かない」という有効な入力。**
			name:   "PlaceBet without the side bet",
			setup:  func(m *interfaces.MockCrazyFourPokerGame) { m.On("PlaceBet", 50, 0).Return(nil) },
			invoke: func(ci *CrazyFourPokerInteractor) string { return ci.PlaceBet(50, 0) },
			method: "PlaceBet", args: []any{50, 0},
		},
		{
			name:   "Play x1",
			setup:  func(m *interfaces.MockCrazyFourPokerGame) { m.On("Play", 1).Return(nil) },
			invoke: func(ci *CrazyFourPokerInteractor) string { return ci.Play(1) },
			method: "Play", args: []any{1},
		},
		{
			name:   "Play x3",
			setup:  func(m *interfaces.MockCrazyFourPokerGame) { m.On("Play", 3).Return(nil) },
			invoke: func(ci *CrazyFourPokerInteractor) string { return ci.Play(3) },
			method: "Play", args: []any{3},
		},
		{
			name:   "Fold",
			setup:  func(m *interfaces.MockCrazyFourPokerGame) { m.On("Fold").Return(nil) },
			invoke: func(ci *CrazyFourPokerInteractor) string { return ci.Fold() },
			method: "Fold", args: nil,
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockCrazyFourPokerGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *CrazyFourPokerInteractor) string { return ci.NextRound() },
			method: "NextRound", args: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newCrazyFourPokerInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(ci))
			mg.AssertCalled(t, tt.method, tt.args...)
		})
	}
}

// **上限倍率の判定はドメインだけが持つ。**
//
// インタラクタが手役を見て 3 倍を通すかどうかを決め直すと、規則が 2 か所に増える。
// エラーは握りつぶさずそのまま届ける。
func TestCrazyFourPokerInteractor_MultiplierErrorReachesThePresenter(t *testing.T) {
	mg, mp, ci := newCrazyFourPokerInteractorForTest()
	multErr := errors.New("その倍率にはエースのペア以上が要ります")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Play", 3).Return(multErr)
	mp.On("Output", mg, multErr).Return("error output")

	assert.Equal(t, "error output", ci.Play(3))
	mp.AssertCalled(t, "Output", mg, multErr)
	mg.AssertNotCalled(t, "MaxPlayMultiplier")
	mg.AssertNotCalled(t, "PlayerHasAcesOrBetter")
}

// **終局後の操作はドメインまで届かない。**
func TestCrazyFourPokerInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*CrazyFourPokerInteractor) string
		method string
	}{
		{"PlaceBet", func(ci *CrazyFourPokerInteractor) string { return ci.PlaceBet(50, 0) }, "PlaceBet"},
		{"Play", func(ci *CrazyFourPokerInteractor) string { return ci.Play(1) }, "Play"},
		{"Fold", func(ci *CrazyFourPokerInteractor) string { return ci.Fold() }, "Fold"},
		{"NextRound", func(ci *CrazyFourPokerInteractor) string { return ci.NextRound() }, "NextRound"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newCrazyFourPokerInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(ci))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestCrazyFourPokerInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, ci := newCrazyFourPokerInteractorForTest()
	cfg := domain.DefaultCrazyFourPokerConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint output", ci.Hint())
	assert.Equal(t, "log output", ci.ActionLog())
}

// **範囲外の設定はドメインまで通さず、ゲームも作り直さない。**
func TestCrazyFourPokerInteractor_ResetWithConfig(t *testing.T) {
	t.Run("正しい設定は通る", func(t *testing.T) {
		mg, mp, ci := newCrazyFourPokerInteractorForTest()
		cfg := domain.CrazyFourPokerConfig{InitialChips: 2000, DefaultAnte: 20}
		mg.On("SetConfig", cfg).Return()
		mg.On("Reset").Return()
		mp.On("Output", mg, nil).Return("reset output")

		assert.Equal(t, "reset output", ci.ResetWithConfig(cfg))
		mg.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("範囲外の設定は弾かれる", func(t *testing.T) {
		mg, mp, ci := newCrazyFourPokerInteractorForTest()
		bad := domain.CrazyFourPokerConfig{
			InitialChips: domain.CrazyFourPokerChipsMax + 1, DefaultAnte: 50,
		}
		mp.On("Output", mg, mock.Anything).Return("bad config")

		assert.NotEmpty(t, ci.ResetWithConfig(bad))
		mg.AssertNotCalled(t, "SetConfig")
		mg.AssertNotCalled(t, "Reset")
	})
}

// **保存して読み直しても同じ盤面になる。**
func TestRestoreCrazyFourPokerInteractor(t *testing.T) {
	mp := new(presenter.MockCrazyFourPokerPresenter)

	game := domain.NewDefaultCrazyFourPoker()
	game.Reset()
	require.NoError(t, game.PlaceBet(50, 20))
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ci, err := RestoreCrazyFourPokerInteractor(data, mp)
	require.NoError(t, err)
	require.NotNil(t, ci)
	assert.Equal(t, game.GetPhase(), ci.Game.GetPhase())
	assert.Equal(t, game.GetAnteBet(), ci.Game.GetAnteBet())
	assert.Equal(t, game.GetSuperBet(), ci.Game.GetSuperBet())
	assert.Equal(t, game.GetQueensUpBet(), ci.Game.GetQueensUpBet())
	assert.Equal(t, game.GetChips(), ci.Game.GetChips())

	snap, err := ci.Snapshot()
	require.NoError(t, err)
	again, err := RestoreCrazyFourPokerInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetPlayerHandRank(), again.Game.GetPlayerHandRank())
}

func TestRestoreCrazyFourPokerInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockCrazyFourPokerPresenter)

	_, err := RestoreCrazyFourPokerInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	_, err = RestoreCrazyFourPokerInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
