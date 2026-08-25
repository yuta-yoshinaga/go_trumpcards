//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// dilotiPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type dilotiPassThrough struct{}

func (dilotiPassThrough) Output(_ interfaces.DilotiGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (dilotiPassThrough) HintOutput(_ interfaces.DilotiGame) string      { return "hint" }
func (dilotiPassThrough) ActionLogOutput(_ interfaces.DilotiGame) string { return "log" }

func newDilotiReal() (*usecase.DilotiInteractor, *domain.Diloti) {
	g := domain.NewDefaultDiloti()
	return usecase.NewDilotiInteractor(g, dilotiPassThrough{}), g
}

func TestNewDilotiInteractor_NilGuards(t *testing.T) {
	dp := new(presenter.MockDilotiPresenter)
	assert.PanicsWithValue(t, "DilotiInteractor: g must not be nil", func() {
		usecase.NewDilotiInteractor(nil, dp)
	})
	assert.PanicsWithValue(t, "DilotiInteractor: dp must not be nil", func() {
		usecase.NewDilotiInteractor(new(interfaces.MockDilotiGame), nil)
	})
}

// **開幕は人間の手番。** 非親が先に打つ規則なので親を席 1 にしてある ──
// 親を 0 にすると人間は最初の 4 枚に一度も手を出せない。
func TestDilotiInteractor_ResetStopsAtTheHuman(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.Reset())
	assert.Equal(t, domain.DilotiPhasePlay, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.DilotiHandSize, g.GetPlayer(0).GetCardsSize())
	assert.Len(t, g.GetTable(), domain.DilotiTableSize)
}

func TestDilotiInteractor_ResetWithConfig(t *testing.T) {
	di, _ := newDilotiReal()
	require.Equal(t, "ok", di.ResetWithConfig(domain.DilotiConfig{
		CpuDifficulty: domain.DilotiCpuDifficultyEasy,
		TargetScore:   41,
	}))
	assert.Equal(t, 41, di.GetConfig().TargetScore)

	out := di.ResetWithConfig(domain.DilotiConfig{TargetScore: 999})
	assert.Contains(t, out, "err:")
	assert.Equal(t, 41, di.GetConfig().TargetScore, "弾いた設定が入ってしまっている")
}

// **打ったら手番が人間に戻ってくる。** インタラクターの仕事は CPU を回し切って
// 人間の番か区切りで止めること ── 止めなければ CUI も Web も入力を待てない。
func TestDilotiInteractor_PlayRunsTheCpuAndComesBack(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.Reset())
	hand := g.GetPlayer(0).GetCardsSize()

	h := g.GetHint()
	require.GreaterOrEqual(t, h.Move.HandIdx, 0)
	require.Equal(t, "ok", di.Play(h.Move.HandIdx, h.Move.Action,
		h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))

	assert.NotEqual(t, hand, g.GetPlayer(0).GetCardsSize(), "手札が減っていない")
	if g.GetPhase() == domain.DilotiPhasePlay {
		assert.True(t, g.IsHumanTurn(), "CPU の手番のまま止まっている")
	}
}

// **弾いた手は盤面を動かさない。**
func TestDilotiInteractor_RejectsAnIllegalMove(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.Reset())
	hand := g.GetPlayer(0).GetCardsSize()
	table := len(g.GetTable())

	assert.Contains(t, di.Play(99, domain.DilotiActionTrail, nil, nil, 0), "err:")
	assert.Contains(t, di.Play(0, "zzz", nil, nil, 0), "err:")
	assert.Contains(t, di.Play(0, domain.DilotiActionCapture, []int{99}, nil, 0), "err:")
	assert.Equal(t, hand, g.GetPlayer(0).GetCardsSize(), "弾いた手で札が減った")
	assert.Len(t, g.GetTable(), table, "弾いた手で場が動いた")
}

// **局の区切りでは勝手に進まない。** 集計を読む時間を人間に渡す。
func TestDilotiInteractor_StopsAtRoundEnd(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.Reset())
	dilotiFinishRound(t, di, g)
	require.Equal(t, domain.DilotiPhaseRoundEnd, g.GetPhase())

	res := g.GetLastResult()
	require.NotNil(t, res)
	assert.Len(t, res.Lines, 5)
	assert.Equal(t, domain.DilotiDeckSize, res.CardCounts[0]+res.CardCounts[1],
		"取り札の合計が 52 枚でない")

	require.Equal(t, "ok", di.NextRound())
	assert.Equal(t, domain.DilotiPhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.True(t, g.IsHumanTurn(), "2 局目で人間の手番に戻らない")
}

