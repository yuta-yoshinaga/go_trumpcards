//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRams(t *testing.T) *Rams {
	t.Helper()
	r := NewDefaultRams()
	r.Reset()
	return r
}

// newTestRamsWith 指定人数で作る
func newTestRamsWith(t *testing.T, n int) *Rams {
	t.Helper()
	r := NewDefaultRams()
	r.SetConfig(RamsConfig{PlayerCnt: n, Rounds: RamsRoundsDefault})
	r.Reset()
	return r
}

// ramsTotalChips 全員の持ちチップ + ポット。**ポット制の不変量。**
func ramsTotalChips(r *Rams) int {
	total := r.GetPot()
	for i := range r.GetPlayerCnt() {
		total += r.GetPlayer(i).GetChips()
	}
	return total
}

// --- 可変人数 ---

// **3〜5 人で成立する。** これがラムスの特徴で、Loo（4人固定）との違い。
func TestRams_SupportsThreeToFivePlayers(t *testing.T) {
	for n := RamsPlayerCntMin; n <= RamsPlayerCntMax; n++ {
		t.Run(string(rune('0'+n))+" players", func(t *testing.T) {
			r := newTestRamsWith(t, n)
			assert.Equal(t, n, r.GetPlayerCnt())
			for i := range n {
				assert.Equal(t, RamsHandSize, r.GetPlayer(i).GetCardsSize(), "player %d", i)
			}
			// 5 人でも 5×5+1=26 枚で 32 枚に収まる。
			assert.NotNil(t, r.GetUpCard(), "切り札を決める 1 枚が残る")
			assert.Equal(t, RamsStartingChips*n, ramsTotalChips(r))
		})
	}
}

func TestRams_ResetDealsAndAntes(t *testing.T) {
	r := newTestRams(t)

	for i := range r.GetPlayerCnt() {
		assert.Equal(t, RamsHandSize, r.GetPlayer(i).GetCardsSize(), "player %d", i)
		assert.Equal(t, RamsStartingChips-RamsAnte, r.GetPlayer(i).GetChips(), "player %d", i)
	}
	assert.Equal(t, RamsAnte*r.GetPlayerCnt(), r.GetPot())
	// **配り終えた直後は選択フェーズ。** いきなりプレイに入らない。
	assert.Equal(t, RamsPhaseDecide, r.GetPhase())
	assert.True(t, r.IsDecidePhase())
	assert.Equal(t, 1, r.GetRoundNumber())
}

// **切り札は表向きの 1 枚で決まる。** issue には切り札の規定が無い。
func TestRams_TrumpComesFromTheUpCard(t *testing.T) {
	r := newTestRams(t)
	up := r.GetUpCard()
	require.NotNil(t, up)
	assert.Equal(t, up.GetDesign(), r.GetTrumpSuit())
}

// --- 参加 / 降りる ---

func TestRams_DecideStartsPlay(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(true))

	assert.True(t, r.GetPlayer(0).GetInRound())
	assert.True(t, r.GetPlayer(0).GetDecided())
	// CPU も全員選び終える。
	for i := range r.GetPlayerCnt() {
		assert.True(t, r.GetPlayer(i).GetDecided(), "player %d", i)
	}
	assert.NotEqual(t, RamsPhaseDecide, r.GetPhase())
}

func TestRams_PassKeepsYouOut(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(false))

	assert.False(t, r.GetPlayer(0).GetInRound(), "降りたら参加者でない")
	assert.True(t, r.GetPlayer(0).GetDecided())
}

func TestRams_DecideRejections(t *testing.T) {
	t.Run("twice", func(t *testing.T) {
		r := newTestRams(t)
		require.NoError(t, r.Decide(true))
		assert.Error(t, r.Decide(true), "二度は選べない")
	})
	t.Run("after game end", func(t *testing.T) {
		r := newTestRams(t)
		r.GiveUp()
		assert.Error(t, r.Decide(true))
	})
}

// 降りた人はプレイに参加しない。
func TestRams_PassedPlayerCannotPlay(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(false))
	r.SetPhaseForTest(RamsPhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, r.PlayerPlay(0), "降りた人は札を出せない")
}

// --- ポット ---

// **参加して 0 トリックだと追加で払う。** これがラムスのリスク。
func TestRams_ZeroTrickPlayerPaysPenalty(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 2
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
	}
	r.GetPlayer(0).SetRoundTricks(0)
	r.GetPlayer(1).SetRoundTricks(RamsTricksPerRound)
	chipsBefore := r.GetPlayer(0).GetChips()

	r.FinishRoundForTest()

	assert.Equal(t, chipsBefore-RamsMissPenalty, r.GetPlayer(0).GetChips(),
		"参加して 0 トリックなら追加支払い")
}

