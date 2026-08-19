//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPolignac(t *testing.T) *Polignac {
	t.Helper()
	p := NewDefaultPolignac()
	p.Reset()
	require.NoError(t, p.PassDeclaration())
	return p
}

// --- デッキと配り ---

func TestPolignac_DeckIsPiquet(t *testing.T) {
	p := newTestPolignac(t)

	bySuit := map[int]map[int]bool{}
	total := 0
	for i := range PolignacPlayerCnt {
		pl := p.GetPlayer(i)
		for j := range pl.GetCardsSize() {
			c := pl.GetCard(j)
			if bySuit[c.GetDesign()] == nil {
				bySuit[c.GetDesign()] = map[int]bool{}
			}
			bySuit[c.GetDesign()][c.GetValue()] = true
			total++
		}
	}

	assert.Equal(t, 32, total)
	want := []int{1, 7, 8, 9, 10, 11, 12, 13}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		require.Len(t, bySuit[suit], len(want), "スート %d は 8 枚", suit)
		for _, v := range want {
			assert.True(t, bySuit[suit][v], "スート %d に値 %d がある", suit, v)
		}
	}
}

// **ジャックが4枚とも配られる。** 回避対象が欠けていたらゲームにならない。
func TestPolignac_AllFourJacksAreDealt(t *testing.T) {
	p := newTestPolignac(t)
	jacks := map[int]bool{}
	for i := range PolignacPlayerCnt {
		pl := p.GetPlayer(i)
		for j := range pl.GetCardsSize() {
			if c := pl.GetCard(j); c.GetValue() == PolignacJackValue {
				jacks[c.GetDesign()] = true
			}
		}
	}
	assert.Len(t, jacks, 4, "4 スートのジャックが揃っている")
}

func TestPolignac_ResetDealsEightEach(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()

	for i := range PolignacPlayerCnt {
		assert.Equal(t, PolignacHandSize, p.GetPlayer(i).GetCardsSize(), "player %d", i)
		assert.Equal(t, 0, p.GetPlayer(i).GetScore())
	}
	// **配り終えた直後は宣言フェーズ。** いきなりプレイに入らない。
	assert.Equal(t, PolignacPhaseDeclare, p.GetPhase())
	assert.True(t, p.IsDeclarePhase())
	assert.Equal(t, -1, p.GetCapotIdx())
	assert.Equal(t, 1, p.GetRoundNumber())
}

// --- 失点の重み付け ---

// **スペードのジャックだけが 2 点。** これがこのゲームの肝。
func TestPolignacCardPenalty(t *testing.T) {
	assert.Equal(t, 2, PolignacCardPenalty(NewCard(CardDesignSpade, 11, false)), "♠J = Polignac = 2点")
	for _, suit := range []int{CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		assert.Equal(t, 1, PolignacCardPenalty(NewCard(suit, 11, false)), "他のJ = 1点 (suit %d)", suit)
	}
	// ジャック以外は 0。近い札を明示的に踏む。
	for _, v := range []int{1, 10, 12, 13} {
		assert.Equal(t, 0, PolignacCardPenalty(NewCard(CardDesignSpade, v, false)), "値 %d は無罰", v)
	}
	assert.Equal(t, 0, PolignacCardPenalty(nil))
}

// 1 ラウンドで動く失点は必ず 5（1+1+1+2）。
func TestPolignac_TotalRoundPenaltyIsFive(t *testing.T) {
	total := 0
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		total += PolignacCardPenalty(NewCard(suit, PolignacJackValue, false))
	}
	assert.Equal(t, 5, total)
}

func TestPolignac_JackPenaltyGoesToTheTrickWinner(t *testing.T) {
	p := newTestPolignac(t)
	p.trickNumber = 2
	p.leadPlayerIdx = 0
	// ♠J を出したのは 3 番だが、A を出した 0 番がトリックを取る。
	p.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 12, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 11, false)},
	}
	p.resolveTrick()

	assert.Equal(t, 2, p.GetPlayer(0).GetRoundPenalty(), "取った人に ♠J の 2 点")
	assert.Equal(t, 0, p.GetPlayer(3).GetRoundPenalty(), "出した人には付かない")
}

