//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newSpeedGame() *domain.Speed {
	tc := domain.NewTrumpCards(0)
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	return domain.NewSpeed(tc, players, domain.DefaultSpeedConfig())
}

func newCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func TestNewSpeed(t *testing.T) {
	s := newSpeedGame()
	assert.NotNil(t, s)
	assert.Equal(t, 2, s.GetPlayerCnt())
	assert.Equal(t, -1, s.GetWinnerIdx())
	assert.False(t, s.GetGameEndFlag())
}

func TestSpeed_Reset(t *testing.T) {
	s := newSpeedGame()
	s.Reset()

	// After reset, phase is Play or Stuck depending on whether the random deal
	// produces a playable state.
	assert.Contains(t, []domain.SpeedPhase{domain.SpeedPhasePlay, domain.SpeedPhaseStuck}, s.GetPhase())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, -1, s.GetWinnerIdx())

	// 各プレイヤー: 手札4枚, 山札21枚
	for i := 0; i < 2; i++ {
		p := s.GetPlayer(i)
		assert.Equal(t, 4, p.GetCardsSize(), "player %d hand", i)
		assert.Equal(t, 21, p.GetDrawPileSize(), "player %d draw", i)
	}
	// 台札2枚
	for i := 0; i < 2; i++ {
		assert.NotNil(t, s.GetCenterPile(i), "center pile %d", i)
	}
}

func TestSpeed_AdjacentRank(t *testing.T) {
	// isAdjacentRank is unexported, so test via CanPlay with controlled state
	tests := []struct {
		name      string
		cardValue int
		pileValue int
		want      bool
	}{
		{"2-1", 2, 1, true},
		{"1-2", 1, 2, true},
		{"K-A wrap", 13, 1, true},
		{"A-K wrap", 1, 13, true},
		{"same", 5, 5, false},
		{"gap of 2", 5, 7, false},
		{"gap of 2b", 3, 5, false},
		{"12-13", 12, 13, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupSpeedManual(
				[]*domain.Card{newCard(domain.CardDesignSpade, tt.cardValue)},
				[]*domain.Card{newCard(domain.CardDesignClover, 1)},
				newCard(domain.CardDesignDiamond, tt.pileValue),
				newCard(domain.CardDesignHeart, 1),
				nil, nil,
			)
			assert.Equal(t, tt.want, s.CanPlay(0, 0, 0))
		})
	}
}

// setupSpeedManual は手動でゲーム状態をセットアップする
func setupSpeedManual(
	humanHand []*domain.Card,
	cpuHand []*domain.Card,
	centerPile0 *domain.Card,
	centerPile1 *domain.Card,
	humanDraw []*domain.Card,
	cpuDraw []*domain.Card,
) *domain.Speed {
	tc := domain.NewTrumpCards(0)
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	s := domain.NewSpeed(tc, players, domain.DefaultSpeedConfig())
	s.Reset()

	// 手動で上書き: JSON round-trip を使ってセットアップ
	// プレイヤー手札をリセットして手動設定
	human := s.GetPlayer(0)
	cpu := s.GetPlayer(1)
	human.Reset()
	human.ResetDrawPile()
	cpu.Reset()
	cpu.ResetDrawPile()

	for _, c := range humanHand {
		human.AddCard(c)
	}
	for _, c := range cpuHand {
		cpu.AddCard(c)
	}
	for _, c := range humanDraw {
		human.AddToDrawPile(c)
	}
	for _, c := range cpuDraw {
		cpu.AddToDrawPile(c)
	}

	// 台札を設定するためにJSON round-trip
	data, _ := json.Marshal(s)
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)

	cp0, _ := json.Marshal(centerPile0)
	cp1, _ := json.Marshal(centerPile1)
	cpArr := []json.RawMessage{cp0, cp1}
	cpData, _ := json.Marshal(cpArr)
	raw["cp"] = cpData

	// game end / winner をリセット
	raw["ge"], _ = json.Marshal(false)
	raw["wi"], _ = json.Marshal(-1)

	newData, _ := json.Marshal(raw)
	_ = json.Unmarshal(newData, s)

	// 手動セットアップ後にフェーズを再計算
	s.UpdatePhase()

	return s
}

func TestSpeed_PlayerPlay_Valid(t *testing.T) {
	// center pile 0 = 5, hand has 4 (adjacent)
	s := setupSpeedManual(
		[]*domain.Card{
			newCard(domain.CardDesignSpade, 4),
			newCard(domain.CardDesignHeart, 10),
		},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5), // center 0
		newCard(domain.CardDesignSpade, 9),   // center 1
		nil, nil,
	)

	err := s.PlayerPlay(0, 0) // play 4 on pile 0 (which has 5)
	assert.NoError(t, err)
	// center pile 0 should now be the played card (4)
	assert.Equal(t, 4, s.GetCenterPile(0).GetValue())
	// hand should now have 1 card (the 10)
	assert.Equal(t, 1, s.GetPlayer(0).GetCardsSize())
}