// **降りた人は 0 トリックでも追加で払わない。** 負のコントロール。
func TestRams_PassedPlayerPaysNoPenalty(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 2
	r.GetPlayer(0).SetInRound(false)
	r.GetPlayer(1).SetInRound(true)
	r.GetPlayer(1).SetRoundTricks(RamsTricksPerRound)
	chipsBefore := r.GetPlayer(0).GetChips()

	r.FinishRoundForTest()

	assert.Equal(t, chipsBefore, r.GetPlayer(0).GetChips(), "降りていれば失うのはアンティだけ")
}

// トリック数に応じてポットを按分する。
func TestRams_PotIsSharedByTrickCount(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 2
	r.SetPotForTest(20)
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
		r.GetPlayer(i).SetChips(0)
	}
	r.GetPlayer(0).SetRoundTricks(3)
	r.GetPlayer(1).SetRoundTricks(2)
	// 2,3 番は 0 トリックなので追加支払いが発生する。

	r.FinishRoundForTest()

	// ポット 20 + 追加 5×2 = 30、5 トリックで割って 1 トリック 6。
	assert.Equal(t, 18, r.GetPlayer(0).GetChips(), "3 トリック × 6")
	assert.Equal(t, 12, r.GetPlayer(1).GetChips(), "2 トリック × 6")
}

// **端数は次ラウンドへ残す。** 配り切れないぶんを消すとチップが減る。
func TestRams_RemainderStaysInThePot(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 2
	r.SetPotForTest(11) // 2 トリックで割ると 5 余り 1
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(false)
	}
	r.GetPlayer(0).SetInRound(true)
	r.GetPlayer(0).SetRoundTricks(2)
	before := ramsTotalChips(r)

	r.FinishRoundForTest()

	assert.Equal(t, 1, r.GetPot(), "端数 1 が残る")
	assert.Equal(t, before, ramsTotalChips(r), "チップの総量は変わらない")
}

// 全員降りたらポットは持ち越し。
func TestRams_AllPassCarriesThePot(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 3
	pot := r.GetPot()
	before := ramsTotalChips(r)

	// 人間が降り、CPU も全員降りた状態を作る。
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetDecided(true)
		r.GetPlayer(i).SetInRound(false)
	}
	r.StartPlayIfReadyForTest()

	assert.Equal(t, 0, r.GetActiveCount())
	assert.Equal(t, pot, r.GetPot(), "ポットはそのまま残る")
	assert.Equal(t, before, ramsTotalChips(r))
}

// **チップは生まれも消えもしない。** ポット制の最重要不変量。
func TestRams_ChipsAreConserved(t *testing.T) {
	for n := RamsPlayerCntMin; n <= RamsPlayerCntMax; n++ {
		for range 10 {
			r := NewDefaultRams()
			r.SetConfig(RamsConfig{PlayerCnt: n, Rounds: RamsRoundsDefault})
			r.Reset()
			want := RamsStartingChips * n
			require.Equal(t, want, ramsTotalChips(r), "配り直後 (%d人)", n)

			guard := 0
			for !r.GetGameEndFlag() && guard < 2000 {
				guard++
				switch {
				case r.IsDecidePhase():
					require.NoError(t, r.Decide(true))
				case r.GetPhase() == RamsPhaseRoundEnd:
					r.NextRound()
				case r.IsHumanTurn():
					valid := r.GetValidPlayIndices(0)
					require.NotEmpty(t, valid)
					require.NoError(t, r.PlayerPlay(valid[0]))
				default:
					r.CpuPlay()
				}
				require.Equal(t, want, ramsTotalChips(r), "1 手ごと (%d人)", n)
			}
			require.True(t, r.GetGameEndFlag(), "%d 人でゲームが終わる", n)
			// **終局時にポットが残っていてはいけない。** 総量が保たれていても、
			// 盤上に取り残されたチップは GetChips() に入らず勝敗に反映されない
			// （保存則の assert だけでは捕まらない）。
			require.Equal(t, 0, r.GetPot(), "終局時にポットは空 (%d人)", n)
		}
	}
}

