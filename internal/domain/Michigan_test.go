//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// michiganCard は指定デザイン・値のカードを生成するテストヘルパー。
func michiganCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// michiganSetHand は player の手札を差し替える。
func michiganSetHand(p *domain.MichiganPlayer, cards ...*domain.Card) {
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// michiganNew3PlayerGame は 3 人構成のゲームを返す (Bet フェーズ、人間のベット待ち)。
func michiganNew3PlayerGame(t *testing.T) *domain.Michigan {
	t.Helper()
	g := domain.NewDefaultMichigan()
	cfg := g.GetConfig()
	cfg.PlayerCount = 3
	g.SetConfig(cfg)
	g.Reset()
	return g
}

func TestMichigan_ResetEntersBetPhase(t *testing.T) {
	g := domain.NewDefaultMichigan()
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.MichiganPhaseBet, g.GetPhase())
	assert.True(t, g.IsHumanTurn())
	assert.False(t, g.GetHumanBetPlaced())
	assert.Equal(t, g.GetConfig().Ante, g.GetBetBudget())
	assert.Equal(t, domain.MichiganBoodleCount, g.GetBoodleCnt())
	// CPUs already bet, so chips accumulated on the boodles.
	total := 0
	for i := 0; i < g.GetBoodleCnt(); i++ {
		total += g.GetBoodle(i).GetChips()
	}
	assert.Greater(t, total, 0)
}

func TestMichigan_PlaceHumanBet_TransitionsToPlay(t *testing.T) {
	g := domain.NewDefaultMichigan()
	// Put the human (seat 0) on lead so no CPU auto-plays before we count the
	// deal — finishBetting drives CPU turns immediately, and a CPU lead would
	// have removed cards from play, making the total non-deterministic.
	g.SetDealerIdx(g.GetPlayerCnt() - 1)
	budget := g.GetBetBudget()
	require.NoError(t, g.PlaceHumanBet(michiganEvenBet(budget)))
	assert.Equal(t, domain.MichiganPhasePlay, g.GetPhase())
	assert.True(t, g.GetHumanBetPlaced())
	// All 52 cards dealt to players + dead hand (nothing played yet: human leads).
	dealt := g.GetDeadHandCount()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		dealt += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, dealt)
	// Lead player is to the dealer's left — seat 0, the human.
	assert.Equal(t, 0, g.GetLeadPlayerIdx())
	assert.True(t, g.IsHumanTurn())
}

// michiganEvenBet は budget を 4 分割した賭けスライスを返す。
func michiganEvenBet(budget int) []int {
	dist := make([]int, domain.MichiganBoodleCount)
	q := budget / domain.MichiganBoodleCount
	r := budget % domain.MichiganBoodleCount
	for i := range dist {
		dist[i] = q
		if i < r {
			dist[i]++
		}
	}
	return dist
}

func TestMichigan_PlaceHumanBet_Errors(t *testing.T) {
	t.Run("wrong length", func(t *testing.T) {
		g := domain.NewDefaultMichigan()
		err := g.PlaceHumanBet([]int{1, 2, 3})
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
	t.Run("wrong total", func(t *testing.T) {
		g := domain.NewDefaultMichigan()
		err := g.PlaceHumanBet([]int{0, 0, 0, 0})
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
	t.Run("negative", func(t *testing.T) {
		g := domain.NewDefaultMichigan()
		bud := g.GetBetBudget()
		err := g.PlaceHumanBet([]int{bud + 1, -1, 0, 0})
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
	t.Run("wrong phase", func(t *testing.T) {
		g := domain.NewDefaultMichigan()
		require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
		err := g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget()))
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})
	t.Run("already placed", func(t *testing.T) {
		g := domain.NewDefaultMichigan()
		g.SetPhase(domain.MichiganPhaseBet)
		// Force humanBetPlaced via a real bet, then attempt again after resetting phase.
		require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
		g.SetPhase(domain.MichiganPhaseBet)
		err := g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget()))
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
}

// michiganSetupPlay は 3 人ゲームをプレイ状態に整える (手札を後から差し替える前提)。
func michiganSetupPlay(t *testing.T) *domain.Michigan {
	t.Helper()
	g := michiganNew3PlayerGame(t)
	require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
	// Clear the randomly dealt hands so tests can install deterministic ones.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).ClearHand()
	}
	return g
}

func TestMichigan_DoPlay_BoodleHitAndStopOnDeadHand(t *testing.T) {
	g := michiganSetupPlay(t)
	michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignHeart, 1), michiganCard(domain.CardDesignSpade, 8))
	michiganSetHand(g.GetPlayer(1), michiganCard(domain.CardDesignSpade, 5))
	michiganSetHand(g.GetPlayer(2), michiganCard(domain.CardDesignClover, 9))
	g.AddDeadCardForTest(michiganCard(domain.CardDesignHeart, 2)) // next in sequence -> stop
	g.SetBoodleForTest(0, 50, -1)                                 // boodle 0 is A of hearts
	g.SetRoundStartChipsForTest([]int{100, 100, 100})
	g.GetPlayer(0).SetChips(100)
	g.SetPlayStateForTest(0, 0, 0, 0)

	g.DoPlayForTest(0, 0) // play A of hearts (index 0)

	assert.Equal(t, 0, g.GetBoodle(0).GetClaimedBy())
	assert.Equal(t, 0, g.GetBoodle(0).GetChips())
	assert.Equal(t, 150, g.GetPlayer(0).GetChips()) // collected 50
	// Next card (heart 2) is in the dead hand -> STOP -> new sequence by last player.
	assert.Equal(t, 0, g.GetSeqSuit())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

func TestMichigan_DoPlay_SequencePassesToHolder(t *testing.T) {
	g := michiganSetupPlay(t)
	michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignHeart, 1), michiganCard(domain.CardDesignSpade, 8))
	michiganSetHand(g.GetPlayer(1), michiganCard(domain.CardDesignHeart, 2), michiganCard(domain.CardDesignClover, 5))
	michiganSetHand(g.GetPlayer(2), michiganCard(domain.CardDesignClover, 9))
	g.SetPlayStateForTest(0, 0, 0, 0)

	g.DoPlayForTest(0, 0) // play heart 1

	assert.Equal(t, domain.CardDesignHeart, g.GetSeqSuit())
	assert.Equal(t, 1, g.GetSeqHighValue())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // seat 1 holds heart 2
}