// 複数のジャックが乗ったトリックは合算される。
func TestPolignac_MultipleJacksInOneTrick(t *testing.T) {
	p := newTestPolignac(t)
	p.trickNumber = 2
	p.leadPlayerIdx = 1
	p.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 11, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 11, false)},
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 11, false)},
	}
	p.resolveTrick()

	assert.Equal(t, 4, p.GetPlayer(1).GetRoundPenalty(), "♥J(1) + ♠J(2) + ♣J(1) = 4")
}

// ジャックの無いトリックは無罰。負のコントロール。
func TestPolignac_TrickWithoutJacksCostsNothing(t *testing.T) {
	p := newTestPolignac(t)
	p.trickNumber = 2
	p.leadPlayerIdx = 0
	p.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 12, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)},
	}
	p.resolveTrick()

	for i := range PolignacPlayerCnt {
		assert.Equal(t, 0, p.GetPlayer(i).GetRoundPenalty(), "player %d", i)
	}
}

// --- 切り札なし・フォロー義務 ---

func TestPolignac_NoTrump(t *testing.T) {
	p := newTestPolignac(t)
	p.leadPlayerIdx = 0
	p.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 1, false)},
	}
	assert.Equal(t, 0, p.trickWinner(), "リードのスートの 7 が他スートの A 3 枚に勝つ")
}

func TestPolignacRank_AceIsHighest(t *testing.T) {
	assert.Greater(t, polignacRank(NewCard(CardDesignSpade, 1, false)), polignacRank(NewCard(CardDesignSpade, 13, false)))
	assert.Greater(t, polignacRank(NewCard(CardDesignSpade, 13, false)), polignacRank(NewCard(CardDesignSpade, 7, false)))
	assert.Equal(t, 0, polignacRank(nil))
}

func TestPolignac_MustFollowSuit(t *testing.T) {
	p := newTestPolignac(t)
	pl := p.GetPlayer(1)
	pl.Reset()
	pl.AddCard(NewCard(CardDesignSpade, 8, false))
	pl.AddCard(NewCard(CardDesignHeart, 9, false))
	p.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}}

	assert.Equal(t, []int{0}, p.GetValidPlayIndices(1))
}

func TestPolignac_VoidPlaysAnything(t *testing.T) {
	p := newTestPolignac(t)
	pl := p.GetPlayer(1)
	pl.Reset()
	pl.AddCard(NewCard(CardDesignHeart, 9, false))
	pl.AddCard(NewCard(CardDesignClover, 10, false))
	p.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}}

	assert.Equal(t, []int{0, 1}, p.GetValidPlayIndices(1))
}

func TestPolignac_GetValidPlayIndicesOutOfRange(t *testing.T) {
	p := newTestPolignac(t)
	assert.Nil(t, p.GetValidPlayIndices(-1))
	assert.Nil(t, p.GetValidPlayIndices(PolignacPlayerCnt))
}

// --- capot ---

func TestPolignac_DeclareCapotStartsPlay(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()
	require.NoError(t, p.DeclareCapot())

	assert.Equal(t, 0, p.GetCapotIdx())
	assert.True(t, p.GetPlayer(0).GetDeclaredCapot())
	assert.Equal(t, PolignacPhasePlay, p.GetPhase())
}

func TestPolignac_PassDeclarationStartsPlay(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()
	require.NoError(t, p.PassDeclaration())

	assert.Equal(t, -1, p.GetCapotIdx())
	assert.Equal(t, PolignacPhasePlay, p.GetPhase())
}

func TestPolignac_DeclarationRejectedOutsidePhase(t *testing.T) {
	p := newTestPolignac(t) // 既に PassDeclaration 済み
	assert.Error(t, p.DeclareCapot())
	assert.Error(t, p.PassDeclaration())
}

