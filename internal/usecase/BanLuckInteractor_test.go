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

func newBanLuckInteractorForTest() (*interfaces.MockBanLuckGame,
	*presenter.MockBanLuckPresenter, *BanLuckInteractor,
) {
	mg := new(interfaces.MockBanLuckGame)
	mp := new(presenter.MockBanLuckPresenter)
	// **既定の期待は個別の期待より後に登録する。** 先に置くと個別が上書き
	// されず、狙った引数の検査が効かなくなる。
	return mg, mp, NewBanLuckInteractor(mg, mp)
}

func TestNewBanLuckInteractor_NilPanics(t *testing.T) {
	mp := new(presenter.MockBanLuckPresenter)
	assert.Panics(t, func() { NewBanLuckInteractor(nil, mp) })

	mg := new(interfaces.MockBanLuckGame)
	assert.Panics(t, func() { NewBanLuckInteractor(mg, nil) })
}

func TestBanLuckInteractor_Reset(t *testing.T) {
	mg, mp, ci := newBanLuckInteractorForTest()
	mg.On("Reset").Return()
	mp.On("Output", mg, nil).Return("reset output")

	assert.Equal(t, "reset output", ci.Reset())
	mg.AssertCalled(t, "Reset")
}

// **どの操作もドメインへそのまま渡る。**
func TestBanLuckInteractor_ActionsReachTheDomain(t *testing.T) {
	for _, tt := range []struct {
		name   string
		setup  func(*interfaces.MockBanLuckGame)
		invoke func(*BanLuckInteractor) string
		method string
		args   []any
	}{
		{
			name:   "PlaceBet",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("PlaceBet", 50).Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.PlaceBet(50) },
			method: "PlaceBet", args: []any{50},
		},
		{
			name:   "Hit",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("Hit").Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.Hit() },
			method: "Hit",
		},
		{
			name:   "Stand",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("Stand").Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.Stand() },
			method: "Stand",
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.NextRound() },
			method: "NextRound",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newBanLuckInteractorForTest()
			tt.setup(mg)
			mg.On("GetGameEndFlag").Return(false)
			mg.On("CpuPlay").Return()
			mp.On("Output", mg, nil).Return("ok")

			assert.Equal(t, "ok", tt.invoke(ci))
			mg.AssertCalled(t, tt.method, tt.args...)
		})
	}
}

// **人間が 1 手指したら CPU が動く。** これを忘れると盤面が止まる。
func TestBanLuckInteractor_DrivesTheCpuAfterEveryAction(t *testing.T) {
	for _, tt := range []struct {
		name   string
		setup  func(*interfaces.MockBanLuckGame)
		invoke func(*BanLuckInteractor) string
	}{
		{
			name:   "PlaceBet",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("PlaceBet", 50).Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.PlaceBet(50) },
		},
		{
			name:   "Hit",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("Hit").Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.Hit() },
		},
		{
			name:   "Stand",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("Stand").Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.Stand() },
		},
		{
			name:   "NextRound",
			setup:  func(m *interfaces.MockBanLuckGame) { m.On("NextRound").Return(nil) },
			invoke: func(ci *BanLuckInteractor) string { return ci.NextRound() },
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mg, mp, ci := newBanLuckInteractorForTest()
			tt.setup(mg)
			mg.On("GetGameEndFlag").Return(false)
			mg.On("CpuPlay").Return()
			mp.On("Output", mg, nil).Return("ok")

			tt.invoke(ci)
			mg.AssertNumberOfCalls(t, "CpuPlay", 1)
		})
	}
}

// **失敗した操作では CPU を進めない。** 拒否された手で盤面が動くと、
// エラーを出しつつ状態だけ先へ行く。
func TestBanLuckInteractor_DoesNotDriveTheCpuOnError(t *testing.T) {
	mg, mp, ci := newBanLuckInteractorForTest()
	boom := errors.New("nope")
	mg.On("Hit").Return(boom)
	mg.On("GetGameEndFlag").Return(false)
	mg.On("CpuPlay").Return()
	mp.On("Output", mg, boom).Return("error output")

	assert.Equal(t, "error output", ci.Hit())
	mg.AssertNumberOfCalls(t, "CpuPlay", 0)
}

// **終局後はドメインに触らない。**
func TestBanLuckInteractor_BlocksAfterGameEnd(t *testing.T) {
	mg, mp, ci := newBanLuckInteractorForTest()
	mg.On("GetGameEndFlag").Return(true)
	mp.On("Output", mg, mock.Anything).Return("finished")

	assert.Equal(t, "finished", ci.Hit())
	mg.AssertNotCalled(t, "Hit")
	mg.AssertNotCalled(t, "CpuPlay")
}

func TestBanLuckInteractor_HintAndLog(t *testing.T) {
	mg, mp, ci := newBanLuckInteractorForTest()
	mp.On("HintOutput", mg).Return("hint")
	mp.On("ActionLogOutput", mg).Return("log")

	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

func TestBanLuckInteractor_Config(t *testing.T) {
	mg, mp, ci := newBanLuckInteractorForTest()
	cfg := domain.DefaultBanLuckConfig()
	mg.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, ci.GetConfig())

	// **範囲外の設定は弾く。** ここを通すと席 0 の卓ができる。
	bad := domain.BanLuckConfig{Seats: 1, InitialChips: 1000, Rounds: 10, DefaultBet: 50}
	mp.On("Output", mg, mock.Anything).Return("bad config")
	assert.Equal(t, "bad config", ci.ResetWithConfig(bad))
	mg.AssertNotCalled(t, "SetConfig", bad)

	good := domain.BanLuckConfig{Seats: 3, InitialChips: 500, Rounds: 5, DefaultBet: 20}
	mg.On("SetConfig", good).Return()
	mg.On("Reset").Return()
	ci.ResetWithConfig(good)
	mg.AssertCalled(t, "SetConfig", good)
}

