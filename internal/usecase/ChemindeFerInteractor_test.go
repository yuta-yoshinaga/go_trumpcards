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

// chemindeFerTestRounds はテストで使う妥当なラウンド数。
const chemindeFerTestRounds = 8

func newChemindeFerInteractorForTest() (*interfaces.MockChemindeFerGame,
	*presenter.MockChemindeFerPresenter, *ChemindeFerInteractor,
) {
	mg := new(interfaces.MockChemindeFerGame)
	mp := new(presenter.MockChemindeFerPresenter)
	return mg, mp, NewChemindeFerInteractor(mg, mp)
}

func TestNewChemindeFerInteractor(t *testing.T) {
	_, _, ci := newChemindeFerInteractorForTest()
	assert.NotNil(t, ci)
}

func TestNewChemindeFerInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockChemindeFerPresenter)
	assert.Panics(t, func() { NewChemindeFerInteractor(nil, mp) })

	mg := new(interfaces.MockChemindeFerGame)
	assert.Panics(t, func() { NewChemindeFerInteractor(mg, nil) })
}

func TestChemindeFerInteractor_Reset(t *testing.T) {
	mg, mp, ci := newChemindeFerInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。**
//
// 金額はどれも int なので、席番号と取り違えても型では気付けない。**引数の中身まで**
// 固定する。
func TestChemindeFerInteractor_ActionsReachTheDomain(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*interfaces.MockChemindeFerGame)
		invoke func(*ChemindeFerInteractor) string
		method string
		args   []any
	}{
		{
			name:   "SetStake",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("SetStake", 250).Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.SetStake(250) },
			method: "SetStake", args: []any{250},
		},
		{
			// **席と金額の順番を取り違えないこと。** どちらも int なので通ってしまう。
			name:   "PlaceBet",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("PlaceBet", 3, 120).Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.PlaceBet(3, 120) },
			method: "PlaceBet", args: []any{3, 120},
		},
		{
			// **0 は「降りる」で、正当な賭け額。** 未指定と同じ扱いにしてはいけない。
			name:   "PlaceBet pass",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("PlaceBet", 2, 0).Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.PlaceBet(2, 0) },
			method: "PlaceBet", args: []any{2, 0},
		},
		{
			name:   "PunterDraw",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("PunterDraw").Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.PunterDraw() },
			method: "PunterDraw", args: nil,
		},
		{
			name:   "PunterStand",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("PunterStand").Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.PunterStand() },
			method: "PunterStand", args: nil,
		},
		{
			name:   "BankerDraw",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("BankerDraw").Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.BankerDraw() },
			method: "BankerDraw", args: nil,
		},
		{
			name:   "BankerStand",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("BankerStand").Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.BankerStand() },
			method: "BankerStand", args: nil,
		},
		{
			name:   "PassBank",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("PassBank").Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.PassBank() },
			method: "PassBank", args: nil,
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockChemindeFerGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *ChemindeFerInteractor) string { return ci.NextRound() },
			method: "NextRound", args: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newChemindeFerInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			tt.setup(mg)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", tt.invoke(ci))
			mg.AssertCalled(t, tt.method, tt.args...)
		})
	}
}

// **選べない合計で引こうとしたエラーはそのまま届く。**
//
// ここで握りつぶすと、0-4 の子に「立つ」ボタンが効いてしまい、規則が UI から消える。
func TestChemindeFerInteractor_ErrorsReachThePresenter(t *testing.T) {
	mg, mp, ci := newChemindeFerInteractorForTest()
	standErr := errors.New("この合計に選択の余地はありません")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("PunterStand").Return(standErr)
	mp.On("Output", mg, standErr).Return("error output")

	assert.Equal(t, "error output", ci.PunterStand())
	mp.AssertCalled(t, "Output", mg, standErr)
}

