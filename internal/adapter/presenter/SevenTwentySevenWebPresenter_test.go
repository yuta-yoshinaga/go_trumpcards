//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func s27Decode(t *testing.T, raw string) controller.SevenTwentySevenWebOutput {
	t.Helper()
	var out controller.SevenTwentySevenWebOutput
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func TestSevenTwentySevenWebPresenter_Output_BaseState(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()

	out := s27Decode(t, p.Output(g, nil))
	assert.Equal(t, int(domain.SevenTwentySevenPhaseDraw), out.Phase)
	assert.Equal(t, 1, out.DrawRound)
	assert.Equal(t, -1, out.LowWinner, "ラウンド途中で勝者が決まっている")
	assert.Equal(t, -1, out.HighWinner)
	assert.Len(t, out.Players, g.GetPlayerCnt())

	// **自分の手札と得点だけが見える。** 相手の得点は手札そのものなので、
	// 出すと配りが丸見えになる。
	for _, pl := range out.Players {
		if pl.IsHuman {
			assert.NotEmpty(t, pl.Cards)
			assert.NotEmpty(t, pl.LowScore, "自分の 7 側の得点が出ていない")
			assert.NotEmpty(t, pl.HighScore)
			continue
		}
		assert.Empty(t, pl.Cards, "player %d の手札が漏れている", pl.ID)
		assert.Empty(t, pl.LowScore, "player %d の得点が漏れている", pl.ID)
		assert.Empty(t, pl.HighScore, "player %d の得点が漏れている", pl.ID)
	}
}

// **両側の得点を返す。** 片方だけではページから狙いが読めない。超過は "-"。
func TestSevenTwentySevenWebPresenter_Output_CarriesBothScores(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	// 4 + K = 4.5 点。両側とも生存。
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 4), s27Card(domain.CardDesignHeart, 13)})
	out := s27Decode(t, p.Output(g, nil))
	assert.Equal(t, "4.5", out.Players[0].LowScore, "0.5 刻みが落ちている")
	assert.Equal(t, "4.5", out.Players[0].HighScore)

	// 19 点 → 7 側は超過。
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 9)})
	out = s27Decode(t, p.Output(g, nil))
	assert.Equal(t, "-", out.Players[0].LowScore, "超過した側が数字で出ている")
	assert.Equal(t, "19", out.Players[0].HighScore)
}

// 決着後は両側の勝者と、誰がどちらを取ったかが届くこと。
func TestSevenTwentySevenWebPresenter_Output_CarriesTheOutcome(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 6)})
	g.SetHandForTest(1, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 9),
		s27Card(domain.CardDesignClover, 8)})
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.SetHandForTest(i, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 10),
			s27Card(domain.CardDesignClover, 10)})
	}
	g.SetPotForTest(100)
	g.StandEveryoneForTest()
	g.SettleForTest()

	out := s27Decode(t, p.Output(g, nil))
	assert.Equal(t, 0, out.LowWinner)
	assert.Equal(t, 1, out.HighWinner)
	assert.True(t, out.Players[0].WonLow)
	assert.False(t, out.Players[0].WonHigh)
	assert.True(t, out.Players[1].WonHigh)
	assert.False(t, out.Players[2].WonLow)
	// 決着後は全員の手札が見える。
	assert.NotEmpty(t, out.Players[1].Cards, "決着後も相手の手札が伏せられている")
	assert.NotEmpty(t, out.Players[1].HighScore)
}

// **受動ヒントは Output にも載る。** hint 専用のレスポンスはページの state に
// マージされないので、ここで載せないと state.hint が常に undefined。
func TestSevenTwentySevenWebPresenter_Output_CarriesThePassiveHint(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 10)})

	out := s27Decode(t, p.Output(g, nil))
	require.NotNil(t, out.Hint)
	assert.True(t, out.Hint.Draw)
	assert.Equal(t, "chase_twentyseven", out.Hint.Reason)

	// 止まっていれば助言しない。
	g.SetStandingForTest(0, true)
	assert.Nil(t, s27Decode(t, p.Output(g, nil)).Hint)
}

func TestSevenTwentySevenWebPresenter_Output_ShowsTheError(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	out := s27Decode(t, p.Output(g, domain.ErrWrongPhase))
	assert.Contains(t, out.Message, domain.ErrWrongPhase.Error())
}

// 試合終了メッセージが出ること。
func TestSevenTwentySevenWebPresenter_Output_AnnouncesTheMatchWinner(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()

	for guard := 0; guard < 60 && !g.GetGameEndFlag(); guard++ {
		switch g.GetPhase() {
		case domain.SevenTwentySevenPhaseDraw:
			require.NoError(t, g.TakeCard(false))
		case domain.SevenTwentySevenPhaseResult:
			g.NextRound()
		}
	}
	require.True(t, g.GetGameEndFlag())

	out := s27Decode(t, p.Output(g, nil))
	assert.True(t, out.GameEndFlag)
	assert.GreaterOrEqual(t, out.MatchWinnerIdx, 0)
	assert.NotEmpty(t, out.MessageCode, "終了のメッセージコードが出ていない")
}

// hint コマンド専用のレスポンス。
func TestSevenTwentySevenWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	g.SetHandForTest(0, []*domain.Card{s27Card(domain.CardDesignSpade, 10), s27Card(domain.CardDesignHeart, 10)})

	out := s27Decode(t, p.HintOutput(g))
	require.NotNil(t, out.Hint)
	assert.True(t, out.Hint.Draw)
	assert.Equal(t, "chase_twentyseven", out.Hint.Reason)

	// 助言できないときも壊れないこと。
	g.SetStandingForTest(0, true)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestSevenTwentySevenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SevenTwentySevenWebPresenter)
	g := domain.NewDefaultSevenTwentySeven()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