func TestMichigan_DoPlay_StopPastKing(t *testing.T) {
	g := michiganSetupPlay(t)
	michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignSpade, 13), michiganCard(domain.CardDesignHeart, 3))
	michiganSetHand(g.GetPlayer(1), michiganCard(domain.CardDesignClover, 5))
	michiganSetHand(g.GetPlayer(2), michiganCard(domain.CardDesignDiamond, 9))
	g.SetPlayStateForTest(0, 0, 0, 0)

	g.DoPlayForTest(0, 0) // play spade King -> past-king stop

	assert.Equal(t, 0, g.GetSeqSuit())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

func TestMichigan_DoPlay_RoundEndOnEmptyHand(t *testing.T) {
	g := michiganSetupPlay(t)
	michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignHeart, 1)) // last card
	michiganSetHand(g.GetPlayer(1), michiganCard(domain.CardDesignClover, 5))
	michiganSetHand(g.GetPlayer(2), michiganCard(domain.CardDesignDiamond, 9))
	g.SetBoodleForTest(0, 20, -1)
	g.SetRoundStartChipsForTest([]int{100, 100, 100})
	// Reset every player's chips to the recorded round-start baseline so the
	// human's net gain (the 20-chip boodle) is the only movement this round;
	// michiganSetupPlay's bet phase left the opponents below their baseline.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetChips(100)
	}
	g.SetPlayStateForTest(0, 0, 0, 0)

	g.DoPlayForTest(0, 0)

	assert.Equal(t, domain.MichiganPhaseResult, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.MichiganResultWin, g.GetResult())
	// Double resolution guard: entering round end again is a no-op.
	g.DoPlayForTest(1, 0)
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestMichigan_RealFlow_ReachesResult(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay; i++ {
		if !g.IsHumanTurn() {
			break
		}
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		require.NoError(t, g.PlayCard(pi[0]))
	}
	assert.Equal(t, domain.MichiganPhaseResult, g.GetPhase())
}

func TestMichigan_PlayCard_Errors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := domain.NewDefaultMichigan() // still in Bet phase
		err := g.PlayCard(0)
		assert.True(t, errors.Is(err, domain.ErrWrongPhase))
	})
	t.Run("illegal card", func(t *testing.T) {
		g := michiganSetupPlay(t)
		michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignHeart, 5), michiganCard(domain.CardDesignHeart, 3))
		g.SetPlayStateForTest(0, 0, 0, 0)
		// index 0 is heart 5, but the lowest heart is heart 3 -> illegal lead.
		err := g.PlayCard(0)
		assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
	})
	t.Run("not human turn", func(t *testing.T) {
		g := michiganSetupPlay(t)
		michiganSetHand(g.GetPlayer(1), michiganCard(domain.CardDesignHeart, 5))
		g.SetPlayStateForTest(0, 0, 1, 1) // current player is a CPU
		err := g.PlayCard(0)
		assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
	})
}

func TestMichigan_Hint(t *testing.T) {
	g := michiganSetupPlay(t)
	michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignHeart, 1), michiganCard(domain.CardDesignSpade, 8))
	g.SetBoodleForTest(0, 30, -1) // A of hearts boodle has chips
	g.SetPlayStateForTest(0, 0, 0, 0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "claim_boodle", hint.Reason)

	// Forced hint when a sequence is active.
	michiganSetHand(g.GetPlayer(0), michiganCard(domain.CardDesignHeart, 4))
	g.SetPlayStateForTest(domain.CardDesignHeart, 3, 0, 0)
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "forced", hint.Reason)

	// No hint outside the human's play turn.
	g.SetPhase(domain.MichiganPhaseResult)
	assert.Nil(t, g.GetHint())
}