// **終局後の操作はドメインまで届かない。**
func TestChemindeFerInteractor_BlocksAfterGameEnd(t *testing.T) {
	for _, tt := range []struct {
		name   string
		invoke func(*ChemindeFerInteractor) string
		method string
	}{
		{"SetStake", func(ci *ChemindeFerInteractor) string { return ci.SetStake(50) }, "SetStake"},
		{"PlaceBet", func(ci *ChemindeFerInteractor) string { return ci.PlaceBet(1, 50) }, "PlaceBet"},
		{"PunterDraw", func(ci *ChemindeFerInteractor) string { return ci.PunterDraw() }, "PunterDraw"},
		{"PunterStand", func(ci *ChemindeFerInteractor) string { return ci.PunterStand() }, "PunterStand"},
		{"BankerDraw", func(ci *ChemindeFerInteractor) string { return ci.BankerDraw() }, "BankerDraw"},
		{"BankerStand", func(ci *ChemindeFerInteractor) string { return ci.BankerStand() }, "BankerStand"},
		{"PassBank", func(ci *ChemindeFerInteractor) string { return ci.PassBank() }, "PassBank"},
		{"NextRound", func(ci *ChemindeFerInteractor) string { return ci.NextRound() }, "NextRound"},
		{"GiveUp", func(ci *ChemindeFerInteractor) string { return ci.GiveUp() }, "GiveUp"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newChemindeFerInteractorForTest()
			mg.On("GetGameEndFlag").Return(true)
			mp.On("Output", mg, nil).Return("game over")

			assert.NotEmpty(t, tt.invoke(ci))
			mg.AssertNotCalled(t, tt.method)
		})
	}
}

func TestChemindeFerInteractor_GiveUp(t *testing.T) {
	mg, mp, ci := newChemindeFerInteractorForTest()
	mg.On("GetGameEndFlag").Return(false)
	mg.On("GiveUp").Return()
	mp.On("Output", mg, nil).Return("gave up")

	assert.Equal(t, "gave up", ci.GiveUp())
	mg.AssertCalled(t, "GiveUp")
}

func TestChemindeFerInteractor_ConfigHintAndLog(t *testing.T) {
	mg, mp, ci := newChemindeFerInteractorForTest()
	cfg := domain.DefaultChemindeFerConfig()
	mg.On("GetConfig").Return(cfg)
	mp.On("HintOutput", mg).Return("hint output")
	mp.On("ActionLogOutput", mg).Return("log output")

	assert.Equal(t, cfg, ci.GetConfig())
	assert.Equal(t, "hint output", ci.Hint())
	assert.Equal(t, "log output", ci.ActionLog())
}