func TestSpeed_PlayerPlay_NotAdjacent(t *testing.T) {
	// Human has 10 and 6. 6 is adj to 5 (keeps phase=Play), but 10 is not adj to 5
	s := setupSpeedManual(
		[]*domain.Card{
			newCard(domain.CardDesignSpade, 10),
			newCard(domain.CardDesignHeart, 6),
		},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 8),
		nil, nil,
	)
	assert.Equal(t, domain.SpeedPhasePlay, s.GetPhase())

	err := s.PlayerPlay(0, 0) // 10 is not adjacent to 5
	assert.Equal(t, domain.ErrInvalidPlay, err)
}

func TestSpeed_PlayerPlay_InvalidCard(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	err := s.PlayerPlay(5, 0) // out of range
	assert.Equal(t, domain.ErrInvalidCard, err)
}

func TestSpeed_PlayerPlay_WrongPhase(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)
	// This state is stuck (no plays), but we need to force stuck phase
	s.UpdatePhase()
	assert.Equal(t, domain.SpeedPhaseStuck, s.GetPhase())

	err := s.PlayerPlay(0, 0)
	assert.Equal(t, domain.ErrWrongPhase, err)
}

func TestSpeed_PlayerPlay_InvalidPileIndex(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	err := s.PlayerPlay(0, 2) // invalid pile index
	assert.Equal(t, domain.ErrInvalidPlay, err)
}

func TestSpeed_PlayerPlay_RefillsHand(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{
			newCard(domain.CardDesignSpade, 4),
			newCard(domain.CardDesignHeart, 8),
			newCard(domain.CardDesignClover, 11),
			newCard(domain.CardDesignDiamond, 2),
		},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		[]*domain.Card{newCard(domain.CardDesignHeart, 3)}, // draw pile
		nil,
	)

	err := s.PlayerPlay(0, 0) // play 4 on pile 0 (5)
	assert.NoError(t, err)
	// Should refill to 4 from draw pile
	assert.Equal(t, 4, s.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 0, s.GetPlayer(0).GetDrawPileSize())
}

func TestSpeed_WinCondition(t *testing.T) {
	// Human has 1 card, no draw pile -> play it -> win
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 6)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7), newCard(domain.CardDesignHeart, 3)},
		newCard(domain.CardDesignDiamond, 5), // 6 is adjacent to 5
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	err := s.PlayerPlay(0, 0)
	assert.NoError(t, err)
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetWinnerIdx())
	assert.Equal(t, domain.SpeedPhaseGameEnd, s.GetPhase())
}

func TestSpeed_CpuPlay_Easy(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{
			newCard(domain.CardDesignClover, 4),
			newCard(domain.CardDesignHeart, 8),
		},
		newCard(domain.CardDesignDiamond, 5), // 4 is adjacent
		newCard(domain.CardDesignSpade, 9),   // 8 is adjacent
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyEasy})

	actions := s.CpuPlay()
	assert.Len(t, actions, 1) // Easy plays only 1
}

func TestSpeed_CpuPlay_Greedy(t *testing.T) {
	// CPU has 4,3 and center is 5,9. Can play 4→pile0, then 3→pile0
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{
			newCard(domain.CardDesignClover, 4),
			newCard(domain.CardDesignHeart, 3),
		},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyNormal})

	actions := s.CpuPlay()
	assert.GreaterOrEqual(t, len(actions), 1) // Should play multiple cards
}

func TestSpeed_CpuPlay_NoValidPlay(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)

	actions := s.CpuPlay()
	assert.Empty(t, actions)
}

func TestSpeed_CpuPlay_WinCondition(t *testing.T) {
	// CPU has 1 card (6), no draw pile, center has 5 -> CPU wins
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 6)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)

	actions := s.CpuPlay()
	assert.Len(t, actions, 1)
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 1, s.GetWinnerIdx())
}

func TestSpeed_IsStuck(t *testing.T) {
	// No valid plays for either player
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)
	assert.True(t, s.IsStuck())
}

func TestSpeed_IsStuck_False(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5), // 4 is adjacent to 5
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)
	assert.False(t, s.IsStuck())
}

