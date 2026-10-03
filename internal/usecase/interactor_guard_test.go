//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPlayableGame struct {
	gameEnd   bool
	humanTurn bool
}

func (m *mockPlayableGame) GetGameEndFlag() bool { return m.gameEnd }
func (m *mockPlayableGame) IsHumanTurn() bool    { return m.humanTurn }

type mockGameEndChecker struct {
	gameEnd bool
}

func (m *mockGameEndChecker) GetGameEndFlag() bool { return m.gameEnd }

type mockPresenter[G any] struct {
	output string
}

func (m *mockPresenter[G]) Output(_ G, _ error) string { return m.output }

func TestGuardNotPlayable(t *testing.T) {
	p := &mockPresenter[*mockPlayableGame]{output: "blocked"}

	t.Run("game ended", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: true, humanTurn: true}
		out, blocked := guardNotPlayable(g, p)
		assert.True(t, blocked)
		assert.Equal(t, "blocked", out)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: false, humanTurn: false}
		out, blocked := guardNotPlayable(g, p)
		assert.True(t, blocked)
		assert.Equal(t, "blocked", out)
	})

	t.Run("playable", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: false, humanTurn: true}
		out, blocked := guardNotPlayable(g, p)
		assert.False(t, blocked)
		assert.Equal(t, "", out)
	})
}

func TestGuardGameEnd(t *testing.T) {
	p := &mockPresenter[*mockGameEndChecker]{output: "ended"}

	t.Run("game ended", func(t *testing.T) {
		g := &mockGameEndChecker{gameEnd: true}
		out, blocked := guardGameEnd(g, p)
		assert.True(t, blocked)
		assert.Equal(t, "ended", out)
	})

	t.Run("game not ended", func(t *testing.T) {
		g := &mockGameEndChecker{gameEnd: false}
		out, blocked := guardGameEnd(g, p)
		assert.False(t, blocked)
		assert.Equal(t, "", out)
	})
}

// advanceRound consolidates 22 byte-identical NextRound implementations
// (Burraco, Canasta, Carioca, Chinchon, Conquian, ContractRummy, …) that
// differed only in receiver name and concrete type — see issue #5368.
type mockRoundGame struct {
	gameEnd    bool
	roundCalls int
}

func (m *mockRoundGame) GetGameEndFlag() bool { return m.gameEnd }
func (m *mockRoundGame) NextRound()           { m.roundCalls++ }

func TestAdvanceRound(t *testing.T) {
	t.Run("advances the round and runs the CPU", func(t *testing.T) {
		g := &mockRoundGame{}
		p := &mockPresenter[*mockRoundGame]{output: "out"}
		cpuRuns := 0

		got := advanceRound(g, p, func() { cpuRuns++ })

		assert.Equal(t, "out", got)
		assert.Equal(t, 1, g.roundCalls)
		assert.Equal(t, 1, cpuRuns)
	})

	// The guard is the whole reason these 22 bodies were identical: without it
	// a finished game would deal another round.
	t.Run("does nothing once the game has ended", func(t *testing.T) {
		g := &mockRoundGame{gameEnd: true}
		p := &mockPresenter[*mockRoundGame]{output: "ended"}
		cpuRuns := 0

		got := advanceRound(g, p, func() { cpuRuns++ })

		assert.Equal(t, "ended", got)
		assert.Equal(t, 0, g.roundCalls, "the round must not advance after the game ends")
		assert.Equal(t, 0, cpuRuns, "the CPU must not play after the game ends")
	})

	// Order matters: the CPU must act on the new round, not the old one.
	t.Run("runs the CPU after the round advances", func(t *testing.T) {
		var order []string
		game := &orderRecordingGame{mockRoundGame: &mockRoundGame{}, order: &order}

		advanceRound(game, &mockPresenter[*orderRecordingGame]{output: "out"},
			func() { order = append(order, "cpu") })

		assert.Equal(t, []string{"nextRound", "cpu"}, order)
	})
}

type actionRecordingPresenter struct {
	order   *[]string
	output  string
	lastErr error
}

func (p *actionRecordingPresenter) Output(_ *mockPlayableGame, err error) string {
	*p.order = append(*p.order, "output")
	p.lastErr = err
	return p.output
}