func TestMichigan_NextRound(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay && g.IsHumanTurn(); i++ {
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		require.NoError(t, g.PlayCard(pi[0]))
	}
	require.Equal(t, domain.MichiganPhaseResult, g.GetPhase())
	if g.GetGameEndFlag() {
		t.Skip("game ended in round 1")
	}
	g.NextRound()
	assert.Equal(t, domain.MichiganPhaseBet, g.GetPhase())
	assert.Equal(t, 2, g.GetRoundNumber())
	// NextRound is a no-op outside the Result phase.
	rn := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, rn, g.GetRoundNumber())
}

func TestMichigan_GameEndByRounds(t *testing.T) {
	g := domain.NewDefaultMichigan()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay && g.IsHumanTurn(); i++ {
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		require.NoError(t, g.PlayCard(pi[0]))
	}
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetMatchWinnerIdx(), 0)
	// NextRound after game end is a no-op.
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestMichiganConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultMichiganConfig().Validate())
	assert.Error(t, domain.MichiganConfig{PlayerCount: 2, Ante: 8, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.MichiganConfig{PlayerCount: 9, Ante: 8, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.MichiganConfig{PlayerCount: 4, Ante: 0, StartingChips: 200, TargetRounds: 10}.Validate())
	assert.Error(t, domain.MichiganConfig{PlayerCount: 4, Ante: 8, StartingChips: 5, TargetRounds: 10}.Validate())
	// Starting chips passes the range check but is below the ante.
	assert.Error(t, domain.MichiganConfig{PlayerCount: 4, Ante: 50, StartingChips: 10, TargetRounds: 10}.Validate())
	assert.Error(t, domain.MichiganConfig{PlayerCount: 4, Ante: 8, StartingChips: 200, TargetRounds: 0}.Validate())
}

func TestMichiganPlayer_JSON(t *testing.T) {
	p := domain.NewMichiganPlayer(true, 300)
	p.AddRoundBet(20)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.MichiganPlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, 300, got.GetChips())
	assert.Equal(t, 20, got.GetRoundBet())
	assert.True(t, got.GetIsHuman())
}

func TestMichiganPlayer_UnmarshalErrors(t *testing.T) {
	var p domain.MichiganPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"ch":-5,"rb":0}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`{"ch":10,"rb":-1}`), &p))
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p))
}

func TestMichigan_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganEvenBet(g.GetBetBudget())))
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var got domain.Michigan
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, g.GetPhase(), got.GetPhase())
	assert.Equal(t, g.GetChips(), got.GetChips())
	assert.Equal(t, g.GetRoundNumber(), got.GetRoundNumber())
	assert.Equal(t, g.GetPlayerCnt(), got.GetPlayerCnt())
	assert.Equal(t, g.GetCurrentPlayerIdx(), got.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetDeadHandCount(), got.GetDeadHandCount())
	assert.Equal(t, g.GetBoodleCnt(), got.GetBoodleCnt())
}

func TestMichigan_UnmarshalValidation(t *testing.T) {
	base := `"cf":{"pc":3,"an":8,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}]`
	cases := map[string]string{
		"not json":        `not json`,
		"invalid config":  `{"cf":{"pc":9,"an":8,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}]}`,
		"player mismatch": `{"cf":{"pc":4,"an":8,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1},{"ch":1}],"ph":0,"rn":1}`,
		"too few players": `{"cf":{"pc":3,"an":8,"sc":200,"tr":10},"ps":[{"ch":1},{"ch":1}],"ph":0,"rn":1}`,
		"invalid phase":   `{` + base + `,"ph":9,"rn":1}`,
		"round zero":      `{` + base + `,"ph":0,"rn":0}`,
		"dealer range":    `{` + base + `,"ph":0,"rn":1,"di":5}`,
		"current range":   `{` + base + `,"ph":0,"rn":1,"ci":5}`,
		"lead range":      `{` + base + `,"ph":0,"rn":1,"li":5}`,
		"seq suit range":  `{` + base + `,"ph":0,"rn":1,"sq":9}`,
		"seq high range":  `{` + base + `,"ph":0,"rn":1,"sh":99}`,
		"winner range":    `{` + base + `,"ph":0,"rn":1,"wi":5}`,
		"result range":    `{` + base + `,"ph":0,"rn":1,"re":9}`,
		"boodle count":    `{` + base + `,"ph":0,"rn":1,"bd":[{"c":{"d":3,"v":1}}]}`,
		"round chips len": `{` + base + `,"ph":0,"rn":1,"rs":[1,2]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.Michigan
			assert.Error(t, json.Unmarshal([]byte(body), &g))
		})
	}
}

func TestMichigan_UnmarshalDefaults(t *testing.T) {
	var g domain.Michigan
	require.NoError(t, json.Unmarshal([]byte(`{"cf":{"pc":3,"an":8,"sc":200,"tr":10},"ps":[{"ch":200},{"ch":200},{"ch":200}],"ph":0,"rn":1}`), &g))
	assert.Equal(t, 3, g.GetPlayerCnt())
	assert.Equal(t, domain.MichiganBoodleCount, g.GetBoodleCnt())
	assert.NotNil(t, g.GetActionLog())
}