func TestPolignac_DoubleCapotRejected(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()
	require.NoError(t, p.DeclareCapot())
	p.SetPhaseForTest(PolignacPhaseDeclare) // フェーズだけ戻す
	assert.Error(t, p.DeclareCapot(), "二重宣言はできない")
}

// **capot は「全8トリック」。** issue は「全4枚のジャック」と書いているが、
// それでは簡単すぎて賭けにならない。7 トリックでは失敗になることを固定する。
func TestPolignac_CapotNeedsAllEightTricks(t *testing.T) {
	t.Run("seven tricks fails", func(t *testing.T) {
		p := NewDefaultPolignac()
		p.Reset()
		p.config.Rounds = 2
		require.NoError(t, p.DeclareCapot())
		p.SetCapotTricksForTest(PolignacTricksPerRound - 1)
		p.FinishRoundForTest()

		assert.Equal(t, PolignacCapotStake, p.GetPlayer(0).GetScore(), "宣言者に 5 失点")
		for i := 1; i < PolignacPlayerCnt; i++ {
			assert.Equal(t, 0, p.GetPlayer(i).GetScore(), "他は無傷 (player %d)", i)
		}
	})

	t.Run("all eight succeeds", func(t *testing.T) {
		p := NewDefaultPolignac()
		p.Reset()
		p.config.Rounds = 2
		require.NoError(t, p.DeclareCapot())
		p.SetCapotTricksForTest(PolignacTricksPerRound)
		p.FinishRoundForTest()

		assert.Equal(t, 0, p.GetPlayer(0).GetScore(), "宣言者は無傷")
		for i := 1; i < PolignacPlayerCnt; i++ {
			assert.Equal(t, PolignacCapotStake, p.GetPlayer(i).GetScore(), "他全員に 5 失点 (player %d)", i)
		}
	})
}

// **capot 成功時はジャックの失点を科さない。** 全部取っているので、
// そのまま足すと成功しても損になってしまう。
func TestPolignac_SuccessfulCapotWaivesJackPenalties(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()
	p.config.Rounds = 2
	require.NoError(t, p.DeclareCapot())
	p.GetPlayer(0).SetRoundPenalty(5) // 全ジャックを取っている
	p.SetCapotTricksForTest(PolignacTricksPerRound)
	p.FinishRoundForTest()

	assert.Equal(t, 0, p.GetPlayer(0).GetScore(), "成功なら 5 点の失点は帳消し")
}

// 失敗した場合はジャックの失点も通常どおり科される。
func TestPolignac_FailedCapotStillPaysJackPenalties(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()
	p.config.Rounds = 2
	require.NoError(t, p.DeclareCapot())
	p.GetPlayer(0).SetRoundPenalty(2)
	p.GetPlayer(1).SetRoundPenalty(3)
	p.SetCapotTricksForTest(3)
	p.FinishRoundForTest()

	assert.Equal(t, PolignacCapotStake+2, p.GetPlayer(0).GetScore(), "5 + ジャック 2")
	assert.Equal(t, 3, p.GetPlayer(1).GetScore())
}

// --- 得点 ---

func TestPolignac_ScoringAtRoundEnd(t *testing.T) {
	p := newTestPolignac(t)
	p.config.Rounds = 2
	p.GetPlayer(0).SetRoundPenalty(2)
	p.GetPlayer(2).SetRoundPenalty(3)

	p.FinishRoundForTest()

	assert.Equal(t, 2, p.GetPlayer(0).GetScore())
	assert.Equal(t, 0, p.GetPlayer(1).GetScore())
	assert.Equal(t, 3, p.GetPlayer(2).GetScore())
	assert.Equal(t, PolignacPhaseRoundEnd, p.GetPhase())
}

// **失点が最も少ないプレイヤーが勝つ。** スロバーハンネスとは符号の向きが逆。
func TestPolignac_LowestScoreWins(t *testing.T) {
	p := newTestPolignac(t)
	p.GetPlayer(0).SetScore(7)
	p.GetPlayer(1).SetScore(2)
	p.GetPlayer(2).SetScore(9)
	p.GetPlayer(3).SetScore(5)

	p.FinishGameForTest()

	assert.True(t, p.GetGameEndFlag())
	assert.Equal(t, PolignacPhaseGameEnd, p.GetPhase())
	assert.Equal(t, 1, p.GetWinnerIdx(), "失点 2 が最少")
}