// **範囲外の設定はドメインまで通さず、ゲームも作り直さない。**
func TestChemindeFerInteractor_ResetWithConfig(t *testing.T) {
	t.Run("正しい設定は通る", func(t *testing.T) {
		mg, mp, ci := newChemindeFerInteractorForTest()
		cfg := domain.ChemindeFerConfig{
			Rounds: chemindeFerTestRounds, InitialChips: domain.ChemindeFerDefaultChips,
		}
		mg.On("SetConfig", cfg).Return()
		mg.On("Reset").Return()
		mp.On("Output", mg, nil).Return("reset output")

		assert.Equal(t, "reset output", ci.ResetWithConfig(cfg))
		mg.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("範囲外の設定は弾かれる", func(t *testing.T) {
		mg, mp, ci := newChemindeFerInteractorForTest()
		bad := domain.ChemindeFerConfig{Rounds: domain.ChemindeFerRoundsMax + 1, InitialChips: 1000}
		mp.On("Output", mg, mock.Anything).Return("bad config")

		assert.NotEmpty(t, ci.ResetWithConfig(bad))
		mg.AssertNotCalled(t, "SetConfig")
		mg.AssertNotCalled(t, "Reset")
	})
}

// **保存して読み直しても同じ盤面になる。**
func TestRestoreChemindeFerInteractor(t *testing.T) {
	mp := new(presenter.MockChemindeFerPresenter)

	game := domain.NewDefaultChemindeFer()
	game.Reset()
	require.NoError(t, game.SetStake(200))
	data, err := json.Marshal(game)
	require.NoError(t, err)

	ci, err := RestoreChemindeFerInteractor(data, mp)
	require.NoError(t, err)
	require.NotNil(t, ci)
	assert.Equal(t, game.GetPhase(), ci.Game.GetPhase())
	assert.Equal(t, game.GetBankerIdx(), ci.Game.GetBankerIdx())
	assert.Equal(t, game.GetStake(), ci.Game.GetStake())
	assert.Equal(t, game.GetBetTurn(), ci.Game.GetBetTurn())

	snap, err := ci.Snapshot()
	require.NoError(t, err)
	again, err := RestoreChemindeFerInteractor(snap, mp)
	require.NoError(t, err)
	assert.Equal(t, game.GetStake(), again.Game.GetStake())
}

func TestRestoreChemindeFerInteractor_RejectsBrokenData(t *testing.T) {
	mp := new(presenter.MockChemindeFerPresenter)

	_, err := RestoreChemindeFerInteractor([]byte(`{`), mp)
	assert.Error(t, err, "壊れた JSON")

	_, err = RestoreChemindeFerInteractor([]byte(`{"ph":99}`), mp)
	assert.Error(t, err, "フェーズが範囲外")
}

// **DrawOrStand の中身を直接踏む。**
//
// コントローラのテストはインタラクタをモックしているので、この振り分けは一度も
// 実行されない。側を決めているのはここだけなので、踏まないと「フェーズを読み違えて
// 相手側の判断を確定させる」バグが誰にも見えない。
func TestChemindeFerInteractor_DrawOrStand(t *testing.T) {
	tests := []struct {
		name    string
		phase   domain.ChemindeFerPhase
		draw    bool
		want    string
		notWant []string
	}{
		{
			name:  "子の判断中の draw は子に届く",
			phase: domain.ChemindeFerPhasePunterDraw, draw: true,
			want: "PunterDraw", notWant: []string{"PunterStand", "BankerDraw", "BankerStand"},
		},
		{
			name:  "子の判断中の stand は子に届く",
			phase: domain.ChemindeFerPhasePunterDraw, draw: false,
			want: "PunterStand", notWant: []string{"PunterDraw", "BankerDraw", "BankerStand"},
		},
		{
			name:  "親の判断中の draw は親に届く",
			phase: domain.ChemindeFerPhaseBankerDraw, draw: true,
			want: "BankerDraw", notWant: []string{"BankerStand", "PunterDraw", "PunterStand"},
		},
		{
			name:  "親の判断中の stand は親に届く",
			phase: domain.ChemindeFerPhaseBankerDraw, draw: false,
			want: "BankerStand", notWant: []string{"BankerDraw", "PunterDraw", "PunterStand"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newChemindeFerInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			mg.On("GetPhase").Return(tt.phase)
			mg.On(tt.want).Return(nil)
			mp.On("Output", mg, nil).Return("ok output")

			assert.Equal(t, "ok output", ci.DrawOrStand(tt.draw))
			mg.AssertCalled(t, tt.want)
			for _, other := range tt.notWant {
				mg.AssertNotCalled(t, other)
			}
		})
	}
}

// **引くか立つかを決める場面でなければ弾く。**
func TestChemindeFerInteractor_DrawOrStandRejectsOtherPhases(t *testing.T) {
	for _, phase := range []domain.ChemindeFerPhase{
		domain.ChemindeFerPhaseStake,
		domain.ChemindeFerPhaseBet,
		domain.ChemindeFerPhaseRoundEnd,
	} {
		t.Run(domain.ChemindeFerPhaseName(phase), func(t *testing.T) {
			mg, mp, ci := newChemindeFerInteractorForTest()
			mg.On("GetGameEndFlag").Return(false)
			mg.On("GetPhase").Return(phase)
			mp.On("Output", mg, ErrChemindeFerNotDrawPhase).Return("not a draw phase")

			assert.Equal(t, "not a draw phase", ci.DrawOrStand(true))
			mp.AssertCalled(t, "Output", mg, ErrChemindeFerNotDrawPhase)
			for _, m := range []string{"PunterDraw", "PunterStand", "BankerDraw", "BankerStand"} {
				mg.AssertNotCalled(t, m)
			}
		})
	}
}

// ドメインが返したエラーはそのままプレゼンタへ届く。
func TestChemindeFerInteractor_DrawOrStandPropagatesErrors(t *testing.T) {
	mg, mp, ci := newChemindeFerInteractorForTest()
	drawErr := errors.New("この合計に選択の余地はありません")
	mg.On("GetGameEndFlag").Return(false)
	mg.On("GetPhase").Return(domain.ChemindeFerPhasePunterDraw)
	mg.On("PunterStand").Return(drawErr)
	mp.On("Output", mg, drawErr).Return("error output")

	assert.Equal(t, "error output", ci.DrawOrStand(false))
}

// **終局後はフェーズを見る前に弾く。**
func TestChemindeFerInteractor_DrawOrStandBlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newChemindeFerInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, nil).Return("game over")

	assert.Equal(t, "game over", ci.DrawOrStand(true))
	mg.AssertNotCalled(t, "GetPhase")
	for _, m := range []string{"PunterDraw", "PunterStand", "BankerDraw", "BankerStand"} {
		mg.AssertNotCalled(t, m)
	}
}
