//go:build test

package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func machiavelliCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func newTestMachiavelli(n int) *domain.Machiavelli {
	players := make([]*domain.MachiavelliPlayer, n)
	players[0] = domain.NewMachiavelliPlayer(true)
	for i := 1; i < n; i++ {
		players[i] = domain.NewMachiavelliPlayer(false)
	}
	cfg := domain.DefaultMachiavelliConfig()
	cfg.PlayerCount = n
	return domain.NewMachiavelli(domain.NewTrumpCardsWithDecks(2, 0), players, cfg)
}

func machiavelliSetHand(p *domain.MachiavelliPlayer, cards []*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func machiavelliRefsOf(melds ...[]*domain.Card) [][]domain.MachiavelliCardRef {
	out := make([][]domain.MachiavelliCardRef, len(melds))
	for i, meld := range melds {
		refs := make([]domain.MachiavelliCardRef, len(meld))
		for j, c := range meld {
			refs[j] = domain.MachiavelliCardRef{Design: c.GetDesign(), Value: c.GetValue()}
		}
		out[i] = refs
	}
	return out
}

func TestNewMachiavelli(t *testing.T) {
	g := newTestMachiavelli(4)
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetRoundWinnerIdx())
}

func TestMachiavelli_DefaultConstructor(t *testing.T) {
	g := domain.NewDefaultMachiavelli()
	g.Reset()
	assert.Equal(t, domain.MachiavelliDefaultPlayerCount, g.GetPlayerCnt())
}

func TestMachiavelli_Reset(t *testing.T) {
	g := newTestMachiavelli(4)
	g.Reset()

	assert.Equal(t, domain.MachiavelliPhaseTurn, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 4, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.MachiavelliHandSize, g.GetPlayer(i).GetCardsSize())
	}
	// 104 - 4*13 = 52 枚が山札
	assert.Equal(t, 104-4*domain.MachiavelliHandSize, g.GetDrawPileCount())
	assert.Empty(t, g.GetTable())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // ディーラー（席 0）の左隣
}

func TestMachiavelliConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultMachiavelliConfig().Validate())

	bad := domain.DefaultMachiavelliConfig()
	bad.PlayerCount = 1
	assert.Error(t, bad.Validate())

	bad2 := domain.DefaultMachiavelliConfig()
	bad2.PlayerCount = 6
	assert.Error(t, bad2.Validate())

	bad3 := domain.DefaultMachiavelliConfig()
	bad3.CpuDifficulty = 9
	assert.Error(t, bad3.Validate())

	bad4 := domain.DefaultMachiavelliConfig()
	bad4.TargetRounds = 0
	assert.Error(t, bad4.Validate())
}

func TestMachiavelli_IsSet(t *testing.T) {
	// 有効: 同ランク別スート 3 枚
	assert.True(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignHeart, 9),
		machiavelliCard(domain.CardDesignClover, 9),
	}))
	// 有効: 4 枚（全スート）
	assert.True(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignHeart, 9),
		machiavelliCard(domain.CardDesignClover, 9),
		machiavelliCard(domain.CardDesignDiamond, 9),
	}))
	// 無効: スート重複
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignHeart, 9),
	}))
	// 無効: ランク不一致（かつラン不成立）
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignHeart, 9),
		machiavelliCard(domain.CardDesignClover, 5),
	}))
	// 無効: 2 枚（枚数不足）
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignHeart, 9),
	}))
	// 無効: ジョーカーは使えない
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignJoker, 0),
		machiavelliCard(domain.CardDesignHeart, 9),
		machiavelliCard(domain.CardDesignClover, 9),
	}))
}

func TestMachiavelli_IsRun(t *testing.T) {
	// 有効: 同スート連続 3 枚
	assert.True(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignSpade, 6),
	}))
	// 有効: Ace-low (A-2-3)
	assert.True(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignHeart, 1),
		machiavelliCard(domain.CardDesignHeart, 2),
		machiavelliCard(domain.CardDesignHeart, 3),
	}))
	// 有効: Ace-high (Q-K-A)
	assert.True(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignDiamond, 12),
		machiavelliCard(domain.CardDesignDiamond, 13),
		machiavelliCard(domain.CardDesignDiamond, 1),
	}))
	// 無効: ラップアラウンド (K-A-2)
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignDiamond, 13),
		machiavelliCard(domain.CardDesignDiamond, 1),
		machiavelliCard(domain.CardDesignDiamond, 2),
	}))
	// 無効: スート不一致
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignHeart, 5),
		machiavelliCard(domain.CardDesignSpade, 6),
	}))
	// 無効: 連続でない
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignSpade, 7),
	}))
	// 無効: 値重複
	assert.False(t, domain.MachiavelliIsValidMeld([]*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	}))
}