func TestPolignac_TieHasNoWinner(t *testing.T) {
	p := newTestPolignac(t)
	for i := range PolignacPlayerCnt {
		p.GetPlayer(i).SetScore(3)
	}
	p.FinishGameForTest()
	assert.Equal(t, -1, p.GetWinnerIdx())
}

func TestPolignac_RoundEndVsGameEnd(t *testing.T) {
	t.Run("more rounds remain", func(t *testing.T) {
		p := newTestPolignac(t)
		p.config.Rounds = 3
		p.roundNumber = 1
		p.FinishRoundForTest()
		assert.Equal(t, PolignacPhaseRoundEnd, p.GetPhase())
		assert.False(t, p.GetGameEndFlag())
	})
	t.Run("final round", func(t *testing.T) {
		p := newTestPolignac(t)
		p.config.Rounds = 3
		p.roundNumber = 3
		p.FinishRoundForTest()
		assert.Equal(t, PolignacPhaseGameEnd, p.GetPhase())
		assert.True(t, p.GetGameEndFlag())
	})
}

func TestPolignac_NextRoundRedealsAndKeepsScore(t *testing.T) {
	p := newTestPolignac(t)
	p.config.Rounds = 3
	p.GetPlayer(0).SetScore(4)
	p.GetPlayer(0).SetDeclaredCapot(true)
	p.SetPhaseForTest(PolignacPhaseRoundEnd)
	dealer := p.GetDealerIdx()

	p.NextRound()

	assert.Equal(t, 2, p.GetRoundNumber())
	// **次のラウンドも宣言フェーズから始まる。**
	assert.Equal(t, PolignacPhaseDeclare, p.GetPhase())
	assert.Equal(t, (dealer+1)%PolignacPlayerCnt, p.GetDealerIdx(), "ディーラーが回る")
	assert.Equal(t, 4, p.GetPlayer(0).GetScore(), "累計失点は持ち越す")
	assert.False(t, p.GetPlayer(0).GetDeclaredCapot(), "宣言はラウンドごとに消える")
	assert.Equal(t, -1, p.GetCapotIdx())
	for i := range PolignacPlayerCnt {
		assert.Equal(t, PolignacHandSize, p.GetPlayer(i).GetCardsSize())
	}
}

func TestPolignac_NextRoundIgnoredOutsideRoundEnd(t *testing.T) {
	p := newTestPolignac(t)
	p.NextRound()
	assert.Equal(t, 1, p.GetRoundNumber())

	p.gameEndFlag = true
	p.SetPhaseForTest(PolignacPhaseRoundEnd)
	p.NextRound()
	assert.Equal(t, 1, p.GetRoundNumber())
}

// --- プレイ ---

func TestPolignac_PlayerPlayRejections(t *testing.T) {
	t.Run("declare phase", func(t *testing.T) {
		p := NewDefaultPolignac()
		p.Reset()
		assert.Error(t, p.PlayerPlay(0), "宣言前は出せない")
	})
	t.Run("not your turn", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(2)
		assert.Error(t, p.PlayerPlay(0))
	})
	t.Run("game over", func(t *testing.T) {
		p := newTestPolignac(t)
		p.gameEndFlag = true
		assert.Error(t, p.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, p.PlayerPlay(99))
		assert.Error(t, p.PlayerPlay(-1))
	})
	t.Run("must follow suit", func(t *testing.T) {
		p := newTestPolignac(t)
		pl := p.GetPlayer(0)
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 8, false))
		pl.AddCard(NewCard(CardDesignHeart, 9, false))
		p.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
		p.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, p.PlayerPlay(1))
		assert.NoError(t, p.PlayerPlay(0))
	})
}

