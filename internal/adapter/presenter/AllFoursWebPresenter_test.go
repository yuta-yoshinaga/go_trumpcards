//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeAllFoursPlayers() []*domain.AllFoursPlayer {
	return []*domain.AllFoursPlayer{
		domain.NewAllFoursPlayer(true),
		domain.NewAllFoursPlayer(false),
	}
}

func setupAllFoursWebMock() *interfaces.MockAllFoursGame {
	m := new(interfaces.MockAllFoursGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetDealerIdx").Return(1)
	m.On("GetNonDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetTurnUp").Return((*domain.Card)(nil))
	m.On("GetRunCount").Return(0)
	m.On("GetLastRunCount").Return(0).Maybe()
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.AllFoursPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultAllFoursConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupAllFoursWebMockWithPlayers() (*interfaces.MockAllFoursGame, []*domain.AllFoursPlayer) {
	m := setupAllFoursWebMock()
	players := makeAllFoursPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestAllFoursWebPresenter_Output(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupAllFoursWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		var resObj controller.AllFoursWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Equal(t, 2, len(resObj.Players))
		assert.Equal(t, int(domain.AllFoursPhasePlay), resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 1, resObj.DealerIdx)
		assert.Equal(t, 0, resObj.NonDealerIdx)
		assert.Equal(t, []int{0, 1}, resObj.ValidPlayIndices)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupAllFoursWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.False(t, resObj.Players[1].IsHuman)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
	})

	t.Run("error message included", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		result := p.Output(m, errors.New("invalid play"))
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "invalid play", resObj.Message)
	})

	t.Run("game end shows winner code", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Equal(t, "allfours.result.humanWin", resObj.MessageCode)
	})
}

func TestAllFoursWebPresenter_RoundBreakdown(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)

	// setupRoundEnd returns a mock at ROUND_END with the given trump suit and
	// captured tricks assigned per player. tricks[i] is the list of card sets
	// captured by player i.
	setupRoundEnd := func(trump int, tricks [][][]*domain.Card) *interfaces.MockAllFoursGame {
		m := setupAllFoursWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetPhase").Return(domain.AllFoursPhaseRoundEnd)
		m.On("GetTrumpSuit").Return(trump)
		m.On("GetPlayerCnt").Return(2)
		players := makeAllFoursPlayers()
		for i, pt := range tricks {
			for _, trick := range pt {
				players[i].AddTrick(trick)
			}
		}
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		return m
	}

	decode := func(m *interfaces.MockAllFoursGame) *controller.AllFoursWebOutputRoundBreakdown {
		var resObj controller.AllFoursWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		return resObj.RoundBreakdown
	}

	// **プレイ中も出す (#4771)。**High / Low / Jack / Game はトリックが進むたびに
	// 途中経過が確定していくのに、ラウンド終了まで隠れていた。この t.Run は
	// その挙動を「仕様」として固定していたので、中身を入れ替えた。
	t.Run("breakdown is present during play, flagged provisional", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		var resObj controller.AllFoursWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		if assert.NotNil(t, resObj.RoundBreakdown) {
			assert.True(t, resObj.RoundBreakdown.Provisional,
				"途中の値を確定値として渡してはいけない")
		}
	})

	t.Run("the settled breakdown is not provisional", func(t *testing.T) {
		m := setupRoundEnd(domain.CardDesignSpade, [][][]*domain.Card{
			{{domain.NewCard(domain.CardDesignSpade, 1, false)}},
			{{domain.NewCard(domain.CardDesignSpade, 2, false)}},
		})
		bd := decode(m)
		if assert.NotNil(t, bd) {
			assert.False(t, bd.Provisional)
		}
	})

	t.Run("no breakdown before the trick play starts", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.AllFoursPhaseBeg)
		var resObj controller.AllFoursWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &resObj))
		assert.Nil(t, resObj.RoundBreakdown)
	})

	t.Run("high/low/jack/game winners", func(t *testing.T) {
		// Trump = Spade. Player 0 captures the trump Jack (11) and Ace (1, high).
		// Player 1 captures the trump 2 (low). Pips: p0 = J(1)+A(4)=5, p1 = 2(0)=0.
		m := setupRoundEnd(domain.CardDesignSpade, [][][]*domain.Card{
			{{domain.NewCard(domain.CardDesignSpade, 11, false), domain.NewCard(domain.CardDesignSpade, 1, false)}},
			{{domain.NewCard(domain.CardDesignSpade, 2, false)}},
		})
		bd := decode(m)
		assert.NotNil(t, bd)
		assert.Equal(t, 0, bd.High.WinnerIdx)
		assert.Equal(t, 1, bd.High.Card.Value) // Ace
		assert.Equal(t, 1, bd.Low.WinnerIdx)
		assert.Equal(t, 2, bd.Low.Card.Value)
		assert.Equal(t, 0, bd.Jack.WinnerIdx)
		assert.Equal(t, 0, bd.Game.WinnerIdx)
		assert.Equal(t, []int{5, 0}, bd.Game.Points)
	})

	t.Run("jack absent yields -1", func(t *testing.T) {
		// Trump = Heart, no Jack captured. Game tie (both 0) yields no game award.
		m := setupRoundEnd(domain.CardDesignHeart, [][][]*domain.Card{
			{{domain.NewCard(domain.CardDesignHeart, 5, false)}},
			{{domain.NewCard(domain.CardDesignHeart, 3, false)}},
		})
		bd := decode(m)
		assert.Equal(t, -1, bd.Jack.WinnerIdx)
		assert.Equal(t, 0, bd.High.WinnerIdx) // Heart 5 is the highest captured trump
		assert.Equal(t, 1, bd.Low.WinnerIdx)  // Heart 3 is the lowest
		assert.Equal(t, -1, bd.Game.WinnerIdx)
	})

	t.Run("trump unset produces empty award slots", func(t *testing.T) {
		m := setupRoundEnd(domain.AllFoursTrumpUnset, [][][]*domain.Card{{}, {}})
		bd := decode(m)
		assert.NotNil(t, bd)
		assert.Equal(t, -1, bd.High.WinnerIdx)
		assert.Nil(t, bd.High.Card)
		assert.Equal(t, -1, bd.Low.WinnerIdx)
		assert.Equal(t, -1, bd.Jack.WinnerIdx)
		assert.Equal(t, -1, bd.Game.WinnerIdx)
	})
}

func TestAllFoursWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.AllFoursHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})

	t.Run("with beg hint", func(t *testing.T) {
		m, _ := setupAllFoursWebMockWithPlayers()
		beg := true
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return(&domain.AllFoursHint{Beg: &beg, Reason: "beg_beg"})
		result := p.HintOutput(m)
		var resObj controller.AllFoursWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.Hint)
		assert.Equal(t, "beg_beg", resObj.Hint.Reason)
		assert.True(t, *resObj.Hint.Beg)
	})
}

func TestAllFoursWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)
	m := new(interfaces.MockAllFoursGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "stand", Detail: "You stand"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "You stand")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// トリックテイキング系は Output 側にゲートを置きません。AllFours.GetHint() が
// 「人間の手番で、かつ行動を選べる状態か」を自分で確かめて nil を返します。
func TestAllFoursWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 1
	afg, _ := setupAllFoursWebMockWithPlayers()
	afg.ExpectedCalls = removeMockCall(afg.ExpectedCalls, "GetHint")
	afg.On("GetHint").Return(&domain.AllFoursHint{CardIndex: &idx, Reason: "lead_low"})

	result := new(presenter.AllFoursWebPresenter).Output(afg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **手札が急に増えて切り札も変わったのに、最初の Beg と同じ文面だった。**
// run は非親へ 3 枚ずつ追加配布して新しいめくり札を出す All Fours 固有の規則で、
// 起きたことを知る手掛かりが画面に何も無かった (#6479)。
func TestAllFoursWebPresenter_TellsThePlayerTheCardsWereRun(t *testing.T) {
	p := new(presenter.AllFoursWebPresenter)

	withRun := func(phase domain.AllFoursPhase, runs, trickNo int) (string, map[string]string) {
		m, _ := setupAllFoursWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastRunCount")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrickNumber")
		m.On("GetLastRunCount").Return(runs)
		m.On("GetPhase").Return(phase)
		m.On("GetTrickNumber").Return(trickNo)

		var out struct {
			MessageCode   string            `json:"messageCode"`
			MessageParams map[string]string `json:"messageParams"`
		}
		require.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		return out.MessageCode, out.MessageParams
	}

	// 配り直して Beg に戻った道。**ここが issue の本体** ── `runCount` は 0 に
	// 戻っているので、区別できるのは `GetLastRunCount` だけ。
	t.Run("beg after a run says so", func(t *testing.T) {
		code, params := withRun(domain.AllFoursPhaseBeg, 1, 1)
		assert.Equal(t, "allfours.begAfterRun", code)
		assert.Equal(t, map[string]string{"count": "1"}, params)
	})

	// 連続した run も回数がそのまま出る (受け入れ条件 3)。
	t.Run("consecutive runs report their count", func(t *testing.T) {
		_, params := withRun(domain.AllFoursPhaseBeg, 4, 1)
		assert.Equal(t, map[string]string{"count": "4"}, params)
	})

	// run が起きていない Beg は今までどおり。
	t.Run("a first beg keeps the plain message", func(t *testing.T) {
		code, params := withRun(domain.AllFoursPhaseBeg, 0, 1)
		assert.Equal(t, "allfours.begPhase", code)
		assert.Nil(t, params)
	})

	// 切り札が変わってプレイに入った道。
	t.Run("the first lead after a run says so", func(t *testing.T) {
		code, params := withRun(domain.AllFoursPhasePlay, 2, 1)
		assert.Equal(t, "allfours.playAfterRun", code)
		assert.Equal(t, map[string]string{"count": "2"}, params)
	})

	// **一度打てば普段の文面に戻る。**毎トリック言い続けると、その局のあいだ
	// ずっと「いま run が起きた」と読める。
	t.Run("later tricks go back to the plain lead message", func(t *testing.T) {
		code, _ := withRun(domain.AllFoursPhasePlay, 2, 3)
		assert.Equal(t, "allfours.playPhase.lead", code)
	})
}
