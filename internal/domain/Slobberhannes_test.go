//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSlobberhannes(t *testing.T) *Slobberhannes {
	t.Helper()
	s := NewDefaultSlobberhannes()
	s.Reset()
	return s
}

// --- デッキと配り ---

// **ピケット32枚であること。** issue は Skat の 32 枚生成の流用を指示して
// いたが、あれは ♠1-13/♣1-13/♥1-6/♦なし を返す (#5296)。
func TestSlobberhannes_DeckIsPiquet(t *testing.T) {
	s := newTestSlobberhannes(t)

	bySuit := map[int]map[int]bool{}
	total := 0
	for i := range SlobberhannesPlayerCnt {
		p := s.GetPlayer(i)
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if bySuit[c.GetDesign()] == nil {
				bySuit[c.GetDesign()] = map[int]bool{}
			}
			bySuit[c.GetDesign()][c.GetValue()] = true
			total++
		}
	}

	assert.Equal(t, 32, total, "32枚がちょうど配りきられる")
	want := []int{1, 7, 8, 9, 10, 11, 12, 13}
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		require.Len(t, bySuit[suit], len(want), "スート %d は 8 枚", suit)
		for _, v := range want {
			assert.True(t, bySuit[suit][v], "スート %d に値 %d がある", suit, v)
		}
	}
}

func TestSlobberhannes_ResetDealsEightEach(t *testing.T) {
	s := newTestSlobberhannes(t)

	for i := range SlobberhannesPlayerCnt {
		assert.Equal(t, SlobberhannesHandSize, s.GetPlayer(i).GetCardsSize(), "player %d", i)
		assert.Equal(t, 0, s.GetPlayer(i).GetScore(), "開始時は 0 点")
	}
	assert.Equal(t, SlobberhannesPhasePlay, s.GetPhase())
	assert.Equal(t, 1, s.GetRoundNumber())
	assert.Equal(t, -1, s.GetWinnerIdx())
	// 手番はディーラーの左隣から。
	assert.Equal(t, 1, s.GetCurrentPlayerIdx())
}

// --- 強さとトリック判定 ---

func TestSlobberhannesRank_AceIsHighest(t *testing.T) {
	ace := NewCard(CardDesignSpade, 1, false)
	king := NewCard(CardDesignSpade, 13, false)
	seven := NewCard(CardDesignSpade, 7, false)

	assert.Greater(t, slobberhannesRank(ace), slobberhannesRank(king), "A は K より強い")
	assert.Greater(t, slobberhannesRank(king), slobberhannesRank(seven))
	assert.Equal(t, 0, slobberhannesRank(nil))
}

// **切り札は無い。** 別スートの A を出してもリードのスートの 7 に勝てない。
func TestSlobberhannes_NoTrump(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.leadPlayerIdx = 0
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 1, false)},
	}

	assert.Equal(t, 0, s.trickWinner(),
		"リードのスートの最弱札でも、他スートの A 3枚に勝つ")
}

func TestSlobberhannes_TrickWinnerHighestOfLeadSuit(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.leadPlayerIdx = 0
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 10, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 12, false)},
	}
	assert.Equal(t, 2, s.trickWinner(), "A が勝つ")
}

func TestSlobberhannes_TrickWinnerEmptyTrick(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.leadPlayerIdx = 2
	s.currentTrick = nil
	assert.Equal(t, 2, s.trickWinner())
}

// --- フォロー義務 ---

func TestSlobberhannes_MustFollowSuit(t *testing.T) {
	s := newTestSlobberhannes(t)
	p := s.GetPlayer(1)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 8, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	s.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}}
	s.currentPlayerIdx = 1

	assert.Equal(t, []int{0}, s.GetValidPlayIndices(1), "スペードを持っていれば他は出せない")
}

func TestSlobberhannes_VoidPlaysAnything(t *testing.T) {
	s := newTestSlobberhannes(t)
	p := s.GetPlayer(1)
	p.Reset()
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignDiamond, 10, false))
	s.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}}

	assert.Equal(t, []int{0, 1}, s.GetValidPlayIndices(1))
}