// --- 本物のドメインで駆動を確かめる ---

// **人間の操作を 1 回だけ実行して、どこで止まったかを見る。**
//
// テストが自分でループを回すと、誰も CPU を進めていなくても全部緑になる。
// ドメイン・usecase・adapter が全部通って E2E だけが落ちる、という形で
// 何度も踏んでいるので、ここは実物を 1 手だけ動かして確かめる。
func TestBanLuckInteractor_OneHumanActionAdvancesTheBoard(t *testing.T) {
	for range 50 {
		g := domain.NewDefaultBanLuck()
		g.Reset()
		ci := NewBanLuckInteractor(g, new(bannedLuckSilentPresenter))

		// 席 0 (人間) が親なので、賭けたら子の CPU が全員打ち終わって
		// 人間の手番に戻るはず。誰も駆動していなければ席 1 で止まる。
		ci.PlaceBet(0)
		if g.GetPhase() != domain.BanLuckPhasePlay {
			continue // その配りで即決着した (特別役)
		}
		require.True(t, g.IsHumanTurn(),
			"人間が 1 手指したのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
	}
}

// **人間が子のときこそ駆動が要る。**
//
// 席 0 が親のラウンドでは人間が最後に打つので、止まった時点で全席が済んでいて
// **CPU を誰も進めなくても決着まで行ってしまう** ── その配置だけで試すと、
// 駆動を外した実装でもテストが通る (実測)。人間が子になるラウンドまで進めて、
// 「自分が止まった後ろにまだ CPU が居る」状態で確かめる。
func TestBanLuckInteractor_CpuActsAfterTheHumanStands(t *testing.T) {
	checked := 0
	for range 50 {
		g := domain.NewDefaultBanLuck()
		g.Reset()
		ci := NewBanLuckInteractor(g, new(bannedLuckSilentPresenter))

		// 人間が子になるラウンドまで進める。
		for steps := 0; g.GetBankerSeat() == g.GetHumanSeat() && !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 100, "人間が子になるラウンドが来ない")
			banLuckPlayOneRound(t, g, ci)
		}
		if g.GetGameEndFlag() {
			continue
		}

		ci.PlaceBet(domain.BanLuckDefaultBet)
		if g.GetPhase() != domain.BanLuckPhasePlay || !g.IsHumanTurn() {
			continue // その配りでは人間に手番が来なかった
		}
		require.Less(t, g.GetTurnSeat(), g.GetBankerSeat(),
			"人間の後ろに席が残っていない配置を掴んでいる")

		ci.Stand()
		require.NotEqual(t, domain.BanLuckPhasePlay, g.GetPhase(),
			"人間が止まったのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
		checked++
	}
	require.Positive(t, checked, "駆動を確かめられる局面が 1 度も出なかった")
}

// banLuckPlayOneRound は 1 ラウンドだけ規則どおりに打ち切る。
func banLuckPlayOneRound(t *testing.T, g *domain.BanLuck, ci *BanLuckInteractor) {
	t.Helper()
	ci.PlaceBet(domain.BanLuckDefaultBet)
	for steps := 0; g.GetPhase() == domain.BanLuckPhasePlay; steps++ {
		require.Less(t, steps, 50, "ラウンドが終わらない")
		if !g.IsHumanTurn() {
			ci.Hit() // 手番が人間でなければ弾かれるだけ。CPU は駆動側が進める。
			break
		}
		if g.MustHit() {
			ci.Hit()
			continue
		}
		ci.Stand()
	}
	if g.GetPhase() == domain.BanLuckPhaseRoundEnd {
		ci.NextRound()
	}
}

// bannedLuckSilentPresenter は駆動だけを見るための無音プレゼンタ。
type bannedLuckSilentPresenter struct{}

func (p *bannedLuckSilentPresenter) Output(interfaces.BanLuckGame, error) string { return "" }
func (p *bannedLuckSilentPresenter) ActionLogOutput(interfaces.BanLuckGame) string {
	return ""
}
func (p *bannedLuckSilentPresenter) HintOutput(interfaces.BanLuckGame) string { return "" }

// --- 永続化 ---

func TestBanLuckInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultBanLuck()
	g.Reset()
	ci := NewBanLuckInteractor(g, new(bannedLuckSilentPresenter))
	ci.PlaceBet(0)

	data, err := ci.Snapshot()
	require.NoError(t, err)
	restored, err := RestoreBanLuckInteractor(data, new(bannedLuckSilentPresenter))
	require.NoError(t, err)
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	assert.Equal(t, g.GetBankerSeat(), restored.Game.GetBankerSeat())
	assert.Equal(t, g.GetRoundNumber(), restored.Game.GetRoundNumber())

	_, err = RestoreBanLuckInteractor([]byte(`{"bk":99}`), new(bannedLuckSilentPresenter))
	assert.Error(t, err, "壊れた保存が復元できてしまった")
}

func TestBanLuckInteractor_SnapshotIsValidJSON(t *testing.T) {
	g := domain.NewDefaultBanLuck()
	g.Reset()
	ci := NewBanLuckInteractor(g, new(bannedLuckSilentPresenter))
	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.True(t, json.Valid(data))
}