func TestMachiavelli_CardPoints(t *testing.T) {
	assert.Equal(t, 1, domain.MachiavelliCardPoints(machiavelliCard(domain.CardDesignSpade, 1)))   // Ace
	assert.Equal(t, 7, domain.MachiavelliCardPoints(machiavelliCard(domain.CardDesignSpade, 7)))   // pip
	assert.Equal(t, 10, domain.MachiavelliCardPoints(machiavelliCard(domain.CardDesignSpade, 10))) // Ten
	assert.Equal(t, 10, domain.MachiavelliCardPoints(machiavelliCard(domain.CardDesignSpade, 13))) // King
	assert.Equal(t, 0, domain.MachiavelliCardPoints(nil))
}

func TestMachiavelli_Play_AcceptNewMeld(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetRoundNumber(1)
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignHeart, 7),
	})
	run := []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	}
	err := g.PlayerPlay(machiavelliRefsOf(run), []int{0, 1, 2})
	require.NoError(t, err)
	assert.Len(t, g.GetTable(), 1)
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize()) // H7 だけ残る
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())       // ターンが進む
}

func TestMachiavelli_Play_RebuildExistingTable(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetTable([][]*domain.Card{{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	}})
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 6),
		machiavelliCard(domain.CardDesignHeart, 9),
	})
	newRun := []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignSpade, 6),
	}
	err := g.PlayerPlay(machiavelliRefsOf(newRun), []int{0})
	require.NoError(t, err)
	assert.Len(t, g.GetTable()[0], 4)
}

func TestMachiavelli_Play_Rejections(t *testing.T) {
	base := func() *domain.Machiavelli {
		g := newTestMachiavelli(2)
		g.SetCurrentPlayerIdx(0)
		g.SetPhase(domain.MachiavelliPhaseTurn)
		machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
			machiavelliCard(domain.CardDesignSpade, 3),
			machiavelliCard(domain.CardDesignSpade, 4),
			machiavelliCard(domain.CardDesignHeart, 5),
		})
		return g
	}

	t.Run("invalid meld", func(t *testing.T) {
		g := base()
		bad := []*domain.Card{
			machiavelliCard(domain.CardDesignSpade, 3),
			machiavelliCard(domain.CardDesignSpade, 4),
			machiavelliCard(domain.CardDesignHeart, 5),
		}
		err := g.PlayerPlay(machiavelliRefsOf(bad), []int{0, 1, 2})
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("card multiset mismatch", func(t *testing.T) {
		g := base()
		// 手札は S3,S4,H5 だが場に S3,S4,S5 を出そうとする（H5 ≠ S5）
		mismatch := []*domain.Card{
			machiavelliCard(domain.CardDesignSpade, 3),
			machiavelliCard(domain.CardDesignSpade, 4),
			machiavelliCard(domain.CardDesignSpade, 5),
		}
		err := g.PlayerPlay(machiavelliRefsOf(mismatch), []int{0, 1, 2})
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("no hand card added", func(t *testing.T) {
		g := base()
		g.SetTable([][]*domain.Card{{
			machiavelliCard(domain.CardDesignClover, 9),
			machiavelliCard(domain.CardDesignHeart, 9),
			machiavelliCard(domain.CardDesignSpade, 9),
		}})
		same := []*domain.Card{
			machiavelliCard(domain.CardDesignClover, 9),
			machiavelliCard(domain.CardDesignHeart, 9),
			machiavelliCard(domain.CardDesignSpade, 9),
		}
		err := g.PlayerPlay(machiavelliRefsOf(same), []int{})
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("bad card ref", func(t *testing.T) {
		g := base()
		refs := [][]domain.MachiavelliCardRef{{{Design: 0, Value: 99}}}
		err := g.PlayerPlay(refs, []int{0})
		assert.ErrorIs(t, err, domain.ErrInvalidCard)
	})

	t.Run("empty meld", func(t *testing.T) {
		g := base()
		refs := [][]domain.MachiavelliCardRef{{}}
		err := g.PlayerPlay(refs, []int{0})
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("wrong phase", func(t *testing.T) {
		g := base()
		g.SetPhase(domain.MachiavelliPhaseRoundEnd)
		err := g.PlayerPlay(machiavelliRefsOf([]*domain.Card{}), []int{0})
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		g := base()
		g.SetCurrentPlayerIdx(1)
		err := g.PlayerPlay(machiavelliRefsOf([]*domain.Card{}), []int{0})
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})
}

func TestMachiavelli_NewMeld_WinAndScore(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetRoundNumber(1)
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	})
	// 対戦相手の手札 = K + 7 → デッドウッド 17
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{
		machiavelliCard(domain.CardDesignHeart, 13),
		machiavelliCard(domain.CardDesignHeart, 7),
	})
	err := g.PlayerNewMeld([]int{0, 1, 2})
	require.NoError(t, err)
	assert.Equal(t, domain.MachiavelliPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetRoundWinnerIdx())
	assert.Equal(t, 0, g.GetPlayer(0).GetRoundScore())
	assert.Equal(t, 17, g.GetPlayer(1).GetRoundScore())
}

func TestMachiavelli_NewMeld_Invalid(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignHeart, 4),
		machiavelliCard(domain.CardDesignClover, 5),
	})
	err := g.PlayerNewMeld([]int{0, 1, 2}) // 無効なメルド
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	// 範囲外インデックス
	assert.Error(t, g.PlayerNewMeld([]int{0, 1, 99}))
}

func TestMachiavelli_Layoff(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetTable([][]*domain.Card{{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	}})
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 6),
		machiavelliCard(domain.CardDesignHeart, 9),
	})
	err := g.PlayerLayoff(0, 0)
	require.NoError(t, err)
	assert.Len(t, g.GetTable()[0], 4)
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())

	// 無効なメルドインデックス
	assert.ErrorIs(t, g.PlayerLayoff(9, 0), domain.ErrInvalidPlay)
	// 範囲外の手札インデックス
	assert.ErrorIs(t, g.PlayerLayoff(0, 99), domain.ErrInvalidCard)
}