// **最終ラウンドには持ち越し先が無い。** 端数をそのまま残すと、チップが
// 盤上に取り残されたままゲームが終わる。
func TestRams_FinalRoundDrainsTheRemainder(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 1
	r.SetPotForTest(11) // 2 トリックで割ると 5 余り 1
	before := ramsTotalChips(r)
	r.GetPlayer(0).SetInRound(true)
	r.GetPlayer(0).SetRoundTricks(2)

	r.FinishRoundForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, 0, r.GetPot(), "端数も配り切る")
	assert.Equal(t, before, ramsTotalChips(r), "チップの総量は変わらない")
}

// 最終ラウンドで全員降りた場合も、ポットを残さず全員に返す。
func TestRams_FinalRoundAllPassSplitsThePot(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 1
	before := ramsTotalChips(r)
	pot := r.GetPot()
	require.Positive(t, pot)

	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetDecided(true)
		r.GetPlayer(i).SetInRound(false)
	}
	r.StartPlayIfReadyForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, 0, r.GetPot(), "誰もトリックを取っていなくても残さない")
	assert.Equal(t, before, ramsTotalChips(r))
}

// 中間ラウンドでは今までどおり持ち越す。負のコントロール。
func TestRams_MidGameRoundStillCarriesTheRemainder(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 3
	r.roundNumber = 1
	r.SetPotForTest(11)
	r.GetPlayer(0).SetInRound(true)
	r.GetPlayer(0).SetRoundTricks(2)

	r.FinishRoundForTest()

	assert.False(t, r.GetGameEndFlag())
	assert.Equal(t, 1, r.GetPot(), "まだ先があるので端数は持ち越す")
}

// --- 切り札とフォロー義務 ---

func TestRams_TrumpBeatsLeadSuit(t *testing.T) {
	r := newTestRams(t)
	r.SetTrumpSuitForTest(CardDesignHeart)
	r.leadPlayerIdx = 0
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetInRound(true)
	}
	r.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 7, false)},
	})
	assert.Equal(t, 1, r.trickWinner(), "切り札の 7 がスペードの A に勝つ")
}

func TestRamsBeats(t *testing.T) {
	trump, lead := CardDesignHeart, CardDesignSpade
	assert.True(t, ramsBeats(NewCard(trump, 7, false), NewCard(lead, 1, false), lead, trump), "切り札が勝つ")
	assert.False(t, ramsBeats(NewCard(lead, 1, false), NewCard(trump, 7, false), lead, trump), "切り札には勝てない")
	assert.True(t, ramsBeats(NewCard(trump, 13, false), NewCard(trump, 12, false), lead, trump), "切り札同士は強さ")
	assert.False(t, ramsBeats(NewCard(CardDesignClover, 1, false), NewCard(lead, 7, false), lead, trump),
		"リードでも切り札でもない札は勝てない")
	assert.True(t, ramsBeats(NewCard(lead, 1, false), NewCard(CardDesignClover, 13, false), lead, trump),
		"リードのスートは無関係のスートに勝つ")
}

func TestRamsRank_AceIsHighest(t *testing.T) {
	assert.Greater(t, ramsRank(NewCard(CardDesignSpade, 1, false)), ramsRank(NewCard(CardDesignSpade, 13, false)))
	assert.Equal(t, 0, ramsRank(nil))
}

func TestRams_MustFollowSuit(t *testing.T) {
	r := newTestRams(t)
	p := r.GetPlayer(1)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 8, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	r.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}})

	assert.Equal(t, []int{0}, r.GetValidPlayIndices(1))
}

func TestRams_GetValidPlayIndicesOutOfRange(t *testing.T) {
	r := newTestRams(t)
	assert.Nil(t, r.GetValidPlayIndices(-1))
	assert.Nil(t, r.GetValidPlayIndices(r.GetPlayerCnt()))
}

// --- プレイ ---

func TestRams_PlayerPlayRejections(t *testing.T) {
	t.Run("decide phase", func(t *testing.T) {
		r := newTestRams(t)
		assert.Error(t, r.PlayerPlay(0), "選択前は出せない")
	})
	t.Run("not your turn", func(t *testing.T) {
		r := newTestRams(t)
		require.NoError(t, r.Decide(true))
		r.SetPhaseForTest(RamsPhasePlay)
		r.SetCurrentPlayerIdxForTest(1)
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("game over", func(t *testing.T) {
		r := newTestRams(t)
		r.gameEndFlag = true
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		r := newTestRams(t)
		require.NoError(t, r.Decide(true))
		r.SetPhaseForTest(RamsPhasePlay)
		r.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, r.PlayerPlay(99))
		assert.Error(t, r.PlayerPlay(-1))
	})
}

func TestRams_CpuPlayIsANoOpOnHumanTurn(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(true))
	r.SetPhaseForTest(RamsPhasePlay)
	r.SetCurrentPlayerIdxForTest(0)
	before := r.GetPlayer(0).GetCardsSize()
	r.CpuPlay()
	assert.Equal(t, before, r.GetPlayer(0).GetCardsSize())
}