func TestPolignac_CpuPlayIsANoOpOnHumanTurn(t *testing.T) {
	p := newTestPolignac(t)
	p.SetCurrentPlayerIdxForTest(0)
	p.CpuPlay()
	assert.Equal(t, PolignacHandSize, p.GetPlayer(0).GetCardsSize())
}

// CPU は合法手しか出さず、ゲームは必ず終わる。
func TestPolignac_CpuAlwaysPlaysLegally(t *testing.T) {
	for range 100 {
		p := NewDefaultPolignac()
		p.Reset()
		guard := 0
		for !p.GetGameEndFlag() && guard < 1000 {
			guard++
			if p.IsDeclarePhase() {
				require.NoError(t, p.PassDeclaration())
				continue
			}
			if p.GetPhase() == PolignacPhaseRoundEnd {
				p.NextRound()
				continue
			}
			if p.IsHumanTurn() {
				valid := p.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, p.PlayerPlay(valid[0]))
				continue
			}
			idx := p.GetCurrentPlayerIdx()
			before := p.GetPlayer(idx).GetCardsSize()
			p.CpuPlay()
			require.Equal(t, before-1, p.GetPlayer(idx).GetCardsSize())
		}
		require.True(t, p.GetGameEndFlag())
	}
}

// **1 ラウンドで配られる失点は必ず 5。** ジャックは 4 枚しかなく、
// 誰かが必ず取る。
func TestPolignac_ExactlyFivePenaltyPointsPerRound(t *testing.T) {
	for range 50 {
		p := NewDefaultPolignac()
		p.Reset()
		p.config.Rounds = 1
		require.NoError(t, p.PassDeclaration())
		guard := 0
		for !p.GetGameEndFlag() && guard < 200 {
			guard++
			if p.IsHumanTurn() {
				valid := p.GetValidPlayIndices(0)
				require.NoError(t, p.PlayerPlay(valid[0]))
				continue
			}
			p.CpuPlay()
		}
		total := 0
		for i := range PolignacPlayerCnt {
			total += p.GetPlayer(i).GetScore()
		}
		require.Equal(t, 5, total, "1+1+1+2 = 5")
	}
}

func TestPolignac_GiveUp(t *testing.T) {
	p := newTestPolignac(t)
	p.GiveUp()
	assert.True(t, p.GetGameEndFlag())
	assert.Equal(t, PolignacPhaseGameEnd, p.GetPhase())
	assert.Equal(t, -1, p.GetWinnerIdx())

	p.GiveUp()
	assert.True(t, p.GetGameEndFlag())
}

func TestPolignac_IsHumanTurnAndDeclarePhase(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()
	assert.True(t, p.IsDeclarePhase())
	assert.False(t, p.IsHumanTurn(), "宣言中は出し手の手番ではない")

	require.NoError(t, p.PassDeclaration())
	p.SetCurrentPlayerIdxForTest(0)
	assert.True(t, p.IsHumanTurn())
	assert.False(t, p.IsDeclarePhase())

	p.SetCurrentPlayerIdxForTest(2)
	assert.False(t, p.IsHumanTurn())

	p.SetCurrentPlayerIdxForTest(0)
	p.gameEndFlag = true
	assert.False(t, p.IsHumanTurn())
	assert.False(t, p.IsDeclarePhase())
}

func TestPolignac_GetPlayerOutOfRange(t *testing.T) {
	p := newTestPolignac(t)
	assert.Nil(t, p.GetPlayer(-1))
	assert.Nil(t, p.GetPlayer(PolignacPlayerCnt))
}

func TestPolignac_Config(t *testing.T) {
	p := newTestPolignac(t)
	assert.Equal(t, PolignacRoundsDefault, p.GetConfig().Rounds)

	p.SetConfig(PolignacConfig{Rounds: 6})
	assert.Equal(t, 6, p.GetConfig().Rounds)

	assert.NoError(t, PolignacConfig{Rounds: PolignacRoundsMin}.Validate())
	assert.NoError(t, PolignacConfig{Rounds: PolignacRoundsMax}.Validate())
	assert.Error(t, PolignacConfig{Rounds: 0}.Validate())
	assert.Error(t, PolignacConfig{Rounds: PolignacRoundsMax + 1}.Validate())
}