func TestSlobberhannes_GetValidPlayIndicesOutOfRange(t *testing.T) {
	s := newTestSlobberhannes(t)
	assert.Nil(t, s.GetValidPlayIndices(-1))
	assert.Nil(t, s.GetValidPlayIndices(SlobberhannesPlayerCnt))
}

// --- 罰点 ---

func TestSlobberhannes_IsPenaltyQueen(t *testing.T) {
	assert.True(t, slobberhannesIsPenaltyQueen(NewCard(CardDesignClover, 12, false)))
	assert.False(t, slobberhannesIsPenaltyQueen(NewCard(CardDesignSpade, 12, false)), "スペードのQは無罰")
	assert.False(t, slobberhannesIsPenaltyQueen(NewCard(CardDesignClover, 13, false)), "クラブのKは無罰")
	assert.False(t, slobberhannesIsPenaltyQueen(nil))
}

// 最初のトリックを取ると罰点。**位置そのものが罰の対象**というのが
// このゲームの肝なので、中身に関係なく付くことを確かめる。
func TestSlobberhannes_FirstTrickPenalty(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.trickNumber = 0
	s.leadPlayerIdx = 0
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 8, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 9, false)},
	}
	s.resolveTrick()

	assert.True(t, s.GetPlayer(0).GetTookFirstTrick())
	assert.False(t, s.GetPlayer(0).GetTookLastTrick())
	assert.False(t, s.GetPlayer(0).GetTookQueen())
	assert.Equal(t, 1, s.GetPlayer(0).PenaltyCount())
	for i := 1; i < SlobberhannesPlayerCnt; i++ {
		assert.Equal(t, 0, s.GetPlayer(i).PenaltyCount(), "player %d", i)
	}
}

// 中間のトリックには位置の罰が付かない。負のコントロール。
func TestSlobberhannes_MiddleTrickHasNoPositionPenalty(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.trickNumber = 3
	s.leadPlayerIdx = 0
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 8, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 9, false)},
	}
	s.resolveTrick()

	assert.Equal(t, 0, s.GetPlayer(0).PenaltyCount(), "3 番目のトリックは無罰")
}

func TestSlobberhannes_QueenPenaltyGoesToTheTrickWinner(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.trickNumber = 2
	s.leadPlayerIdx = 0
	// ♣Q を出したのは 3 番だが、トリックを取るのは A を出した 0 番。
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 7, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 8, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 12, false)},
	}
	s.resolveTrick()

	assert.True(t, s.GetPlayer(0).GetTookQueen(), "罰は出した人でなく取った人に付く")
	assert.False(t, s.GetPlayer(3).GetTookQueen())
}

func TestSlobberhannes_LastTrickPenalty(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.trickNumber = SlobberhannesTricksPerRound - 1
	s.leadPlayerIdx = 0
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 7, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 8, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 9, false)},
	}
	s.resolveTrick()

	assert.True(t, s.GetPlayer(1).GetTookLastTrick())
}

// --- 得点 ---

func TestSlobberhannes_ScoringAtRoundEnd(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.config.Rounds = 2 // まだ終わらせない
	s.GetPlayer(0).tookFirstTrick = true
	s.GetPlayer(0).tookQueen = true
	s.GetPlayer(1).tookLastTrick = true
	// 2 番と 3 番は無罰 → +1

	s.finishRound()

	assert.Equal(t, -2, s.GetPlayer(0).GetScore(), "罰 2 つで -2")
	assert.Equal(t, -1, s.GetPlayer(1).GetScore())
	assert.Equal(t, 1, s.GetPlayer(2).GetScore(), "全回避で +1")
	assert.Equal(t, 1, s.GetPlayer(3).GetScore())
	assert.Equal(t, SlobberhannesPhaseRoundEnd, s.GetPhase())
}

// 1 人が 3 つとも背負う最悪ケース。
func TestSlobberhannes_AllThreePenaltiesOnOnePlayer(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.config.Rounds = 2
	p := s.GetPlayer(2)
	p.tookFirstTrick, p.tookLastTrick, p.tookQueen = true, true, true

	s.finishRound()
	assert.Equal(t, 3, p.PenaltyCount())
	assert.Equal(t, -3, p.GetScore())
}