// --- ゲーム終了 ---

func TestRams_MostChipsWins(t *testing.T) {
	r := newTestRams(t)
	r.GetPlayer(0).SetChips(30)
	r.GetPlayer(1).SetChips(90)
	r.GetPlayer(2).SetChips(45)
	r.GetPlayer(3).SetChips(45)

	r.FinishGameForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, RamsPhaseGameEnd, r.GetPhase())
	assert.Equal(t, 1, r.GetWinnerIdx())
}

func TestRams_TieHasNoWinner(t *testing.T) {
	r := newTestRams(t)
	for i := range r.GetPlayerCnt() {
		r.GetPlayer(i).SetChips(50)
	}
	r.FinishGameForTest()
	assert.Equal(t, -1, r.GetWinnerIdx())
}

func TestRams_NextRoundRedealsAndKeepsChips(t *testing.T) {
	r := newTestRams(t)
	r.config.Rounds = 3
	r.GetPlayer(0).SetChips(80)
	r.GetPlayer(0).SetInRound(true)
	r.SetPhaseForTest(RamsPhaseRoundEnd)
	dealer := r.GetDealerIdx()

	r.NextRound()

	assert.Equal(t, 2, r.GetRoundNumber())
	// **次のラウンドも選択フェーズから始まる。**
	assert.Equal(t, RamsPhaseDecide, r.GetPhase())
	assert.Equal(t, (dealer+1)%r.GetPlayerCnt(), r.GetDealerIdx(), "ディーラーが回る")
	assert.Equal(t, 80-RamsAnte, r.GetPlayer(0).GetChips(), "チップは持ち越し、アンティを払う")
	assert.False(t, r.GetPlayer(0).GetInRound(), "参加状態はラウンドごとに消える")
	for i := range r.GetPlayerCnt() {
		assert.Equal(t, RamsHandSize, r.GetPlayer(i).GetCardsSize())
	}
}

func TestRams_NextRoundIgnoredOutsideRoundEnd(t *testing.T) {
	r := newTestRams(t)
	r.NextRound()
	assert.Equal(t, 1, r.GetRoundNumber())

	r.gameEndFlag = true
	r.SetPhaseForTest(RamsPhaseRoundEnd)
	r.NextRound()
	assert.Equal(t, 1, r.GetRoundNumber())
}

func TestRams_GiveUp(t *testing.T) {
	r := newTestRams(t)
	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, RamsPhaseGameEnd, r.GetPhase())
	assert.Equal(t, -1, r.GetWinnerIdx())

	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
}

func TestRams_GetPlayerOutOfRange(t *testing.T) {
	r := newTestRams(t)
	assert.Nil(t, r.GetPlayer(-1))
	assert.Nil(t, r.GetPlayer(r.GetPlayerCnt()))
}

func TestRams_Config(t *testing.T) {
	r := newTestRams(t)
	assert.Equal(t, RamsPlayerCntDefault, r.GetConfig().PlayerCnt)
	assert.Equal(t, RamsRoundsDefault, r.GetConfig().Rounds)

	assert.NoError(t, RamsConfig{PlayerCnt: 3, Rounds: 1}.Validate())
	assert.NoError(t, RamsConfig{PlayerCnt: 5, Rounds: 12}.Validate())
	assert.Error(t, RamsConfig{PlayerCnt: 2, Rounds: 4}.Validate(), "2 人は不可")
	assert.Error(t, RamsConfig{PlayerCnt: 6, Rounds: 4}.Validate(), "6 人は不可")
	assert.Error(t, RamsConfig{PlayerCnt: 4, Rounds: 0}.Validate())
	assert.Error(t, RamsConfig{PlayerCnt: 4, Rounds: 13}.Validate())
}

// --- ヒント ---

// **選択フェーズでは出す札ではなく、参加するかどうかを助言する。**
func TestRams_GetHint_DecidePhaseAdvisesPlayOrPass(t *testing.T) {
	r := newTestRams(t)
	h := r.GetHint()
	require.NotNil(t, h)
	assert.Nil(t, h.CardIndex, "選択フェーズでは札を指さない")
	assert.Contains(t, []string{"ramsPlayIn", "ramsPassOut"}, h.Reason)
}