// --- ヒント ---

func TestPolignac_GetHint_NilWhenNotHumanTurn(t *testing.T) {
	p := newTestPolignac(t)
	p.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, p.GetHint())

	p.SetCurrentPlayerIdxForTest(0)
	p.gameEndFlag = true
	assert.Nil(t, p.GetHint())
}

func TestPolignac_GetHint_SuggestsALegalCard(t *testing.T) {
	p := newTestPolignac(t)
	p.SetCurrentPlayerIdxForTest(0)

	h := p.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, p.GetValidPlayIndices(0), *h.CardIndex)
		assert.NotEmpty(t, h.Reason)
	}
}

// 4 つの理由キーがそれぞれ出る条件を全部踏む。
func TestPolignac_GetHint_Reasons(t *testing.T) {
	spadeTen := []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)}}

	t.Run("lead", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.currentTrick = nil
		assert.Equal(t, "polignacLeadSafe", p.GetHint().Reason)
	})

	t.Run("a jack is on the table", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 11, false)}}
		assert.Equal(t, "polignacAvoidJack", p.GetHint().Reason)
	})

	t.Run("no jack on the table", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.currentTrick = spadeTen
		assert.Equal(t, "polignacDumpJack", p.GetHint().Reason)
	})

	// 誰かが capot を宣言していれば、狙いは失点回避より妨害に変わる。
	t.Run("blocking someone else's capot", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.SetCapotIdxForTest(2)
		p.currentTrick = spadeTen
		assert.Equal(t, "polignacBlockCapot", p.GetHint().Reason)
	})

	// **自分が宣言していたら狙いは丸ごと反転する。** 全トリック取るしかないので、
	// 「取らないように」と助言してはいけない。リードでもフォローでも同じ。
	t.Run("your own capot inverts the aim", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			trick []*TrickCard
		}{
			{"leading", nil},
			{"following", spadeTen},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p := newTestPolignac(t)
				p.SetCurrentPlayerIdxForTest(0)
				p.SetCapotIdxForTest(0)
				p.currentTrick = tc.trick
				assert.Equal(t, "polignacWinCapot", p.GetHint().Reason)
			})
		}
	})
}

// capot 宣言者へのヒントは**実際に取りに行く札**を指す。理由キーだけ直して
// 選ぶ札が回避のままでは意味がない。
func TestPolignac_CapotDeclarerPlaysToWin(t *testing.T) {
	t.Run("leads the strongest card", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.SetCapotIdxForTest(0)
		p.currentTrick = nil
		pl := p.GetPlayer(0)
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 7, false))
		pl.AddCard(NewCard(CardDesignSpade, 1, false)) // A が最強
		pl.AddCard(NewCard(CardDesignSpade, 9, false))

		h := p.GetHint()
		require.NotNil(t, h)
		require.NotNil(t, h.CardIndex)
		assert.Equal(t, 1, *h.CardIndex, "A を出す")
	})

	// フォローでは「取れる中で一番安い札」。無駄に強い札を使わない。
	t.Run("follows with the cheapest winner", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.SetCapotIdxForTest(0)
		p.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)}}
		pl := p.GetPlayer(0)
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 7, false))  // 負ける
		pl.AddCard(NewCard(CardDesignSpade, 12, false)) // 勝てる最安
		pl.AddCard(NewCard(CardDesignSpade, 1, false))  // 勝てるが高い

		h := p.GetHint()
		require.NotNil(t, h)
		require.NotNil(t, h.CardIndex)
		assert.Equal(t, 1, *h.CardIndex, "Q で取る（A は温存）")
	})

	// 取れないなら一番弱い札を捨てる。そのトリックはどのみち落ちる。
	t.Run("discards low when it cannot win", func(t *testing.T) {
		p := newTestPolignac(t)
		p.SetCurrentPlayerIdxForTest(0)
		p.SetCapotIdxForTest(0)
		p.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 1, false)}}
		pl := p.GetPlayer(0)
		pl.Reset()
		pl.AddCard(NewCard(CardDesignSpade, 13, false))
		pl.AddCard(NewCard(CardDesignSpade, 7, false))

		h := p.GetHint()
		require.NotNil(t, h)
		require.NotNil(t, h.CardIndex)
		assert.Equal(t, 1, *h.CardIndex, "7 を捨てる")
	})
}