func TestSpeed_Flip(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		[]*domain.Card{newCard(domain.CardDesignHeart, 7)},
		[]*domain.Card{newCard(domain.CardDesignDiamond, 8)},
	)
	s.UpdatePhase()
	assert.Equal(t, domain.SpeedPhaseStuck, s.GetPhase())

	err := s.Flip()
	assert.NoError(t, err)
	// Center piles should have new cards
	assert.Equal(t, 7, s.GetCenterPile(0).GetValue())
	assert.Equal(t, 8, s.GetCenterPile(1).GetValue())
}

func TestSpeed_Flip_WrongPhase(t *testing.T) {
	// Playable state: 4 is adjacent to 5, so phase=Play
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 6)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 8),
		nil, nil,
	)
	err := s.Flip()
	assert.Equal(t, domain.ErrWrongPhase, err)
}

func TestSpeed_Flip_DrawPileExhausted(t *testing.T) {
	// Both draw piles empty, stuck -> resolve by card count
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{
			newCard(domain.CardDesignClover, 10),
			newCard(domain.CardDesignHeart, 11),
		},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil, // no draw piles
	)
	s.UpdatePhase()
	assert.Equal(t, domain.SpeedPhaseStuck, s.GetPhase())

	err := s.Flip()
	assert.NoError(t, err)
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetWinnerIdx()) // human has 1, cpu has 2 -> human wins
}

func TestSpeed_GetHint(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{
			newCard(domain.CardDesignSpade, 10),
			newCard(domain.CardDesignHeart, 6),
		},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5), // 6 is adjacent to 5
		newCard(domain.CardDesignSpade, 9),   // 10 is adjacent to 9
		nil, nil,
	)

	ci, pi, found := s.GetHint()
	assert.True(t, found)
	// Should find first valid play (card 0=10 on pile 1=9, or card 1=6 on pile 0=5)
	assert.True(t, ci >= 0 && pi >= 0)
}

func TestSpeed_GetHint_NoPlay(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)

	_, _, found := s.GetHint()
	assert.False(t, found)
}

func TestSpeed_CanPlay(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	assert.True(t, s.CanPlay(0, 0, 0))  // 4 adj to 5
	assert.False(t, s.CanPlay(0, 0, 1)) // 4 not adj to 9
	assert.False(t, s.CanPlay(1, 0, 1)) // 7 not adj to 9 (diff=2)
}

func TestSpeed_CanPlay_EdgeCases(t *testing.T) {
	s := newSpeedGame()
	s.Reset()

	assert.False(t, s.CanPlay(-1, 0, 0))
	assert.False(t, s.CanPlay(0, -1, 0))
	assert.False(t, s.CanPlay(0, 0, -1))
	assert.False(t, s.CanPlay(3, 0, 0))
	assert.False(t, s.CanPlay(0, 0, 3))
}

func TestSpeed_GetPlayer_OutOfRange(t *testing.T) {
	s := newSpeedGame()
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(2))
}

func TestSpeed_GetCenterPile_OutOfRange(t *testing.T) {
	s := newSpeedGame()
	assert.Nil(t, s.GetCenterPile(-1))
	assert.Nil(t, s.GetCenterPile(2))
}

func TestSpeed_IsHumanTurn(t *testing.T) {
	// Playable state: 4 is adjacent to 5
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 8),
		nil, nil,
	)
	assert.True(t, s.IsHumanTurn())
}

func TestSpeed_ActionLog(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	err := s.PlayerPlay(0, 0)
	require.NoError(t, err)
	log := s.GetActionLog()
	require.Len(t, log, 1)
	assert.Equal(t, 0, log[0].PlayerIdx)
	assert.Equal(t, "play", log[0].ActionType)
}

func TestSpeed_JSON(t *testing.T) {
	s := newSpeedGame()
	s.Reset()

	data, err := json.Marshal(s)
	require.NoError(t, err)

	tc := domain.NewTrumpCards(0)
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	restored := domain.NewSpeed(tc, players, domain.DefaultSpeedConfig())
	err = json.Unmarshal(data, restored)
	require.NoError(t, err)

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, s.GetWinnerIdx(), restored.GetWinnerIdx())
	for i := 0; i < 2; i++ {
		assert.Equal(t, s.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
		assert.Equal(t, s.GetPlayer(i).GetDrawPileSize(), restored.GetPlayer(i).GetDrawPileSize())
		assert.Equal(t, s.GetCenterPile(i).GetValue(), restored.GetCenterPile(i).GetValue())
	}
}

func TestSpeed_PlayerHasAnyPlay(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)

	assert.True(t, s.PlayerHasAnyPlay(0))  // 4 adj 5
	assert.False(t, s.PlayerHasAnyPlay(1)) // 10 not adj 5 or 2
}

func TestSpeed_PlayerHasAnyPlay_OutOfRange(t *testing.T) {
	s := newSpeedGame()
	assert.False(t, s.PlayerHasAnyPlay(-1))
	assert.False(t, s.PlayerHasAnyPlay(5))
}