func TestMachiavelli_Draw(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{machiavelliCard(domain.CardDesignSpade, 3)})
	g.SetDrawPile([]*domain.Card{machiavelliCard(domain.CardDesignHeart, 8)})

	err := g.PlayerDraw()
	require.NoError(t, err)
	assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 0, g.GetDrawPileCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestMachiavelli_Draw_StockOut(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetRoundNumber(1)
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{machiavelliCard(domain.CardDesignSpade, 3)})
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{machiavelliCard(domain.CardDesignHeart, 4)})
	g.SetDrawPile(nil)

	err := g.PlayerDraw()
	require.NoError(t, err)
	assert.Equal(t, domain.MachiavelliPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetRoundWinnerIdx())
	// 全員デッドウッドを採点
	assert.Equal(t, 3, g.GetPlayer(0).GetRoundScore())
	assert.Equal(t, 4, g.GetPlayer(1).GetRoundScore())
}

func TestMachiavelli_GameEnd(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetRoundNumber(g.GetTargetRounds()) // 最終ラウンド
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	})
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{machiavelliCard(domain.CardDesignHeart, 13)})

	err := g.PlayerNewMeld([]int{0, 1, 2})
	require.NoError(t, err)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.MachiavelliPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx()) // 累計最少（0 点）

	// ゲーム終了後のアクションはブロックされる
	assert.ErrorIs(t, g.PlayerDraw(), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerNewMeld([]int{0}), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerLayoff(0, 0), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerPlay(nil, []int{0}), domain.ErrGameEnded)
}

func TestMachiavelli_NextRound(t *testing.T) {
	g := newTestMachiavelli(2)
	g.Reset()
	g.SetPhase(domain.MachiavelliPhaseRoundEnd)
	g.SetRoundNumber(1)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.MachiavelliPhaseTurn, g.GetPhase())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.MachiavelliHandSize, g.GetPlayer(i).GetCardsSize())
	}

	// フェーズが RoundEnd でなければ no-op
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestMachiavelli_IsHumanTurn(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MachiavelliPhaseRoundEnd)
	assert.False(t, g.IsHumanTurn())
}