// **札は湧かない。** 山・手札・場・宣言・取り札の総和は常に 52 枚。
func TestDilotiInteractor_CardsAreConserved(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.Reset())
	total := func() int {
		n := g.GetDeckRemaining() + len(g.GetTable())
		for _, x := range g.GetDeclarations() {
			n += len(x.AllCards())
		}
		for i := 0; i < g.GetPlayerCnt(); i++ {
			n += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetCaptured())
		}
		return n
	}
	for round := 0; round < 20 && !g.GetGameEndFlag(); round++ {
		assert.Equal(t, domain.DilotiDeckSize, total(), "局 %d の頭で札が合わない", round+1)
		dilotiFinishRound(t, di, g)
		assert.Equal(t, domain.DilotiDeckSize, total(), "局 %d の終わりで札が合わない", round+1)
		require.Equal(t, "ok", di.NextRound())
	}
}

// **目標点に届けば終わる。**
func TestDilotiInteractor_ReachesTheTargetScore(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.ResetWithConfig(domain.DilotiConfig{
		CpuDifficulty: domain.DilotiCpuDifficultyEasy,
		TargetScore:   domain.DilotiMinTarget,
	}))
	for round := 0; round < 200 && !g.GetGameEndFlag(); round++ {
		dilotiFinishRound(t, di, g)
		require.Equal(t, "ok", di.NextRound())
	}
	require.True(t, g.GetGameEndFlag(), "21 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, g.GetWinnerIdx(), 0)
	// 終局後の操作は盤面を触らない。
	assert.Equal(t, "ok", di.NextRound())
	assert.Equal(t, "ok", di.Play(0, domain.DilotiActionTrail, nil, nil, 0))
	assert.True(t, g.GetGameEndFlag())
}

func TestDilotiInteractor_HintAndActionLog(t *testing.T) {
	di, _ := newDilotiReal()
	require.Equal(t, "ok", di.Reset())
	assert.Equal(t, "hint", di.Hint())
	assert.Equal(t, "log", di.ActionLog())
}

// **保存した盤で打ち続けられる。**
func TestDilotiInteractor_SnapshotRestoreKeepsPlaying(t *testing.T) {
	di, g := newDilotiReal()
	require.Equal(t, "ok", di.ResetWithConfig(domain.DilotiConfig{
		CpuDifficulty: domain.DilotiCpuDifficultyEasy,
		TargetScore:   domain.DilotiMinTarget,
	}))
	h := g.GetHint()
	require.GreaterOrEqual(t, h.Move.HandIdx, 0)
	require.Equal(t, "ok", di.Play(h.Move.HandIdx, h.Move.Action,
		h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))

	data, err := di.Snapshot()
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	restored, err := usecase.RestoreDilotiInteractor(data, dilotiPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetDeckRemaining(), rg.GetDeckRemaining(), "山の位置が消えている")
	assert.Equal(t, g.GetLastCapturer(), rg.GetLastCapturer())
	assert.Equal(t, len(g.GetTable()), len(rg.GetTable()))
	assert.Equal(t, len(g.GetDeclarations()), len(rg.GetDeclarations()))
	assert.Equal(t, domain.DilotiMinTarget, rg.GetConfig().TargetScore, "設定が消えている")
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, len(g.GetPlayer(i).GetCaptured()), len(rg.GetPlayer(i).GetCaptured()),
			"席 %d の取り札", i)
		assert.Equal(t, g.GetPlayer(i).GetXeri(), rg.GetPlayer(i).GetXeri(), "席 %d のクセリ", i)
	}

	for round := 0; round < 200 && !rg.GetGameEndFlag(); round++ {
		dilotiFinishRound(t, restored, rg)
		require.Equal(t, "ok", restored.NextRound())
	}
	assert.True(t, rg.GetGameEndFlag(), "復元した盤で終局に届かない")
}

func TestRestoreDilotiInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreDilotiInteractor([]byte("{"), dilotiPassThrough{})
	assert.Error(t, err)
}

// dilotiFinishRound は現在の局を終了まで打つ。
func dilotiFinishRound(t *testing.T, di *usecase.DilotiInteractor, g interfaces.DilotiGame) {
	t.Helper()
	for step := 0; step < 500 && g.GetPhase() == domain.DilotiPhasePlay; step++ {
		h := g.GetHint()
		require.GreaterOrEqual(t, h.Move.HandIdx, 0, "手番なのに打てる手が無い")
		require.Equal(t, "ok", di.Play(h.Move.HandIdx, h.Move.Action,
			h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))
	}
	require.NotEqual(t, domain.DilotiPhasePlay, g.GetPhase(), "局が終わらない")
}