func TestSpeed_SetConfig(t *testing.T) {
	s := newSpeedGame()
	cfg := domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard}
	s.SetConfig(cfg)
	assert.Equal(t, cfg, s.GetConfig())
}

func TestSpeed_Flip_PartialDrawPile(t *testing.T) {
	// Only player 0 has draw pile
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		[]*domain.Card{newCard(domain.CardDesignHeart, 7)},
		nil, // CPU has no draw pile
	)
	s.UpdatePhase()

	err := s.Flip()
	assert.NoError(t, err)
	// center pile 0 updated, pile 1 unchanged
	assert.Equal(t, 7, s.GetCenterPile(0).GetValue())
}

func TestSpeed_CpuPlay_GameEndDoesNothing(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 6)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)
	// Human wins
	_ = s.PlayerPlay(0, 0)
	assert.True(t, s.GetGameEndFlag())

	// CPU should not play after game end
	actions := s.CpuPlay()
	assert.Empty(t, actions)
}

func TestSpeed_KingAceWrap(t *testing.T) {
	// K(13) adjacent to A(1)
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 13)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 1), // A
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	err := s.PlayerPlay(0, 0) // K on A
	assert.NoError(t, err)
	assert.Equal(t, 13, s.GetCenterPile(0).GetValue())
}

func TestSpeed_CpuPlay_Hard_Basic(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 4)},
		newCard(domain.CardDesignDiamond, 5), // 4 adj to 5
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard})

	actions := s.CpuPlay()
	assert.NotEmpty(t, actions)
	assert.Equal(t, 4, s.GetCenterPile(0).GetValue())
}

func TestSpeed_CpuPlay_Hard_MultipleCards(t *testing.T) {
	// CPU has chain 4,3,2 and pile 0 = 5. Should play all 3 in sequence.
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{
			newCard(domain.CardDesignClover, 4),
			newCard(domain.CardDesignHeart, 3),
			newCard(domain.CardDesignDiamond, 2),
		},
		newCard(domain.CardDesignSpade, 5),
		newCard(domain.CardDesignHeart, 12),
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard})

	actions := s.CpuPlay()
	assert.GreaterOrEqual(t, len(actions), 2, "should play multiple cards in a chain")
}

func TestSpeed_CpuPlay_Hard_BlockingChoice(t *testing.T) {
	// CPU has 5 and 8. Human has 6. Pile 0=6, Pile 1=9.
	// 5 on pile 0 (6->5): human's 6 adj to 5 -> opp=1 (bad)
	// 8 on pile 1 (9->8): human's 6 not adj to 8 -> opp=0 (blocks)
	// Hard should prefer 8 on pile 1.
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignHeart, 6)},
		[]*domain.Card{
			newCard(domain.CardDesignClover, 5),
			newCard(domain.CardDesignDiamond, 8),
		},
		newCard(domain.CardDesignSpade, 6), // pile 0: 5 adj (6->5, human 6 adj to 5)
		newCard(domain.CardDesignHeart, 9), // pile 1: 8 adj (9->8, human 6 NOT adj to 8)
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard})

	actions := s.CpuPlay()
	require.NotEmpty(t, actions)
	// First action should prefer the blocking play (8 on pile 1)
	assert.Equal(t, 1, actions[0].PileIndex, "should prefer pile that blocks opponent")
}

func TestSpeed_CpuPlay_Hard_WinCondition(t *testing.T) {
	// CPU has 1 card (6), no draw pile, center has 5 -> CPU wins
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 6)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard})

	actions := s.CpuPlay()
	assert.Len(t, actions, 1)
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 1, s.GetWinnerIdx())
}

func TestSpeed_CpuPlay_Hard_NoValidPlay(t *testing.T) {
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 10)},
		[]*domain.Card{newCard(domain.CardDesignClover, 10)},
		newCard(domain.CardDesignDiamond, 5),
		newCard(domain.CardDesignSpade, 2),
		nil, nil,
	)
	s.SetConfig(domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard})

	actions := s.CpuPlay()
	assert.Empty(t, actions)
}

func TestSpeed_AceKingWrap(t *testing.T) {
	// A(1) adjacent to K(13)
	s := setupSpeedManual(
		[]*domain.Card{newCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{newCard(domain.CardDesignClover, 7)},
		newCard(domain.CardDesignDiamond, 13), // K
		newCard(domain.CardDesignSpade, 9),
		nil, nil,
	)

	err := s.PlayerPlay(0, 0) // A on K
	assert.NoError(t, err)
	assert.Equal(t, 1, s.GetCenterPile(0).GetValue())
}