// 強い手なら参加、弱い手なら降りるを勧める。両側を踏む。
func TestRams_GetHint_DecideBothWays(t *testing.T) {
	strong := newTestRams(t)
	strong.SetTrumpSuitForTest(CardDesignHeart)
	p := strong.GetPlayer(0)
	p.Reset()
	for _, v := range []int{1, 13, 12, 11, 10} {
		p.AddCard(NewCard(CardDesignHeart, v, false)) // 切り札ばかり
	}
	assert.Equal(t, "ramsPlayIn", strong.GetHint().Reason)

	weak := newTestRams(t)
	weak.SetTrumpSuitForTest(CardDesignHeart)
	q := weak.GetPlayer(0)
	q.Reset()
	for _, v := range []int{7, 8, 9, 10, 11} {
		q.AddCard(NewCard(CardDesignSpade, v, false)) // 切り札もエースも無い
	}
	assert.Equal(t, "ramsPassOut", weak.GetHint().Reason)
}

func TestRams_GetHint_PlayPhaseSuggestsACard(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(true))
	r.SetPhaseForTest(RamsPhasePlay)
	r.SetCurrentPlayerIdxForTest(0)

	h := r.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, r.GetValidPlayIndices(0), *h.CardIndex)
		assert.Equal(t, "ramsTakeTrick", h.Reason, "まだ 1 トリックも取っていない")
	}
}

// 1 トリック取ったあとは追加支払いを免れているので、狙いが変わる。
func TestRams_GetHint_ChangesOnceSafe(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(true))
	r.SetPhaseForTest(RamsPhasePlay)
	r.SetCurrentPlayerIdxForTest(0)
	r.GetPlayer(0).SetRoundTricks(1)

	h := r.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "ramsAlreadySafe", h.Reason)
}

func TestRams_GetHint_NilAfterGameEnd(t *testing.T) {
	r := newTestRams(t)
	r.GiveUp()
	assert.Nil(t, r.GetHint())
}

// --- JSON 往復 ---

func TestRams_JSONRoundTrip(t *testing.T) {
	r := newTestRams(t)
	require.NoError(t, r.Decide(true))
	r.GetPlayer(0).SetChips(77)
	r.GetPlayer(1).SetInRound(false)
	r.SetPotForTest(31)
	r.roundNumber = 2
	r.SetTrickNumberForTest(3)

	data, err := json.Marshal(r)
	require.NoError(t, err)

	restored := NewDefaultRams()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 77, restored.GetPlayer(0).GetChips())
	assert.True(t, restored.GetPlayer(0).GetInRound(), "参加状態が往復する")
	assert.False(t, restored.GetPlayer(1).GetInRound(), "降りた人は降りたまま")
	assert.Equal(t, 31, restored.GetPot())
	assert.Equal(t, r.GetTrumpSuit(), restored.GetTrumpSuit(), "切り札が往復する")
	assert.Equal(t, 2, restored.GetRoundNumber())
	assert.Equal(t, r.GetConfig().PlayerCnt, restored.GetConfig().PlayerCnt)
	if up := r.GetUpCard(); up != nil {
		require.NotNil(t, restored.GetUpCard())
		assert.Equal(t, up.GetDesign(), restored.GetUpCard().GetDesign())
	}
}

// **5 人の局面を 4 人の器へ復元しても壊れない。** 可変人数ゲーム固有の危険。
func TestRams_JSONRoundTripWithFivePlayers(t *testing.T) {
	r := newTestRamsWith(t, 5)
	data, err := json.Marshal(r)
	require.NoError(t, err)

	restored := NewDefaultRams() // 4 人で作られている
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, 5, restored.GetPlayerCnt(), "席数も復元される")
}

func TestRams_UnmarshalRejectsGarbage(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("not json"), NewDefaultRams()))
}

// 範囲外の席数は壊れた KV。読み込みを拒否する。
func TestRams_UnmarshalRejectsBadPlayerCount(t *testing.T) {
	r := newTestRams(t)
	data, err := json.Marshal(r)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["pl"] = []any{} // 0 人
	broken, err := json.Marshal(raw)
	require.NoError(t, err)

	assert.Error(t, json.Unmarshal(broken, NewDefaultRams()))
}

func TestRams_ActionLog(t *testing.T) {
	r := newTestRams(t)
	assert.NotEmpty(t, r.GetActionLog())
}