// **合計が最大のプレイヤーが勝つ。** issue #5233 は「最も少ない」と書いて
// いるが、罰が -1・ボーナスが +1 なので符号が合わない。
func TestSlobberhannes_HighestScoreWins(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.GetPlayer(0).SetScore(-3)
	s.GetPlayer(1).SetScore(2)
	s.GetPlayer(2).SetScore(-1)
	s.GetPlayer(3).SetScore(0)

	s.finishGame()

	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, SlobberhannesPhaseGameEnd, s.GetPhase())
	assert.Equal(t, 1, s.GetWinnerIdx(), "+2 が最高")
}

func TestSlobberhannes_TieHasNoWinner(t *testing.T) {
	s := newTestSlobberhannes(t)
	for i := range SlobberhannesPlayerCnt {
		s.GetPlayer(i).SetScore(1)
	}
	s.finishGame()
	assert.Equal(t, -1, s.GetWinnerIdx())
}

// 最終ラウンドを終えるとゲームが終わる。ラウンドが残っていれば続く。
func TestSlobberhannes_RoundEndVsGameEnd(t *testing.T) {
	t.Run("more rounds remain", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		s.config.Rounds = 3
		s.roundNumber = 1
		s.finishRound()
		assert.Equal(t, SlobberhannesPhaseRoundEnd, s.GetPhase())
		assert.False(t, s.GetGameEndFlag())
	})
	t.Run("final round", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		s.config.Rounds = 3
		s.roundNumber = 3
		s.finishRound()
		assert.Equal(t, SlobberhannesPhaseGameEnd, s.GetPhase())
		assert.True(t, s.GetGameEndFlag())
	})
}

func TestSlobberhannes_NextRoundDealsAgainAndKeepsScore(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.config.Rounds = 3
	s.GetPlayer(0).SetScore(-2)
	s.GetPlayer(0).tookQueen = true
	s.phase = SlobberhannesPhaseRoundEnd
	dealer := s.GetDealerIdx()

	s.NextRound()

	assert.Equal(t, 2, s.GetRoundNumber())
	assert.Equal(t, SlobberhannesPhasePlay, s.GetPhase())
	assert.Equal(t, (dealer+1)%SlobberhannesPlayerCnt, s.GetDealerIdx(), "ディーラーが回る")
	assert.Equal(t, -2, s.GetPlayer(0).GetScore(), "累計得点は持ち越す")
	assert.False(t, s.GetPlayer(0).GetTookQueen(), "罰の内訳はラウンドごとに消える")
	for i := range SlobberhannesPlayerCnt {
		assert.Equal(t, SlobberhannesHandSize, s.GetPlayer(i).GetCardsSize())
	}
}

func TestSlobberhannes_NextRoundIgnoredOutsideRoundEnd(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.phase = SlobberhannesPhasePlay
	s.NextRound()
	assert.Equal(t, 1, s.GetRoundNumber(), "プレイ中は進まない")

	s.gameEndFlag = true
	s.phase = SlobberhannesPhaseRoundEnd
	s.NextRound()
	assert.Equal(t, 1, s.GetRoundNumber(), "終局後も進まない")
}

// --- プレイ ---

func TestSlobberhannes_PlayerPlayRejections(t *testing.T) {
	t.Run("not your turn", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		s.currentPlayerIdx = 2
		assert.Error(t, s.PlayerPlay(0))
	})
	t.Run("game over", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		s.gameEndFlag = true
		assert.Error(t, s.PlayerPlay(0))
	})
	t.Run("round ended", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		s.phase = SlobberhannesPhaseRoundEnd
		s.currentPlayerIdx = 0
		assert.Error(t, s.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		s.currentPlayerIdx = 0
		assert.Error(t, s.PlayerPlay(99))
		assert.Error(t, s.PlayerPlay(-1))
	})
	t.Run("must follow suit", func(t *testing.T) {
		s := newTestSlobberhannes(t)
		p := s.GetPlayer(0)
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignHeart, 9, false))
		s.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
		s.currentPlayerIdx = 0
		assert.Error(t, s.PlayerPlay(1), "ハートは出せない")
		assert.NoError(t, s.PlayerPlay(0))
	})
}