func TestMachiavelli_CpuPlay_NewMeld(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 9),
		machiavelliCard(domain.CardDesignHeart, 9),
		machiavelliCard(domain.CardDesignClover, 9),
	})
	g.CpuPlay()
	assert.Len(t, g.GetTable(), 1)
	assert.Equal(t, 0, g.GetPlayer(1).GetCardsSize())
}

func TestMachiavelli_CpuPlay_Run(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignSpade, 6),
		machiavelliCard(domain.CardDesignHeart, 2),
	})
	g.CpuPlay()
	assert.Len(t, g.GetTable(), 1)
	assert.Len(t, g.GetTable()[0], 3)
}

func TestMachiavelli_CpuPlay_Layoff(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetTable([][]*domain.Card{{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	}})
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 6),
		machiavelliCard(domain.CardDesignHeart, 2),
	})
	g.CpuPlay()
	assert.Len(t, g.GetTable()[0], 4) // S6 がレイオフされる
}

func TestMachiavelli_CpuPlay_Draw(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	machiavelliSetHand(g.GetPlayer(1), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignHeart, 7),
		machiavelliCard(domain.CardDesignDiamond, 11),
	})
	g.SetDrawPile([]*domain.Card{machiavelliCard(domain.CardDesignClover, 2)})
	g.CpuPlay()
	assert.Empty(t, g.GetTable())
	assert.Equal(t, 4, g.GetPlayer(1).GetCardsSize()) // 引いた
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())       // ターンが進む
}

func TestMachiavelli_JSON_RoundTrip(t *testing.T) {
	g := newTestMachiavelli(3)
	g.Reset()
	g.SetTable([][]*domain.Card{{
		machiavelliCard(domain.CardDesignSpade, 3),
		machiavelliCard(domain.CardDesignSpade, 4),
		machiavelliCard(domain.CardDesignSpade, 5),
	}})
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	restored := &domain.Machiavelli{}
	require.NoError(t, restored.UnmarshalJSON(data))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Len(t, restored.GetTable(), 1)
}

func machiavelliJSONWith(t *testing.T, g *domain.Machiavelli, key, rawVal string) []byte {
	t.Helper()
	data, err := g.MarshalJSON()
	require.NoError(t, err)
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m))
	m[key] = json.RawMessage(rawVal)
	out, err := json.Marshal(m)
	require.NoError(t, err)
	return out
}

func TestMachiavelli_JSON_ValidationBranches(t *testing.T) {
	valid := func() *domain.Machiavelli {
		g := newTestMachiavelli(2)
		g.Reset()
		return g
	}

	t.Run("bad json", func(t *testing.T) {
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON([]byte("not json")))
	})
	t.Run("bad phase", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "ps", "9")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("nil player", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "pl", "[null]")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("player count out of range", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "pl", "[]")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("currentPlayerIdx out of range", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "ci", "99")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("dealerIdx out of range", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "di", "99")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("winnerIdx out of range", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "wi", "99")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("roundWinnerIdx sentinel ok", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "rw", "-1")
		assert.NoError(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("negative round number", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "rn", "-1")
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("invalid config", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "cf", `{"pc":1,"cd":1,"tr":3}`)
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
	t.Run("nil card in table is filtered", func(t *testing.T) {
		data := machiavelliJSONWith(t, valid(), "tb", "[[null]]")
		restored := &domain.Machiavelli{}
		require.NoError(t, restored.UnmarshalJSON(data))
		assert.Empty(t, restored.GetTable())
	})
	t.Run("oversize array", func(t *testing.T) {
		big := "[" + strings.Repeat("null,", 1001) + "null]"
		data := machiavelliJSONWith(t, valid(), "wp", big)
		assert.Error(t, (&domain.Machiavelli{}).UnmarshalJSON(data))
	})
}

func TestMachiavelli_PlayerDeadwoodValue(t *testing.T) {
	g := newTestMachiavelli(2)
	machiavelliSetHand(g.GetPlayer(0), []*domain.Card{
		machiavelliCard(domain.CardDesignSpade, 13), // 10
		machiavelliCard(domain.CardDesignSpade, 1),  // 1
		machiavelliCard(domain.CardDesignSpade, 7),  // 7
	})
	assert.Equal(t, 18, g.PlayerDeadwoodValue(0))
	assert.Equal(t, 0, g.PlayerDeadwoodValue(99))
}