// --- JSON 往復 ---

func TestPolignac_JSONRoundTrip(t *testing.T) {
	p := newTestPolignac(t)
	p.GetPlayer(0).SetScore(4)
	p.GetPlayer(0).SetRoundPenalty(2)
	p.GetPlayer(1).SetScore(1)
	p.SetCapotIdxForTest(1)
	p.GetPlayer(1).SetDeclaredCapot(true)
	p.SetCapotTricksForTest(3)
	p.roundNumber = 2
	p.trickNumber = 5
	p.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 11, false)}}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	restored := NewDefaultPolignac()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 4, restored.GetPlayer(0).GetScore())
	assert.Equal(t, 2, restored.GetPlayer(0).GetRoundPenalty(), "ラウンド失点が往復する")
	assert.Equal(t, 1, restored.GetPlayer(1).GetScore())
	assert.True(t, restored.GetPlayer(1).GetDeclaredCapot(), "capot 宣言が往復する")
	assert.Equal(t, 1, restored.GetCapotIdx())
	assert.Equal(t, 3, restored.GetCapotTricks())
	assert.Equal(t, 2, restored.GetRoundNumber())
	assert.Equal(t, 5, restored.GetTrickNumber())
	assert.Equal(t, p.GetPhase(), restored.GetPhase())
	assert.Equal(t, p.GetConfig().Rounds, restored.GetConfig().Rounds)
	require.Len(t, restored.GetCurrentTrick(), 1)
	assert.Equal(t, PolignacJackValue, restored.GetCurrentTrick()[0].Card.GetValue())
}

func TestPolignac_UnmarshalRejectsGarbage(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("not json"), NewDefaultPolignac()))
}

func TestPolignac_ActionLog(t *testing.T) {
	p := newTestPolignac(t)
	assert.NotEmpty(t, p.GetActionLog())
}

// **合計失点と内訳が食い違わないこと** (#5746)。内訳は獲得済みトリックから
// 数え、失点は resolveTrick が積む。別々に育つ 2 つの値なので、一致は
// 偶然ではなく規則。
func TestPolignacTakenJackSuits_MatchTheRoundPenalty(t *testing.T) {
	p := NewDefaultPolignac()
	p.Reset()

	player := p.GetPlayer(0)
	// ♠J (2点) と ♥J (1点)、それに失点しない札を混ぜて取らせる。
	player.AddTrick([]*Card{
		NewCard(CardDesignHeart, PolignacJackValue, true),
		NewCard(CardDesignHeart, 3, true),
	})
	player.AddTrick([]*Card{
		NewCard(CardDesignSpade, PolignacJackValue, true),
		NewCard(CardDesignClover, 9, true),
	})
	player.AddRoundPenalty(PolignacSpadeJackPenalty + PolignacJackPenalty)

	suits := player.GetTakenJackSuits()
	// **♠ が先頭。**重い方から読めないと、内訳を出す意味が薄い。
	if len(suits) != 2 || suits[0] != CardDesignSpade || suits[1] != CardDesignHeart {
		t.Fatalf("suits = %v, want [spade heart]", suits)
	}

	total := 0
	for _, suit := range suits {
		total += PolignacCardPenalty(NewCard(suit, PolignacJackValue, true))
	}
	if total != player.GetRoundPenalty() {
		t.Errorf("the breakdown sums to %d but the round penalty is %d", total, player.GetRoundPenalty())
	}

	// 失点しない札しか取っていない席は内訳も空。
	if got := p.GetPlayer(1).GetTakenJackSuits(); len(got) != 0 {
		t.Errorf("suits = %v, want none", got)
	}
}