func TestSlobberhannes_PlayAdvancesTurn(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 0
	s.currentTrick = nil

	require.NoError(t, s.PlayerPlay(0))
	assert.Equal(t, 1, s.GetCurrentPlayerIdx())
	assert.Len(t, s.GetCurrentTrick(), 1)
	assert.Equal(t, SlobberhannesHandSize-1, s.GetPlayer(0).GetCardsSize())
}

func TestSlobberhannes_CpuPlayIsANoOpOnHumanTurn(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 0
	s.CpuPlay()
	assert.Equal(t, SlobberhannesHandSize, s.GetPlayer(0).GetCardsSize())
}

// CPU は合法手しか出さない。1000 回まわして 1 度も違反しないこと。
func TestSlobberhannes_CpuAlwaysPlaysLegally(t *testing.T) {
	for range 100 {
		s := NewDefaultSlobberhannes()
		s.Reset()
		guard := 0
		for !s.GetGameEndFlag() && guard < 1000 {
			guard++
			if s.IsHumanTurn() {
				valid := s.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, s.PlayerPlay(valid[0]))
				continue
			}
			if s.GetPhase() == SlobberhannesPhaseRoundEnd {
				s.NextRound()
				continue
			}
			before := s.GetPlayer(s.GetCurrentPlayerIdx()).GetCardsSize()
			idx := s.GetCurrentPlayerIdx()
			s.CpuPlay()
			require.Equal(t, before-1, s.GetPlayer(idx).GetCardsSize(), "CPU が必ず 1 枚出す")
		}
		require.True(t, s.GetGameEndFlag(), "ゲームが終わる")
	}
}

// 1 ラウンドの罰点は必ず 3 つ配られる（最初・最後・♣Q）。同一人物に
// 重なることはあっても、合計は常に 3。
func TestSlobberhannes_ExactlyThreePenaltiesPerRound(t *testing.T) {
	for range 50 {
		s := NewDefaultSlobberhannes()
		s.config.Rounds = 1
		s.Reset()
		guard := 0
		for !s.GetGameEndFlag() && guard < 200 {
			guard++
			if s.IsHumanTurn() {
				valid := s.GetValidPlayIndices(0)
				require.NoError(t, s.PlayerPlay(valid[0]))
				continue
			}
			s.CpuPlay()
		}
		total := 0
		for i := range SlobberhannesPlayerCnt {
			total += s.GetPlayer(i).PenaltyCount()
		}
		require.Equal(t, 3, total, "罰は毎ラウンドちょうど 3 つ")
	}
}

func TestSlobberhannes_GiveUp(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.GiveUp()
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, SlobberhannesPhaseGameEnd, s.GetPhase())
	assert.Equal(t, -1, s.GetWinnerIdx())

	// 2 度目は何もしない。
	s.GiveUp()
	assert.True(t, s.GetGameEndFlag())
}

func TestSlobberhannes_IsHumanTurn(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 0
	assert.True(t, s.IsHumanTurn())
	s.currentPlayerIdx = 2
	assert.False(t, s.IsHumanTurn())
	s.currentPlayerIdx = 0
	s.phase = SlobberhannesPhaseRoundEnd
	assert.False(t, s.IsHumanTurn(), "ラウンド終了中は手番ではない")
	s.phase = SlobberhannesPhasePlay
	s.gameEndFlag = true
	assert.False(t, s.IsHumanTurn())
}

func TestSlobberhannes_GetPlayerOutOfRange(t *testing.T) {
	s := newTestSlobberhannes(t)
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(SlobberhannesPlayerCnt))
}

func TestSlobberhannes_Config(t *testing.T) {
	s := newTestSlobberhannes(t)
	assert.Equal(t, SlobberhannesRoundsDefault, s.GetConfig().Rounds)

	s.SetConfig(SlobberhannesConfig{Rounds: 6})
	assert.Equal(t, 6, s.GetConfig().Rounds)

	assert.NoError(t, SlobberhannesConfig{Rounds: SlobberhannesRoundsMin}.Validate())
	assert.NoError(t, SlobberhannesConfig{Rounds: SlobberhannesRoundsMax}.Validate())
	assert.Error(t, SlobberhannesConfig{Rounds: 0}.Validate())
	assert.Error(t, SlobberhannesConfig{Rounds: SlobberhannesRoundsMax + 1}.Validate())
}