func TestHumanAction(t *testing.T) {
	t.Run("does not act when ended or not the human turn", func(t *testing.T) {
		for _, game := range []*mockPlayableGame{
			{gameEnd: true, humanTurn: true},
			{gameEnd: false, humanTurn: false},
		} {
			var order []string
			p := &actionRecordingPresenter{order: &order, output: "blocked"}
			got := humanAction(game, p, func() error { order = append(order, "act"); return nil }, func() { order = append(order, "cpu") })
			assert.Equal(t, "blocked", got)
			assert.Equal(t, []string{"output"}, order)
			assert.NoError(t, p.lastErr)
		}
	})

	t.Run("presents rejected action without running CPU", func(t *testing.T) {
		var order []string
		wantErr := errors.New("rejected")
		p := &actionRecordingPresenter{order: &order, output: "rejected output"}
		got := humanAction(&mockPlayableGame{humanTurn: true}, p,
			func() error { order = append(order, "act"); return wantErr },
			func() { order = append(order, "cpu") })
		assert.Equal(t, "rejected output", got)
		assert.Equal(t, []string{"act", "output"}, order)
		assert.ErrorIs(t, p.lastErr, wantErr)
	})

	t.Run("runs action then CPU then presents", func(t *testing.T) {
		var order []string
		p := &actionRecordingPresenter{order: &order, output: "success"}
		got := humanAction(&mockPlayableGame{humanTurn: true}, p,
			func() error { order = append(order, "act"); return nil },
			func() { order = append(order, "cpu") })
		assert.Equal(t, "success", got)
		assert.Equal(t, []string{"act", "cpu", "output"}, order)
		assert.NoError(t, p.lastErr)
	})
}

// orderRecordingGame records when NextRound runs relative to the CPU callback.
type orderRecordingGame struct {
	*mockRoundGame
	order *[]string
}

func (o *orderRecordingGame) NextRound() {
	*o.order = append(*o.order, "nextRound")
	o.mockRoundGame.NextRound()
}

// **上限そのものが検査対象。** 上限が無いと、ドメインが局を終わらせないときに
// CLI はプロンプトを返さず Web はリクエストを返さない (#5414 / #5416)。
func TestRunCpuTurnsCapped(t *testing.T) {
	t.Run("stops when the game ends", func(t *testing.T) {
		g := &mockPlayableGame{}
		calls := 0
		ok := runCpuTurnsCapped(g, func() {
			calls++
			if calls == 3 {
				g.gameEnd = true
			}
		})
		assert.True(t, ok, "上限に当たっていないのに false")
		assert.Equal(t, 3, calls)
	})

	t.Run("stops when it becomes the human's turn", func(t *testing.T) {
		g := &mockPlayableGame{}
		calls := 0
		ok := runCpuTurnsCapped(g, func() {
			calls++
			if calls == 2 {
				g.humanTurn = true
			}
		})
		assert.True(t, ok)
		assert.Equal(t, 2, calls)
	})

	t.Run("does not call play at all when already finished", func(t *testing.T) {
		g := &mockPlayableGame{gameEnd: true}
		calls := 0
		ok := runCpuTurnsCapped(g, func() { calls++ })
		assert.True(t, ok)
		assert.Zero(t, calls)
	})

	// **これが本題。** 進まないゲームでも必ず戻り、false を返す。
	t.Run("gives up at the cap on a game that never progresses", func(t *testing.T) {
		g := &mockPlayableGame{}
		calls := 0
		ok := runCpuTurnsCapped(g, func() { calls++ })
		assert.False(t, ok, "上限に当たったのに true を返している")
		assert.Equal(t, MaxCpuIterations, calls)
	})
}

// runCpuTurnsUntil のほうは 18 ゲームが共有するので、上限に当たる分岐を
// ここで踏んでおけば全部が守られる。各ゲームに書くと、実ゲームは 1000 手より
// ずっと手前でフェーズガードに当たるため **その return は永久に未実行**になる。
func TestRunCpuTurnsUntil(t *testing.T) {
	t.Run("stops on the game end flag", func(t *testing.T) {
		g := &mockGameEndChecker{gameEnd: true}
		calls := 0
		assert.True(t, runCpuTurnsUntil(g, func() bool { return false }, func() { calls++ }))
		assert.Zero(t, calls, "終局しているのに play が呼ばれた")
	})

	t.Run("stops when the phase guard says so", func(t *testing.T) {
		g := &mockGameEndChecker{}
		calls := 0
		stopAfter := 3
		ok := runCpuTurnsUntil(g, func() bool { return calls >= stopAfter }, func() { calls++ })
		assert.True(t, ok)
		assert.Equal(t, stopAfter, calls)
	})

	// **本題。** 進まないゲームでも必ず戻り、false を返す。
	t.Run("gives up at the cap", func(t *testing.T) {
		g := &mockGameEndChecker{}
		calls := 0
		assert.False(t, runCpuTurnsUntil(g, func() bool { return false }, func() { calls++ }),
			"上限に当たったのに true を返している")
		assert.Equal(t, MaxCpuIterations, calls)
	})

	// stop は play の前に読む -- 開始時点で止まる局面なら 1 手も進めない。
	t.Run("reads stop before playing", func(t *testing.T) {
		g := &mockGameEndChecker{}
		calls := 0
		assert.True(t, runCpuTurnsUntil(g, func() bool { return true }, func() { calls++ }))
		assert.Zero(t, calls)
	})
}
