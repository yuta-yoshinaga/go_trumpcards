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

func newDoubleAttackInteractorForTest() (*interfaces.MockDoubleAttackBlackjackGame,
	*presenter.MockDoubleAttackBlackjackPresenter, *DoubleAttackBlackjackInteractor,
) {
	mg := new(interfaces.MockDoubleAttackBlackjackGame)
	mp := new(presenter.MockDoubleAttackBlackjackPresenter)
	return mg, mp, NewDoubleAttackBlackjackInteractor(mg, mp)
}

func TestNewDoubleAttackBlackjackInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockDoubleAttackBlackjackPresenter)
	assert.Panics(t, func() { NewDoubleAttackBlackjackInteractor(nil, mp) })

	mg := new(interfaces.MockDoubleAttackBlackjackGame)
	assert.Panics(t, func() { NewDoubleAttackBlackjackInteractor(mg, nil) })
}

func TestDoubleAttackInteractor_Reset(t *testing.T) {
	mg, mp, ci := newDoubleAttackInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。**
//
// アンティ・サイドベット・追加ベットがすべて int なので、順番を取り違えても型では
// 気付けない。**引数の中身まで**固定する。
func TestDoubleAttackInteractor_ActionsReachTheDomain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*interfaces.MockDoubleAttackBlackjackGame)
		invoke func(*DoubleAttackBlackjackInteractor) string
		method string
		args   []any
	}{
		{
			name:   "PlaceBet",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("PlaceBet", 50, 20).Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.PlaceBet(50, 20) },
			method: "PlaceBet", args: []any{50, 20},
		},
		{
			// **Bust It 0 は「置かない」という有効な入力。**
			name:   "PlaceBet without the side bet",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("PlaceBet", 50, 0).Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.PlaceBet(50, 0) },
			method: "PlaceBet", args: []any{50, 0},
		},
		{
			name:   "Attack",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("Attack", 50).Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.Attack(50) },
			method: "Attack", args: []any{50},
		},
		{
			// **見送り (0) も有効な入力。** 送らないのとは違う。
			name:   "Attack declined",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("Attack", 0).Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.Attack(0) },
			method: "Attack", args: []any{0},
		},
		{
			name:   "Hit",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("Hit").Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.Hit() },
			method: "Hit", args: nil,
		},
		{
			name:   "Stand",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("Stand").Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.Stand() },
			method: "Stand", args: nil,
		},
		{
			name:   "Double",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("Double").Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.Double() },
			method: "Double", args: nil,
		},
		{
			name:   "Split",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("Split").Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.Split() },
			method: "Split", args: nil,
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockDoubleAttackBlackjackGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *DoubleAttackBlackjackInteractor) string { return ci.NextRound() },
			method: "NextRound", args: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newDoubleAttackInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(ci))
			mg.AssertCalled(t, tt.method, tt.args...)
		})
	}
}

// **上限や可否の判定はドメインだけが持つ。**
//
// インタラクタが MaxAttackBet や CanDouble を見て自分で弾くと、規則が 2 か所に増える。
// エラーは握りつぶさずそのまま届ける。
func TestDoubleAttackInteractor_RuleChecksStayInTheDomain(t *testing.T) {
	mg, mp, ci := newDoubleAttackInteractorForTest()
	attackErr := errors.New("追加ベットはアンティを超えられません")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("Attack", 999).Return(attackErr)
	mp.On("Output", mg, attackErr).Return("error output")

	assert.Equal(t, "error output", ci.Attack(999))
	mp.AssertCalled(t, "Output", mg, attackErr)
	mg.AssertNotCalled(t, "MaxAttackBet")
	mg.AssertNotCalled(t, "CanDouble")
	mg.AssertNotCalled(t, "CanSplit")
}

// **終局後の操作はドメインまで届かない。**
func TestDoubleAttackInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*DoubleAttackBlackjackInteractor) string
		method string
	}{
		{"PlaceBet", func(ci *DoubleAttackBlackjackInteractor) string { return ci.PlaceBet(50, 0) }, "PlaceBet"},
		{"Attack", func(ci *DoubleAttackBlackjackInteractor) string { return ci.Attack(10) }, "Attack"},
		{"Hit", func(ci *DoubleAttackBlackjackInteractor) string { return ci.Hit() }, "Hit"},
		{"Stand", func(ci *DoubleAttackBlackjackInteractor) string { return ci.Stand() }, "Stand"},
		{"Double", func(ci *DoubleAttackBlackjackInteractor) string { return ci.Double() }, "Double"},
		{"Split", func(ci *DoubleAttackBlackjackInteractor) string { return ci.Split() }, "Split"},
		{"NextRound", func(ci *DoubleAttackBlackjackInteractor) string { return ci.NextRound() }, "NextRound"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newDoubleAttackInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(ci))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestDoubleAttackInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, ci := newDoubleAttackInteractorForTest()
	cfg := domain.DefaultDoubleAttackBlackjackConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint output", ci.Hint())
	assert.Equal(t, "log output", ci.ActionLog())
}

func TestDoubleAttackInteractor_ResetWithConfig(t *testing.T) {
	t.Run("正しい設定は通る", func(t *testing.T) {
		mg, mp, ci := newDoubleAttackInteractorForTest()
		cfg := domain.DoubleAttackBlackjackConfig{InitialChips: 2000, DefaultAnte: 20}
		mg.On("SetConfig", cfg).Return()
		mg.On("Reset").Return()
		mp.On("Output", mg, nil).Return("reset output")

		assert.Equal(t, "reset output", ci.ResetWithConfig(cfg))
		mg.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("範囲外の設定は弾かれる", func(t *testing.T) {
		mg, mp, ci := newDoubleAttackInteractorForTest()
		bad := domain.DoubleAttackBlackjackConfig{
			InitialChips: domain.DoubleAttackChipsMax + 1, DefaultAnte: 50,
		}
		mp.On("Output", mg, mock.Anything).Return("bad config")

		assert.NotEmpty(t, ci.ResetWithConfig(bad))
		mg.AssertNotCalled(t, "SetConfig")
		mg.AssertNotCalled(t, "Reset")
	})
}

// **保存して読み直しても同じ盤面になる。**
func TestRestoreDoubleAttackBlackjackInteractor(t *testing.T) {
	mp := new(presenter.MockDoubleAttackBlackjackPresenter)

	game := domain.NewDefaultDoubleAttackBlackjack()
	game.Reset()
	require.NoError(t, game.PlaceBet(50, 20))
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ci, err := RestoreDoubleAttackBlackjackInteractor(data, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetPhase(), ci.Game.GetPhase())
	assert.Equal(t, game.GetAnteBet(), ci.Game.GetAnteBet())
	assert.Equal(t, game.GetBustItBet(), ci.Game.GetBustItBet())
	assert.Equal(t, game.GetChips(), ci.Game.GetChips())
	// **アップカードだけの状態も保たれる。**
	assert.False(t, ci.Game.IsDealerHoleDealt())
	assert.Len(t, ci.Game.GetDealerCards(), 1)

	snap, err := ci.Snapshot()
	require.NoError(t, err)
	again, err := RestoreDoubleAttackBlackjackInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetAnteBet(), again.Game.GetAnteBet())
}

func TestRestoreDoubleAttackBlackjackInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockDoubleAttackBlackjackPresenter)

	_, err := RestoreDoubleAttackBlackjackInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	_, err = RestoreDoubleAttackBlackjackInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}