// --- JSON 往復 ---

// Worker はリクエストごとに KV から作り直す。**得点と罰の内訳が往復しないと
// ラウンド途中で罰が消える** (#4478)。
func TestSlobberhannes_JSONRoundTrip(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.GetPlayer(0).SetScore(-2)
	s.GetPlayer(0).tookQueen = true
	s.GetPlayer(1).SetScore(1)
	s.roundNumber = 2
	s.trickNumber = 5
	s.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 10, false)}}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewDefaultSlobberhannes()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, -2, restored.GetPlayer(0).GetScore())
	assert.True(t, restored.GetPlayer(0).GetTookQueen(), "罰の内訳が往復する")
	assert.Equal(t, 1, restored.GetPlayer(1).GetScore())
	assert.Equal(t, 2, restored.GetRoundNumber())
	assert.Equal(t, 5, restored.GetTrickNumber())
	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, s.GetConfig().Rounds, restored.GetConfig().Rounds)
	require.Len(t, restored.GetCurrentTrick(), 1)
	assert.Equal(t, 10, restored.GetCurrentTrick()[0].Card.GetValue())
	assert.Equal(t, s.GetPlayer(0).GetCardsSize(), restored.GetPlayer(0).GetCardsSize())
}

func TestSlobberhannes_UnmarshalRejectsGarbage(t *testing.T) {
	s := NewDefaultSlobberhannes()
	assert.Error(t, json.Unmarshal([]byte("not json"), s))
}

func TestSlobberhannes_ActionLog(t *testing.T) {
	s := newTestSlobberhannes(t)
	assert.NotEmpty(t, s.GetActionLog(), "配りが棋譜に残る")
}

// --- ヒント ---

func TestSlobberhannes_GetHint_NilWhenNotHumanTurn(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 2
	assert.Nil(t, s.GetHint())

	s.currentPlayerIdx = 0
	s.gameEndFlag = true
	assert.Nil(t, s.GetHint())
}

func TestSlobberhannes_GetHint_SuggestsALegalCard(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 0
	s.currentTrick = nil

	h := s.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, s.GetValidPlayIndices(0), *h.CardIndex)
		assert.NotEmpty(t, h.Reason)
	}
}

// 理由は「危険なトリックを避ける」と「安全に捨てる」で入れ替わる。両側を踏む。
func TestSlobberhannes_GetHint_ReasonReflectsDanger(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 0
	p := s.GetPlayer(0)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 7, false))
	p.AddCard(NewCard(CardDesignSpade, 1, false))

	// ♣Q が場に出ている＝取ったら罰点。
	s.trickNumber = 3
	s.currentTrick = []*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 12, false)},
	}
	danger := s.GetHint()
	if assert.NotNil(t, danger) {
		assert.Equal(t, "slobberhannesAvoid", danger.Reason)
	}

	// 罰点の絡まない中間のトリック。
	s.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)}}
	safe := s.GetHint()
	if assert.NotNil(t, safe) {
		assert.Equal(t, "slobberhannesDump", safe.Reason)
	}
}

// 最初と最後のトリックは中身に関係なく危険。
func TestSlobberhannes_GetHint_PositionIsDangerous(t *testing.T) {
	for _, tc := range []struct {
		name string
		n    int
	}{
		{"first trick", 0},
		{"last trick", SlobberhannesTricksPerRound - 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestSlobberhannes(t)
			s.currentPlayerIdx = 0
			s.trickNumber = tc.n
			s.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)}}
			h := s.GetHint()
			if assert.NotNil(t, h) {
				assert.Equal(t, "slobberhannesAvoid", h.Reason)
			}
		})
	}
}

func TestSlobberhannes_GetHint_LeadReason(t *testing.T) {
	s := newTestSlobberhannes(t)
	s.currentPlayerIdx = 0
	s.trickNumber = 3
	s.currentTrick = nil

	h := s.GetHint()
	if assert.NotNil(t, h) {
		assert.Equal(t, "slobberhannesLeadLow", h.Reason)
	}
}
